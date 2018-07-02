package revcommands

import (
	"github.com/strongo/bots-framework/core"
	"net/url"
	"github.com/prizarena/turn-based"
	"bytes"
	"github.com/prizarena/reversi/server-go/revgame"
	"strconv"
	"github.com/pkg/errors"
	"github.com/prizarena/prizarena-public/pamodels"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/platforms/telegram"
	"strings"
	"github.com/strongo/log"
	"github.com/strongo/emoji/go/emoji"
	"github.com/strongo/app"
	"fmt"
)

const placeDiskCommandCode = "place"

func getPlaceDiskSinglePlayerCallbackData(board revgame.Board, address turnbased.CellAddress, nextDisk revgame.Disk, lang, tournamentID string) string {
	var s bytes.Buffer
	s.WriteString("place?d=" + string(nextDisk))
	s.WriteString("&ca=" + string(address))
	s.WriteString("&b=" + strconv.FormatInt(int64(board.Blacks), 36))
	s.WriteString("&w=" + strconv.FormatInt(int64(board.Whites), 36))
	if tournamentID != "" {
		s.WriteString("&t=" + tournamentID)
	}
	if lang != "" {
		s.WriteString("&l=" + lang)
	}
	return s.String()
}

var placeDiskCommand = bots.NewCallbackCommand(
	placeDiskCommandCode,
	func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		c := whc.Context()
		var board revgame.Board
		var disks int64
		q := callbackUrl.Query()
		if disks, err = strconv.ParseInt(q.Get("b"), 36, 64); err != nil {
			return
		} else {
			board.Blacks = revgame.Disks(disks)
		}
		if disks, err = strconv.ParseInt(q.Get("w"), 36, 64); err != nil {
			return
		} else {
			board.Whites = revgame.Disks(disks)
		}
		ca := turnbased.CellAddress(q.Get("ca"))
		var currentPlayer, nextPlayer revgame.Disk
		x, y := ca.XY()

		switch q.Get("d") {
		case "w":
			currentPlayer = revgame.White
		case "b":
			currentPlayer = revgame.Black
		default:
			err = errors.New("unknown disk: " + q.Get("d"))
		}

		possibleMove := ""

		// -- Start[ Make move ]--
		if board, err = board.MakeMove(currentPlayer, x, y); err != nil {
			if cause := errors.Cause(err); cause == revgame.ErrNotValidMove || cause == revgame.ErrAlreadyOccupied {
				log.Debugf(c, "Wrong move: %v", cause)
				m.BotMessage = telegram.CallbackAnswer(tgbotapi.AnswerCallbackQueryConfig{
					Text: strings.Title(cause.Error()) + ".",
				})
				if _, err = whc.Responder().SendMessage(c, m, bots.BotAPISendMessageOverHTTPS); err != nil {
					log.Errorf(c, err.Error())
					err = nil // Non critical
				}
				if cause == revgame.ErrNotValidMove {
					possibleMove = emoji.SoccerBall
				}
				m.BotMessage = nil
			} else {
				return
			}
			nextPlayer = currentPlayer
		} else {
			nextPlayer = revgame.OtherPlayer(currentPlayer)
		}
		//-- End[ Make move ]--

		var tournament pamodels.Tournament
		tournament.ID = q.Get("t")
		lang := q.Get("l")
		if lang != "" {
			if err = whc.SetLocale(lang); err != nil {
				return
			}
		}
		m.IsEdit = true
		m.Format = bots.MessageFormatHTML
		m.Text = renderReversiBoardText(whc, board)
		m.Keyboard = renderReversiTgKeyboard(board, nextPlayer, possibleMove, lang, tournament.ID)
		return
	},
)

func renderReversiBoardText(t strongo.SingleLocaleTranslator, board revgame.Board) string {
	text := new(bytes.Buffer)
	text.WriteString("<b>Reversi game</b>\n")
	blacksScore, whitesScore := board.Score()
	nextMove := board.NextMove()
	writeScore := func(player revgame.Disk, disk string, score int) {
		fmt.Fprintf(text, "%v: %v", disk, score)
		if nextMove == player {
			text.WriteString(" ← next move")
		}
		text.WriteString("\n")
	}
	writeScore(revgame.Black, emoji.BlackCircle, blacksScore)
	writeScore(revgame.White, emoji.WhiteCircle, whitesScore)
	return text.String()
}
