package revmodels

import (
	"testing"
	"github.com/prizarena/turn-based"
)

func TestPlayerDisks_Add(t *testing.T) {
	var pd Disks
	var err error

	testCases := []struct {
		target turnbased.CellAddress
		expectedErr error
		// expetectState PlayerDisks
	}{
		{"A1", nil},
		{"A1", ErrAlreadyOccupied},
	}

	for i, testCase := range testCases {
		if pd, err = pd.Add(testCase.target); err != testCase.expectedErr {
			t.Fatalf("case #%v: expected err=%v, got=%v", i+1, testCase.expectedErr, err)
		}
	}
}

func TestPlayerDisks_Remove(t *testing.T) {
	var pd Disks
	var err error

	testCases := []struct {
		do string
		target turnbased.CellAddress
		expectedErr error
		// expetectState PlayerDisks
	}{
		{"add", "A1", nil},
		{"add", "A1", ErrAlreadyOccupied},
		{"remove", "A1", nil},
		{"add", "A1", nil},
	}

	for i, testCase := range testCases {
		switch testCase.do {
		case "add":
			pd, err = pd.Add(testCase.target)
		case "remove":
			pd, err = pd.Remove(testCase.target)
		}
		if err != testCase.expectedErr {
			t.Fatalf("case #%v: expected err=%v, got=%v", i+1, testCase.expectedErr, err)
		}
	}
}
