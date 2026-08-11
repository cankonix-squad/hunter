package notifications

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type WebhookPayload struct {
	ID              int64     `json:"id"`
	EventType       string    `json:"event_type"`
	OpportunityID   string    `json:"opportunity_id"`
	PackageIdentity string    `json:"package_identity"`
	Title           string    `json:"title"`
	ChangedFields   []string  `json:"changed_fields"`
	CanonicalURL    string    `json:"canonical_url"`
	OccurredAt      time.Time `json:"occurred_at"`
}
type WebhookSender interface {
	Send(WebhookPayload) (int, error)
}
type WebhookConfig struct {
	URL, Secret string
	Timeout     time.Duration
}

func (c WebhookConfig) Validate() error {
	parsed, err := url.Parse(c.URL)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil {
		return errors.New("webhook URL must be an absolute HTTPS URL without credentials")
	}
	if len(c.Secret) < 32 {
		return errors.New("webhook secret must contain at least 32 characters")
	}
	if c.Timeout <= 0 {
		return errors.New("webhook timeout must be positive")
	}
	return nil
}

type HTTPWebhookSender struct {
	Config WebhookConfig
	Client *http.Client
}

func (s HTTPWebhookSender) Send(payload WebhookPayload) (int, error) {
	if err := s.Config.Validate(); err != nil {
		return 0, err
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}
	request, err := http.NewRequest(http.MethodPost, s.Config.URL, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	signature := hmac.New(sha256.New, []byte(s.Config.Secret))
	_, _ = signature.Write(body)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "CANKONIX-Notifier/1.0")
	request.Header.Set("X-Cankonix-Signature", "sha256="+hex.EncodeToString(signature.Sum(nil)))
	request.Header.Set("Idempotency-Key", "alert-event-"+strconv.FormatInt(payload.ID, 10))
	client := s.Client
	if client == nil {
		client = &http.Client{Timeout: s.Config.Timeout, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	}
	response, err := client.Do(request)
	if err != nil {
		return 0, fmt.Errorf("send webhook: %w", err)
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return response.StatusCode, fmt.Errorf("webhook returned HTTP %d", response.StatusCode)
	}
	return response.StatusCode, nil
}

type DryRunWebhookSender struct{}

func (DryRunWebhookSender) Send(WebhookPayload) (int, error) { return http.StatusNoContent, nil }
