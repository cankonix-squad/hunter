package ai

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

const (
	SchemaVersion  = "1.0"
	TaskClassifyIT = "classify_it"
)

var (
	ErrProviderUnavailable = errors.New("AI provider unavailable")
	ErrBudgetExceeded      = errors.New("AI budget exceeded")
)

type ClassifyInput struct {
	Title           string `json:"title"`
	ProcurementType string `json:"procurement_type"`
}

type Evidence struct {
	Field   string `json:"field"`
	Snippet string `json:"snippet"`
}

type Classification struct {
	Disposition string `json:"disposition"`
	Category    string `json:"category_code,omitempty"`
	Explanation string `json:"explanation"`
}

type Envelope struct {
	SchemaVersion    string         `json:"schema_version"`
	Task             string         `json:"task"`
	Result           Classification `json:"result"`
	ConfidenceBP     int            `json:"confidence_bp"`
	Evidence         []Evidence     `json:"evidence"`
	Provider         string         `json:"provider"`
	Model            string         `json:"model"`
	PromptVersion    string         `json:"prompt_version"`
	InputFingerprint string         `json:"input_fingerprint"`
	CreatedAt        time.Time      `json:"created_at"`
	Fallback         bool           `json:"fallback"`
}

type Prompt struct {
	Version      string
	System       string
	MaxCostMicro int64
}

type Registry interface {
	Get(task string) (Prompt, bool)
}

type StaticRegistry map[string]Prompt

func (r StaticRegistry) Get(task string) (Prompt, bool) {
	prompt, ok := r[task]
	return prompt, ok
}

func DefaultRegistry() StaticRegistry {
	return StaticRegistry{TaskClassifyIT: {
		Version:      "classify-it/v1",
		System:       "Classify only from supplied fields. Return strict JSON. Never infer missing qualifications.",
		MaxCostMicro: 25_000,
	}}
}

type GenerateRequest struct {
	Task       string
	Prompt     Prompt
	Input      ClassifyInput
	Repair     bool
	PriorError string
}

type GenerateResponse struct {
	Content   []byte
	Model     string
	InputTok  int
	OutputTok int
	CostMicro int64
}

type Provider interface {
	Name() string
	Generate(context.Context, GenerateRequest) (GenerateResponse, error)
}

type CacheKey struct {
	Task, PromptVersion, InputFingerprint, Provider, Model string
}

type Store interface {
	Get(context.Context, CacheKey) (Envelope, bool, error)
	Put(context.Context, CacheKey, Envelope, GenerateResponse) error
}

type Budget interface {
	Allow(context.Context, string, int64) (bool, error)
	Record(context.Context, string, GenerateResponse) error
}

type Service struct {
	provider Provider
	registry Registry
	store    Store
	budget   Budget
	now      func() time.Time
}

func NewService(provider Provider, registry Registry, store Store, budget Budget) *Service {
	return &Service{provider: provider, registry: registry, store: store, budget: budget, now: time.Now}
}

func (s *Service) Classify(ctx context.Context, input ClassifyInput) (Envelope, error) {
	prompt, ok := s.registry.Get(TaskClassifyIT)
	if !ok {
		return Envelope{}, errors.New("classify prompt is not registered")
	}
	fingerprint, err := Fingerprint(input)
	if err != nil {
		return Envelope{}, err
	}
	providerName := s.provider.Name()
	prefix := CacheKey{Task: TaskClassifyIT, PromptVersion: prompt.Version, InputFingerprint: fingerprint, Provider: providerName}
	if cached, found, err := s.store.Get(ctx, prefix); err != nil {
		return Envelope{}, fmt.Errorf("load AI cache: %w", err)
	} else if found {
		return cached, nil
	}
	allowed, err := s.budget.Allow(ctx, providerName, prompt.MaxCostMicro)
	if err != nil {
		return Envelope{}, fmt.Errorf("check AI budget: %w", err)
	}
	if !allowed {
		return fallback(input, fingerprint, providerName, prompt.Version, s.now(), "budget_exceeded"), ErrBudgetExceeded
	}

	var response GenerateResponse
	var envelope Envelope
	var validationErr error
	for attempt := 0; attempt < 2; attempt++ {
		response, err = s.provider.Generate(ctx, GenerateRequest{Task: TaskClassifyIT, Prompt: prompt, Input: input,
			Repair: attempt == 1, PriorError: errorText(validationErr)})
		if err != nil {
			return fallback(input, fingerprint, providerName, prompt.Version, s.now(), "provider_unavailable"), fmt.Errorf("%w: %v", ErrProviderUnavailable, err)
		}
		envelope, validationErr = decodeAndValidate(response.Content, input)
		if validationErr == nil {
			break
		}
	}
	if validationErr != nil {
		return fallback(input, fingerprint, providerName, prompt.Version, s.now(), "invalid_provider_output"), validationErr
	}
	envelope.Provider, envelope.Model = providerName, response.Model
	envelope.PromptVersion, envelope.InputFingerprint = prompt.Version, fingerprint
	envelope.CreatedAt = s.now().UTC()
	key := prefix
	key.Model = response.Model
	if err := s.store.Put(ctx, key, envelope, response); err != nil {
		return Envelope{}, fmt.Errorf("persist AI result: %w", err)
	}
	if err := s.budget.Record(ctx, providerName, response); err != nil {
		return Envelope{}, fmt.Errorf("record AI usage: %w", err)
	}
	return envelope, nil
}

func Fingerprint(input ClassifyInput) (string, error) {
	payload, err := json.Marshal(input)
	if err != nil {
		return "", fmt.Errorf("encode AI input: %w", err)
	}
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func decodeAndValidate(payload []byte, input ClassifyInput) (Envelope, error) {
	var value Envelope
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return Envelope{}, fmt.Errorf("AI output is not valid schema JSON: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return Envelope{}, errors.New("AI output contains trailing data")
	}
	if value.SchemaVersion != SchemaVersion || value.Task != TaskClassifyIT {
		return Envelope{}, errors.New("AI output schema_version or task is unsupported")
	}
	if value.ConfidenceBP < 0 || value.ConfidenceBP > 10000 {
		return Envelope{}, errors.New("AI confidence_bp must be between 0 and 10000")
	}
	if value.Result.Disposition != "relevant" && value.Result.Disposition != "excluded" && value.Result.Disposition != "needs_review" {
		return Envelope{}, errors.New("AI disposition is invalid")
	}
	if value.Result.Disposition == "relevant" && value.Result.Category == "" {
		return Envelope{}, errors.New("AI relevant result requires category_code")
	}
	if strings.TrimSpace(value.Result.Explanation) == "" {
		return Envelope{}, errors.New("AI explanation is required")
	}
	fields := map[string]string{"title": input.Title, "procurement_type": input.ProcurementType}
	for _, evidence := range value.Evidence {
		source, ok := fields[evidence.Field]
		if !ok || strings.TrimSpace(evidence.Snippet) == "" || !strings.Contains(strings.ToLower(source), strings.ToLower(evidence.Snippet)) {
			return Envelope{}, errors.New("AI evidence must quote a supplied source field")
		}
	}
	if value.Result.Disposition == "relevant" && len(value.Evidence) == 0 {
		return Envelope{}, errors.New("AI relevant result requires evidence")
	}
	value.Provider, value.Model, value.PromptVersion, value.InputFingerprint = "", "", "", ""
	value.CreatedAt, value.Fallback = time.Time{}, false
	return value, nil
}

func fallback(_ ClassifyInput, fingerprint, provider, promptVersion string, now time.Time, reason string) Envelope {
	return Envelope{SchemaVersion: SchemaVersion, Task: TaskClassifyIT,
		Result:   Classification{Disposition: "needs_review", Explanation: reason},
		Provider: provider, Model: "unavailable", PromptVersion: promptVersion,
		InputFingerprint: fingerprint, CreatedAt: now.UTC(), Fallback: true, Evidence: []Evidence{}}
}

func errorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
