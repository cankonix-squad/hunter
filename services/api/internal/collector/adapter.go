package collector

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

type Source struct {
	ID             string
	Name           string
	BaseURL        string
	AdapterName    string
	AdapterVersion string
	Policy         json.RawMessage
}

type Candidate struct {
	PackageIdentity string         `json:"package_identity"`
	CanonicalURL    string         `json:"canonical_url"`
	SourceTimestamp *time.Time     `json:"source_timestamp,omitempty"`
	Metadata        map[string]any `json:"metadata"`
}

func (c Candidate) Fingerprint() (string, error) {
	payload, err := json.Marshal(c.Metadata)
	if err != nil {
		return "", fmt.Errorf("encode candidate fingerprint: %w", err)
	}
	hash := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(hash[:]), nil
}

type Adapter interface {
	Name() string
	Version() string
	Discover(context.Context, Source) ([]Candidate, error)
}

type Registry struct{ adapters map[string]Adapter }

func NewRegistry(adapters ...Adapter) *Registry {
	r := &Registry{adapters: make(map[string]Adapter, len(adapters))}
	for _, adapter := range adapters {
		r.adapters[key(adapter.Name(), adapter.Version())] = adapter
	}
	return r
}

func (r *Registry) Resolve(name, version string) (Adapter, error) {
	adapter, ok := r.adapters[key(name, version)]
	if !ok {
		return nil, fmt.Errorf("unsupported adapter %s@%s", name, version)
	}
	return adapter, nil
}

func key(name, version string) string { return name + "@" + version }
