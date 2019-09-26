package swagsock

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubmitAsyncOptionIs(t *testing.T) {
	testoptions := []SubmitAsyncOption{SubmitAsyncOptionNone, SubmitAsyncOptionSubscribe, SubmitAsyncOptionUnsubscribe("123")}
	testmodes := []SubmitAsyncMode{SubmitAsyncModeNone, SubmitAsyncModeSubscribe, SubmitAsyncModeUnsubscribe}

	for i, opt := range testoptions {
		for j, mode := range testmodes {
			if i == j {
				assert.True(t, opt.Is(mode))
			} else {
				assert.False(t, opt.Is(mode))
			}
		}
	}
}

func TestSubmitAsyncOptionParam(t *testing.T) {
	assert.Equal(t, "", SubmitAsyncOptionNone.Param())
	assert.Equal(t, "", SubmitAsyncOptionSubscribe.Param())
	assert.Equal(t, "123", SubmitAsyncOptionUnsubscribe("123").Param())
}

func TestHandshakeJson(t *testing.T) {
	hreq := &HandshakeRequest{Version: "2.0"}
	b, _ := json.Marshal(hreq)
	assert.Equal(t, `{"version":"2.0"}`, string(b))
	json.Unmarshal(b, &hreq)
	assert.Equal(t, "2.0", hreq.Version)

	hresp := &HandshakeResponse{Version: "2.0", TrackingID: "b0cbb3b4-aaee-a63a-49ae-0d5a31af9c93"}
	b, _ = json.Marshal(hresp)
	assert.Equal(t, `{"version":"2.0","trackingID":"b0cbb3b4-aaee-a63a-49ae-0d5a31af9c93"}`, string(b))
	json.Unmarshal(b, &hresp)
	assert.Equal(t, "2.0", hresp.Version)
	assert.Equal(t, "b0cbb3b4-aaee-a63a-49ae-0d5a31af9c93", hresp.TrackingID)

}
