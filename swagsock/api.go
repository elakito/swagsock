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
	// Subscribe subscribes to the specified name. Parameters responder, hello, and bye represent the response sent
	// to the requester, the optional hello and bye messages sent to the subscribers.
	Subscribe(key string, name string, responder middleware.Responder, hello []byte, bye []byte) middleware.Responder
	SubscribeTopic(key string, topic string, name string, responder middleware.Responder, hello []byte, bye []byte) middleware.Responder

	// Unsubscribe cancels the subscription associated with the subscription id (i.e., the request id used for the subscription)
	Unsubscribe(key string, subid string)

	// UnsubscribeAll cancels all the subscriptions associated with the swaggersocket key.
	UnsubscribeAll(key string)

	// Subscribed returns the list of active subscriber names.
	Subscribed() []string
	SubscribedTopics() []string
	SubscribedTopic(topic string) []string

	// Write writes to the specified subscriber or to all subscribers when name is '*'.
	Write(name string, data []byte) error
	WriteTopic(topic string, data []byte) error
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
