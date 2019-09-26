package swagsock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/go-openapi/runtime"

	"github.com/stretchr/testify/assert"

	"github.com/elakito/swagsock/examples/greeter/models"
	"github.com/elakito/swagsock/examples/greeter/restapi/operations"
)

var (
	testTrackingID = "b0cbb3b4-aaee-a63a-49ae-0d5a31af9c93"

	testMessageStrings = []string{
		`{"id": "123", "method": "GET",
			"path": "/ws/consumers/my_group/instances/my_instance/topics/my_topic",
			"accept": "application/vnd.kafka.json.v1+json"}`,
		`{"id": "125", "method": "POST", "path": "/topics/test",
			"type": "application/vnd.kafka.binary.v1+json"}{"records":[{"value":"S2Fma2E="}]}`,
		`{"id": "127", "code": 200, "type": "application/vnd.kafka.binary.v1+json"}[{"partition":0,"leader":0,"replicas":[{"broker":0,"leader":true,"in_sync":true}]}]`,
		`{"id": "131", "method": "GET", "headers":{"x-shopping-ref":"4126"}, "path": "/shopping/bag"}`,
	}
	testMessageMaps = []map[string]interface{}{
		map[string]interface{}{
			"id":     "123",
			"method": "GET",
			"path":   "/ws/consumers/my_group/instances/my_instance/topics/my_topic",
			"accept": "application/vnd.kafka.json.v1+json",
		},
		map[string]interface{}{
			"id":     "125",
			"method": "POST",
			"path":   "/topics/test",
			"type":   "application/vnd.kafka.binary.v1+json",
		},
		map[string]interface{}{
			"id":   "127",
			"code": 200,
			"type": "application/vnd.kafka.binary.v1+json",
		},
		map[string]interface{}{
			"id":      "131",
			"method":  "GET",
			"path":    "/shopping/bag",
			"headers": map[string]interface{}{"x-shopping-ref": "4126"},
		},
	}
	testMessageBodies = []string{
		``,
		`{"records":[{"value":"S2Fma2E="}]}`,
		`[{"partition":0,"leader":0,"replicas":[{"broker":0,"leader":true,"in_sync":true}]}]`,
		``,
	}
	testMessageIsRequest = []bool{
		true,
		true,
		false,
		true,
	}

	testContinuedMessageStrings = []string{
		`{"id": "129", "method": "POST",
			"path": "/topics/bigtopic",
			"type": "application/vnd.kafka.binary.v1+json", "continue": true}{"recor`,
		`{"id": "129", "method": "POST",
			"path": "/topics/bigtopic",
			"type": "application/vnd.kafka.binary.v1+json", "continue": true}ds":[{"va`,
		`{"id": "129", "method": "POST",
			"path": "/topics/bigtopic",
			"type": "application/vnd.kafka.binary.v1+json", "continue": true}lue":"S2Fma`,
		`{"id": "129", "method": "POST",
			"path": "/topics/bigtopic",
			"type": "application/vnd.kafka.binary.v1+json"}2E="}]}`,
	}

	testContinuedMessageMap = map[string]interface{}{
		"id":     "129",
		"method": "POST",
		"path":   "/topics/bigtopic",
		"type":   "application/vnd.kafka.binary.v1+json",
	}

	testContinuedMessageBody = `{"records":[{"value":"S2Fma2E="}]}`

	testInvalidMessageStrings = []string{
		`{"id": "123", "method": "GET", invalid}`,
	}

	testHandshakeReqString     = `{"version":"2.0"}`
	testHandshakeReqBadString  = `{"version":"1.0"}`
	testHandshakeRespString    = `{"version":"2.0","trackingID":"b0cbb3b4-aaee-a63a-49ae-0d5a31af9c93"}`
	testHandshakeRespBadString = `{"version":"2.0","error":"version_mismatch"}`

	testHandshakeReqInvalidString = `{"version": invalid}`
)

func TestIsWebsocketUpgradeRequested(t *testing.T) {
	req, _ := http.NewRequest("GET", "/v1/demo", nil)
	assert.False(t, IsWebsocketUpgradeRequested(req))

	req.Header.Add("Connection", "Upgrade")
	req.Header.Add("Upgrade", "websocket")
	req.Header.Add("Sec-WebSocket-Version", "13")
	req.Header.Add("Sec-WebSocket-Key", "aZpgbJ4FWyDiKaQiZx3yvw==")
	assert.True(t, IsWebsocketUpgradeRequested(req))
}

func TestGetBaseURI(t *testing.T) {
	req, _ := http.NewRequest("GET", "/v1/service", nil)
	assert.Equal(t, "/v1/service", getBaseURI(req))

	req, _ = http.NewRequest("GET", "/v1/service/", nil)
	assert.Equal(t, "/v1/service", getBaseURI(req))
}

func TestDefaultCodecEncode(t *testing.T) {
	codec := NewDefaultCodec()
	for i, tmsgmap := range testMessageMaps {
		data, err := codec.EncodeSwaggerSocketMessage(tmsgmap, []byte(testMessageBodies[i]))
		assert.Nil(t, err, "Failed to decode '%s': %v", tmsgmap, err)

		//use the decoder to verify the encoded data (as the decoder's behavior has been validated)
		headers, body, err := codec.DecodeSwaggerSocketMessage(data)
		assert.Nil(t, err, "The encoded data %s not decodable: %v", data, err)

		assert.Equal(t, testMessageMaps[i], headers, "[%d] Expected headers %v but was %v", i, testMessageMaps[i], headers)
		assert.Equal(t, testMessageBodies[i], string(body), "[%d] Expected body %v but was %s", i, testMessageBodies[i], body)
	}
}

func TestDefaultCodecDecode(t *testing.T) {
	codec := NewDefaultCodec()
	for i, tmsgstr := range testMessageStrings {
		headers, body, err := codec.DecodeSwaggerSocketMessage([]byte(tmsgstr))
		assert.Nil(t, err, "Failed to decode '%s': %v", tmsgstr, err)
		assert.Equal(t, testMessageMaps[i], headers, "[%d] Expected headers %v but was %v", i, testMessageMaps[i], headers)
		assert.Equal(t, testMessageBodies[i], string(body), "[%d] Expected body %v but was %s", i, testMessageBodies[i], body)
	}
}

func TestDefaultCodecDecodeInvalid(t *testing.T) {
	codec := NewDefaultCodec()
	for _, tmsgstr := range testInvalidMessageStrings {
		_, _, err := codec.DecodeSwaggerSocketMessage([]byte(tmsgstr))
		assert.NotNil(t, err, "Didn't Fail to decode '%s': %v", tmsgstr, err)
	}
}

func TestCreateDefaultProtocolHandler(t *testing.T) {
	conf := NewConfig()
	ph, ok := CreateProtocolHandler(conf).(*protocolHandler)
	assert.True(t, ok)
	assert.NotNil(t, ph.codec)
	assert.NotNil(t, ph.GetCodec())
	assert.NotNil(t, ph.continued)
	assert.NotNil(t, ph.mediator)
}

func TestBuildHeaders(t *testing.T) {
	resp := &responseWriter{id: "513", code: 200, headers: make(http.Header)}
	assert.Equal(t, map[string]interface{}{
		"id":   "513",
		"code": 200,
	}, resp.buildHeaders())

	resp.headers.Add("Content-Type", "application/json")
	assert.Equal(t, map[string]interface{}{
		"id":   "513",
		"code": 200,
		"type": "application/json",
	}, resp.buildHeaders())
}

func TestNewHTTPRequest(t *testing.T) {
	codec := NewDefaultCodec()
	for i, tmsgstr := range testMessageStrings {
		if !testMessageIsRequest[i] {
			continue
		}
		headers, body, _ := codec.DecodeSwaggerSocketMessage([]byte(tmsgstr))
		rid := getStringHeader(headers, "id")
		req := newHTTPRequest("/test", "default", rid, headers, bytes.NewReader(body))
		mmap := testMessageMaps[i]
		assert.Equal(t, mmap["method"].(string), req.Method)
		assert.True(t, strings.HasPrefix(req.RequestURI, "/test"))
		assert.Equal(t, mmap["path"].(string), req.RequestURI[5:])
		assert.Equal(t, buildRequestKey("default", rid), GetRequestKey(req))
	}
}

func TestHandshake(t *testing.T) {
	conf := NewConfig()
	ph, ok := CreateProtocolHandler(conf).(*protocolHandler)
	assert.True(t, ok)
	writer := &testConnectionWriter{}

	// good handshake
	err := ph.handshake([]byte(testHandshakeReqString), testTrackingID, writer)
	assert.NoError(t, err)
	assert.Equal(t, testHandshakeRespString, writer.data.String())

	// bad handshake
	writer.data.Reset()
	err = ph.handshake([]byte(testHandshakeReqBadString), testTrackingID, writer)
	assert.Error(t, err)
	assert.Equal(t, testHandshakeRespBadString, writer.data.String())

	// invalid handshake
	writer.data.Reset()
	err = ph.handshake([]byte(testHandshakeReqInvalidString), testTrackingID, writer)
	assert.Error(t, err)
}

func TestServeNormal(t *testing.T) {
	conf := NewConfig()
	ph, ok := CreateProtocolHandler(conf).(*protocolHandler)
	assert.True(t, ok)

	// for the non-continued requests
	hh := &testHTTPHandler{}
	count := 0
	for i, tmsgstr := range testMessageStrings {
		if !testMessageIsRequest[i] {
			continue
		}
		ph.serve(hh, "/service", testTrackingID, 1, []byte(tmsgstr), nil, nil)
		count++
		mmap := testMessageMaps[i]

		assert.Equal(t, mmap["method"].(string), hh.req.Method)
		assert.True(t, strings.HasPrefix(hh.req.RequestURI, "/service"))
		assert.Equal(t, mmap["path"].(string), hh.req.RequestURI[8:])
		assert.Equal(t, buildRequestKey(testTrackingID, mmap["id"].(string)), GetRequestKey(hh.req))

		hhresp, ok := hh.resp.(*responseWriter)
		assert.True(t, ok)
		assert.Equal(t, mmap["id"].(string), hhresp.id)
	}
	assert.Equal(t, count, hh.served)
}

func TestServeContinued(t *testing.T) {
	conf := NewConfig()
	ph, ok := CreateProtocolHandler(conf).(*protocolHandler)
	assert.True(t, ok)

	// for the continued requests
	done := make(chan struct{})
	hh := &testHTTPHandler{done: done}
	for i, tmsgstr := range testContinuedMessageStrings {
		ph.serve(hh, "/service", testTrackingID, 1, []byte(tmsgstr), nil, nil)
		// the handler will be only invoked once after the first segment is served
		assert.Equal(t, 1, hh.served)
		if i == 0 {
			assert.Equal(t, testContinuedMessageMap["method"].(string), hh.req.Method)
			assert.True(t, strings.HasPrefix(hh.req.RequestURI, "/service"))
			assert.Equal(t, testContinuedMessageMap["path"].(string), hh.req.RequestURI[8:])
			assert.Equal(t, buildRequestKey(testTrackingID, testContinuedMessageMap["id"].(string)), GetRequestKey(hh.req))
		}
	}
	select {
	case <-done:
		assert.Equal(t, testContinuedMessageBody, hh.body.String())
	case <-time.After(2 * time.Second):
		assert.Fail(t, "timeout unexpected")
	}
}

func TestServe(t *testing.T) {
	conf := NewConfig()
	conf.Heartbeat = 5
	ph := CreateProtocolHandler(conf)
	defer ph.Destroy()
	hh := &testEchoHandler{mediator: conf.ResponseMediator}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ph.Serve(hh, w, r)
	}))
	defer ts.Close()

	// Connect to the server
	ws, _, err := websocket.DefaultDialer.Dial("ws"+ts.URL[4:]+"?x-tracking-id="+testTrackingID, nil)
	defer ws.Close()
	assert.NoError(t, err)

	go func() {
		ws.WriteMessage(websocket.TextMessage, []byte(`{"version":"2.0"}`))
		ws.WriteMessage(websocket.TextMessage, []byte(`{"id":"1","method":"GET","path":"/v1/ping"}`))
		ws.WriteMessage(websocket.TextMessage, []byte(`{"id":"2","method":"POST","path":"/v1/echo","type":"text/plain"}hola`))
		ws.WriteMessage(websocket.TextMessage, []byte(`{"id":"3","method":"GET","path":"/v1/peek"}`))
	}()

	done := make(chan struct{})
	step := 0
	go func() {
		defer func() {
			done <- struct{}{}
		}()
		_, message, err := ws.ReadMessage()
		assert.NoError(t, err)

		var hresp HandshakeResponse
		err = json.Unmarshal(message, &hresp)
		assert.NoError(t, err)
		assert.Equal(t, "2.0", hresp.Version)
		assert.Equal(t, testTrackingID, hresp.TrackingID)
		step++

		_, message, err = ws.ReadMessage()
		assert.Equal(t, `{"code":200,"id":"1","type":"application/json"}{"pong":0}`, string(message))
		step++

		_, message, err = ws.ReadMessage()
		assert.Equal(t, `{"code":200,"id":"2","type":"application/json"}{"echo":"hola"}`, string(message))
		step++

		_, message, err = ws.ReadMessage()
		assert.Equal(t, `{"code":404,"id":"3"}`, string(message))
		step++
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		assert.Fail(t, "not all messages received at step %d", step)
	}
}

func TestAddDeleteConnection(t *testing.T) {
	conf := NewConfig()
	ph, ok := CreateProtocolHandler(conf).(*protocolHandler)
	assert.True(t, ok)
	assert.Equal(t, 0, len(ph.connections))
	c := &websocket.Conn{}
	ph.addConnetion(c, "dummy")
	assert.Equal(t, 1, len(ph.connections))
	id := ph.deleteConnection(c)
	assert.Equal(t, 0, len(ph.connections))
	assert.Equal(t, "dummy", id)
}

func TestDefaultResponseMediator(t *testing.T) {
	mediator := NewDefaultResponseMediator().(*defaultResponseMediator)
	r1 := &testOK{}
	w1 := &testWriter{}
	rr1 := mediator.Subscribe("foo#0", "naranja", r1, nil, nil)
	rr1.WriteResponse(w1, nil)

	r2 := &testOK{}
	w2 := &testWriter{}
	rr2 := mediator.Subscribe("bar#2", "manzana", r2, []byte("hi"), []byte("adios"))
	rr2.WriteResponse(w2, nil)

	r3 := &testOK{}
	w3 := &testWriter{}
	rr3 := mediator.Subscribe("bar#5", "orange", r3, nil, nil)
	rr3.WriteResponse(w3, nil)

	subscribed := mediator.Subscribed()
	sort.Strings(subscribed)
	assert.Equal(t, []string{"manzana", "naranja", "orange"}, subscribed)

	mediator.Write("naranja", []byte("hola"))
	mediator.Write("orange", []byte("hallo"))
	mediator.Write("*", []byte("#swagger"))

	assert.Equal(t, "hihola#swagger", w1.buf.String())
	assert.Equal(t, "hi#swagger", w2.buf.String())
	assert.Equal(t, "hallo#swagger", w3.buf.String())
	assert.Equal(t, 3, len(mediator.responders))

	mediator.Unsubscribe("bar#10", "2")
	mediator.Write("manzana", []byte("bien"))
	mediator.Write("*", []byte("hello"))

	subscribed = mediator.Subscribed()
	sort.Strings(subscribed)
	assert.Equal(t, []string{"naranja", "orange"}, subscribed)

	assert.Equal(t, "hihola#swaggeradioshello", w1.buf.String())
	assert.Equal(t, "hi#swaggeradios", w2.buf.String())
	assert.Equal(t, "hallo#swaggeradioshello", w3.buf.String())
	assert.Equal(t, 2, len(mediator.responders))

	mediator.UnsubscribeAll("foo")
	mediator.UnsubscribeAll("bar")
	assert.Equal(t, 0, len(mediator.responders))
	subscribed = mediator.Subscribed()
	assert.Empty(t, subscribed)
}

func TestDefaultResponseMediatorTopics(t *testing.T) {
	mediator := NewDefaultResponseMediator().(*defaultResponseMediator)
	r1 := &testOK{}
	w1 := &testWriter{}
	rr1 := mediator.SubscribeTopic("foo#0", "general", "naranja", r1, nil, nil)
	rr1.WriteResponse(w1, nil)

	r2 := &testOK{}
	w2 := &testWriter{}
	rr2 := mediator.SubscribeTopic("bar#2", "general", "manzana", r2, []byte("hi"), []byte("adios"))
	rr2.WriteResponse(w2, nil)

	r3 := &testOK{}
	w3 := &testWriter{}
	rr3 := mediator.SubscribeTopic("bar#5", "general", "orange", r3, nil, nil)
	rr3.WriteResponse(w3, nil)

	r4 := &testOK{}
	w4 := &testWriter{}
	rr4 := mediator.SubscribeTopic("bar#7", "private", "manzana", r4, []byte("hi"), []byte("adios"))
	rr4.WriteResponse(w4, nil)

	r5 := &testOK{}
	w5 := &testWriter{}
	rr5 := mediator.SubscribeTopic("bar#8", "private", "orange", r5, nil, nil)
	rr5.WriteResponse(w5, nil)

	subscribedtopics := mediator.SubscribedTopics()
	sort.Strings(subscribedtopics)
	assert.Equal(t, []string{"general", "private"}, subscribedtopics)
	subscribed := mediator.SubscribedTopic("general")
	sort.Strings(subscribed)
	assert.Equal(t, []string{"manzana", "naranja", "orange"}, subscribed)

	mediator.WriteTopic("general", []byte("hola"))
	mediator.WriteTopic("general", []byte("hallo"))
	mediator.WriteTopic("*", []byte("#swagger"))

	assert.Equal(t, "hihiholahallo#swagger", w1.buf.String())
	assert.Equal(t, "hihiholahallo#swagger", w2.buf.String())
	assert.Equal(t, "hiholahallo#swagger", w3.buf.String())
	assert.Equal(t, "hi#swagger", w4.buf.String())
	assert.Equal(t, "#swagger", w5.buf.String())
	assert.Equal(t, 5, len(mediator.responders))

	mediator.Unsubscribe("bar#10", "2")
	mediator.WriteTopic("private", []byte("bien"))
	mediator.WriteTopic("*", []byte("hello"))

	subscribedtopics = mediator.SubscribedTopics()
	sort.Strings(subscribedtopics)
	assert.Equal(t, []string{"general", "private"}, subscribedtopics)
	subscribed = mediator.SubscribedTopic("general")
	sort.Strings(subscribed)
	assert.Equal(t, []string{"naranja", "orange"}, subscribed)

	assert.Equal(t, "hihiholahallo#swaggeradioshello", w1.buf.String())
	assert.Equal(t, "hihiholahallo#swaggeradios", w2.buf.String())
	assert.Equal(t, "hiholahallo#swaggeradioshello", w3.buf.String())
	assert.Equal(t, "hi#swaggeradiosbienhello", w4.buf.String())
	assert.Equal(t, "#swaggeradiosbienhello", w5.buf.String())
	assert.Equal(t, 4, len(mediator.responders))

	mediator.UnsubscribeAll("foo")
	mediator.UnsubscribeAll("bar")
	assert.Equal(t, 0, len(mediator.responders))
	subscribed = mediator.SubscribedTopic("general")
	assert.Empty(t, subscribed)
}

type testOK struct {
}

func (o *testOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {
}

type testWriter struct {
	buf bytes.Buffer
}

func (w *testWriter) Header() http.Header {
	return nil
}
func (w *testWriter) Write(b []byte) (int, error) {
	return w.buf.Write(b)
}
func (w *testWriter) WriteHeader(statusCode int) {
}

type testProducer struct {
}

func (p *testProducer) Produce(w io.Writer, b interface{}) error {
	switch v := b.(type) {
	case string:
		w.Write([]byte(v))
	case []byte:
		w.Write(v)
	default:
		bv, _ := json.Marshal(v)
		w.Write(bv)
	}
	return nil
}

type testHTTPHandler struct {
	req    *http.Request
	resp   http.ResponseWriter
	body   bytes.Buffer
	served int
	done   chan struct{}
}

func (h *testHTTPHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	h.req = req
	h.resp = resp
	h.body.Reset()
	h.served++
	if h.req.Body != nil {
		b := make([]byte, 64)
		for {
			if n, err := h.req.Body.Read(b); err == io.EOF {
				if h.done != nil {
					h.done <- struct{}{}
				}
				break
			} else {
				h.body.Write(b[0:n])
			}
		}
	}
}

type testEchoHandler struct {
	mediator ResponseMediator
	count    int
}

func (h *testEchoHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	// this is a simplified test service with ping, echo, subscribe, unsubscribe and using echo for subscription
	switch req.RequestURI {
	case "/v1/ping":
		resp.Header().Set("Content-Type", "application/json")
		resp.WriteHeader(http.StatusOK)
		resp.Write([]byte(fmt.Sprintf(`{"pong":%d}`, h.count)))
		h.count++
	case "/v1/echo":
		resp.Header().Set("Content-Type", "application/json")
		resp.WriteHeader(http.StatusOK)
		b, _ := ioutil.ReadAll(req.Body)
		pb := []byte(fmt.Sprintf(`{"from":"system","name":"everyone","text":"%s"}`, b))
		h.mediator.Write("*", pb)
		resp.Write([]byte(fmt.Sprintf(`{"echo":"%s"}`, b)))
	default:
		if strings.HasPrefix(req.RequestURI, "/v1/subscribe") {
			user := req.RequestURI[len("/v1/subscribe")+1:]
			payload := &models.GreetingReply{}
			responder := h.mediator.Subscribe(GetRequestKey(req), user, operations.NewSubscribeOK().WithPayload(payload), nil, nil)
			resp.Header().Set("Content-Type", "application/json")
			resp.WriteHeader(http.StatusOK)
			responder.WriteResponse(resp, &testProducer{})
		} else if strings.HasPrefix(req.RequestURI, "/v1/unsubscribe") {
			sid := req.RequestURI[len("/v1/unsubscribe")+1:]
			h.mediator.Unsubscribe(GetRequestKey(req), sid)
			resp.Header().Set("Content-Type", "application/json")
			resp.WriteHeader(http.StatusOK)
			resp.Write([]byte(`{"greeted":[]}`))
		} else {
			resp.WriteHeader(http.StatusNotFound)
		}
	}
}

type testConnectionWriter struct {
	data bytes.Buffer
}

func (w *testConnectionWriter) WriteMessage(messageType int, data []byte) error {
	w.data.Write(data)
	return nil
}
func (w *testConnectionWriter) WriteJSON(v interface{}) error {
	var b []byte
	var err error
	if b, err = json.Marshal(v); err == nil {
		w.data.Write(b)
		return nil
	}
	return err
}
