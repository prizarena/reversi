package revgaedal

import (
	"github.com/strongo/db/gaedb"
	"github.com/prizarena/reversi/server-go/revdal"
)

func RegisterDal() {
	revdal.DB = gaedb.NewDatabase()
}
