package revbot

import "testing"

func TestInitBot(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("panic expected")
		}
	}()
	InitBot(nil, nil, nil)
}
