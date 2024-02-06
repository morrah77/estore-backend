package restapi

import (
	"encoding/json"

	"estore-backend/server/models"
	"fmt"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"io"
	"log"
	"net/http"
)

var (
	config *oauth2.Config
	ACL    *map[string]string
)

type UserInfo struct {
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Locale        string `json:"locale"`
	User          *models.User
	ACLRole       string
}

/*
Authenticate user checking the scopes from less restrictive to the most restrictive:
  - user (allowed access to the data of 'user' scope)
  - private (allowed access to data related to the user making the request)
  - admin (allowed access to any data)

Return UserInfo structure pointer as a principal
the 1st return value of this function will be available in all handler functions
as a 2nd parameter
as well as a request context value (use middleware.SecurityPrincipalFrom(request) to fetch it)
*/
func authenticate(token string, scopes []string) (*models.UserInfo, error) {
	Logger.Debug("auth_impl authenticate\nApiConfiguration: %s\ntoken: %s\nscopes:%s\n", ApiConfiguration, token, scopes)

	//Logger.Debug("auth_impl authenticate\nGetting user info for token %s\n", token)
	info, err := getUserInfo(token)
	if err != nil {
		return nil, fmt.Errorf("could not validate token. Error: %v", err)
	}
	Logger.Debug("auth_impl authenticate\nGot user info: %s\n", info)
	userDTO, err := registerUserIfNeeded(info)
	if err != nil {
		Logger.Info("Could not register user %s; error: %s", info, err.Error())
	} else {
		Logger.Debug("Registered user %s", userDTO)
		info.User = userDTO
	}
	if isUser(scopes) {
		// the endpoint requested authorization is in the user scope, allowing
		Logger.Debug("auth_impl authenticate\nScopes: %s. Allowing.", scopes)
		return info, nil
	}
	Logger.Debug("auth_impl authenticate\nACL from config:\n%s\n", ApiConfiguration.ACL)
	// TODO Implement real private scope authorization check!
	if isPrivate(scopes) {
		// the endpoint requested authorization is in the private scope. Actions:
		// - allowing for users listed in the ACL under "admin" or "private" scope,
		// - checking authorization for other users
		if (info.ACLRole == "admin") || (info.ACLRole == "private") {
			Logger.Debug("auth_impl authenticate\nScopes: %s, user %s role: %s. Allowing.", scopes, info.Email, info.ACLRole)
			return info, nil
		}
		if info.User != nil {
			Logger.Debug("Private scope user is checked. Scopes: %s, user: [%d] %s. Allowing.", scopes, info.User.ID, info.Email)
			return info, nil
		} else {
			Logger.Debug("Private scope user not allowed. Scopes: %s, user: %s. Denying.", scopes, info.Email)
			return nil, nil
		}
	}

	if isAdmin(scopes) {
		// the endpoint requested authorization is in the admin scope, allowing if the user is in the ACL
		if info.ACLRole == "admin" {
			Logger.Debug("Admin scope user is present in ACL. Scopes: %s, user: %s. Allowing.", scopes, info.Email)
			return info, nil
		}
		Logger.Debug("Admin scope user is absent in ACL. Scopes: %s, user: %s. Denying.", scopes, info.Email)
		return nil, nil
	}
	Logger.Debug("No known scopes provided. Scopes: %s, user: %s. Allowing.", scopes, info.Email)
	return info, nil
}

func login(r *http.Request) middleware.Responder {
	Logger.Debug("auth_impl login ApiConfiguration: %s", ApiConfiguration)
	return middleware.ResponderFunc(
		func(w http.ResponseWriter, pr runtime.Producer) {
			http.Redirect(w, r, getConfig().AuthCodeURL(ApiConfiguration.OAuthState), http.StatusFound)
		})
}

func oauthRedirectCallback(r *http.Request) (*models.UserInfo, error) {
	Logger.Debug("auth_impl oauthRedirectCallback ApiConfiguration: %s", ApiConfiguration)

	/* OAuth2 service redirects clients to this "callback" endpoint;
	"?state=<>&code=<>" are provided in request query params */
	state := r.URL.Query().Get("state")
	if state != ApiConfiguration.OAuthState {
		logMessage := fmt.Sprintf("Wrong OAuth state %s!", state)
		Logger.Info(logMessage)
		return nil, fmt.Errorf(logMessage)
	}

	/* oauth2.Config#Exchange: requests OAUth2 token endpoint, then calls the redirect endpoint,
	i. e., exchanges an authorization code for an access token */
	authCode := r.URL.Query().Get("code")
	Logger.Debug("Authorization code: %v\n", authCode)
	openIDContext := oidc.ClientContext(context.Background(), &http.Client{})
	oauth2Token, err := getConfig().Exchange(openIDContext, authCode)
	if err != nil {
		log.Println("failed to exchange token", err.Error())
		return nil, fmt.Errorf("failed to exchange token")
	}

	info, err := getUserInfo(oauth2Token.AccessToken)
	Logger.Debug("Got user info in the callback:\ninfo: %s\nerror: %v\n", info, err)
	userDTO, err := registerUserIfNeeded(info)
	if err != nil {
		Logger.Info("Could not register user %s; error: %s", info, err.Error())
	} else {
		Logger.Debug("Registered user %s", userDTO)
	}

	/*access token from OAuth2 authorization service*/
	log.Println("Raw token data:", oauth2Token)
	return &models.UserInfo{
		AccessToken: oauth2Token.AccessToken,
		Email:       info.Email,
		Name:        info.Name,
	}, nil
}

func registerUserIfNeeded(userInfo *models.UserInfo) (user *models.User, err error) {
	user, err = getUserByEmail(userInfo.Email)
	if err == nil {
		Logger.Debug("User %s is already present in the DB. Will not register.\n", userInfo.Email)
		return user, nil
	}
	Logger.Debug("Could not find user %s!\nERROR: %s\nRegistering...\n", userInfo.Email, err.Error())
	dbUser := &models.User{
		Email: &userInfo.Email,
		Name:  userInfo.Name,
	}
	err = addDBUser(dbUser)
	if err != nil {
		return nil, errors.New(http.StatusInternalServerError, "Could not register user %s!\nERROR: %s\n", userInfo.Email, err.Error())
	}
	return dbUser, nil
}

func getConfig() *oauth2.Config {
	if config == nil {

		var endpoint = oauth2.Endpoint{
			AuthURL:  ApiConfiguration.OAuthAuthURL,
			TokenURL: ApiConfiguration.OAuthTokenURL,
		}

		config = &oauth2.Config{
			ClientID:     ApiConfiguration.OAuthClientID,
			ClientSecret: ApiConfiguration.OAuthClientSecret,
			Endpoint:     endpoint,
			RedirectURL:  ApiConfiguration.OAuthCallbackURL,
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		}

	}
	return config
}

func getACL() *map[string]string {
	if ACL == nil {
		acl := map[string]string{}
		if len(ApiConfiguration.ACL.Admin) > 0 {
			for _, email := range ApiConfiguration.ACL.Admin {
				acl[email] = "admin"
			}
		}
		ACL = &acl
	}
	Logger.Debug("getACL: ACL: %s\n", ACL)
	return ACL
}

func getUserInfo(token string) (userInfo *models.UserInfo, err error) {
	bearToken := "Bearer " + token
	req, err := http.NewRequest("GET", ApiConfiguration.OAuthUserInfoURL, nil)

	if err != nil {
		return nil, fmt.Errorf("getUserInfo: http request error: %v", err)
	}

	req.Header.Add("Authorization", bearToken)

	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("getUserInfo: http request error: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("getUserInfo: Invalid status code: %v", resp.StatusCode)
	}
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("getUserInfo: fail to get response: %v", err)
	}
	Logger.Debug("Got user info: %s", string(bytes))

	userInfo = &models.UserInfo{}
	err = json.Unmarshal(bytes, userInfo)
	if err != nil {
		Logger.Info("getUserInfo: Error while parsing user info: %v", err)
		return nil, fmt.Errorf("getUserInfo: Error while parsing user info: %s", err.Error())
	} else {
		user, err := getUserByEmail(userInfo.Email)
		if err != nil {
			Logger.Info("getUserInfo: WARN: %v\nCould not find a registered user by email %s!", err, userInfo.Email)
		} else {
			userInfo.User = user
		}
		ACLScope, ok := (*(getACL()))[userInfo.Email]
		if ok {
			userInfo.ACLRole = ACLScope
			Logger.Debug("getUserInfo: Added ACL scope %s to user info for user %s", ACLScope, userInfo.Email)
		} else {
			Logger.Debug("getUserInfo: Could not find ACL scope for user %s", userInfo.Email)
		}
		Logger.Debug("getUserInfo: Parsed user info: %s\n", userInfo)
	}
	return userInfo, nil
}

func isAdmin(scopes []string) bool {
	return isScope(scopes, "admin")
}

func isPrivate(scopes []string) bool {
	return isScope(scopes, "private")
}

func isUser(scopes []string) bool {
	return isScope(scopes, "user")
}

func isScope(scopes []string, scope string) bool {
	for _, s := range scopes {
		if s == scope {
			return true
		}
	}
	return false
}
