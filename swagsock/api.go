package swagsock

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// Codec is the interface that wraps the decode and encode methods.
type Codec interface {
	// Decodes the swaggersocket wire message into its map part and the body part
	DecodeSwaggerSocketMessage(data []byte) (map[string]interface{}, []byte, error)
	// Encodes the map part and the body into its swaggersocket wire message
	EncodeSwaggerSocketMessage(headers map[string]interface{}, body []byte) ([]byte, error)
}

// ProtocolHandler is the interface to interact with the protocol handler
type ProtocolHandler interface {
	// Returns the codec instance used by this protocol handler
	GetCodec() Codec
	// Serve the request using the protocol
	Serve(handler http.Handler, w http.ResponseWriter, r *http.Request)
	// Destroy the handler
	Destroy()
}

// ResponseMediator is the interface to manage responders and delivery of responses to the subscribers
type ResponseMediator interface {
	Subscribe(key string, name string, responder middleware.Responder, hello []byte, bye []byte) middleware.Responder
	Unsubscribe(key string, subid string)
	UnsubscribeAll(key string)
	Subscribed() []string
	Write(name string, data []byte) error
}

// Logger is the interface for logging
type Logger interface {
	Print(...interface{})
	Printf(string, ...interface{})
	Println(...interface{})
}

// Config is the configuration object
type Config struct {
	Codec            Codec
	ResponseMediator ResponseMediator
	Log              Logger
}
