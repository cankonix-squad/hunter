package notifications

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDeliveryJSONDoesNotExposeLeaseOrSecret(t *testing.T) {
	raw, err := json.Marshal(Delivery{ID: 1, Channel: "webhook", Destination: "https://hooks.example.test", Status: "dead_letter", SafeError: "HTTP 500"})
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	for _, forbidden := range []string{"lease_owner", "lease_expires_at", "secret"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("delivery JSON exposed %s", forbidden)
		}
	}
}
