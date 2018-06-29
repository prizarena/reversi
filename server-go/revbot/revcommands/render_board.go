package revcommands

import (
	"github.com/strongo/app"
	"github.com/prizarena/prizarena-public/pamodels"
	"github.com/prizarena/reversi/server-go/revmodels"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-api-telegram"
	"fmt"
	"strings"
	"github.com/prizarena/turn-based"
	"bytes"
	"github.com/prizarena/reversi/server-go/revtrans"
	"github.com/strongo/emoji/go/emoji"
	"context"
	"github.com/strongo/log"
	"github.com/strongo/emoji/go"
)

func renderPairsBoardMessage(c context.Context, t strongo.SingleLocaleTranslator, tournament pamodels.Tournament, board revmodels.PairsBoard, matchedTile, userID string, players []revmodels.PairsPlayer) (m bots.MessageFromBot, err error) {
	isCompleted := board.IsCompleted(players)
	log.Debugf(c, "renderPairsBoardMessage(): isCompleted=%v", isCompleted)
	lang := t.Locale().Code5
	m.IsEdit = true
	m.Format = bots.MessageFormatHTML
	text := new(bytes.Buffer)
	fmt.Fprintf(text, `<a href="https://t.me/PairMatchingGameBot">%v</a>`, t.Translate(revtrans.GameCardTitle))
	fmt.Fprintln(text, "")
	fmt.Fprintln(text, t.Translate(revtrans.FindFast))
	if board.UsersMax == 1 && len(players) == 1 {
		switch players[0].MatchedCount {
		case 0: // Nothing
		case 1:
			fmt.Fprintf(text, t.Translate(revtrans.SinglePlayerMatchedOne))
			fmt.Fprint(text, "; ")
		default:
			fmt.Fprintf(text, t.Translate(revtrans.SinglePlayerMatchedCount, players[0].MatchedCount))
			fmt.Fprint(text, "; ")
		}
		fmt.Fprintf(text, t.Translate(revtrans.Flips, board.PairsPlayerEntity.FlipsCount))
		fmt.Fprint(text, "\n")
	} else {
		for i, p := range players {
			fmt.Fprintf(text, "%d. <b>%v</b>: %v\n", i+1, p.UserName, p.MatchedCount)
		}
	}
	if matchedTile != "" {
		if info, ok := emojis.All[matchedTile]; ok {
			fmt.Fprintf(text, "%v - %v\n", matchedTile, info.Description)
			if info.Category == "Flags" {
				fmt.Fprintf(text, "%v\n", t.Translate(revtrans.FlagOfTheDay))
			}
		}
	}
	if isCompleted {
		fmt.Fprintf(text,"\n<b>%v:</b>", t.Translate(revtrans.Board))
		text.WriteString(board.DrawBoard("", "\n"))
		fmt.Fprintf(text, "\n<b>%v</b>", t.Translate(revtrans.ChooseSizeOfNextBoard))

		var keyboard *tgbotapi.InlineKeyboardMarkup
		if board.UsersMax == 1 || tournament.ID != "" {
			keyboard = getNewPlayTgInlineKbMarkup(lang, tournament.ID, board.UsersMax)
		} else {
			keyboard = newNonTournamentBoardSizesKeyboards[lang]
		}
		switchInlinePlay := t.Locale().Code5[:2]
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{
			{Text: t.Translate(revtrans.MultiPlayer), SwitchInlineQuery: &switchInlinePlay},
		})
		m.Keyboard = keyboard
	} else {
		width, height := board.Size.WidthHeight()
		kbRows := make([][]tgbotapi.InlineKeyboardButton, height)
		for y, row := range board.Rows() {
			if len(row) != width {
				err = fmt.Errorf("len(board.Rows()[%v]) != board.SizeX: %v != %v", y, len(row), width)
				return
			}
			kbRow := make([]tgbotapi.InlineKeyboardButton, width)
			const (
				isMatched = " "
				closed = emoji.WhiteLargeSquare
			)
			for x, cell := range row {
				var text string

				for _, player := range players {
					if player.Open1.IsXY(x, y) || player.Open2.IsXY(x, y) {
						text = string(cell)
						break
					} else if strings.Contains(player.MatchedItems, string(cell)) {
						text = isMatched
						break
					}
				}
				if text == "" {
					text = closed
				}
				var boardID string
				if board.UsersMax == 1 {
					boardID = board.ID
				}
				kbRow[x] = tgbotapi.InlineKeyboardButton{Text: text, CallbackData: openCellCallbackData(turnbased.NewCellAddress(x, y), len(board.UserIDs), boardID, userID, lang)}
			}
			kbRows[y] = kbRow
		}
		m.Keyboard = tgbotapi.NewInlineKeyboardMarkup(kbRows...)
	}
	m.Text = text.String()
	return
}

