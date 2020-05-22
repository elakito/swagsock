package swagsock

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"

	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
)

const (
	// Revisit the Name of this internal header that represents the tracking-id and request-id pair
	headerRequestKey = "X-Request-Key"
	// ProtocolVersion specifies the current protocol version
	ProtocolVersion = "2.0"
)

var (
	// knownTrackingIDs lists the known tracking-id query parameters used in the websocket upgrade request
	knownTrackingIDs = []string{"x-tracking-id", "X-Atmosphere-tracking-id"}
	defaultLogger    = log.New(ioutil.Discard, "[swagsocket] ", log.LstdFlags)

	//Revisit defining known errors
	errVersionMismatch = errors.New("version_mismatch")
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
	return &protocolHandler{
		codec: conf.Codec, mediator: conf.ResponseMediator, heartbeat: conf.Heartbeat, log: conf.Log, connections: make(map[*websocket.Conn]string), continued: make(map[string]io.WriteCloser)}
}

type protocolHandler struct {
	codec       Codec
	connections map[*websocket.Conn]string
	mediator    ResponseMediator
	continued   map[string]io.WriteCloser // temporary
	heartbeat   int
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
	if ph.heartbeat > 0 {
		heartbeatwait := time.Duration(ph.heartbeat*2) * time.Second
		if err = conn.SetReadDeadline(time.Now().Add(heartbeatwait)); err != nil {
			ph.log.Printf("Failed to set heartbeat deadline: %s", err.Error())
		}

		conn.SetPongHandler(func(string) error {
			if err = conn.SetReadDeadline(time.Now().Add(heartbeatwait)); err != nil {
				ph.log.Printf("Failed to set heartbeat deadline: %s", err.Error())
			}
			return nil
		})
	}

	connlock := &sync.Mutex{}
	baseURI := getBaseURI(r)
	trackingID := getTrackingID(r)
	if trackingID == "" {
		trackingID = uuid.NewV4().String()
	}
	ph.addConnetion(conn, trackingID)

	ph.log.Printf("connected at baseURI=%s, trackingID=%s", baseURI, trackingID)

	var heartbeatstop chan struct{}
	if ph.heartbeat > 0 {
		heartbeatstop = make(chan struct{})
		go func() {
			ticker := time.NewTicker(time.Duration(ph.heartbeat) * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					if err = conn.WriteControl(websocket.PingMessage, []byte{}, time.Time{}); err != nil {
						// this could just be a temporary heartbeat ping failure
						ph.log.Printf("Failed to ping: %s", err.Error())
					}
				case <-heartbeatstop:
					return
				}
			}
		}()
	}

	var handshaked bool
	go func() {
		for {
			mt, p, err := conn.ReadMessage()
			if err != nil {
				conn.Close()
				break
			}
			if handshaked {
				ph.serve(handler, baseURI, trackingID, mt, p, conn, connlock)
			} else if err := ph.handshake(p, trackingID, conn); err != nil {
				conn.Close()
				break
			} else {
				handshaked = true
			}
		}
		if heartbeatstop != nil {
			heartbeatstop <- struct{}{}
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

type connectionWriter interface {
	WriteMessage(messageType int, data []byte) error
	WriteJSON(v interface{}) error
}

func (ph *protocolHandler) handshake(p []byte, trackingID string, conn connectionWriter) error {
	var hr *HandshakeRequest
	if err := json.Unmarshal(p, &hr); err != nil {
		return err
	}
	if hr.Version != ProtocolVersion {
		if err := conn.WriteJSON(&HandshakeResponse{Version: ProtocolVersion, Error: "version_mismatch"}); err != nil {
			ph.log.Printf("Failed to write the handshake response: %s", err.Error())
		}
		return errVersionMismatch
	}
	if err := conn.WriteJSON(&HandshakeResponse{Version: ProtocolVersion, TrackingID: trackingID}); err != nil {
		ph.log.Printf("Failed to write the handshake response: %s", err.Error())
	}
	return nil
}

func (ph *protocolHandler) serve(handler http.Handler, baseURI string, trackingID string, mtype int, p []byte, conn connectionWriter, connlock *sync.Mutex) {
	headers, body, err := ph.codec.DecodeSwaggerSocketMessage(p)
	if err != nil {
		ph.log.Printf("Error %s at decoding swaggersocket message. Skipping it", err.Error())
		return
	}

	cont := getBoolHeader(headers, "continue")
	rid := getStringHeader(headers, "id")
	if cwriter, ok := ph.continued[rid]; !ok && cont {
		// for the first segment of a new continued series, dispatch it asynchronously to the handler and write the data to its writer
		creader, cwriter := io.Pipe()
		ph.continued[rid] = cwriter
		go func() {
			req := newHTTPRequest(baseURI, trackingID, rid, headers, creader)
			resp := newHTTPResponse(rid, mtype, conn, connlock, ph.codec)
			handler.ServeHTTP(resp, req)
		}()
		if _, err = cwriter.Write(body); err != nil {
			ph.log.Printf("Failed to write the first part: %s", err.Error())
		}
	} else if ok {
		// for one of the subsequent segments of a continued series, write the data to its writer
		if _, err = cwriter.Write(body); err != nil {
			ph.log.Printf("Failed to write a subsequent part: %s", err.Error())
		}
		if !cont {
			// delete the cwriter
			if err = cwriter.Close(); err != nil {
				ph.log.Printf("Failed to close the writer: %s", err.Error())
			}
			delete(ph.continued, rid)
		}
	} else {
		// for a non-continued single request, dispatch it the handler
		req := newHTTPRequest(baseURI, trackingID, rid, headers, bytes.NewReader(body))
		resp := newHTTPResponse(rid, mtype, conn, connlock, ph.codec)
		handler.ServeHTTP(resp, req)
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
				// log error TODO use the cofigured logger instead
				defaultLogger.Printf("failed to write: %s", err.Error())
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
					// log error TODO use the cofigured logger instead
					defaultLogger.Printf("failed to write: %s", err.Error())
				}
			}
		}
	} else if ss, ok := m.topicsubs[topic]; ok {
		for s := range ss {
			if _, err := m.responders[s].Write(data); err != nil {
				// log error TODO use the cofigured logger instead
				defaultLogger.Printf("failed to write: %s", err.Error())
			}
		}
	}
}

func newHTTPRequest(baseURI string, trackingID string, rid string, headers map[string]interface{}, body io.Reader) *http.Request {
	uri := fmt.Sprintf("%s%s", baseURI, getStringHeader(headers, "path"))
	req, _ := http.NewRequest(getStringHeader(headers, "method"), uri, body) //nolint:errcheck
	req.RequestURI = uri
	req.Header.Add(headerRequestKey, buildRequestKey(trackingID, rid))
	copyHeaderToHTTPHeaders(headers, "type", req.Header, "Content-Type")
	copyHeaderToHTTPHeaders(headers, "accept", req.Header, "Accept")
	if needsChunkEnabled(body) {
		req.TransferEncoding = append(req.TransferEncoding, "chunked")
	}
	if aheaders, ok := headers["headers"].(map[string]interface{}); ok {
		for aheader, avalue := range aheaders {
			req.Header.Add(aheader, avalue.(string))
		}
	}
	return req
}

func needsChunkEnabled(body io.Reader) bool {
	switch body.(type) {
	case *bytes.Buffer, *bytes.Reader, *strings.Reader:
		return false
	default:
		return true
	}
}

func newHTTPResponse(id string, messageType int, conn connectionWriter, connlock *sync.Mutex, codec Codec) http.ResponseWriter {
	resp := &responseWriter{id: id, messageType: messageType, headers: make(http.Header), conn: conn, connlock: connlock, codec: codec}
	return resp
}

type responseWriter struct {
	id          string
	headers     http.Header
	code        int
	conn        connectionWriter
	messageType int
	codec       Codec
	connlock    *sync.Mutex
}

func (r *responseWriter) Header() http.Header {
	return r.headers
}

func (r *responseWriter) Write(body []byte) (int, error) {
	// flush the buffer when the content-type header is not set
	data, err := r.codec.EncodeSwaggerSocketMessage(r.buildHeaders(), body)
	if err != nil {
		return 0, err
	}
	r.connlock.Lock()
	if err = r.conn.WriteMessage(r.messageType, data); err != nil {
		return 0, err
	}
	r.connlock.Unlock()
	return len(body), nil
}

func (r *responseWriter) WriteHeader(code int) {
	r.code = code
	ctype := r.headers.Get("Content-Type")
	if ctype == "" {
		// flush the buffer when the content-type header is not set
		data, _ := r.codec.EncodeSwaggerSocketMessage(r.buildHeaders(), nil)
		if err := r.conn.WriteMessage(r.messageType, data); err != nil {
			// TODO use the cofigured logger instead
			defaultLogger.Printf("Failed to flush the buffer: %s", err.Error())
		}
	}
}

func (r *responseWriter) buildHeaders() map[string]interface{} {
	headers := make(map[string]interface{})
	headers["id"] = r.id
	headers["code"] = r.code
	copyHTTPHeaderToHeaders(r.headers, "Content-Type", headers, "type")
	//TODO fill other headers
	return headers
}

func getStringHeader(headers map[string]interface{}, key string) string {
	// we know that the value is of string if present
	if v, ok := headers[key]; ok {
		return v.(string)
	}
	return ""
}

func getBoolHeader(headers map[string]interface{}, key string) bool {
	// we know that the value is of string if present
	if v, ok := headers[key]; ok {
		return v.(bool)
	}
	return false
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

// WriteResponse writes the initial responseWriter to the responseWriter writer
func (r *ReusableResponder) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {
	if r.writer == nil {
		r.writer = rw
	}
	r.responder.WriteResponse(rw, producer)
	if r.hello != nil {
		if r.topic == "" {
			if err := r.mediator.Write("*", r.hello); err != nil {
				// TODO use the cofigured logger instead
				defaultLogger.Printf("Failed to broadcast hello to subscribers: %s", err.Error())
			}
		} else {
			if err := r.mediator.WriteTopic("*", r.hello); err != nil {
				// TODO use the cofigured logger instead
				defaultLogger.Printf("Failed to broadcast hello to topics: %s", err.Error())
			}
		}
	}
}

// Write writes the subsequent responseWriter to the responseWriter writer
func (r *ReusableResponder) Write(b []byte) (int, error) {
	return r.writer.Write(b)
}
