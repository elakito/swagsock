package swagsock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"

	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
)

const (
	// Revisit the Name of this internal header that represents the tracking-id and request-id pair
	headerRequestKey = "X-Request-Key"
)

var (
	// knownTrackingIDs lists the known tracking-id query parameters used in the websocket upgrade request
	knownTrackingIDs = []string{"x-tracking-id", "X-Atmosphere-tracking-id"}
	defaultLogger    = log.New(ioutil.Discard, "[swagsocket] ", log.LstdFlags)
)

var websocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// allow all connections by default
		return true
	},
}

// NewDefaultCodec returns an instance of the default Codec
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

func (c *defaultCodec) EncodeSwaggerSocketMessage(headers map[string]interface{}, body []byte) ([]byte, error) {
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

// NewConfig returns a default Config object
func NewConfig() *Config {
	conf := &Config{}
	conf.Codec = NewDefaultCodec()
	conf.ResponseMediator = NewDefaultResponseMediator()
	conf.Log = defaultLogger
	return conf
}

// CreateProtocolHandler creates a new ProtocolHandler with the specified codec. If codec is nil, the defaultCodec is used
func CreateProtocolHandler(conf *Config) ProtocolHandler {
	return &protocolHandler{codec: conf.Codec, mediator: conf.ResponseMediator, log: conf.Log, connections: make(map[*websocket.Conn]string)}
}

type protocolHandler struct {
	codec       Codec
	connections map[*websocket.Conn]string
	mediator    ResponseMediator
	log         Logger
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
	conn.SetCloseHandler(func(code int, text string) error {
		conn.Close()
		trackingID := ph.deleteConnection(conn)
		ph.log.Printf("disconnected code=%d, trackingID=%s", code, trackingID)
		ph.mediator.UnsubscribeAll(trackingID)
		return nil
	})
	connlock := &sync.Mutex{}
	baseURI := getBaseURI(r)
	trackingID := getTrackingID(r)
	if trackingID == "" {
		trackingID = uuid.NewV4().String()
	}
	ph.addConnetion(conn, trackingID)

	ph.log.Printf("connected at baseURI=%s, trackingID=%s", baseURI, trackingID)

	go func() {
		for {
			mt, p, err := conn.ReadMessage()
			if err != nil {
				conn.Close()
				break
			}
			id, req := newHTTPRequest(baseURI, trackingID, p, ph.codec)
			resp := newHTTPResponse(id, mt, conn, connlock, ph.codec)
			handler.ServeHTTP(resp, req)
		}
	}()
}

func (ph *protocolHandler) Destroy() {
	ph.RLock()
	defer ph.RUnlock()
	for con := range ph.connections {
		con.Close()
	}
}

func (ph *protocolHandler) addConnetion(conn *websocket.Conn, trackingID string) {
	ph.Lock()
	defer ph.Unlock()
	ph.connections[conn] = trackingID
}

func (ph *protocolHandler) deleteConnection(conn *websocket.Conn) string {
	ph.Lock()
	defer ph.Unlock()
	trackingID := ph.connections[conn]
	delete(ph.connections, conn)
	return trackingID
}

type defaultResponseMediator struct {
	// subscriptionid -> reusableresponder
	responders map[string]*ReusableResponder
	// topic -> subscriptionid set
	topicsubs map[string]map[string]struct{}
	// subscriptionid -> topic
	substopics map[string]string
	sync.RWMutex
}

// NewDefaultResponseMediator returns a new default ResponseMediator
func NewDefaultResponseMediator() ResponseMediator {
	return &defaultResponseMediator{responders: make(map[string]*ReusableResponder), topicsubs: make(map[string]map[string]struct{}), substopics: make(map[string]string)}
}

func (m *defaultResponseMediator) Subscribe(key string, name string, responder middleware.Responder, hello []byte, bye []byte) middleware.Responder {
	rr := NewReusableResponder(name, "", responder, m, hello, bye)
	m.Lock()
	defer m.Unlock()
	m.responders[key] = rr
	return rr
}

func (m *defaultResponseMediator) SubscribeTopic(key string, topic string, name string, responder middleware.Responder, hello []byte, bye []byte) middleware.Responder {
	rr := NewReusableResponder(name, topic, responder, m, hello, bye)
	m.Lock()
	defer m.Unlock()
	var subs map[string]struct{}
	if s, ok := m.topicsubs[topic]; !ok {
		subs = make(map[string]struct{})
		m.topicsubs[topic] = subs
	} else {
		subs = s
	}
	subs[key] = struct{}{}
	m.substopics[key] = topic
	m.responders[key] = rr
	return rr
}

func (m *defaultResponseMediator) Unsubscribe(key string, subid string) {
	unsubid := getDerivedRequestKey(key, subid)
	m.Lock()
	defer m.Unlock()
	if r := m.responders[unsubid]; r != nil {
		if r.bye != nil {
			if r.topic == "" {
				m.write("*", r.bye)
			} else {
				m.writeTopic("*", r.bye)
			}
		}
		if t, ok := m.substopics[unsubid]; ok {
			delete(m.topicsubs[t], unsubid)
		}
		delete(m.responders, unsubid)

	}
}

func (m *defaultResponseMediator) UnsubscribeAll(trackingID string) {
	m.Lock()
	defer m.Unlock()
	var byebye [][]byte
	for key := range m.responders {
		if strings.HasPrefix(key, trackingID) {
			if r := m.responders[key]; r != nil {
				if t, ok := m.substopics[key]; ok {
					delete(m.topicsubs[t], key)
				}
				delete(m.responders, key)
				if r.bye != nil {
					byebye = append(byebye, r.bye)
				}
			}
		}
	}
	for _, bye := range byebye {
		m.write("*", bye)
	}
}

func (m *defaultResponseMediator) Subscribed() []string {
	subs := make([]string, 0)
	seen := make(map[string]struct{})
	m.RLock()
	defer m.RUnlock()
	for _, r := range m.responders {
		if _, ok := seen[r.name]; !ok {
			seen[r.name] = struct{}{}
			subs = append(subs, r.name)
		}
	}
	return subs
}

func (m *defaultResponseMediator) SubscribedTopics() []string {
	topics := make([]string, 0)
	m.RLock()
	defer m.RUnlock()
	for t := range m.topicsubs {
		topics = append(topics, t)
	}
	return topics
}

func (m *defaultResponseMediator) SubscribedTopic(topic string) []string {
	subs := make([]string, 0)
	seen := make(map[string]struct{})
	m.RLock()
	defer m.RUnlock()
	if ss, ok := m.topicsubs[topic]; ok {
		for s := range ss {
			rname := m.responders[s].name
			if _, ok := seen[rname]; !ok {
				subs = append(subs, rname)
			}
		}
	}
	return subs
}

func (m *defaultResponseMediator) Write(name string, data []byte) error {
	m.RLock()
	defer m.RUnlock()
	m.write(name, data)
	return nil
}
func (m *defaultResponseMediator) write(name string, data []byte) {
	for _, r := range m.responders {
		if name == "*" || name == r.name {
			if _, err := r.Write(data); err != nil {
				// log error
			}
		}
	}
}

func (m *defaultResponseMediator) WriteTopic(topic string, data []byte) error {
	m.RLock()
	defer m.RUnlock()
	m.writeTopic(topic, data)
	return nil
}

func (m *defaultResponseMediator) writeTopic(topic string, data []byte) {
	if topic == "*" {
		for _, ss := range m.topicsubs {
			for s := range ss {
				if _, err := m.responders[s].Write(data); err != nil {
					// log error
				}
			}
		}
	} else if ss, ok := m.topicsubs[topic]; ok {
		for s := range ss {
			if _, err := m.responders[s].Write(data); err != nil {
				// log error
			}
		}
	}
}

func newHTTPRequest(baseURI string, trackingID string, data []byte, codec Codec) (string, *http.Request) {
	headers, body, err := codec.DecodeSwaggerSocketMessage(data)
	if err != nil {
		//TODO return the error
	}
	uri := fmt.Sprintf("%s%s", baseURI, getHeader(headers, "path"))
	req, _ := http.NewRequest(getHeader(headers, "method"), uri, bytes.NewReader(body))
	rid := getHeader(headers, "id")
	req.RequestURI = uri
	req.Header.Add(headerRequestKey, buildRequestKey(trackingID, rid))
	copyHeaderToHTTPHeaders(headers, "type", req.Header, "Content-Type")
	copyHeaderToHTTPHeaders(headers, "accept", req.Header, "Accept")
	if aheaders, ok := headers["headers"].(map[string]interface{}); ok {
		for aheader, avalue := range aheaders {
			req.Header.Add(aheader, avalue.(string))
		}
	}
	return rid, req
}

func newHTTPResponse(id string, messageType int, conn *websocket.Conn, connlock *sync.Mutex, codec Codec) http.ResponseWriter {
	resp := &response{id: id, messageType: messageType, headers: make(http.Header), conn: conn, connlock: connlock, codec: codec}
	return resp
}

type response struct {
	id          string
	headers     http.Header
	code        int
	conn        *websocket.Conn
	messageType int
	codec       Codec
	connlock    *sync.Mutex
}

func (r *response) Header() http.Header {
	return r.headers
}

func (r *response) Write(body []byte) (int, error) {
	// flush the buffer when the content-type header is not set
	data, _ := r.codec.EncodeSwaggerSocketMessage(r.buildHeaders(), body)
	r.connlock.Lock()
	r.conn.WriteMessage(r.messageType, data)
	r.connlock.Unlock()
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
	copyHTTPHeaderToHeaders(r.headers, "Content-Type", headers, "type")
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

func copyHeaderToHTTPHeaders(src map[string]interface{}, srckey string, target http.Header, targetkey string) {
	if v, ok := src[srckey]; ok {
		target.Add(targetkey, v.(string))
	}
}

func copyHTTPHeaderToHeaders(src http.Header, srckey string, target map[string]interface{}, targetkey string) {
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

func getTrackingID(r *http.Request) string {
	query := r.URL.Query()
	for _, tname := range knownTrackingIDs {
		if tid := query.Get(tname); tid != "" {
			return tid
		}
	}
	return ""
}

// IsWebsocketUpgradeRequested checks if the request is an upgrade request (based on the jetty's websocket check)
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

// utilities

// GetRequestKey returns the swagger socket request name
func GetRequestKey(req *http.Request) string {
	return req.Header.Get(headerRequestKey)
}

// getDerivedRequestKey returns the swagger socket request name derived from the specified request name and the client-local request id
func getDerivedRequestKey(rkey string, rid string) string {
	return strings.Split(rkey, "#")[0] + "#" + rid
}

// buildRequestKey returns the string consisting of tracking-id and request-id separaterd by '#'
func buildRequestKey(trackingid string, reqid string) string {
	return fmt.Sprintf("%s#%s", trackingid, reqid)
}

// NewReusableResponder wraps the original responder to capture the underlining durable connection for later use
func NewReusableResponder(key string, topic string, r middleware.Responder, mediator ResponseMediator, hello []byte, bye []byte) *ReusableResponder {
	return &ReusableResponder{name: key, topic: topic, responder: r, mediator: mediator, hello: hello, bye: bye}
}

// ReusableResponder is a middleware.Responder which grab the http.ResponseWriter for later reuse
type ReusableResponder struct {
	name      string
	topic     string
	responder middleware.Responder
	hello     []byte
	bye       []byte
	writer    http.ResponseWriter
	mediator  ResponseMediator
}

// WriteResponse writes the initial response to the response writer
func (r *ReusableResponder) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {
	if r.writer == nil {
		r.writer = rw
	}
	r.responder.WriteResponse(rw, producer)
	if r.hello != nil {
		if r.topic == "" {
			r.mediator.Write("*", r.hello)
		} else {
			r.mediator.WriteTopic("*", r.hello)
		}
	}
}

// Write writes the subsequent response to the response writer
func (r *ReusableResponder) Write(b []byte) (int, error) {
	return r.writer.Write(b)
}
