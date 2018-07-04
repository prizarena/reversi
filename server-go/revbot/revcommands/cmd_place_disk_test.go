package revcommands

import "testing"

func TestSwitch(t *testing.T) {
	switch "+1"[0] {
	case '+', '-':
		t.Log("OK")
	default:
		t.Fatal("Not OK!")
	}
}
