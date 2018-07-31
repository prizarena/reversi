package revcommands

import (
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/log"
)

var newChatMembersCommand = bots.Command{
	Code: "new-chat-members",
	InputTypes: []bots.WebhookInputType{bots.WebhookInputNewChatMembers},
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		log.Debugf(whc.Context(), "newChatMembersCommand()")
		return
	},
}
