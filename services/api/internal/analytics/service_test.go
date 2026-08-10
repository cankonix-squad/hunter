package analytics

import (
	"encoding/json"
	"testing"
)

func TestPipelineJSONContract(t *testing.T) {
	raw, err := json.Marshal(Pipeline{Total: 1, Active: 1, Undecided: 1})
	if err != nil {
		t.Fatal(err)
	}
	const want = `{"total":1,"active":1,"shortlisted":0,"assigned":0,"pursue":0,"pass":0,"undecided":1,"overdue_actions":0,"actions_due_seven_days":0}`
	if string(raw) != want {
		t.Fatalf("unexpected analytics contract: %s", raw)
	}
}
