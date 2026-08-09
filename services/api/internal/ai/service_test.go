package ai

import (
	"context"
	"errors"
	"testing"
	"time"
)

type scriptedProvider struct {
	responses []GenerateResponse
	errors    []error
	calls     int
}

func (*scriptedProvider) Name() string { return "test-provider" }

func (p *scriptedProvider) Generate(_ context.Context, _ GenerateRequest) (GenerateResponse, error) {
	index := p.calls
	p.calls++
	if index < len(p.errors) && p.errors[index] != nil {
		return GenerateResponse{}, p.errors[index]
	}
	return p.responses[index], nil
}

func TestClassifyValidatesPersistsAndCaches(t *testing.T) {
	provider := &scriptedProvider{responses: []GenerateResponse{{Model: "small-v1", CostMicro: 100,
		Content: []byte(`{"schema_version":"1.0","task":"classify_it","result":{"disposition":"relevant","category_code":"data-ai","explanation":"Data integration work."},"confidence_bp":8700,"evidence":[{"field":"title","snippet":"Integrasi Data"}],"provider":"ignored","model":"ignored","prompt_version":"ignored","input_fingerprint":"ignored","created_at":"2025-01-01T00:00:00Z","fallback":false}`)}}}
	store, budget := NewMemoryStore(), &MemoryBudget{Limit: 100_000}
	service := NewService(provider, DefaultRegistry(), store, budget)
	service.now = func() time.Time { return time.Date(2026, 8, 9, 0, 0, 0, 0, time.UTC) }
	input := ClassifyInput{Title: "Pengadaan Integrasi Data", ProcurementType: "Jasa Lainnya"}

	first, err := service.Classify(context.Background(), input)
	if err != nil || first.Provider != "test-provider" || first.Model != "small-v1" || first.InputFingerprint == "" {
		t.Fatalf("unexpected first result: %#v, err=%v", first, err)
	}
	second, err := service.Classify(context.Background(), input)
	if err != nil || second.InputFingerprint != first.InputFingerprint || provider.calls != 1 || budget.Responses != 1 {
		t.Fatalf("cache miss: calls=%d responses=%d err=%v", provider.calls, budget.Responses, err)
	}
}

func TestClassifyRepairsInvalidOutputOnce(t *testing.T) {
	provider := &scriptedProvider{responses: []GenerateResponse{
		{Model: "small-v1", Content: []byte(`{"schema_version":"1.0","task":"classify_it","result":{"disposition":"relevant","category_code":"data-ai","explanation":"x"},"confidence_bp":9000,"evidence":[]}`)},
		{Model: "small-v1", Content: []byte(`{"schema_version":"1.0","task":"classify_it","result":{"disposition":"needs_review","explanation":"Insufficient evidence."},"confidence_bp":4000,"evidence":[]}`)},
	}}
	service := NewService(provider, DefaultRegistry(), NewMemoryStore(), &MemoryBudget{})
	result, err := service.Classify(context.Background(), ClassifyInput{Title: "Transformasi"})
	if err != nil || result.Result.Disposition != "needs_review" || provider.calls != 2 {
		t.Fatalf("result=%#v calls=%d err=%v", result, provider.calls, err)
	}
}

func TestClassifyRejectsInventedEvidence(t *testing.T) {
	provider := &scriptedProvider{responses: []GenerateResponse{{Model: "small-v1", Content: []byte(`{"schema_version":"1.0","task":"classify_it","result":{"disposition":"relevant","category_code":"data-ai","explanation":"x"},"confidence_bp":9000,"evidence":[{"field":"title","snippet":"invented"}]}`)}, {Model: "small-v1", Content: []byte(`{}`)}}}
	service := NewService(provider, DefaultRegistry(), NewMemoryStore(), &MemoryBudget{})
	result, err := service.Classify(context.Background(), ClassifyInput{Title: "Integrasi Data"})
	if err == nil || !result.Fallback || result.Result.Disposition != "needs_review" || provider.calls != 2 {
		t.Fatalf("result=%#v calls=%d err=%v", result, provider.calls, err)
	}
}

func TestClassifyBudgetAndProviderFallback(t *testing.T) {
	input := ClassifyInput{Title: "Ambiguous package"}
	budgetService := NewService(&scriptedProvider{}, DefaultRegistry(), NewMemoryStore(), &MemoryBudget{Limit: 1})
	result, err := budgetService.Classify(context.Background(), input)
	if !errors.Is(err, ErrBudgetExceeded) || !result.Fallback {
		t.Fatalf("budget result=%#v err=%v", result, err)
	}

	provider := &scriptedProvider{errors: []error{errors.New("offline")}}
	providerService := NewService(provider, DefaultRegistry(), NewMemoryStore(), &MemoryBudget{})
	result, err = providerService.Classify(context.Background(), input)
	if !errors.Is(err, ErrProviderUnavailable) || !result.Fallback || provider.calls != 1 {
		t.Fatalf("provider result=%#v calls=%d err=%v", result, provider.calls, err)
	}
}

func TestFingerprintIsStableAndSensitive(t *testing.T) {
	a, _ := Fingerprint(ClassifyInput{Title: "A", ProcurementType: "B"})
	b, _ := Fingerprint(ClassifyInput{Title: "A", ProcurementType: "B"})
	c, _ := Fingerprint(ClassifyInput{Title: "C", ProcurementType: "B"})
	if a != b || a == c || len(a) != 71 {
		t.Fatalf("unexpected fingerprints: %q %q %q", a, b, c)
	}
}
