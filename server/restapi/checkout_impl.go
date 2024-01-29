package restapi

import (
	"encoding/json"
	dbModels "estore-backend/server/database/models"
	"estore-backend/server/models"
	"estore-backend/server/restapi/operations/checkout"
	"estore-backend/server/restapi/operations/webhooks"
	"github.com/go-openapi/errors"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/webhook"
	"io"
)

func createCheckoutSession(params *checkout.AddCheckoutSessionParams, principal *models.Principal) (*string, errors.Error) {
	err := isPrincipalOwnerOrAdmin(principal, *params.Body.ID)
	if err != nil {
		return nil, err
	}
	isAdmin, err := isPrincipalAdmin(principal)
	if err != nil {
		return nil, err
	}

	order, err := getOrderFromDB(*params.Body.ID, isAdmin, principal.User.ID)
	if err != nil {
		return nil, err
	}

	var lineItems []*stripe.CheckoutSessionLineItemParams = make([]*stripe.CheckoutSessionLineItemParams, len(order.Products))
	for i, p := range order.Products {
		lineItems[i] = &stripe.CheckoutSessionLineItemParams{
			// TODO add a product price or fetch products from the DB

			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Currency: stripe.String("usd"),
				ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
					Name: stripe.String(p.ProductName),
				},
				UnitAmountDecimal: stripe.Float64(*p.TotalPrice / float64(*p.Quantity) * 100),
			},

			Quantity: stripe.Int64(*p.Quantity),
		}
	}

	sessionParams := &stripe.CheckoutSessionParams{
		UIMode:               stripe.String("embedded"),
		ReturnURL:            stripe.String(ApiConfiguration.AppFrontEndHost + "/cart?session_id={CHECKOUT_SESSION_ID}"),
		RedirectOnCompletion: stripe.String("if_required"),
		LineItems:            lineItems,
		Mode:                 stripe.String(string(stripe.CheckoutSessionModePayment)),
		AutomaticTax:         &stripe.CheckoutSessionAutomaticTaxParams{Enabled: stripe.Bool(true)},
	}
	stripe.Key = ApiConfiguration.Payments.Stripe.Secret

	s, sessionErr := session.New(sessionParams)

	if sessionErr != nil {
		Logger.Error("ERROR: session.New: %v", err)
		return nil, errors.New(500, err.Error())
	}

	var payment *dbModels.Payment = &dbModels.Payment{
		Amount:            order.TotalPrice,
		OrderID:           order.ID,
		Status:            "intended",
		UserID:            principal.User.ID,
		CheckoutSessionID: s.ID,
	}
	res, err := createDBPayment(payment)
	if err != nil {
		Logger.Error("Could not process successful payment for intent %v\n%s\nError: %d %s", res, res, err.Code(), err.Error())
		return nil, err
	}

	return &s.ClientSecret, nil
}

func retrieveCheckoutSession(params *checkout.GetCheckoutSessionParams, principal *models.Principal) (*models.CheckoutSession, errors.Error) {
	s, err := session.Get(*params.SessionID, nil)
	if err != nil {
		Logger.Error("ERROR: session.Get: %v", err)
		return nil, errors.New(500, err.Error())
	}

	return &models.CheckoutSession{
		Status:        string(s.Status),
		CustomerEmail: s.CustomerDetails.Email,
	}, nil
}

func processStripePaymentEvent(params *webhooks.ProcessStripePaymentParams) errors.Error {
	stripe.Key = ApiConfiguration.Payments.Stripe.Secret
	headerString := params.HTTPRequest.Header.Get("Stripe-Signature")
	Logger.Debug("processStripePaymentEvent: stripe signature %s\nStripe.PaymentWebhookSecret %s\nheaderString %s\n",
		params.StripeSignature, ApiConfiguration.Payments.Stripe.PaymentWebhookSecret, headerString)

	payload, err := io.ReadAll(params.HTTPRequest.Body)
	if err != nil {
		Logger.Error("Error reading request body: %v\n", err)
		return errors.New(500, "Error reading request body")
	}

	// If you are testing your webhook locally with the Stripe CLI you
	// can find the endpoint's secret by running `stripe listen`
	// Otherwise, find your endpoint's secret in your webhook settings
	// in the Developer Dashboard
	endpointSecret := ApiConfiguration.Payments.Stripe.PaymentWebhookSecret
	event, err := webhook.ConstructEvent(payload, params.HTTPRequest.Header.Get("Stripe-Signature"),
		endpointSecret)

	Logger.Debug(string(payload))

	if err != nil {
		Logger.Error("processStripePaymentEvent: Could not parse event! error: %s (%v)", err.Error(), err)
		return errors.New(500, "Could not parse event")
	}

	switch event.Type {
	/**
	canceled
	processing
	requires_action
	requires_capture
	requires_confirmation
	requires_payment_method
	succeeded
	*/
	case "payment_intent.created":
		intent, err := parsePaymentIntent(&event)
		if err != nil {
			return errors.New(500, "Could not parse event intent")
		}
		go processCreatedPayment(intent)
		return nil
	case "payment_intent.succeeded":
		intent, err := parsePaymentIntent(&event)
		if err != nil {
			return errors.New(500, "Could not parse event intent")
		}
		go processSuccessfulPayment(intent)
		return nil
	case "payment_intent.canceled":
		intent, err := parsePaymentIntent(&event)
		if err != nil {
			return errors.New(500, "Could not parse event intent")
		}
		go processCanceledPayment(intent)
		return nil
	case "checkout.session.completed":
		sess, err := parseCheckoutSession(&event)
		if err != nil {
			return errors.New(500, "Could not parse checkout session")
		}
		go processCompletedCheckoutSession(sess)
	case "payment_intent.processing":
	case "payment_intent.requires_action":
	case "payment_intent.requires_capture":
	case "payment_intent.requires_confirmation":
	case "payment_intent.requires_payment_method":
		intent, err := parsePaymentIntent(&event)
		if err != nil {
			return errors.New(500, "Could not parse event intent")
		}
		go processOtherPaymentEvent(intent)
		return nil
	default:
	}
	return nil
}

func parseCheckoutSession(event *stripe.Event) (*stripe.CheckoutSession, errors.Error) {
	var sess stripe.CheckoutSession
	err := json.Unmarshal(event.Data.Raw, &sess)
	if err != nil {
		return nil, errors.New(500, "Could not parse event session")
	}
	return &sess, nil
}

func processCompletedCheckoutSession(sess *stripe.CheckoutSession) {
	sessPaymentStatus := string(sess.PaymentStatus)
	sessPaymentIntentId := sess.PaymentIntent.ID

	if len(sessPaymentStatus) <= 0 || len(sessPaymentIntentId) <= 0 {
		if sess.PaymentIntent != nil && len(sess.PaymentIntent.Status) > 0 && len(sess.PaymentIntent.ID) > 0 {
			Logger.Debug("processCompletedCheckoutSession: setting sessPaymentStatus, sessPaymentIntentId from sess.PaymentIntent: %s, %s",
				sess.PaymentIntent.Status, sess.PaymentIntent.ID)
			sessPaymentStatus = string(sess.PaymentIntent.Status)
			sessPaymentIntentId = sess.PaymentIntent.ID
		} else {
			Logger.Debug("processCompletedCheckoutSession: setting sessPaymentStatus, sessPaymentIntentId from session payment intent: %s",
				sess.PaymentIntent)
			var expand []*string = []*string{stripe.String("payment_intent_data")}
			fullSess, err := session.Get(sess.ID, &stripe.CheckoutSessionParams{
				Expand: expand,
			})
			if err != nil {
				Logger.Error("processCompletedCheckoutSession: Could not retrieve session with payment data! error: %s\n%v\n", err.Error(), err)
			} else {
				bt, err := json.Marshal(fullSess)
				ls := ""
				if err != nil {
					Logger.Debug("Could not marshal full session, error: %s", err)
				} else {
					ls = string(bt)
				}
				Logger.Debug("processCompletedCheckoutSession: Full session: %s\n%v\n%s", fullSess, fullSess, ls)
				sessPaymentStatus = string(fullSess.PaymentIntent.Status)
				sessPaymentIntentId = fullSess.PaymentIntent.ID
			}
		}
	}

	Logger.Debug("processCompletedCheckoutSession: session status %s, payment intent status: %s, payment intent ID: %s",
		sess.Status, sessPaymentStatus, sessPaymentIntentId)

	processCheckoutSession(sess, string(sess.Status), sessPaymentStatus, sessPaymentIntentId)
}

func processCheckoutSession(sess *stripe.CheckoutSession, paymentStatus string, orderStatus string, paymentIntentId string) {
	processPaidOrder(sess.ID, paymentStatus, orderStatus, paymentIntentId)
}

func processPayment(intent *stripe.PaymentIntent, paymentStatus string, orderStatus string) {
	Logger.Debug("processPayment: intent: %s\n")
}

func processPaidOrder(checkoutSessionId string, paymentStatus string, orderStatus string, paymentIntentId string) {
	Logger.Debug("processPayment: checkout session ID %s, payment status %s, order status %s, payment intent ID %s",
		checkoutSessionId, paymentStatus, orderStatus, paymentIntentId)
	var payment *dbModels.Payment
	payment, err := getDBPaymentByCheckoutSessionId(checkoutSessionId)
	if err != nil {
		Logger.Error("Could not get payment by checkout session ID %s; error: ", checkoutSessionId, err.Error())
		return
	}
	payment.Status = paymentStatus
	payment.PaymentIntentId = paymentIntentId

	dbPayment, err := updateDBPayment(payment)
	if err != nil {
		Logger.Error("Could not process paid order for checkout session ID: ID %v\n%s\npayment  %v\n%s\n",
			checkoutSessionId, checkoutSessionId, dbPayment, dbPayment)
		return
	} else {
		Logger.Debug("Processed paid order successfully: checkout session ID %v\n%s\npayment  %v\n%s\n",
			checkoutSessionId, checkoutSessionId, dbPayment, dbPayment)
	}

	order, err := getOrderFromDB(payment.OrderID, true, -1)
	if err != nil {
		Logger.Error("Could not find order for payment: error %v\n%s\npayment  %v\n%s\n", err, err, payment, payment)
		return
	} else {
		Logger.Debug("Found order for payment: order %v\n%s\npayment  %v\n%s\n", order, order, payment, payment)
	}
	order.Status = orderStatus
	orderRes, err := updateDBOrder(order)
	if err != nil {
		Logger.Error("Could not update order for payment: error %v\n%s\norder  %v\n%s\npayment  %v\n%s\n",
			err, err, order, order, payment, payment)
	} else {
		Logger.Debug("Updated order for payment: order %v\n%s\npayment  %v\n%s\n", orderRes, orderRes, payment, payment)
	}
}

func processCreatedPayment(intent *stripe.PaymentIntent) {
	processPayment(intent, "created", "intended")
}

func processSuccessfulPayment(intent *stripe.PaymentIntent) {
	processPayment(intent, "success", "paid")
}

func processCanceledPayment(intent *stripe.PaymentIntent) {
	processPayment(intent, "canceled", "not_paid")
}

func processOtherPaymentEvent(intent *stripe.PaymentIntent) {
	Logger.Debug("processOtherPaymentEvent: client secret %s", intent.ClientSecret)
}

func parsePaymentIntent(event *stripe.Event) (*stripe.PaymentIntent, errors.Error) {
	var intent stripe.PaymentIntent
	err := json.Unmarshal(event.Data.Raw, &intent)
	if err != nil {
		return nil, errors.New(500, "Could not parse event intent")
	}
	return &intent, nil
}
