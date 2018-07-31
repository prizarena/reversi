package revcommands

import (
	"github.com/strongo/bots-framework/core"
	"github.com/prizarena/prizarena-public/pabot"
				"github.com/prizarena/reversi/server-go/revsecrets"
)

func RegisterRevCommands(router bots.WebhooksRouter) {
	router.RegisterCommands([]bots.Command{
		startCommand,
		inlineQueryCommand,
		placeDiskCommand,
		newBoardCommand,
		newBoardSinglePlayerCommand,
		newChatMembersCommand,
		// newBoardWithAICommand,
	})

	pabot.InitPrizarenaInGameBot(revsecrets.PrizarenaGameID, revsecrets.PrizarenaToken, router)
}
