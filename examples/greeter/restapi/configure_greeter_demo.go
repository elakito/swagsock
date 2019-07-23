// Code generated by go-swagger; DO NOT EDIT.

package restapi

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/elakito/swagsock/examples/greeter/models"
	"github.com/elakito/swagsock/examples/greeter/restapi/operations"
	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"

	"github.com/elakito/swagsock/swagsock"
)

// This file is safe to edit. Once it exists it will not be overwritten

//go:generate swagger generate server --target .. --name greeter-demo --spec ../swagger.yaml

var (
	counter          int32
	greeted          = make(map[string]int32)
	statuslock       = &sync.RWMutex{}
	subscriptions    = make(map[string]*greetingSubscription)
	subscriptionlock = &sync.RWMutex{}
)

type greetingSubscription struct {
	name string
	sub  *swagsock.ReusableResponder
}

func getGreetSummary() (int32, []string) {
	var total int32
	names := make([]string, 0, len(greeted))
	statuslock.RLock()
	defer statuslock.RUnlock()
	for name, count := range greeted {
		names = append(names, name)
		total += count
	}
	return total, names
}

func getGreetStatus(name string) int32 {
	statuslock.RLock()
	defer statuslock.RUnlock()
	return greeted[name]
}

func updateGreet(name string) {
	statuslock.Lock()
	defer statuslock.Unlock()
	greeted[name]++
}

func configureFlags(api *operations.GreeterDemoAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.GreeterDemoAPI) http.Handler {

	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// api.Logger = log.Printf

	api.JSONConsumer = runtime.JSONConsumer()

	api.MultipartformConsumer = runtime.DiscardConsumer

	api.TxtConsumer = runtime.TextConsumer()

	api.JSONProducer = runtime.JSONProducer()

	api.EchoHandler = operations.EchoHandlerFunc(func(params operations.EchoParams) middleware.Responder {
		return operations.NewEchoOK().WithPayload(&models.EchoReply{Echo: params.Body})
	})
	api.GetGreetStatusHandler = operations.GetGreetStatusHandlerFunc(func(params operations.GetGreetStatusParams) middleware.Responder {
		count := getGreetStatus(params.Name)
		return operations.NewGetGreetStatusOK().WithPayload(&models.GreetingStatus{Name: params.Name, Count: count})
	})
	api.GetGreetSummaryHandler = operations.GetGreetSummaryHandlerFunc(func(params operations.GetGreetSummaryParams) middleware.Responder {
		total, gnames := getGreetSummary()
		return operations.NewGetGreetSummaryOK().WithPayload(&models.GreetingSummary{Greeted: gnames, Total: total})
	})
	api.GreetHandler = operations.GreetHandlerFunc(func(params operations.GreetParams) middleware.Responder {
		if params.Name == "" || params.Body == nil {
			return &errorResp{http.StatusBadRequest, "operation .Greet requires name parameter and json body", make(http.Header)}
		}
		updateGreet(params.Name)
		payload := &models.GreetingReply{From: params.Name, Name: params.Body.Name, Text: params.Body.Text}
		pb, _ := payload.MarshalBinary()
		//TODO do the writing and purging of the subscriptions asynchronously
		var dead []string
		subscriptionlock.RLock()
		for subid, subscription := range subscriptions {
			if params.Body.Name == "*" || subscription.name == params.Body.Name {
				_, err := subscription.sub.Write(pb)
				if err != nil {
					dead = append(dead, subid)
				}
			}
		}
		subscriptionlock.RUnlock()
		if len(dead) > 0 {
			subscriptionlock.Lock()
			for _, subid := range dead {
				delete(subscriptions, subid)
			}
			subscriptionlock.Unlock()
		}
		return operations.NewGreetOK().WithPayload(payload)
	})
	api.PingHandler = operations.PingHandlerFunc(func(params operations.PingParams) middleware.Responder {
		return operations.NewPingOK().WithPayload(&models.Pong{Pong: atomic.AddInt32(&counter, 1)})
	})
	api.UploadHandler = operations.UploadHandlerFunc(func(params operations.UploadParams) middleware.Responder {
		var size int
		if data, err := ioutil.ReadAll(params.File); err == nil {
			size = len(data)
			params.File.Close()
		}
		return operations.NewUploadOK().WithPayload(&models.UploadStatus{Name: params.Name, Size: int32(size)})
	})
	api.SubscribeHandler = operations.SubscribeHandlerFunc(func(params operations.SubscribeParams) middleware.Responder {
		responder := swagsock.NewReusableResponder(operations.NewSubscribeOK())
		id := swagsock.GetRequestKey(params.HTTPRequest)
		subscriptionlock.Lock()
		subscriptions[id] = &greetingSubscription{name: params.Name, sub: responder}
		subscriptionlock.Unlock()
		return responder
	})
	api.UnsubscribeHandler = operations.UnsubscribeHandlerFunc(func(params operations.UnsubscribeParams) middleware.Responder {
		id := swagsock.GetRequestKey(params.HTTPRequest)
		unsubid := swagsock.GetDerivedRequestKey(id, params.Sid)
		subscriptionlock.Lock()
		delete(subscriptions, unsubid)
		subscriptionlock.Unlock()
		total, gnames := getGreetSummary()
		return operations.NewGetGreetSummaryOK().WithPayload(&models.GreetingSummary{Greeted: gnames, Total: total})
	})

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
// scheme value will be set accordingly: "http", "https" or "unix"
func configureServer(s *http.Server, scheme, addr string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return globalMiddleware(handler)
}

// The globalMiddleware uses the swaggersocket handler to handle websocket requests
func globalMiddleware(handler http.Handler) http.Handler {

	// instantiate the protocol handler
	protocolHandler := swagsock.CreateProtocolHandler(nil)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// use the protocol handler to handle websocket requests
		if swagsock.IsWebsocketUpgradeRequested(r) {
			protocolHandler.Serve(handler, w, r)
			return
		}

		// handle normal http requests
		handler.ServeHTTP(w, r)
	})
}

// copied from middleware/not_implemented.go
type errorResp struct {
	code     int
	response interface{}
	headers  http.Header
}

func (e *errorResp) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {
	for k, v := range e.headers {
		for _, val := range v {
			rw.Header().Add(k, val)
		}
	}
	if e.code > 0 {
		rw.WriteHeader(e.code)
	} else {
		rw.WriteHeader(http.StatusInternalServerError)
	}
	if err := producer.Produce(rw, e.response); err != nil {
		panic(err)
	}
}
