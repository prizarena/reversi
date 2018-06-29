package revapp

import "testing"

func TestInitApp(t *testing.T) {
	t.Run("nil_params", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("panic expected")
			}
		}()
		InitApp(nil)
	})
}
