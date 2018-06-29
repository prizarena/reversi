package revgaeroot

import (
	"github.com/strongo/log"
	"github.com/strongo/bots-framework/hosts/appengine"
	"github.com/prizarena/reversi/server-go/revapp"
)

func init() {
	log.AddLogger(gaehost.GaeLogger)
	revapp.InitApp(gaehost.GaeBotHost{})
}
