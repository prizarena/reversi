package revgame

import (
	"testing"
)

func TestTranscript_ToBase64(t *testing.T) {
	transcript := NewTranscript("")
	if s := transcript.ToBase64(); s != "" {
		t.Error("Expected empty string, got: " + s)
	}

	for i, step := range []struct {
		ca      CellAddress
		expects string
	}{
		{ca: "B1", expects: "B"},
		{ca: "C2", expects: "BK"},
	} {
		transcript = append(transcript, byte(CellAddressToRevAddress(step.ca).Index()))
		if s := transcript.ToBase64(); s != step.expects {
			t.Fatalf("step #%v: expected [%v], got: [%v]", i+1, step.expects, s)
		}
	}
}

func TestTranscript_String(t *testing.T) {
	transcript := NewTranscript("BK")
	if s := transcript.String(); s != "B1C2" {
		t.Errorf("Expected B1C2, got: [%v]", s)
	}
}

func TestNewTranscript_UnknownCodePanics(t *testing.T) {
	// '!' is not one of the base64url transcript codes, so NewTranscript panics.
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on an unknown transcript code")
		}
	}()
	NewTranscript("!")
}

func TestTranscript_Equal(t *testing.T) {
	a := NewTranscript("BK")
	b := NewTranscript("BK")
	if !a.Equal(b) {
		t.Error("Equal returned false for identical transcripts")
	}
	if a.Equal(NewTranscript("B")) {
		t.Error("Equal returned true for transcripts of different length")
	}
	if a.Equal(NewTranscript("KB")) {
		t.Error("Equal returned true for transcripts with different contents")
	}
	if !EmptyTranscript().Equal(NewTranscript("")) {
		t.Error("two empty transcripts should be Equal")
	}
}

func TestEmptyTranscript(t *testing.T) {
	if tr := EmptyTranscript(); len(tr) != 0 {
		t.Errorf("EmptyTranscript() len = %d, want 0", len(tr))
	}
}

func TestNewTranscriptFromHumanReadable(t *testing.T) {
	// "B1C2" -> cells (x=1,y=0) and (x=2,y=1) -> move indices 1 and 10.
	// That is exactly what NewTranscript("BK") produces, so the two must be Equal.
	got := NewTranscriptFromHumanReadable("B1C2")
	want := NewTranscript("BK")
	if !got.Equal(want) {
		t.Errorf("NewTranscriptFromHumanReadable(\"B1C2\") = %v, want %v", []byte(got), []byte(want))
	}
	if s := got.String(); s != "B1C2" {
		t.Errorf("round-trip String() = %q, want %q", s, "B1C2")
	}
	if tr := NewTranscriptFromHumanReadable(""); len(tr) != 0 {
		t.Errorf("empty input should yield empty transcript, got len %d", len(tr))
	}
}

func TestNewTranscriptFromHumanReadable_OddLengthPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on odd-length human-readable transcript")
		}
	}()
	NewTranscriptFromHumanReadable("B1C")
}

func TestTranscript_LastMove_Pop_NextMove(t *testing.T) {
	tr := NewTranscript("BK") // move indices [1, 10]

	if lm := tr.LastMove(); lm != Move(10) {
		t.Errorf("LastMove() = %d, want 10", lm)
	}

	last, rest := tr.Pop()
	if last != Move(10) {
		t.Errorf("Pop() move = %d, want 10", last)
	}
	if !rest.Equal(NewTranscript("B")) {
		t.Errorf("Pop() rest = %v, want [1]", []byte(rest))
	}

	first, tail := tr.NextMove()
	if first != Move(1) {
		t.Errorf("NextMove() move = %d, want 1", first)
	}
	if !tail.Equal(NewTranscript("K")) {
		t.Errorf("NextMove() tail = %v, want [10]", []byte(tail))
	}
}

func TestTranscript_Pop_EmptyPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when popping an empty transcript")
		}
	}()
	EmptyTranscript().Pop()
}

func TestTranscript_NextMove_EmptyPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when NextMove on an empty transcript")
		}
	}()
	EmptyTranscript().NextMove()
}
