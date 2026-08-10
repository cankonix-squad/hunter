package collector

import (
	"strings"
	"testing"
	"time"
)

func TestScheduleKeyIsStableWithinBucketAndChangesAcrossBuckets(t *testing.T) {
	base := time.Unix(1_800_000_000, 0).UTC()
	a := scheduleKey("source", "policy-v1", 60, base)
	b := scheduleKey("source", "policy-v1", 60, base.Add(30*time.Minute))
	c := scheduleKey("source", "policy-v1", 60, base.Add(61*time.Minute))
	if a != b || a == c || !strings.Contains(a, "policy-v1") {
		t.Fatalf("unexpected keys: %q %q %q", a, b, c)
	}
}
