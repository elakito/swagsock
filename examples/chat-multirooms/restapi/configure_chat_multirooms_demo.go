// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"

	"github.com/elakito/swagsock/examples/chat-multirooms/models"
	"github.com/elakito/swagsock/examples/chat-multirooms/restapi/operations"
	"github.com/elakito/swagsock/swagsock"
)

//go:generate swagger generate server --target ../../chat-multiroom --name ChatMultiroomsDemo --spec ../swagger.json

var (
	usersstat        = make(map[string]int32)
	statuslock       = &sync.RWMutex{}
	responseMediator swagsock.ResponseMediator
)

func getChatSummary() (int32, []string) {
	var total int32
	names := make([]string, 0, len(usersstat))
	statuslock.RLock()
	defer statuslock.RUnlock()
	for name, count := range usersstat {
		names = append(names, name)
		total += count
	}
	return total, names
}
func updateChatSummary(name string) {
	statuslock.Lock()
	defer statuslock.Unlock()
	usersstat[name]++
}
func configureFlags(api *operations.ChatMultiroomsDemoAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.ChatMultiroomsDemoAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// api.Logger = log.Printf

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	api.ChatHandler = operations.ChatHandlerFunc(func(params operations.ChatParams) middleware.Responder {
		if params.Name == "" {
			return &errorResp{http.StatusBadRequest, "operation .Greet requires name parameter and json body", make(http.Header)}
		}
		updateChatSummary(params.Name)
		payload := &models.Message{Name: params.Name, Room: params.Room, Text: params.Body.Text, Type: "message"}
		pb, _ := payload.MarshalBinary()
		responseMediator.WriteTopic(params.Room, pb)
		return operations.NewChatOK().WithPayload(payload)
	})

	api.RoomMembersHandler = operations.RoomMembersHandlerFunc(func(params operations.RoomMembersParams) middleware.Responder {
		return operations.NewRoomMembersOK().WithPayload(responseMediator.SubscribedTopic(params.Room))
	})

	api.RoomsHandler = operations.RoomsHandlerFunc(func(params operations.RoomsParams) middleware.Responder {
		return operations.NewRoomsOK().WithPayload(responseMediator.SubscribedTopics())
	})

	api.SubscribeHandler = operations.SubscribeHandlerFunc(func(params operations.SubscribeParams) middleware.Responder {
		hello := &models.Message{Name: params.Name, Room: params.Room, Type: "joined"}
		hb, _ := hello.MarshalBinary()
		bye := &models.Message{Name: params.Name, Room: params.Room, Type: "left"}
		bb, _ := bye.MarshalBinary()
		responder := responseMediator.SubscribeTopic(swagsock.GetRequestKey(params.HTTPRequest), params.Room, params.Name, operations.NewSubscribeOK(), hb, bb)
		return responder
	})

	api.UnsubscribeHandler = operations.UnsubscribeHandlerFunc(func(params operations.UnsubscribeParams) middleware.Responder {
		responseMediator.Unsubscribe(swagsock.GetRequestKey(params.HTTPRequest), params.Sid)
		total, cnames := getChatSummary()
		return operations.NewUnsubscribeOK().WithPayload(&models.Summary{Authors: cnames, Total: total})
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
	conf := swagsock.NewConfig()
	conf.Log = log.New(os.Stdout, "[swagsocket] ", log.LstdFlags)
	responseMediator = conf.ResponseMediator

	protocolHandler := swagsock.CreateProtocolHandler(conf)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// use the protocol handler to handle websocket requests
		if swagsock.IsWebsocketUpgradeRequested(r) {
			protocolHandler.Serve(handler, w, r)
			return
		}
		// Shortcut helpers for swagger-ui
		if r.URL.Path == "/ui" {
			http.Redirect(w, r, "/ui/", http.StatusFound)
			return
		}

		// Serving ./ui/
		if strings.Index(r.URL.Path, "/ui/") == 0 {
			http.StripPrefix("/ui/", http.FileServer(http.Dir("ui"))).ServeHTTP(w, r)
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
