package revgaedal

import (
	"testing"
	"github.com/prizarena/reversi/server-go/revdal"
)

func TestRegisterDal(t *testing.T) {
	RegisterDal()
	if revdal.DB == nil {
		t.Fatal("revdal.DB == nil")
	}
}
