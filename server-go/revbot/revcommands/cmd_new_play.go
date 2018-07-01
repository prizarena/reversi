package revcommands

import (
	"github.com/strongo/bots-framework/core"
	"net/url"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/app"
	"bytes"
	"github.com/prizarena/reversi/server-go/revtrans"
	"github.com/prizarena/prizarena-public/pamodels"
	"github.com/prizarena/reversi/server-go/revgame"
)

const newSinleplayerCommandCode = "singleplayer"

var newPlayCommand = bots.Command{
	Code:     newSinleplayerCommandCode,
	Commands: []string{"/singleplayer"},
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		return newPlayAction(whc, "", 1)
	},
	CallbackAction: func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		tournamentID := callbackUrl.Query().Get("t")
		return newPlayAction(whc, tournamentID, 1)
	},
}

func newPlayAction(whc bots.WebhookContext, tournamentID string, maxUsersLimit int) (m bots.MessageFromBot, err error) {
	var tournament pamodels.Tournament
	m.Text = getNewPlayText(whc, tournament)
	m.Format = bots.MessageFormatHTML
	m.Keyboard = renderReversiTgKeyboard(revgame.OthelloBoard, revgame.Black, "", whc.Locale().Code5, tournamentID)
	return
}

var newNonTournamentBoardSizesKeyboards = map[string]*tgbotapi.InlineKeyboardMarkup{
	"en-US": renderReversiTgKeyboard(revgame.OthelloBoard, revgame.Black, "", "en-US", ""),
	"ru-RU": renderReversiTgKeyboard(revgame.OthelloBoard, revgame.Black, "", "ru-RU", ""),
}


func getNewPlayText(t strongo.SingleLocaleTranslator, tournament pamodels.Tournament) string {
	text := new(bytes.Buffer)
	text.WriteString(t.Translate(revtrans.NewGameText))
	return text.String()
}