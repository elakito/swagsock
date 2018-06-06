package swagsock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

var websocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// allow all connections by default
		return true
	},
}

func NewDefaultCodec() Codec {
	return &defaultCodec{}
}

type defaultCodec struct {
}

func (c *defaultCodec) DecodeSwaggerSocketMessage(data []byte) (map[string]interface{}, []byte, error) {
	reader := json.NewDecoder(bytes.NewReader(data))
	var headers map[string]interface{}
	err := reader.Decode(&headers)
	if err != nil {
		return nil, nil, err
	}
	if v, ok := headers["code"]; ok {
		headers["code"] = int(v.(float64))
	}
	body, err := ioutil.ReadAll(reader.Buffered())
	if err != nil {
		return nil, nil, err
	}
	return headers, body, nil
}

func (c *defaultCodec) EncodeSwaggerSocketMessage(headers map[string]interface{}, body []byte) ([]byte, error){
	buf := new(bytes.Buffer)
	hb, err := json.Marshal(headers)
	if err != nil {
		return nil, err
	}
	buf.Write(hb)
	if body != nil {
		buf.Write(body)
	}
	return buf.Bytes(), nil
}

func copyValue(src map[string]interface{}, target map[string]interface{}, key string) {
	if v, ok := src[key]; ok {
		target[key] = v
	}
}

// Create a new ProtocolHandler with the specified codec. If codec is nil, the defaultCodec is used
func CreateProtocolHandler(codec Codec) ProtocolHandler {
	if codec == nil {
		codec = NewDefaultCodec()
	}
	return &protocolHandler{codec: codec, connections: make(map[*websocket.Conn]struct{})}
}

type protocolHandler struct {
	codec       Codec
	connections map[*websocket.Conn]struct{}
	sync.RWMutex
}

func (ph *protocolHandler) GetCodec() Codec {
	return ph.codec
}

func (ph *protocolHandler) Serve(handler http.Handler, w http.ResponseWriter, r *http.Request) {
	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to upgrade: %v", err), http.StatusInternalServerError)
		return
	}
	ph.addConnetion(conn)
	baseURI := getBaseURI(r)

	go func() {
		for {
			mt, p, err := conn.ReadMessage()
			if err != nil {
				conn.Close()
				ph.deleteConnection(conn)
				break
			}
			id, req := newHttpRequest(baseURI, p, ph.codec)
			resp := newHttpResponse(id, mt, conn, ph.codec)
			handler.ServeHTTP(resp, req)
		}
	}()
}

func (ph *protocolHandler) Destroy() {
	ph.RLock()
	defer ph.RUnlock()
	for con, _ := range ph.connections {
		con.Close()
	}
}

func (ph *protocolHandler) addConnetion(conn *websocket.Conn) {
	ph.Lock()
	defer ph.Unlock()
	ph.connections[conn] = struct{}{}
}

func (ph *protocolHandler) deleteConnection(conn *websocket.Conn) {
	ph.Lock()
	defer ph.Unlock()
	delete(ph.connections, conn)
}

func newHttpRequest(baseURI string, data []byte, codec Codec) (string, *http.Request) {
	headers, body, err := codec.DecodeSwaggerSocketMessage(data)
	if err != nil {
		//TODO return the error
	}
	uri := fmt.Sprintf("%s%s", baseURI, getHeader(headers, "path"))
	req, _ := http.NewRequest(getHeader(headers, "method"), uri, bytes.NewReader(body))
	req.RequestURI = uri
	copyHeaderToHttpHeaders(headers, "type", req.Header, "Content-Type")
	copyHeaderToHttpHeaders(headers, "accept", req.Header, "Accept")
	return getHeader(headers, "id"), req
}

func newHttpResponse(id string, messageType int, conn *websocket.Conn, codec Codec) http.ResponseWriter {
	resp := &response{id: id, messageType: messageType, headers: make(http.Header), conn: conn, codec: codec}
	return resp
}

type response struct {
	id          string
	headers     http.Header
	code        int
	conn        *websocket.Conn
	messageType int
	codec       Codec
}

func (r *response) Header() http.Header {
	return r.headers
}

func (r *response) Write(body []byte) (int, error) {
	// flush the buffer when the content-type header is not set
	data, _ := r.codec.EncodeSwaggerSocketMessage(r.buildHeaders(), body)
	r.conn.WriteMessage(r.messageType, data)
	return len(body), nil
}

func (r *response) WriteHeader(code int) {
	r.code = code
	ctype := r.headers.Get("Content-Type")
	if ctype == "" {
		// flush the buffer when the content-type header is not set
		data, _ := r.codec.EncodeSwaggerSocketMessage(r.buildHeaders(), nil)
		r.conn.WriteMessage(r.messageType, data)
	}
}

func (r *response) buildHeaders() map[string]interface{} {
	headers := make(map[string]interface{})
	headers["id"] = r.id
	headers["code"] = r.code
	copyHttpHeaderToHeaders(r.headers, "Content-Type", headers, "type")
	//TODO fill other headers
	return headers
}

func getHeader(headers map[string]interface{}, key string) string {
	// we know that the value is of string if present
	if v, ok := headers[key]; ok {
		return v.(string)
	}
	return ""
}

func copyHeaderToHttpHeaders(src map[string]interface{}, srckey string, target http.Header, targetkey string) {
	if v, ok := src[srckey]; ok {
		target.Add(targetkey, v.(string))
	}
}

func copyHttpHeaderToHeaders(src http.Header, srckey string, target map[string]interface{}, targetkey string) {
	if v := src.Get(srckey); v != "" {
		target[targetkey] = v
	}
}

func getBaseURI(r *http.Request) string {
	baseURI := r.URL.Path
	if baseURI == "" {
		requri, _ := url.ParseRequestURI(r.RequestURI)
		baseURI = requri.Path
	}
	if strings.HasSuffix(baseURI, "/") {
		return baseURI[:len(baseURI)-1]
	}
	return baseURI
}

// based on the jetty's websocket check
func IsWebsocketUpgradeRequested(r *http.Request) bool {
	if "GET" != r.Method {
		return false
	}
	upgrading := false
	for _, iconnection := range strings.Split(r.Header.Get("Connection"), ",") {
		if strings.EqualFold("Upgrade", iconnection) {
			upgrading = true
			break
		}
	}
	if !upgrading {
		return false
	}
	if !strings.EqualFold("Websocket", r.Header.Get("Upgrade")) {
		return false
	}
	return true
}
