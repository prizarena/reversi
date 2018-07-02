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
	newBoardSinglePlayerCommandCode = "singleplayer"
	newBoardMultiPlayerCommandCode  = "multiplayer"
	newBoardWithAICommandCode       = "ai"
)

func newBoardCallbackData(mode revgame.Mode) string {
	switch mode {
	case revgame.SinglePlayer:
		return newBoardSinglePlayerCommandCode
	case revgame.WithAI:
		return newBoardWithAICommandCode
	default:
		panic("unknown mode: " + string(mode))
	}
}

var newBoardSinleplayerCommand = bots.Command{
	Code:     newBoardSinglePlayerCommandCode,
	Commands: []string{"/singleplayer"},
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		return newPlayAction(whc, "", revgame.SinglePlayer, revgame.Black)
	},
	CallbackAction: func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		tournamentID := callbackUrl.Query().Get("t")
		return newPlayAction(whc, tournamentID, revgame.SinglePlayer, revgame.Black)
	},
}

var newBoardWithAICommand = bots.Command{
	Code:     newBoardWithAICommandCode,
	Commands: []string{"/ai"},
	Action: func(whc bots.WebhookContext) (m bots.MessageFromBot, err error) {
		return newPlayAction(whc, "", revgame.WithAI, revgame.Black)
	},
	CallbackAction: func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		tournamentID := callbackUrl.Query().Get("t")
		player := getPlayerFromString(callbackUrl.Query().Get("p"))
		if player != revgame.Black && player != revgame.White {
			player = revgame.Black
		}
		return newPlayAction(whc, tournamentID, revgame.WithAI, player)
	},
}

func newPlayAction(whc bots.WebhookContext, tournamentID string, mode revgame.Mode, player revgame.Disk) (m bots.MessageFromBot, err error) {
	var tournament pamodels.Tournament
	m.Text = getNewPlayText(whc, tournament)
	m.Format = bots.MessageFormatHTML
	m.Keyboard = renderReversiTgKeyboard(revgame.OthelloBoard, mode, player, false, "", whc.Locale().Code5, tournamentID)
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
