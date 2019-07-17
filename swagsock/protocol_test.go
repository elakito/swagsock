package swagsock

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	TEST_MSG_STRS = []string{
		`{"id": "123", "method": "GET",
			"path": "/ws/consumers/my_group/instances/my_instance/topics/my_topic",
			"accept": "application/vnd.kafka.json.v1+json"}`,
		`{"id": "125", "method": "POST", "path": "/topics/test",
			"type": "application/vnd.kafka.binary.v1+json"}{"records":[{"value":"S2Fma2E="}]}`,
		`{"id": "127", "code": 200, "type": "application/vnd.kafka.binary.v1+json"}[{"partition":0,"leader":0,"replicas":[{"broker":0,"leader":true,"in_sync":true}]}]`,
	}
	TEST_MSG_MAPS = []map[string]interface{}{
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
	TEST_MSG_BODIES = []string{
		``,
		`{"records":[{"value":"S2Fma2E="}]}`,
		`[{"partition":0,"leader":0,"replicas":[{"broker":0,"leader":true,"in_sync":true}]}]`,
	}
	TEST_MSG_ISREQUERST = []bool{
		true,
		true,
		false,
	}
)

func TestDefaultCodecEncode(t *testing.T) {
	codec := NewDefaultCodec()
	for i, tmsgmap := range TEST_MSG_MAPS {
		data, err := codec.EncodeSwaggerSocketMessage(tmsgmap, []byte(TEST_MSG_BODIES[i]))
		assert.Nil(t, err, "Failed to decode '%s': %v", tmsgmap, err)

		//use the decoder to verify the encoded data (as the decoder's behavior has been validated)
		headers, body, err := codec.DecodeSwaggerSocketMessage(data)
		assert.Nil(t, err, "The encoded data %s not decodable: %v", data, err)

		assert.Equal(t, TEST_MSG_MAPS[i], headers, "[%d] Expected headers %v but was %v", i, TEST_MSG_MAPS[i], headers)
		assert.Equal(t, TEST_MSG_BODIES[i], string(body), "[%d] Expected body %v but was %s", i, TEST_MSG_BODIES[i], body)
	}
}

func TestDefaultCodecDecode(t *testing.T) {
	codec := NewDefaultCodec()
	for i, tmsgstr := range TEST_MSG_STRS {
		headers, body, err := codec.DecodeSwaggerSocketMessage([]byte(tmsgstr))
		assert.Nil(t, err, "Failed to decode '%s': %v", tmsgstr, err)
		assert.Equal(t, TEST_MSG_MAPS[i], headers, "[%d] Expected headers %v but was %v", i, TEST_MSG_MAPS[i], headers)
		assert.Equal(t, TEST_MSG_BODIES[i], string(body), "[%d] Expected body %v but was %s", i, TEST_MSG_BODIES[i], body)
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

func TestNewHttpRequest(t *testing.T) {
	codec := NewDefaultCodec()
	for i, tmsgstr := range TEST_MSG_STRS {
		if !TEST_MSG_ISREQUERST[i] {
			continue
		}

		id, req := newHttpRequest("/test", []byte(tmsgstr), codec)
		mmap := TEST_MSG_MAPS[i]
		assert.Equal(t, mmap["method"].(string), req.Method)
		assert.True(t, strings.HasPrefix(req.RequestURI, "/test"))
		assert.Equal(t, mmap["path"].(string), req.RequestURI[5:])
		assert.Equal(t, id, GetRequestID(req))
	}
}
