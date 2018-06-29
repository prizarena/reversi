package revdal

import "testing"

func TestVars(t *testing.T) {
	if DB != nil {
		t.Fatalf("DB != nil: %v", DB)
	}
}
