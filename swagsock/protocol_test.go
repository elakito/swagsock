package swagsock

import (
	"bytes"
	"net/http"
	"strings"
	"testing"

	"github.com/go-openapi/runtime"

	"github.com/stretchr/testify/assert"
)

var (
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

func TestBuildHeaders(t *testing.T) {
	resp := &response{id: "513", code: 200, headers: make(http.Header)}
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

		id, req := newHTTPRequest("/test", "default", []byte(tmsgstr), codec)
		mmap := testMessageMaps[i]
		assert.Equal(t, mmap["method"].(string), req.Method)
		assert.True(t, strings.HasPrefix(req.RequestURI, "/test"))
		assert.Equal(t, mmap["path"].(string), req.RequestURI[5:])
		assert.Equal(t, buildRequestKey("default", id), GetRequestKey(req))
	}
}

func TestDefaultResponseMediator(t *testing.T) {
	mediator := NewDefaultResponseMediator().(*defaultResponseMediator)
	r1 := &testOK{}
	w1 := &testWriter{}
	rr1 := mediator.Subscribe("foo#0", "naranja", r1, nil)
	rr1.WriteResponse(w1, nil)

	r2 := &testOK{}
	w2 := &testWriter{}
	rr2 := mediator.Subscribe("bar#2", "manzana", r2, []byte("adios"))
	rr2.WriteResponse(w2, nil)

	r3 := &testOK{}
	w3 := &testWriter{}
	rr3 := mediator.Subscribe("bar#5", "orange", r3, nil)
	rr3.WriteResponse(w3, nil)

	mediator.Write("naranja", []byte("hola"))
	mediator.Write("orange", []byte("hallo"))
	mediator.Write("*", []byte("#swagger"))

	assert.Equal(t, "hola#swagger", w1.buf.String())
	assert.Equal(t, "#swagger", w2.buf.String())
	assert.Equal(t, "hallo#swagger", w3.buf.String())
	assert.Equal(t, 3, len(mediator.responders))

	mediator.Unsubscribe("bar#10", "2")
	mediator.Write("manzana", []byte("bien"))
	mediator.Write("*", []byte("hello"))

	assert.Equal(t, "hola#swaggeradioshello", w1.buf.String())
	assert.Equal(t, "#swagger", w2.buf.String())
	assert.Equal(t, "hallo#swaggeradioshello", w3.buf.String())
	assert.Equal(t, 2, len(mediator.responders))

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
