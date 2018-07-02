package revcommands

import (
	"github.com/strongo/bots-framework/core"
	"net/url"
	"github.com/strongo/app"
	"bytes"
	"github.com/prizarena/reversi/server-go/revtrans"
	"github.com/prizarena/prizarena-public/pamodels"
	"github.com/prizarena/reversi/server-go/revgame"
)

const (
	newBoardSinglePlayer = "singleplayer"
	newBoardWithAI = "ai"
)

func newBoardCallbackData(mode revgame.Mode) string {
	switch mode {
	case revgame.SinglePlayer:
		return newBoardSinglePlayer
	case revgame.WithAI:
		return newBoardWithAI
	default:
		panic("unknown mode: " + string(mode))
	}
}

var newBoardSinleplayerCommand = bots.Command{
	Code:     newBoardSinglePlayer,
	Commands: []string{"/singleplayer"},
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		return newPlayAction(whc, "", revgame.SinglePlayer)
	},
	CallbackAction: func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		tournamentID := callbackUrl.Query().Get("t")
		return newPlayAction(whc, tournamentID, revgame.SinglePlayer)
	},
}

var newBoardWithAICommand = bots.Command{
	Code:     newBoardWithAI,
	Commands: []string{"/ai"},
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		return newPlayAction(whc, "", revgame.WithAI)
	},
	CallbackAction: func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		tournamentID := callbackUrl.Query().Get("t")
		return newPlayAction(whc, tournamentID, revgame.WithAI)
	},
}


func newPlayAction(whc bots.WebhookContext, tournamentID string, mode revgame.Mode) (m bots.MessageFromBot, err error) {
	var tournament pamodels.Tournament
	m.Text = getNewPlayText(whc, tournament)
	m.Format = bots.MessageFormatHTML
	m.Keyboard = renderReversiTgKeyboard(revgame.OthelloBoard, mode, "", whc.Locale().Code5, tournamentID)
	return
}

// var newNonTournamentBoardSizesKeyboards = map[string]*tgbotapi.InlineKeyboardMarkup{
// 	"en-US": renderReversiTgKeyboard(revgame.OthelloBoard, revgame.Black, "", "en-US", ""),
// 	"ru-RU": renderReversiTgKeyboard(revgame.OthelloBoard, revgame.Black, "", "ru-RU", ""),
// }


func getNewPlayText(t strongo.SingleLocaleTranslator, tournament pamodels.Tournament) string {
	text := new(bytes.Buffer)
	text.WriteString(t.Translate(revtrans.NewGameText))
	return text.String()
}