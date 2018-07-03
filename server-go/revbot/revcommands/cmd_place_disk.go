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
	"github.com/prizarena/reversi/server-go/revtrans"
	"unicode/utf8"
)

const placeDiskCommandCode = "place"

func getPlaceDiskSinglePlayerCallbackData(board revgame.Board, mode revgame.Mode, player revgame.Disk, address turnbased.CellAddress, lastMoves revgame.Transcript, lang, tournamentID string) string {
	s := new(bytes.Buffer)
	s.WriteString("place?a=" + string(address))
	if mode != revgame.MultiPlayer {
		s.WriteString("&m=" + string(mode))
		s.WriteString("&c=" + strconv.Itoa(board.Turns()))
		if mode == revgame.WithAI {
			switch player {
			case revgame.Black, revgame.White:
				s.WriteString("&p=" + string(player))
			default:
				panic("mode=WithAI has unexpected player: " + string(player))
			}
		}
	}

	fmt.Fprintf(s, "&b=%v_%v_%v",
		strconv.FormatInt(int64(board.Blacks), 36),
		strconv.FormatInt(int64(board.Whites), 36),
		string(board.Last))
	if tournamentID != "" {
		s.WriteString("&t=" + tournamentID)
	}
	if lang != "" {
		s.WriteString("&l=" + lang)
	}
	if mode != revgame.MultiPlayer && lastMoves != "" {
		s.WriteString("&h=" + string(lastMoves))
	}
	return s.String()
}

func getPlayerFromString(s string) (player revgame.Disk) {
	p, _ := utf8.DecodeRuneInString(s)
	player = revgame.Disk(p)
	return
}

var placeDiskCommand = bots.NewCallbackCommand(
	placeDiskCommandCode,
	func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {
		q := callbackUrl.Query()
		mode := revgame.Mode(q.Get("m"))
		var player revgame.Disk // Needed for AI mode only so we can swap sides each turn
		switch mode {
		case revgame.WithAI:
			player = getPlayerFromString(q.Get("p"))
		case revgame.SinglePlayer, revgame.MultiPlayer: // OK
		case "":
			mode = revgame.MultiPlayer
		default:
			err = fmt.Errorf("unknown mode: [%v]", mode)
		}

		var board revgame.Board
		var disks int64

		{
			b := strings.Split(q.Get("b"), "_")
			if disks, err = strconv.ParseInt(b[0], 36, 64); err != nil {
				return
			} else {
				board.Blacks = revgame.Disks(disks)
			}
			if disks, err = strconv.ParseInt(b[1], 36, 64); err != nil {
				return
			} else {
				board.Whites = revgame.Disks(disks)
			}
			board.Last = getPlayerFromString(b[2])
		}

		a := q.Get("a")
		switch a {
		case "+1", "-1":
			return replayAction(whc, callbackUrl, board, mode, player)
		default:
			return placeDiskAction(whc, callbackUrl, board, mode, player)
		}
	},
)

func replayAction(whc bots.WebhookContext, callbackUrl *url.URL, board revgame.Board, mode revgame.Mode, player revgame.Disk) (m bots.MessageFromBot, err error) {
	q := callbackUrl.Query()
	replay := q.Get("a")
	var lastMoves revgame.Transcript
	if lastMoves, err = revgame.NewTranscript(q.Get("h")); err != nil {
		return
	}
	if replay == "-1" {
		lastMove := turnbased.CellAddress(lastMoves[len(lastMoves)-2:])
		if board, err = board.UndoMove(lastMove.XY()); err != nil {
			return
		}
	}
	return renderTelegramMessage(whc, callbackUrl, board, mode, player, lastMoves, "")
}

func placeDiskAction(whc bots.WebhookContext, callbackUrl *url.URL, board revgame.Board, mode revgame.Mode, player revgame.Disk) (m bots.MessageFromBot, err error) {
	c := whc.Context()
	q := callbackUrl.Query()
	ca := turnbased.CellAddress(q.Get("a"))
	x, y := ca.XY()

	currentPlayer := board.NextPlayer()
	if currentPlayer == revgame.Completed {
		m.BotMessage = telegram.CallbackAnswer(tgbotapi.AnswerCallbackQueryConfig{
			Text:      "This game has been completed",
			ShowAlert: true,
		})
		return
	}

	possibleMove := ""

	// -- Start[ Make move ]--
	if mode == revgame.WithAI && player != currentPlayer {
		move := revgame.SimpleAI{}.GetMove(board, currentPlayer)
		board, err = board.MakeMove(currentPlayer, move.X, move.Y)
	} else {
		board, err = board.MakeMove(currentPlayer, x, y)
	}
	// -- End[ Make move ]--

	if err != nil {
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
		// nextPlayer = currentPlayer
	} else {
		// nextPlayer = revgame.OtherPlayer(currentPlayer)
	}

	var lastMoves revgame.Transcript
	if mode != revgame.MultiPlayer {
		if lastMoves, err = revgame.NewTranscript(q.Get("h") + string(ca)); err != nil {
			return
		}
		if len(lastMoves) > 8 {
			lastMoves = lastMoves[len(lastMoves)-8:]
		}
	}

	return renderTelegramMessage(whc, callbackUrl, board, mode, player, lastMoves, possibleMove)
}

func renderTelegramMessage(whc bots.WebhookContext, callbackUrl *url.URL, board revgame.Board, mode revgame.Mode, player revgame.Disk, lastMoves revgame.Transcript, possibleMove string) (m bots.MessageFromBot, err error) {
	q := callbackUrl.Query()
	lang := q.Get("l")
	if lang != "" {
		if err = whc.SetLocale(lang); err != nil {
			return
		}
	}
	var tournament pamodels.Tournament
	tournament.ID = q.Get("t")

	m.IsEdit = true
	m.Format = bots.MessageFormatHTML
	isCompleted := board.IsCompleted()
	m.Text = renderReversiBoardText(whc, board, mode, player, isCompleted, nil)
	m.Keyboard = renderReversiTgKeyboard(board, mode, player, isCompleted, lastMoves, possibleMove, lang, tournament.ID)
	return
}

func renderReversiBoardText(t strongo.SingleLocaleTranslator, board revgame.Board, mode revgame.Mode, player revgame.Disk, isCompleted bool, userNames []string) string {
	text := new(bytes.Buffer)
	text.WriteString(fmt.Sprintf("<b>%v</b>\n", t.Translate(revtrans.GameCardTitle)))
	blacksScore, whitesScore := board.Scores()
	nextMove := board.NextPlayer()
	writeScore := func(p revgame.Disk, disk string, score int) {
		switch mode {
		case revgame.SinglePlayer:
			fmt.Fprintf(text, "%v: %v", disk, score)
		case revgame.WithAI:
			var name string
			if p == player {
				name = "me"
			} else {
				name = "AI"
			}
			fmt.Fprintf(text, "<code>%v (%v):</code> <b>%v</b>", disk, name, score)
		case revgame.MultiPlayer:
			switch p {
			case revgame.Black, revgame.White:
				fmt.Fprintf(text, "<code>%v (%v):</code> <b>%v</b>", disk, userNames[0], score)
			default:
				panic("unknown player: " + string(p))
			}
		default:
			panic("unknown mode: " + string(mode))
		}

		if nextMove == p {
			text.WriteString(" ← next move")
		}
		text.WriteString("\n")
	}
	writeScore(revgame.Black, emoji.BlackCircle, blacksScore)
	writeScore(revgame.White, emoji.WhiteCircle, whitesScore)
	if isCompleted {
		text.WriteString("Game is completed!\n")
	}
	return text.String()
}
