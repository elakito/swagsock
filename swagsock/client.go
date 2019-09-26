package swagsock

import (
	"bytes"
	"encoding"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.wdf.sap.corp/velocity/vflow/src/components/openapi/swagsock"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/gorilla/websocket"
)

// NewTransport creates a new ClientTransport for swaggersocket
func NewTransport(url string) ClientTransport {
	t := &wstransport{url: url, codec: NewDefaultCodec(), pending: make(map[string]asyncResponse)}

	t.consumers = map[string]runtime.Consumer{
		runtime.JSONMime:    runtime.JSONConsumer(),
		runtime.XMLMime:     runtime.XMLConsumer(),
		runtime.TextMime:    runtime.TextConsumer(),
		runtime.DefaultMime: runtime.ByteStreamConsumer(),
	}
	t.producers = map[string]runtime.Producer{
		runtime.JSONMime:    runtime.JSONProducer(),
		runtime.XMLMime:     runtime.XMLProducer(),
		runtime.TextMime:    runtime.TextProducer(),
		runtime.DefaultMime: runtime.ByteStreamProducer(),
	}

	t.connect()
	return t
}

type wstransport struct {
	url       string
	codec     swagsock.Codec
	conn      *websocket.Conn
	consumers map[string]runtime.Consumer
	producers map[string]runtime.Producer

	nextid  int32
	pending map[string]asyncResponse
	lock    sync.RWMutex
}

func (t *wstransport) getNextID() string {
	return strconv.Itoa(int(atomic.AddInt32(&t.nextid, 1)))
}

func (t *wstransport) connect() error {
	c, _, err := websocket.DefaultDialer.Dial(t.url, nil)
	if err != nil {
		return err
	}
	t.conn = c
	if err := t.conn.WriteJSON(&HandshakeRequest{Version: ProtocolVersion}); err != nil {
		return err
	}
	go func() {
		var handshaked bool
		for {
			_, message, err := t.conn.ReadMessage()
			if err == nil {
				if handshaked {
					headers, body, err := t.codec.DecodeSwaggerSocketMessage(message)
					if err == nil {
						reqid := headers["id"].(string)
						// only handle if there is a pending request
						if fresp := t.removeAsyncResponse(reqid, false); fresp != nil {
							res := &response{code: headers["code"].(int), id: reqid}
							if mediaType, found := headers["type"].(string); found {
								res.mediaType = mediaType
							}
							if len(body) > 0 {
								res.body = ioutil.NopCloser(bytes.NewBuffer(body))
							}
							fresp.set(res)
						}
					}
				} else {
					var hr *HandshakeResponse
					if err := json.Unmarshal(message, &hr); err != nil {
						log.Println("Error invalid handshake:", err)
						return
					}
					if hr.Error != "" {
						log.Println("Error handshake:", hr.Error)
						return
					}
					handshaked = true
				}
			} else {
				log.Println("Error read:", err)
				return
			}
		}
	}()

	return nil
}

func (t *wstransport) createRequest(reqid string, operation *runtime.ClientOperation) ([]byte, error) {
	req := &request{
		pathPattern: operation.PathPattern,
		method:      operation.Method,
		writer:      operation.Params,
		header:      make(http.Header),
		query:       make(url.Values),
		timeout:     client.DefaultTimeout,
	}

	// build the data
	if err := req.writer.WriteToRequest(req, strfmt.Default); err != nil {
		return nil, err
	}

	rawpath := req.GetPath()
	rawquery := req.query.Encode()
	if rawquery != "" {
		rawpath = fmt.Sprintf("%s?%s", rawpath, rawquery)
	}

	var body string
	if req.payload != nil {
		switch req.payload.(type) {
		case string:
			body = req.payload.(string)
		case encoding.BinaryMarshaler:
			binbody, _ := req.payload.(encoding.BinaryMarshaler).MarshalBinary()
			body = string(binbody)
		}
	}

	headers := map[string]string{
		"id":     reqid,
		"method": operation.Method,
		"path":   rawpath,
	}

	for _, mediaType := range operation.ConsumesMediaTypes {
		// Pick first non-empty media type
		if mediaType != "" {
			headers["type"] = mediaType
			break
		}
	}
	rawheaders, err := json.Marshal(headers)
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf(`%s%s`, rawheaders, body)), nil
}

func (t *wstransport) Submit(operation *runtime.ClientOperation) (interface{}, error) {
	reqid := t.getNextID()
	rawmessage, err := t.createRequest(reqid, operation)
	if err != nil {
		return nil, err
	}

	fresp := newFutureResponse(reqid)
	t.putAsyncResponse(reqid, fresp)

	t.writeMessage(rawmessage)

	// TODO make the timeout for the synchronous response configurable
	response, err := fresp.Get(5 * time.Second)
	if err != nil {
		return nil, err
	}
	cons, ok := t.consumers[response.mediaType]
	if !ok {
		// scream about not knowing what to do
		return nil, fmt.Errorf("no consumer: %q", response.mediaType)
	}
	return operation.Reader.ReadResponse(response, cons)
}

func (t *wstransport) SubmitAsync(operation *runtime.ClientOperation, cb func(string, interface{}), sao SubmitAsyncOption) (string, error) {
	reqid := t.getNextID()
	rawmessage, err := t.createRequest(reqid, operation)
	if err != nil {
		return "", err
	}

	if sao.Is(SubmitAsyncModeUnsubscribe) {
		t.removeAsyncResponse(sao.Param(), true)
	}
	fresp := newCallbackResponse(reqid, func(r *response) {
		if cons, ok := t.consumers[r.mediaType]; ok {
			if resp, err := operation.Reader.ReadResponse(r, cons); err == nil {
				cb(reqid, resp)
			} else {
				cb(reqid, err)
			}
		}
	}, sao.Is(SubmitAsyncModeSubscribe))
	t.putAsyncResponse(reqid, fresp)

	t.writeMessage(rawmessage)
	return reqid, nil
}

func (t *wstransport) putAsyncResponse(reqid string, aresp asyncResponse) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.pending[reqid] = aresp
}

func (t *wstransport) removeAsyncResponse(reqid string, force bool) asyncResponse {
	t.lock.Lock()
	defer t.lock.Unlock()
	if aresp, found := t.pending[reqid]; found {
		if !aresp.isSticky() || force {
			delete(t.pending, reqid)
		}
		return aresp
	}
	return nil
}

func (t *wstransport) writeMessage(data []byte) error {
	if t.conn == nil {
		return fmt.Errorf("not connected")
	}
	return t.conn.WriteMessage(websocket.TextMessage, data)
}

func (t *wstransport) Close() {
	if t.conn != nil {
		t.conn.Close()
		t.conn = nil
	}
}

type response struct {
	code      int
	message   string
	id        string
	mediaType string
	body      io.ReadCloser
}

func (r response) Code() int {
	return r.code
}

func (r response) Message() string {
	return r.message
}

func (r response) GetHeader(name string) string {
	//TODO keep headers and return the matched header
	return ""
}

func (r response) Body() io.ReadCloser {
	return r.body
}

type asyncResponse interface {
	set(r *response)
	isSticky() bool
}

func newFutureResponse(id string) *futureResponse {
	return &futureResponse{id: id, received: make(chan struct{}, 1)}
}
func newCallbackResponse(id string, cb func(*response), sticky bool) *callbackResponse {
	return &callbackResponse{id: id, cb: cb, sticky: sticky}
}

type futureResponse struct {
	id       string
	resp     *response
	received chan struct{}
}

func (m *futureResponse) Get(timeout time.Duration) (*response, error) {
	select {
	case <-m.received:
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout reached for %s", m.id)
	}
	return m.resp, nil
}
func (m *futureResponse) isSticky() bool {
	return false
}
func (m *futureResponse) set(r *response) {
	m.resp = r
	close(m.received)
}

type callbackResponse struct {
	id     string
	cb     func(*response)
	sticky bool
}

func (m *callbackResponse) set(r *response) {
	m.cb(r)
}
func (m *callbackResponse) isSticky() bool {
	return m.sticky
}

//
// a copy of the package private go-openapi/runtime/runtime/client/request is included below as there is no public method
// to instantiate this instance and no public method to access some methods which are needed to customize the protocol binding.
//
// func newRequest(method, pathPattern string, writer runtime.ClientRequestWriter) (*request, error)
//
// request.writer.WriteRequest
// request.query.Encode()
// request.payload
//
//////////////////// DO NOT EDIT THIS PART
func newRequest(method, pathPattern string, writer runtime.ClientRequestWriter) (*request, error) {
	return &request{
		pathPattern: pathPattern,
		method:      method,
		writer:      writer,
		header:      make(http.Header),
		query:       make(url.Values),
		timeout:     client.DefaultTimeout,
	}, nil
}

type request struct {
	pathPattern string
	method      string
	writer      runtime.ClientRequestWriter

	pathParams map[string]string
	header     http.Header
	query      url.Values
	formFields url.Values
	fileFields map[string][]runtime.NamedReadCloser
	payload    interface{}
	timeout    time.Duration
	buf        *bytes.Buffer
}

func (r *request) GetMethod() string {
	return r.method
}

func (r *request) GetPath() string {
	path := r.pathPattern
	for k, v := range r.pathParams {
		path = strings.Replace(path, "{"+k+"}", v, -1)
	}
	return path
}

func (r *request) GetBody() []byte {
	if r.buf == nil {
		return nil
	}
	return r.buf.Bytes()
}

func (r *request) GetBodyParam() interface{} {
	return r.GetBody()
}

func (r *request) GetHeaderParams() http.Header {
	return r.header
}

// SetHeaderParam adds a header param to the request
// when there is only 1 value provided for the varargs, it will set it.
// when there are several values provided for the varargs it will add it (no overriding)
func (r *request) SetHeaderParam(name string, values ...string) error {
	if r.header == nil {
		r.header = make(http.Header)
	}
	r.header[http.CanonicalHeaderKey(name)] = values
	return nil
}

// SetQueryParam adds a query param to the request
// when there is only 1 value provided for the varargs, it will set it.
// when there are several values provided for the varargs it will add it (no overriding)
func (r *request) SetQueryParam(name string, values ...string) error {
	if r.query == nil {
		r.query = make(url.Values)
	}
	r.query[name] = values
	return nil
}

// GetQueryParams returns a copy of all query params currently set for the request
func (r *request) GetQueryParams() url.Values {
	var result = make(url.Values)
	for key, value := range r.query {
		result[key] = append([]string{}, value...)
	}
	return result
}

// SetFormParam adds a forn param to the request
// when there is only 1 value provided for the varargs, it will set it.
// when there are several values provided for the varargs it will add it (no overriding)
func (r *request) SetFormParam(name string, values ...string) error {
	if r.formFields == nil {
		r.formFields = make(url.Values)
	}
	r.formFields[name] = values
	return nil
}

// SetPathParam adds a path param to the request
func (r *request) SetPathParam(name string, value string) error {
	if r.pathParams == nil {
		r.pathParams = make(map[string]string)
	}

	r.pathParams[name] = value
	return nil
}

func (r *request) GetFileParam() map[string][]runtime.NamedReadCloser {
	return r.fileFields
}

// SetFileParam adds a file param to the request
func (r *request) SetFileParam(name string, files ...runtime.NamedReadCloser) error {
	for _, file := range files {
		if actualFile, ok := file.(*os.File); ok {
			fi, err := os.Stat(actualFile.Name())
			if err != nil {
				return err
			}
			if fi.IsDir() {
				return fmt.Errorf("%q is a directory, only files are supported", file.Name())
			}
		}
	}

	if r.fileFields == nil {
		r.fileFields = make(map[string][]runtime.NamedReadCloser)
	}
	if r.formFields == nil {
		r.formFields = make(url.Values)
	}

	r.fileFields[name] = files
	return nil
}

// SetBodyParam sets a body parameter on the request.
// This does not yet serialze the object, this happens as late as possible.
func (r *request) SetBodyParam(payload interface{}) error {
	r.payload = payload
	return nil
}

// SetTimeout sets the timeout for a request
func (r *request) SetTimeout(timeout time.Duration) error {
	r.timeout = timeout
	return nil
}

//////////////////// DO NOT EDIT END
