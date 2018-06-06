package swagsock

import (
	"net/http"
	"reflect"
	"testing"
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
			"id":       "125",
			"method":   "POST",
			"path":     "/topics/test",
			"type":     "application/vnd.kafka.binary.v1+json",
		},
		map[string]interface{}{
			"id":       "127",
			"code":     200,
			"type":     "application/vnd.kafka.binary.v1+json",
		},
	}
	TEST_MSG_BODIES = []string{
		``,
		`{"records":[{"value":"S2Fma2E="}]}`,
		`[{"partition":0,"leader":0,"replicas":[{"broker":0,"leader":true,"in_sync":true}]}]`,
	}
)

func TestDefaultCodecEncode(t *testing.T) {
	codec := NewDefaultCodec()
	for i, tmsgmap := range TEST_MSG_MAPS {
		data, err := codec.EncodeSwaggerSocketMessage(tmsgmap, []byte(TEST_MSG_BODIES[i]))
		if err != nil {
			t.Fatalf("Failed to decode '%s': %v", tmsgmap, err)
		}

		//use the decoder to verify the encoded data (as the decoder's behavior has been validated)
		headers, body, err := codec.DecodeSwaggerSocketMessage(data)
		if err != nil {
			t.Fatalf("The encoded data %s not decodable: %v", data, err)
		}
		if !reflect.DeepEqual(TEST_MSG_MAPS[i], headers) {
			t.Fatalf("[%d] Expected headers %v but was %v", i, TEST_MSG_MAPS[i], headers)
		}
		if TEST_MSG_BODIES[i] != string(body) {
			t.Fatalf("[%d] Expected body %v but was %s", i, TEST_MSG_BODIES[i], body)
		}
	}
}

func TestDefaultCodecDecode(t *testing.T) {
	codec := NewDefaultCodec()
	for i, tmsgstr := range TEST_MSG_STRS {
		headers, body, err := codec.DecodeSwaggerSocketMessage([]byte(tmsgstr))
		if err != nil {
			t.Fatalf("Failed to decode '%s': %v", tmsgstr, err)
		}
		if !reflect.DeepEqual(TEST_MSG_MAPS[i], headers) {
			t.Fatalf("[%d] Expected headers %v but was %v", i, TEST_MSG_MAPS[i], headers)
		}
		if TEST_MSG_BODIES[i] != string(body) {
			t.Fatalf("[%d] Expected body %v but was %s", i, TEST_MSG_BODIES[i], body)
		}
	}
}


func TestBuildHeaders(t *testing.T) {
	resp := &response{id: "513", code: 200, headers: make(http.Header)}
	verifier := func(expected map[string]interface{}, actual map[string]interface{}) {
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("headers expected %v but got %v", expected, actual)
		}
	}

	verifier(map[string]interface{}{
		"id":   "513",
		"code": 200,
	}, resp.buildHeaders())

	resp.headers.Add("Content-Type", "application/json")
	verifier(map[string]interface{}{
		"id":   "513",
		"code": 200,
		"type": "application/json",
	}, resp.buildHeaders())
}
