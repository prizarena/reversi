package revtgbot

import (
	"testing"
	"github.com/strongo/app"
	"github.com/strongo/bots-framework/core"
)

func TestBots(t *testing.T) {
	Bots(nil, strongo.EnvProduction, bots.WebhooksRouter{})
}
