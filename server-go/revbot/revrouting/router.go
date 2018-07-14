package revrouting

import (
	"github.com/strongo/bots-framework/core"
	"github.com/prizarena/reversi/server-go/revbot/revcommands"
)

var WebhooksRouter = bots.NewWebhookRouter(
	map[bots.WebhookInputType][]bots.Command{},
	func() string { return "Oops..." },
)

func init() {
	revcommands.RegisterRevCommands(WebhooksRouter)
}
