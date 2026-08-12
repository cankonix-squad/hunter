package collector

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestRetryBackoffSpansTransientOutageAndIsBounded(t *testing.T) {
	wants := []time.Duration{30 * time.Second, time.Minute, 2 * time.Minute, 4 * time.Minute, 8 * time.Minute, 16 * time.Minute, 30 * time.Minute}
	for index, want := range wants {
		if got := retryBackoff(index + 1); got != want {
			t.Fatalf("attempt %d backoff = %s, want %s", index+1, got, want)
		}
	}
	if got := retryBackoff(20); got != 30*time.Minute {
		t.Fatalf("bounded backoff = %s", got)
	}
}

func TestPermanentSourceErrorSurvivesWrapping(t *testing.T) {
	err := fmt.Errorf("discover: %w", permanentSourceError{errors.New("source access stopped")})
	var permanent interface{ Permanent() bool }
	if !errors.As(err, &permanent) || !permanent.Permanent() {
		t.Fatal("wrapped source policy error must remain permanent")
	}
}
