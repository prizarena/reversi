package revgame

import (
	"testing"
	"github.com/pkg/errors"
)

func TestDisks_bit(t *testing.T) {
	// TODO: add unit test
}

func TestPlayerDisks_Add(t *testing.T) {
	var pd Disks
	var err error

	testCases := []struct {
		target      address
		expectedErr error
		// expetectState PlayerDisks
	}{
		{address{0, 0}, nil},
		{address{0, 0}, ErrAlreadyOccupied},
	}

	for i, testCase := range testCases {
		if pd, err = pd.add(testCase.target); errors.Cause(err) != testCase.expectedErr {
			t.Fatalf("case #%v: expected err=%v, got=%v", i+1, testCase.expectedErr, err)
		}
	}
}

func TestPlayerDisks_Add_All(t *testing.T) {
	for x := int8(0); x < 8; x++ {
		for y := int8(0); y < 8; y++ {
			a := address{x, y}
			var disks Disks
			var err error
			disks, err = disks.add(a)
			if err != nil {
				t.Errorf("address=%v => error: %v", a, err)
				continue
			}
			if disks == 0 {
				t.Errorf("address=%v => disks: 0", a)
			}
			expects := 1 << uint(y*8+x)
			if disks != Disks(expects) {
				t.Errorf("address=%v, expected %v => got: %v", a, expects, disks)
			}
		}
	}
}

func TestPlayerDisks_Remove(t *testing.T) {
	var pd Disks
	var err error

	testCases := []struct {
		do          string
		target      address
		expectedErr error
		// expetectState PlayerDisks
	}{
		{"add", address{0, 0}, nil},
		{"add", address{0, 0}, ErrAlreadyOccupied},
		{"remove", address{0, 0}, nil},
		{"add", address{0, 0}, nil},
	}

	for i, testCase := range testCases {
		err = nil
		switch testCase.do {
		case "add":
			pd, err = pd.add(testCase.target)
		case "remove":
			pd = pd.remove(testCase.target)
		}
		if errors.Cause(err) != testCase.expectedErr {
			t.Fatalf("case #%v: expected err=%v, got=%v", i+1, testCase.expectedErr, err)
		}
	}
}
