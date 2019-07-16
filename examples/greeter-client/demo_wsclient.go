package main

import (
	"bytes"
	"encoding"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-openapi/runtime"
	runtimeclient "github.com/go-openapi/runtime/client"
	"github.com/gorilla/websocket"

	"github.com/go-openapi/strfmt"
	"github.com/satori/go.uuid"

	apiclient "github.com/elakito/swagsock/examples/greeter-client/client"
	apioperations "github.com/elakito/swagsock/examples/greeter-client/client/operations"
	"github.com/elakito/swagsock/swagsock"
)

//
// This is not the final picture of wsclient but demonstrates how the greeter requests can be sent over websocket and
// gives a hint to how it can be integrated to the current client that only supports http.
//
func main() {
	target_path := flag.String("target_path", "ws://localhost:8091/samples/greeter", "target_path")
	flag.Parse()

	u, err := url.Parse(*target_path)
	if err != nil {
		log.Fatalf("failed to parse the target url %s: %v", *target_path, err)
	}
	//transport := httpclient.New(u.Host, u.Path, []string{u.Scheme})
	transport := newWSTransport(u.String())
	defer transport.Close()

	// create the API client, with the transport
	client := newWSClient(transport, strfmt.Default)
	fmt.Printf("demo_client invoking some operations at target_path %s ...\n", *target_path)

	perform(client)
}

/////////
// currently, this demo client works with the unmodified go-openapi/runtime et al.
// the below code are are modified code from go-openapi/runtime that are directly used
// in this demo at the momement.
// at this moment, i'm still trying to figure out how the websocket based approach (or potentially other  protocols) could be
// optimally integrated into the current go-openapi/runtime.
//
// TODO consolidate the wsclient runtime
func newWSClient(transport *wstransport, formats strfmt.Registry) *apiclient.GreeterDemo {
	cli := new(apiclient.GreeterDemo)
	cli.Operations = apioperations.New(transport, formats)
	return cli
}

func newWSTransport(url string) *wstransport {
	t := &wstransport{url: url, codec: swagsock.NewDefaultCodec(), pending: make(map[string]*futureResponse)}

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

	pending map[string]*futureResponse
	lock    sync.RWMutex
}

func (t *wstransport) connect() error {
	c, _, err := websocket.DefaultDialer.Dial(t.url, nil)
	if err != nil {
		return err
	}
	t.conn = c
	go func() {
		for {
			_, message, err := t.conn.ReadMessage()
			if err == nil {
				headers, body, err := t.codec.DecodeSwaggerSocketMessage(message)
				if err == nil {
					reqid := headers["id"].(string)

					// only handle if there is a pending request
					if fresp := t.removeFutureResponse(reqid); fresp != nil {
						delete(t.pending, reqid)
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
				log.Println("Error read:", err)
				return
			}
		}
	}()

	return nil
}

func (t *wstransport) Submit(operation *runtime.ClientOperation) (interface{}, error) {
	req := &request{
		pathPattern: operation.PathPattern,
		method:      operation.Method,
		writer:      operation.Params,
		header:      make(http.Header),
		query:       make(url.Values),
		timeout:     runtimeclient.DefaultTimeout,
	}

	// build the data
	if err := req.writer.WriteToRequest(req, strfmt.Default); err != nil {
		return nil, err
	}

	reqid := uuid.NewV4().String()
	fresp := newFutureReponse(reqid)
	t.putFutureResponse(reqid, fresp)

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

	rawheaders, _ := json.Marshal(headers)
	rawmessage := fmt.Sprintf(`%s%s`, rawheaders, body)
	t.writeMessage([]byte(rawmessage))

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

func (t *wstransport) putFutureResponse(reqid string, fresp *futureResponse) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.pending[reqid] = fresp
}

func (t *wstransport) removeFutureResponse(reqid string) *futureResponse {
	t.lock.Lock()
	defer t.lock.Unlock()
	if fresp, found := t.pending[reqid]; found {
		delete(t.pending, reqid)
		return fresp
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

func newFutureReponse(id string) *futureResponse {
	return &futureResponse{id: id, received: make(chan struct{}, 1)}
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

func (m *futureResponse) set(r *response) {
	m.resp = r
	close(m.received)
}

// a copy of the package private runtime.request (used now until figuring out how everything can be integrated)
//////////////////// DO NOT EDIT THIS PART
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
