package decisions

import (
	"context"
	"testing"
)

func TestDecisionValidation(t *testing.T) {
	service := Service{}
	if _, err := service.Update(context.Background(), "invalid", "actor", UpdateInput{Decision: "maybe", Reason: "valid reason"}); err == nil {
		t.Fatal("accepted invalid decision")
	}
	if _, err := service.Update(context.Background(), "invalid", "actor", UpdateInput{Decision: "pursue", Reason: "x"}); err == nil {
		t.Fatal("accepted short reason")
	}
}
