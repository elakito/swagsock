package swagsock

import (
	"net/http"
	"strings"

	"github.com/go-openapi/runtime"
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
	// Subscribe subscribes to the specified name. Parameters responder, hello, and bye represent the responseWriter sent
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
	Heartbeat        int
	Log              Logger
}

// HandshakeRequest is the handshake request message that is sent from the client
type HandshakeRequest struct {
	Version string `json:"version"`
}

// HandshakeResponse is the handshake responseWriter message that is sent from the server
type HandshakeResponse struct {
	Version    string `json:"version"`
	TrackingID string `json:"trackingID,omitempty"`
	Error      string `json:"error,omitempty"`
}

// ClientTransport is the interface for submitting requests and it defines Submit or SubmitAsync for the synchronous
// and asynchronous invocation supported in swagsock
type ClientTransport interface {
	//Submit(string, RequestWriter, ResponseReader, AuthInfoWriter) (interface{}, error)
	Submit(*runtime.ClientOperation) (interface{}, error)
	//SubmitAsync(string, RequestWriter, ResponseReader, AuthInfoWriter, func(string, interface{})) (string, error)
	SubmitAsync(*runtime.ClientOperation, func(string, interface{}), SubmitAsyncOption) (string, error)
	//Close closes the socket
	Close()
}

// SubmitAsyncMode represents one of the async submit mode none, subscribe, or unsubscribe
type SubmitAsyncMode int

const (
	// SubmitAsyncModeNone represents no option
	SubmitAsyncModeNone SubmitAsyncMode = iota
	// SubmitAsyncModeSubscribe represents the subscription option
	SubmitAsyncModeSubscribe
	// SubmitAsyncModeUnsubscribe represents the unsubscription option
	SubmitAsyncModeUnsubscribe
)

// SubmitAsyncOption represents one of the async submit option none, subscribe, or unsubscribe
type SubmitAsyncOption string

// Is tests if this option is of the specified mode
func (o SubmitAsyncOption) Is(mode SubmitAsyncMode) bool {
	if o == SubmitAsyncOptionNone {
		return mode == SubmitAsyncModeNone
	} else if o == SubmitAsyncOptionSubscribe {
		return mode == SubmitAsyncModeSubscribe
	} else {
		return mode == SubmitAsyncModeUnsubscribe
	}
}

// Param returns the parameter associated with this option
func (o SubmitAsyncOption) Param() string {
	ostr := string(o)
	if p := strings.Index(ostr, ":"); p > 0 {
		return ostr[p+1:]
	}
	return ""
}

// SubmitAsyncOptionNone represents no specific option
const SubmitAsyncOptionNone = ""

// SubmitAsyncOptionSubscribe represents the subscription option
const SubmitAsyncOptionSubscribe = "subscribe"

// SubmitAsyncOptionUnsubscribe represents the unsubscription option to unsubscribe from the specified subscription id
func SubmitAsyncOptionUnsubscribe(sid string) SubmitAsyncOption {
	return SubmitAsyncOption("unsubscribe:" + sid)
}
