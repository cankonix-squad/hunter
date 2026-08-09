package collector

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed fixtures/spse_v1_candidates.json
var spseV1Fixture []byte

type FixtureAdapter struct{}

func (FixtureAdapter) Name() string    { return "spse-fixture" }
func (FixtureAdapter) Version() string { return "v1" }

func (FixtureAdapter) Discover(ctx context.Context, _ Source) ([]Candidate, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var candidates []Candidate
	if err := json.Unmarshal(spseV1Fixture, &candidates); err != nil {
		return nil, fmt.Errorf("parse embedded SPSE fixture: %w", err)
	}
	return candidates, nil
}
