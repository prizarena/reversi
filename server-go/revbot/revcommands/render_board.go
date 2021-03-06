package revcommands

import (
	"github.com/strongo/app"
	"github.com/prizarena/prizarena-public/pamodels"
	"github.com/prizarena/reversi/server-go/revmodels"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-api-telegram"
	"context"
	"github.com/prizarena/reversi/server-go/revgame"
	"github.com/strongo/emoji/go/emoji"
	"github.com/prizarena/turn-based"
	"bytes"
	"fmt"
	"github.com/prizarena/reversi/server-go/revtrans"
	"github.com/prizarena/prizarena-public/patrans"
)

func renderReversiBoardMessage(c context.Context, t strongo.SingleLocaleTranslator, tournament pamodels.Tournament, board revmodels.RevBoard, matchedTile, userID string) (m bots.MessageFromBot, err error) {
	// isCompleted := board.IsCompleted(players)
	// log.Debugf(c, "renderPairsBoardMessage(): isCompleted=%v", isCompleted)
	// lang := t.Locale().Code5
	// m.IsEdit = true
	// m.Format = bots.MessageFormatHTML
	// text := new(bytes.Buffer)
	// fmt.Fprintf(text, `<a href="https://t.me/PairMatchingGameBot">%v</a>`, t.Translate(revtrans.GameCardTitle))
	// fmt.Fprintln(text, "")
	// fmt.Fprintln(text, t.Translate(revtrans.FindFast))
	// if board.UsersMax == 1 && len(players) == 1 {
	// 	switch players[0].MatchedCount {
	// 	case 0: // Nothing
	// 	case 1:
	// 		fmt.Fprintf(text, t.Translate(revtrans.SinglePlayerMatchedOne))
	// 		fmt.Fprint(text, "; ")
	// 	default:
	// 		fmt.Fprintf(text, t.Translate(revtrans.SinglePlayerMatchedCount, players[0].MatchedCount))
	// 		fmt.Fprint(text, "; ")
	// 	}
	// 	fmt.Fprintf(text, t.Translate(revtrans.Flips, board.PairsPlayerEntity.FlipsCount))
	// 	fmt.Fprint(text, "\n")
	// } else {
	// 	for i, p := range players {
	// 		fmt.Fprintf(text, "%d. <b>%v</b>: %v\n", i+1, p.UserName, p.MatchedCount)
	// 	}
	// }
	// if matchedTile != "" {
	// 	if info, ok := emojis.All[matchedTile]; ok {
	// 		fmt.Fprintf(text, "%v - %v\n", matchedTile, info.Description)
	// 		if info.Category == "Flags" {
	// 			fmt.Fprintf(text, "%v\n", t.Translate(revtrans.FlagOfTheDay))
	// 		}
	// 	}
	// }
	// if isCompleted {
	// 	fmt.Fprintf(text,"\n<b>%v:</b>", t.Translate(revtrans.Board))
	// 	text.WriteString(board.DrawBoard("", "\n"))
	// 	fmt.Fprintf(text, "\n<b>%v</b>", t.Translate(revtrans.ChooseSizeOfNextBoard))
	//
	// 	var keyboard *tgbotapi.InlineKeyboardMarkup
	// 	if board.UsersMax == 1 || tournament.ID != "" {
	// 		keyboard = getNewPlayTgInlineKbMarkup(lang, tournament.ID, board.UsersMax)
	// 	} else {
	// 		keyboard = newNonTournamentBoardSizesKeyboards[lang]
	// 	}
	// 	switchInlinePlay := t.Locale().Code5[:2]
	// 	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{
	// 		{Text: t.Translate(revtrans.MultiPlayer), SwitchInlineQuery: &switchInlinePlay},
	// 	})
	// 	m.Keyboard = keyboard
	// } else {
	// 	width, height := board.Size.WidthHeight()
	// 	kbRows := make([][]tgbotapi.InlineKeyboardButton, height)
	// 	for y, row := range board.Rows() {
	// 		if len(row) != width {
	// 			err = fmt.Errorf("len(board.Rows()[%v]) != board.SizeX: %v != %v", y, len(row), width)
	// 			return
	// 		}
	// 		kbRow := make([]tgbotapi.InlineKeyboardButton, width)
	// 		const (
	// 			isMatched = " "
	// 			closed = emoji.WhiteLargeSquare
	// 		)
	// 		for x, cell := range row {
	// 			var text string
	//
	// 			for _, player := range players {
	// 				if player.Open1.IsXY(x, y) || player.Open2.IsXY(x, y) {
	// 					text = string(cell)
	// 					break
	// 				} else if strings.Contains(player.MatchedItems, string(cell)) {
	// 					text = isMatched
	// 					break
	// 				}
	// 			}
	// 			if text == "" {
	// 				text = closed
	// 			}
	// 			var boardID string
	// 			if board.UsersMax == 1 {
	// 				boardID = board.ID
	// 			}
	// 			kbRow[x] = tgbotapi.InlineKeyboardButton{Text: text, CallbackData: openCellCallbackData(turnbased.NewCellAddress(x, y), len(board.UserIDs), boardID, userID, lang)}
	// 		}
	// 		kbRows[y] = kbRow
	// 	}
	// 	m.Keyboard = tgbotapi.NewInlineKeyboardMarkup(kbRows...)
	// }
	// m.Text = text.String()
	return
}

func renderReversiBoardText(t strongo.SingleLocaleTranslator, p placeDiskPayload, isCompleted bool, possibleMove string) string {
	text := new(bytes.Buffer)
	text.WriteString(fmt.Sprintf("<b>%v</b>\n", t.Translate(revtrans.GameCardTitle)))
	blacksScore, whitesScore := p.currentBoard.Scores()
	nextMove := p.currentBoard.NextPlayer()

	writeScore := func(player revgame.Disk, disk string, score int) {
		switch p.mode {
		case revgame.SinglePlayer:
			var name string
			if player == revgame.Black {
				name = t.Translate(revtrans.Me)
			} else {
				name = emoji.RobotFace
			}
			fmt.Fprintf(text, "%v (%v): <b>%v</b>", disk, name, score)
		case revgame.MultiPlayer:
			var userName string
			switch player {
			case revgame.Black:
				userName = p.userNames[0]
			case revgame.White:
				if len(p.userNames) > 1 {
					userName = p.userNames[1]
				} else {
					userName = "<i>not joined yet</i>"
				}
			default:
				panic("unknown player: " + string(player))
			}
			if userName != "" {
				fmt.Fprintf(text, "%v %v: <b>%v</b>", disk, userName, score)
			}
		default:
			panic("unknown mode: " + string(p.mode))
		}

		if nextMove == player {
			if t.Locale().IsRtl {
				fmt.Fprintf(text, " <code>%v →</code>", t.Translate(revtrans.NextMove))
			} else {
				fmt.Fprintf(text, " <code>← %v</code>", t.Translate(revtrans.NextMove))
			}
		}
		text.WriteString("\n")
	}
	writeScore(revgame.Black, emoji.BlackCircle, blacksScore)
	writeScore(revgame.White, emoji.WhiteCircle, whitesScore)
	if possibleMove != "" {
		fmt.Fprintf(text, "%v - possile moves\n", possibleMove)
	}

	text.Write([]byte("\n<b>Transcript</b>: <i>"))
	text.WriteString(p.transcript.String())
	if p.backSteps > 0 {
		fmt.Fprintf(text, "-%v", p.backSteps)
	}
	text.Write([]byte("</i>"))

	if isCompleted {
		text.WriteString(t.Translate(revtrans.GameCompleted))
	}
	return text.String()
}

func renderReversiTgKeyboard(whc bots.WebhookContext, p placeDiskPayload, isCompleted bool, possibleMove, lang, tournamentID string) (kb *tgbotapi.InlineKeyboardMarkup) {
	if isCompleted {
		playAgainCallbackData := new(bytes.Buffer)
		switch p.mode {
		case revgame.SinglePlayer:
			playAgainCallbackData.WriteString(newBoardSinglePlayerCommandCode + "?")
		case revgame.MultiPlayer:
			playAgainCallbackData.WriteString(newBoardMultiPlayerCommandCode)
		}
		if lang != "" {
			playAgainCallbackData.WriteString("&l=" + lang)
		}
		if tournamentID != "" {
			playAgainCallbackData.WriteString("&t=" + tournamentID)
		}
		switchInlinePlay := whc.Locale().Code5[:2]
		kb = &tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{{
					Text: whc.Translate(patrans.PlayAgain), CallbackData: playAgainCallbackData.String(),
				}},
				{
					{Text: whc.Translate(patrans.MultiPlayer), SwitchInlineQuery: &switchInlinePlay},
				},
			},
		}
		return
	}

	rows := p.currentBoard.Rows(emoji.BlackCircle, emoji.WhiteCircle, possibleMove, " ")

	kb = &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			make([]tgbotapi.InlineKeyboardButton, 8),
			make([]tgbotapi.InlineKeyboardButton, 8),
			make([]tgbotapi.InlineKeyboardButton, 8),
			make([]tgbotapi.InlineKeyboardButton, 8),
			make([]tgbotapi.InlineKeyboardButton, 8),
			make([]tgbotapi.InlineKeyboardButton, 8),
			make([]tgbotapi.InlineKeyboardButton, 8),
			make([]tgbotapi.InlineKeyboardButton, 8),
		},
	}

	getButton := func(x, y int, text string) tgbotapi.InlineKeyboardButton {
		ca := turnbased.NewCellAddress(x, y)
		callbackData := getPlaceDiskSinglePlayerCallbackData(p, ca, lang, tournamentID)
		// if a.X == int8(x) && a.Y == int8(y) {
		// 	text = "(" + text + ")"
		// }
		return tgbotapi.NewInlineKeyboardButtonData(text, callbackData)
	}

	for y, row := range rows {
		for x, cell := range row {
			kb.InlineKeyboard[y][x] = getButton(x, y, cell)
		}
	}

	if p.mode == revgame.SinglePlayer {
		if lastMovesCount := len(p.transcript); lastMovesCount == 0 {
			// No additional buttons
		} else {
			replayRow := make([]tgbotapi.InlineKeyboardButton, 0, 3)

			locale := whc.Locale()

			if p.backSteps+1 < lastMovesCount || (p.backSteps+1 == lastMovesCount && p.currentBoard.Turns() == 1) {
				backButton := tgbotapi.InlineKeyboardButton{
					CallbackData: getPlaceDiskSinglePlayerCallbackData(p, turnbased.CellAddress("-1"), lang, tournamentID),
				}
				if locale.IsRtl {
					backButton.Text = emoji.PlayButton + " -1"
				} else {
					backButton.Text = "-1 " + emoji.ReverseButton
				}

				replayRow = append(replayRow, backButton)
			}

			if p.backSteps > 0 {
				forwardButton := tgbotapi.InlineKeyboardButton{
					CallbackData: getPlaceDiskSinglePlayerCallbackData(p, turnbased.CellAddress("+1"), lang, tournamentID),
				}
				if locale.IsRtl {
					forwardButton.Text = "+1 " + emoji.ReverseButton
				} else {
					forwardButton.Text = emoji.PlayButton + " +1"
				}
				replayRow = append(replayRow, forwardButton)
			}

			aiButton := tgbotapi.InlineKeyboardButton{
				Text:         emoji.RobotFace + " AI",
				CallbackData: getPlaceDiskSinglePlayerCallbackData(p, turnbased.CellAddress("~"), lang, tournamentID),
			}
			replayRow = append(replayRow, aiButton)
			if locale.IsRtl {
				replayRow[0], replayRow[len(replayRow)-1] = replayRow[len(replayRow)-1], replayRow[0]
			}
			kb.InlineKeyboard = append(kb.InlineKeyboard, replayRow)
		}
	}

	return
}
