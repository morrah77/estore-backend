// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"context"
	"crypto/tls"
	"embed"
	"encoding/json"
	"estore-backend/server/logger"
	"estore-backend/server/restapi/operations/categories"
	"estore-backend/server/restapi/operations/category"
	"estore-backend/server/restapi/operations/checkout"
	"estore-backend/server/restapi/operations/order"
	"estore-backend/server/restapi/operations/orders"
	"estore-backend/server/restapi/operations/user"
	"estore-backend/server/restapi/operations/users"
	"estore-backend/server/restapi/operations/webhooks"
	"fmt"
	"github.com/go-openapi/swag"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
	"log"
	"net/http"
	"os"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"

	"estore-backend/server/database"
	dbModels "estore-backend/server/database/models"
	"estore-backend/server/models"
	"estore-backend/server/restapi/operations"
	"estore-backend/server/restapi/operations/auth"
	"estore-backend/server/restapi/operations/product"
	"estore-backend/server/restapi/operations/products"
)

//go:generate swagger generate server --target ../../server --name EStoreMain --spec ../../swagger.yml --principal models.Principal

var apiConf = struct {
	ConfigFile string `short:"c" long:"config-file" description:"Config File"`
}{}

var ApiConfiguration struct {
	AppHost string `json:"AppHost"`

	AppFrontEndHost string `json:"AppFrontEndHost"`

	// "state" request parameter used as an internal token in OAuth2 request exchanges
	OAuthState string `json:"OAuthState"`

	// OAuth2 credentials
	OAuthClientID     string
	OAuthClientSecret string

	// token signer on OAuth2
	OAuthIssuer string // for Google OAuth2: "https://accounts.google.com"

	// OAuth2 login URL
	OAuthAuthURL string // for Google OAuth2: "https://accounts.google.com/o/oauth2/auth"

	// OAuth2 endpoints for access tokens and user info
	OAuthTokenURL    string // for Google OAuth2: "https://oauth2.googleapis.com/token"
	OAuthUserInfoURL string // for Google OAuth2: "https://www.googleapis.com/oauth2/v3/userinfo"

	// The application endpoint OAuth2 service redirects to (i. e., calls back)
	OAuthCallbackURL string

	// TODO implement storing roles in the database or in a separate service
	// Access control list for administrators
	ACL struct {
		Admin []string `json:"admin"`
	}

	DBDriver           string // Database driver name
	DBConnectionString string // Database conection string

	LogLevel string

	AccessControlAllowOrigin string

	Payments struct {
		Stripe struct {
			Secret               string `json:"secret"`
			PaymentWebhookSecret string `json:"paymentWebhookSecret"`
			PaymentWebhookId     string `json:"paymentWebhookId"`
		} `json:"Stripe"`
	} `json:"Payments"`
}

var logLevel = logger.LOG_DEBUG

var Logger logger.SimpleLogger

var db *bun.DB

// TODO implement storing migration in non-embedded files
//
//go:embed *.sql
var sqlMigrations embed.FS

func configureFlags(api *operations.EStoreMainAPI) {
	api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{
		{
			ShortDescription: "configuration",
			Options:          &apiConf,
		},
	}
}

func configureAPI(api *operations.EStoreMainAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// api.Logger = log.Printf
	Logger = logger.New()
	Logger.SetLevel(logLevel)
	api.Logger = Logger

	log.Printf("Log level: %s", logLevel)
	api.UseSwaggerUI()
	// To continue using redoc as your UI, uncomment the following line
	// api.UseRedoc()

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	// Reading the main application configuration
	configFileName := apiConf.ConfigFile
	if configFileName == "" {
		configFileName = "config.json"
	}
	if configFileName != "" {
		bytes, err := os.ReadFile(configFileName)
		if err != nil {
			panic(fmt.Sprintf("Error: %s\nNo config file at %s", err.Error(), configFileName))
		}
		err = json.Unmarshal(bytes, &ApiConfiguration)
		if err != nil {
			panic(fmt.Sprintf("Error: %s\nould not parse config from file %s", err.Error(), configFileName))
		}
		Logger.Debug("Parsed ApiConfiguration: %s", ApiConfiguration)
		if ApiConfiguration.LogLevel != "" {
			Logger.Log(logger.LOG_INFO, "Parsing log level %s", ApiConfiguration.LogLevel)
			Logger.SetLevel(logger.ParseLogLevel(ApiConfiguration.LogLevel))
			Logger.Log(logger.LOG_INFO, "Set log level %s from config file", ApiConfiguration.LogLevel)
		}
	}

	// Initializing and setting up the database
	db = database.InitDB(ApiConfiguration.DBDriver, ApiConfiguration.DBConnectionString)
	SetUpDB()

	api.OauthSecurityAuth = func(token string, scopes []string) (*models.Principal, error) {
		Logger.Debug("OauthSecurityAuth: Scopes %s\n", scopes)
		userInfo, err := authenticate(token, scopes)
		if err != nil {
			Logger.Info("OauthSecurityAuth: Authentication error: %s\n%v", err, err)
			return nil, errors.New(http.StatusUnauthorized, "Authentication error")
		}
		if userInfo == nil {
			Logger.Info("OauthSecurityAuth: Invalid token: userinfo is nil\ntoken: %s", token)
			return nil, errors.New(http.StatusUnauthorized, "Invalid token")
		}

		principal := models.Principal{
			UserInfo: *userInfo,
		}
		Logger.Debug("OauthSecurityAuth: Authenticated.\nUserinfo: %s\nPrincipal: %s\n", userInfo, principal)
		return &principal, nil
	}

	api.AuthGetAccessTokenHandler = auth.GetAccessTokenHandlerFunc(func(params auth.GetAccessTokenParams) middleware.Responder {
		userInfo, err := oauthRedirectCallback(params.HTTPRequest)
		if err != nil {
			Logger.Info("Error: GetAccessToken: %s\n", err.Error())
			return auth.NewGetAccessTokenDefault(http.StatusForbidden).
				WithPayload(&models.Error{
					Httpcode: http.StatusForbidden,
					Message:  swag.String("Could not get access token"),
				})
		}
		Logger.Debug("userInfo: %s\n", userInfo)
		return auth.NewGetAccessTokenOK().WithPayload(&models.Principal{UserInfo: *userInfo})
	})

	api.AuthLoginHandler = auth.LoginHandlerFunc(func(params auth.LoginParams) middleware.Responder {
		return login(params.HTTPRequest)
	})

	// Set your custom authorizer if needed. Default one is security.Authorized()
	// Expected interface runtime.Authorizer
	//
	// Example:
	// api.APIAuthorizer = security.Authorized()

	// Products
	api.ProductsAddProductHandler = products.AddProductHandlerFunc(func(params products.AddProductParams, principal *models.Principal) middleware.Responder {
		Logger.Debug("Calling addProduct with %s\n%s\n%s %s %s %s\n", params, params.Body, params.Body.ID, params.Body.Title, params.Body.Description, params.Body.Images)
		if err := addProduct(params.Body); err != nil {
			return products.NewAddProductDefault(500).WithPayload(&models.Error{Httpcode: 500, Message: swag.String(err.Error())})
		}
		return products.NewAddProductCreated().WithPayload(params.Body)
	})

	api.ProductDeleteProductHandler = product.DeleteProductHandlerFunc(func(params product.DeleteProductParams, principal *models.Principal) middleware.Responder {
		if err := deleteProduct(params.ID); err != nil {
			return product.NewDeleteProductDefault(500).WithPayload(&models.Error{Httpcode: 500, Message: swag.String(err.Error())})
		}
		return product.NewDeleteProductNoContent()
	})

	api.ProductEditProductHandler = product.EditProductHandlerFunc(func(params product.EditProductParams, principal *models.Principal) middleware.Responder {
		if err := updateProduct(params.ID, params.Body); err != nil {
			return product.NewEditProductDefault(500).WithPayload(&models.Error{Httpcode: 500, Message: swag.String(err.Error())})
		}
		return product.NewEditProductOK().WithPayload(params.Body)
	})

	api.ProductGetProductHandler = product.GetProductHandlerFunc(func(params product.GetProductParams) middleware.Responder {
		var result *models.Product
		result, err := getProduct(params.ID)
		if err != nil {
			return product.NewGetProductDefault(500).
				WithPayload(&models.Error{Httpcode: 500, Message: swag.String(err.Error())})
		}
		return product.NewGetProductOK().WithPayload(result)
	})

	api.ProductsGetProductsHandler = products.GetProductsHandlerFunc(func(params products.GetProductsParams) middleware.Responder {
		cleanParams := products.NewGetProductsParams()
		cleanParams.Offset = swag.Int64(0)
		if params.Offset != nil {
			cleanParams.Offset = params.Offset
		}
		if params.Limit != nil {
			cleanParams.Limit = params.Limit
		}
		if params.Search != nil {
			cleanParams.Search = params.Search
		}
		if params.CategoryIds != nil {
			cleanParams.CategoryIds = params.CategoryIds
		}
		Logger.Debug("Calling allProducts with limit %s, offset %s, search %sCategoryIds %s",
			params.Limit, params.Offset, params.Search, params.CategoryIds)
		result, err := allProducts(&cleanParams)
		if err != nil {
			return products.NewGetProductsDefault(500).
				WithPayload(&models.Error{Httpcode: 500, Message: swag.String(err.Error())})
		}
		return products.NewGetProductsOK().WithPayload(result)
	})

	// Categories

	api.CategoriesAddCategoryHandler = categories.AddCategoryHandlerFunc(func(params categories.AddCategoryParams, principal *models.Principal) middleware.Responder {
		Logger.Debug("Calling addCategory with %v\n%s\n", params, params.Body)
		if err := addCategory(params.Body); err != nil {
			return categories.NewAddCategoryDefault(500).WithPayload(&models.Error{Httpcode: 500, Message: swag.String(err.Error())})
		}
		return categories.NewAddCategoryCreated().WithPayload(params.Body)
	})

	api.CategoryDeleteCategoryHandler = category.DeleteCategoryHandlerFunc(func(params category.DeleteCategoryParams, principal *models.Principal) middleware.Responder {
		if err := deleteCategory(params.ID); err != nil {
			return category.NewDeleteCategoryDefault(500).WithPayload(&models.Error{Httpcode: 500, Message: swag.String(err.Error())})
		}
		return category.NewDeleteCategoryNoContent()
	})

	api.CategoryEditCategoryHandler = category.EditCategoryHandlerFunc(func(params category.EditCategoryParams, principal *models.Principal) middleware.Responder {
		if err := updateCategory(params.ID, params.Body); err != nil {
			return category.NewEditCategoryDefault(500).WithPayload(&models.Error{Httpcode: 500, Message: swag.String(err.Error())})
		}
		return category.NewEditCategoryOK().WithPayload(params.Body)
	})

	api.CategoryGetCategoryHandler = category.GetCategoryHandlerFunc(func(params category.GetCategoryParams) middleware.Responder {
		var result *models.Category
		result, err := getCategory(params.ID)
		if err != nil {
			return category.NewGetCategoryDefault(500).
				WithPayload(&models.Error{Httpcode: 500, Message: swag.String(err.Error())})
		}
		return category.NewGetCategoryOK().WithPayload(result)
	})

	api.CategoriesListCategoriesHandler = categories.ListCategoriesHandlerFunc(func(params categories.ListCategoriesParams) middleware.Responder {
		cleanParams := categories.NewListCategoriesParams()
		cleanParams.Offset = swag.Int64(0)
		if params.Offset != nil {
			cleanParams.Offset = params.Offset
		}
		if params.Limit != nil {
			cleanParams.Limit = params.Limit
		}
		if params.Search != nil {
			cleanParams.Search = params.Search
		}
		Logger.Debug("Calling allCategories with limit %s, offset %s, search %s",
			params.Limit, params.Offset, params.Search)
		result, err := allCategories(&cleanParams)
		if err != nil {
			return categories.NewListCategoriesDefault(500).
				WithPayload(&models.Error{Httpcode: 500, Message: swag.String(err.Error())})
		}
		return categories.NewListCategoriesOK().WithPayload(result)
	})

	// Users

	api.UsersAddUserHandler = users.AddUserHandlerFunc(func(params users.AddUserParams, principal *models.Principal) middleware.Responder {
		Logger.Debug("Calling addUser with %v\n%s\n", params, params.Body)
		if err := addUser(&params, principal); err != nil {
			return users.NewAddUserDefault(int(err.Code())).
				WithPayload(&models.Error{Httpcode: int64(err.Code()), Message: swag.String(err.Error())})
		}
		return users.NewAddUserCreated().WithPayload(params.Body)
	})

	api.UserDeleteUserHandler = user.DeleteUserHandlerFunc(func(params user.DeleteUserParams, principal *models.Principal) middleware.Responder {
		if err := deleteUser(&params, principal); err != nil {
			return user.NewDeleteUserDefault(int(err.Code())).
				WithPayload(&models.Error{Httpcode: int64(err.Code()), Message: swag.String(err.Error())})
		}
		return user.NewDeleteUserNoContent()
	})

	api.UserEditUserHandler = user.EditUserHandlerFunc(func(params user.EditUserParams, principal *models.Principal) middleware.Responder {
		if err := updateUser(&params, principal); err != nil {
			return user.NewEditUserDefault(int(err.Code())).
				WithPayload(&models.Error{Httpcode: int64(err.Code()), Message: swag.String(err.Error())})
		}
		return user.NewEditUserOK().WithPayload(params.Body)
	})

	api.UserGetUserHandler = user.GetUserHandlerFunc(func(params user.GetUserParams, principal *models.Principal) middleware.Responder {
		var result *models.User
		result, err := getUser(&params, principal)
		if err != nil {
			return user.NewGetUserDefault(int(err.Code())).
				WithPayload(&models.Error{Httpcode: int64(err.Code()), Message: swag.String(err.Error())})
		}
		return user.NewGetUserOK().WithPayload(result)
	})

	api.UserGetOwnUserHandler = user.GetOwnUserHandlerFunc(func(params user.GetOwnUserParams, principal *models.Principal) middleware.Responder {
		var result *models.User
		result, err := getOwnUserInfo(&params, principal)
		if err != nil {
			return user.NewGetOwnUserDefault(int(err.Code())).
				WithPayload(&models.Error{Httpcode: int64(err.Code()), Message: swag.String(err.Error())})
		}
		return user.NewGetOwnUserOK().WithPayload(result)
	})

	api.UsersListUsersHandler = users.ListUsersHandlerFunc(func(params users.ListUsersParams, principal *models.Principal) middleware.Responder {
		Logger.Debug("Calling allUsers with limit %s, offset %s, search %s",
			params.Limit, params.Offset, params.Search)
		result, err := allUsers(&params, principal)
		if err != nil {
			return users.NewListUsersDefault(int(err.Code())).
				WithPayload(&models.Error{Httpcode: int64(err.Code()), Message: swag.String(err.Error())})
		}
		return users.NewListUsersOK().WithPayload(result)
	})

	// Orders

	api.OrdersAddOrderHandler = orders.AddOrderHandlerFunc(func(params orders.AddOrderParams, principal *models.Principal) middleware.Responder {
		Logger.Debug("Calling addOrder with %v\n%s\n", params, params.Body)
		orderDTO, err := addOrder(&params, principal)
		if err != nil {
			return orders.NewAddOrderDefault(int(err.Code())).
				WithPayload(&models.Error{Httpcode: int64(err.Code()), Message: swag.String(err.Error())})
		}
		Logger.Debug("Responding with Order %s", orderDTO)
		return orders.NewAddOrderCreated().WithPayload(orderDTO)
	})

	api.OrderDeleteOrderHandler = order.DeleteOrderHandlerFunc(func(params order.DeleteOrderParams, principal *models.Principal) middleware.Responder {
		if err := deleteOrder(&params, principal); err != nil {
			return order.NewDeleteOrderDefault(int(err.Code())).
				WithPayload(&models.Error{Httpcode: int64(err.Code()), Message: swag.String(err.Error())})
		}
		return order.NewDeleteOrderNoContent()
	})

	api.OrderEditOrderHandler = order.EditOrderHandlerFunc(func(params order.EditOrderParams, principal *models.Principal) middleware.Responder {
		if err := updateOrder(&params, principal); err != nil {
			return order.NewEditOrderDefault(int(err.Code())).
				WithPayload(&models.Error{Httpcode: int64(err.Code()), Message: swag.String(err.Error())})
		}
		return order.NewEditOrderOK().WithPayload(params.Body)
	})

	api.OrderGetOrderHandler = order.GetOrderHandlerFunc(func(params order.GetOrderParams, principal *models.Principal) middleware.Responder {
		var result *models.Order
		result, err := getOrder(&params, principal)
		if err != nil {
			return order.NewGetOrderDefault(int(err.Code())).
				WithPayload(&models.Error{Httpcode: int64(err.Code()), Message: swag.String(err.Error())})
		}
		return order.NewGetOrderOK().WithPayload(result)
	})

	api.OrdersListOrdersHandler = orders.ListOrdersHandlerFunc(func(params orders.ListOrdersParams, principal *models.Principal) middleware.Responder {
		Logger.Debug("Calling allOrders with limit %s, offset %s",
			params.Limit, params.Offset)
		result, err := allOrders(&params, principal)
		if err != nil {
			return orders.NewListOrdersDefault(int(err.Code())).
				WithPayload(&models.Error{Httpcode: int64(err.Code()), Message: swag.String(err.Error())})
		}
		return orders.NewListOrdersOK().WithPayload(result)
	})

	// Payments

	//Checkout
	api.CheckoutAddCheckoutSessionHandler = checkout.AddCheckoutSessionHandlerFunc(func(params checkout.AddCheckoutSessionParams, principal *models.Principal) middleware.Responder {
		Logger.Debug("Calling createCheckoutSession with %v\n%s\n", params, params.Body)
		clientSecret, err := createCheckoutSession(&params, principal)
		if err != nil {
			return checkout.NewAddCheckoutSessionDefault(int(err.Code())).
				WithPayload(&models.Error{Httpcode: int64(err.Code()), Message: swag.String(err.Error())})
		}
		return checkout.NewAddCheckoutSessionCreated().WithPayload(&models.CheckoutSessionSecret{ClientSecret: *clientSecret})
	})

	api.CheckoutGetCheckoutSessionHandler = checkout.GetCheckoutSessionHandlerFunc(func(params checkout.GetCheckoutSessionParams, principal *models.Principal) middleware.Responder {
		Logger.Debug("Calling retrieveCheckoutSession with SessionID %s",
			params.SessionID)
		result, err := retrieveCheckoutSession(&params, principal)
		if err != nil {
			return checkout.NewGetCheckoutSessionDefault(int(err.Code())).
				WithPayload(&models.Error{Httpcode: int64(err.Code()), Message: swag.String(err.Error())})
		}
		return checkout.NewGetCheckoutSessionOK().WithPayload(result)
	})

	// Stripe payment webhook
	api.WebhooksProcessStripePaymentHandler = webhooks.ProcessStripePaymentHandlerFunc(func(params webhooks.ProcessStripePaymentParams) middleware.Responder {
		Logger.Debug("Calling WebhooksProcessStripePaymentHandler with Stripe Signature %s",
			params.StripeSignature)
		err := processStripePaymentEvent(&params)
		if err != nil {
			return webhooks.NewProcessStripePaymentDefault(int(err.Code())).
				WithPayload(&models.Error{Httpcode: int64(err.Code()), Message: swag.String(err.Error())})
		}
		return webhooks.NewProcessStripePaymentOK()
	})

	api.PreServerShutdown = func() {}

	api.ServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix".
func configureServer(s *http.Server, scheme, addr string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation.
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics.
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("ApiConfiguration.AccessControlAllowOrigin: %s", ApiConfiguration.AccessControlAllowOrigin)
		w.Header().Add("Access-Control-Allow-Origin", ApiConfiguration.AccessControlAllowOrigin)
		handler.ServeHTTP(w, r)
	})
}

// TODO Refactor this ugly code!!!
func SetUpDB() {
	//query := db.NewCreateTable().Model(&dbModels.Product{}).IfNotExists()
	//log.Printf("Built the query %s\n", query)
	//result, err := query.Exec(context.Background())
	//if err != nil {
	//	panic(fmt.Sprintf("Error: %s, result: %s", err.Error(), result))
	//}
	//
	//dropColumnQuery := db.NewCreateTable().NewDropColumn().TableExpr("products").Column("category_ids")
	//log.Printf("Built the query %s\n", dropColumnQuery)
	//result, err = dropColumnQuery.Exec(context.Background())
	//if err != nil {
	//	log.Printf("Error: %s, result: %s", err.Error(), result)
	//}

	db.RegisterModel((*dbModels.ProductToCategory)(nil))
	var err error
	modelTables := []interface{}{&dbModels.ProductToCategory{}, &dbModels.Product{}, &dbModels.Category{},
		&dbModels.User{}, &dbModels.OrderedProduct{}, &dbModels.Order{}, &dbModels.Payment{}}
	for _, m := range modelTables {
		query := db.NewCreateTable().Model(m).IfNotExists()
		Logger.Debug("Built the query %s\n", query)
		result, err := query.Exec(context.Background())
		if err != nil {
			panic(fmt.Sprintf("Error: %s, result: %s", err.Error(), result))
		}
	}

	// Migrating the database according to the embedded migration files if needed
	var Migrations = migrate.NewMigrations()

	if err := Migrations.Discover(sqlMigrations); err != nil {
		panic(err)
	}
	migrator := migrate.NewMigrator(db, Migrations)
	err = migrator.Init(context.Background())
	if err != nil {
		panic(fmt.Sprintf("Error: %s", err.Error()))
	}

	var migrationsToRollBack = migrate.NewMigrations()
	var registeredMigrations migrate.MigrationSlice
	registeredMigrations, err = migrator.MigrationsWithStatus(context.Background())
	if err != nil {
		panic(fmt.Sprintf("Error: %s", err.Error()))
	}
	Logger.Info("There are %d discovered migrations", len(registeredMigrations))
	for i, m := range registeredMigrations {
		Logger.Debug("Discovered migration %d: %d, %s, %b", i, m.ID, m, m.IsApplied())
	}
	appliedMigrations := registeredMigrations.Applied()
	Logger.Info("There are %d applied migrations", len(appliedMigrations))
	var counter = 0
	var group *migrate.MigrationGroup
	for _, m := range appliedMigrations {
		Logger.Debug("Processing migration %s", m)
		if m.Down != nil {
			Logger.Debug("Adding migration %s to rolback list", m)
			migrationsToRollBack.Add(m)
			counter++
		}
	}
	Logger.Info("There are %d migrations to rollback", len(migrationsToRollBack.Sorted()))
	if counter > 0 {
		rollbackMigrator := migrate.NewMigrator(db, migrationsToRollBack)
		Logger.Debug("Rolling back migrations: %s", migrationsToRollBack)
		group, err = rollbackMigrator.Rollback(context.Background())
		if err != nil {
			panic(fmt.Sprintf("Error: %s, result: %s", err.Error(), group))
		}
		Logger.Debug("Rolled back %d migrations\ngroup: %s\n", len(group.Migrations),
			group)
		rollbackMigrator = nil
	}

	var migrationsToMigrate = migrate.NewMigrations()
	discoveredMigrations := Migrations.Sorted()
	Logger.Info("Applying migrations: %s", discoveredMigrations)
	var isRolledBack = false
	counter = 0
	for i, dm := range discoveredMigrations {
		isRolledBack = false
		for _, rm := range migrationsToRollBack.Sorted() {
			if dm.Name == rm.Name {
				isRolledBack = true
				Logger.Debug("Filtering out migration %d with name %s: %s", i, dm.Name, dm)
				break
			}
		}
		if !isRolledBack {
			migrationsToMigrate.Add(dm)
			counter++
		}
	}
	Logger.Debug("Migration to apply: %d ( %s )", counter, migrationsToMigrate)
	if counter > 0 {
		migrator = migrate.NewMigrator(db, migrationsToMigrate)
		group, err = migrator.Migrate(context.Background())
		if err != nil {
			panic(fmt.Sprintf("Error: %s, result: %s", err.Error(), group))
		}
		Logger.Debug("Applied %d migrations\n", len(group.Migrations),
			group)
		if group.ID == 0 {
			Logger.Debug("there are no new migrations to run\ngroup: %s\n", group)
		}
		Logger.Debug("migrated to %s\n", group)
	}
}
