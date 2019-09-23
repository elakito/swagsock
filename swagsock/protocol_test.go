package swagsock

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/go-openapi/runtime"

	"github.com/stretchr/testify/assert"
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
	}
	testMessageBodies = []string{
		``,
		`{"records":[{"value":"S2Fma2E="}]}`,
		`[{"partition":0,"leader":0,"replicas":[{"broker":0,"leader":true,"in_sync":true}]}]`,
	}
	testMessageIsRequest = []bool{
		true,
		true,
		false,
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

	testHandshakeReqString     = `{"version":"2.0"}`
	testHandshakeReqBadString  = `{"version":"1.0"}`
	testHandshakeRespString    = `{"version":"2.0","trackingID":"b0cbb3b4-aaee-a63a-49ae-0d5a31af9c93"}`
	testHandshakeRespBadString = `{"version":"2.0","error":"version_mismatch"}`
)

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

func TestCreateDefaultProtocolHandler(t *testing.T) {
	conf := NewConfig()
	ph, ok := CreateProtocolHandler(conf).(*protocolHandler)
	assert.True(t, ok)
	assert.NotNil(t, ph.codec)
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
	ph.handshake([]byte(testHandshakeReqString), testTrackingID, writer)
	assert.Equal(t, testHandshakeRespString, writer.data.String())

	// bad handshake
	writer.data.Reset()
	ph.handshake([]byte(testHandshakeReqBadString), testTrackingID, writer)
	assert.Equal(t, testHandshakeRespBadString, writer.data.String())
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

	assert.Equal(t, "hihola#swaggeradioshello", w1.buf.String())
	assert.Equal(t, "hi#swaggeradios", w2.buf.String())
	assert.Equal(t, "hallo#swaggeradioshello", w3.buf.String())
	assert.Equal(t, 2, len(mediator.responders))

	mediator.UnsubscribeAll("foo")
	mediator.UnsubscribeAll("bar")
	assert.Equal(t, 0, len(mediator.responders))
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

	assert.Equal(t, "hihiholahallo#swaggeradioshello", w1.buf.String())
	assert.Equal(t, "hihiholahallo#swaggeradios", w2.buf.String())
	assert.Equal(t, "hiholahallo#swaggeradioshello", w3.buf.String())
	assert.Equal(t, "hi#swaggeradiosbienhello", w4.buf.String())
	assert.Equal(t, "#swaggeradiosbienhello", w5.buf.String())
	assert.Equal(t, 4, len(mediator.responders))

	mediator.UnsubscribeAll("foo")
	mediator.UnsubscribeAll("bar")
	assert.Equal(t, 0, len(mediator.responders))
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
