package notifications

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestWebhookConfigRequiresHTTPSAndStrongSecret(t *testing.T) {
	valid := WebhookConfig{URL: "https://hooks.example.test/cankonix", Secret: "01234567890123456789012345678901", Timeout: time.Second}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	valid.URL = "http://hooks.example.test"
	if valid.Validate() == nil {
		t.Fatal("expected non-HTTPS rejection")
	}
}
func TestWebhookSenderSignsPayloadAndSetsIdempotencyKey(t *testing.T) {
	secret := "01234567890123456789012345678901"
	payload := WebhookPayload{ID: 42, EventType: "opportunity.new", Title: "Test"}
	body, _ := json.Marshal(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		received, _ := io.ReadAll(r.Body)
		if string(received) != string(body) {
			t.Fatalf("body=%s", received)
		}
		if r.Header.Get("X-Cankonix-Signature") != "sha256="+hex.EncodeToString(mac.Sum(nil)) || r.Header.Get("Idempotency-Key") != "alert-event-42" {
			t.Fatal("missing signature or idempotency key")
		}
		return &http.Response{StatusCode: 204, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header)}, nil
	})}
	sender := HTTPWebhookSender{Config: WebhookConfig{URL: "https://hooks.example.test/cankonix", Secret: secret, Timeout: time.Second}, Client: client}
	code, err := sender.Send(payload)
	if err != nil || code != 204 {
		t.Fatalf("code=%d err=%v", code, err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
