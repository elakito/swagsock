package swagsock

import (
	"net/http"
)

type Codec interface {
	// Decodes the swaggersocket wire message into its map part and the body part
	DecodeSwaggerSocketMessage(data []byte) (map[string]interface{}, []byte, error)
	// Encodes the map part and the body into its swaggersocket wire message
	EncodeSwaggerSocketMessage(headers map[string]interface{}, body []byte) ([]byte, error)
}

type ProtocolHandler interface {
	// Returns the codec instance used by this protocol handler
	GetCodec() Codec
	// Serve the request using the protocol
	Serve(handler http.Handler, w http.ResponseWriter, r *http.Request)
	// Destroy the handler
	Destroy()
}

