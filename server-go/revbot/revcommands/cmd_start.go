package revcommands

import (
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-api-telegram"
	"net/url"
		"github.com/prizarena/reversi/server-go/revtrans"
				"github.com/prizarena/prizarena-public/pabot"
	"github.com/prizarena/reversi/server-go/revsecrets"
	"github.com/prizarena/reversi/server-go/revgame"
	"github.com/prizarena/prizarena-public/patrans"
	"github.com/DebtsTracker/translations/emoji"
)

const startCommandCommandCode = "start"

var startCommand = bots.Command{
	Code:     startCommandCommandCode,
	Commands: []string{"/" + startCommandCommandCode},
	Action:   startAction,
	CallbackAction: startCallbackAction,
}

func startCallbackAction(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
	m, err = startAction(whc)
	m.IsEdit = true
	return
}

func startAction(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
	if m, err = pabot.OnStartIfTournamentLink(whc, revsecrets.PrizarenaGameID, revsecrets.PrizarenaToken); m.Text != "" || err != nil {
		return
	}
	m.Text = whc.Translate(revtrans.OnStartWelcome)
	m.Format = bots.MessageFormatHTML
	m.DisableWebPagePreview = true
	switchInlinePlay := whc.Locale().Code5[:2]
	m.Keyboard = tgbotapi.NewInlineKeyboardMarkup(
		pabot.GetLangButtons(startCommandCommandCode, whc.Locale().Code5),
		[]tgbotapi.InlineKeyboardButton{
			{Text: whc.Translate(patrans.SinglePlayer), CallbackData: newBoardCallbackData(revgame.SinglePlayer)},
		},
		[]tgbotapi.InlineKeyboardButton{
			{Text: whc.Translate(patrans.MultiPlayer), SwitchInlineQuery: &switchInlinePlay},
		},
		[]tgbotapi.InlineKeyboardButton{
			{Text: whc.Translate(patrans.TournamentsButton), CallbackData: "tournaments"},
		},
		[]tgbotapi.InlineKeyboardButton{
			{Text: emoji.STAR_ICON + emoji.STAR_ICON + emoji.STAR_ICON + emoji.STAR_ICON + emoji.STAR_ICON, URL: "http://storebot.me/bot/reversigamebot"},
		},
	)
	return
}
