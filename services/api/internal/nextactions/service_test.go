package nextactions

import (
	"context"
	"testing"
)

func TestUpdateRejectsPartialInput(t *testing.T) {
	action := "Follow up"
	if _, err := (&Service{}).Update(context.Background(), "invalid", "actor", UpdateInput{Action: &action}); err == nil {
		t.Fatal("accepted partial action")
	}
}
