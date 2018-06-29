package revmodels

import (
	"testing"
	)

func TestBoard_Kind(t *testing.T) {
	b := Board{}
	if b.Kind() != BoardKind {
		t.Fatal("b.Kind() != BoardKind")
	}
}
