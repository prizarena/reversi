package revgame

import (
	"testing"
	"github.com/prizarena/turn-based"
)

func TestTranscript_ToBase64(t *testing.T) {
	transcript := NewTranscript("")
	if s := transcript.ToBase64(); s != "" {
		t.Error("Expected empty string, got: " + s)
	}

	for i, step := range []struct {
		ca turnbased.CellAddress
		expects string
	}{
		{ca: "B1", expects: "B"},
		{ca: "C2", expects: "BK"},
	}{
		transcript = append(transcript, byte(CellAddressToRevAddress(step.ca).Index()))
		if s := transcript.ToBase64(); s != step.expects {
			t.Fatalf("step #%v: expected [%v], got: [%v]", i+1, step.expects, s)
		}
	}
}
