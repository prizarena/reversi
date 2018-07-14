package revcommands

import (
	"testing"
	"github.com/strongo/bots-framework/core"
)

func TestRegisterRevCommands(t *testing.T) {
	router := bots.NewWebhookRouter(map[bots.WebhookInputType][]bots.Command{}, nil)
	RegisterRevCommands(router)
	if router.CommandsCount() == 0 {
		t.Fatal("router.CommandsCount() == 0")
	}
}
