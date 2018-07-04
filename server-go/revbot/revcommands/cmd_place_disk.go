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
		"fmt"
		"unicode/utf8"
)

const placeDiskCommandCode = "place"

func getPlaceDiskSinglePlayerCallbackData(board revgame.Board, mode revgame.Mode, address turnbased.CellAddress, lastMoves revgame.Transcript, backSteps int, lang, tournamentID string) string {
	s := new(bytes.Buffer)
	s.WriteString("place?a=" + string(address))
	if mode != revgame.MultiPlayer {
		s.WriteString("&m=" + string(mode))
		if turns := board.Turns(); turns > 0 {
			s.WriteString("&c=" + strconv.Itoa(turns))
		}
		// if mode == revgame.WithAI {
		// 	switch player {
		// 	case revgame.Black, revgame.White:
		// 		s.WriteString("&p=" + string(player))
		// 	default:
		// 		panic("mode=WithAI has unexpected player: " + string(player))
		// 	}
		// }
	}

	fmt.Fprintf(s, "&b=%v_%v",
		strconv.FormatInt(int64(board.Blacks), 36),
		strconv.FormatInt(int64(board.Whites), 36),
	)
	if tournamentID != "" {
		s.WriteString("&t=" + tournamentID)
	}
	if mode == revgame.MultiPlayer && lang != "" {
		s.WriteString("&l=" + lang)
	}
	if mode != revgame.MultiPlayer && len(lastMoves) != 0 {
		if backSteps > 0 {
			s.WriteString("&r=" + strconv.Itoa(backSteps))
		}
		s.WriteString("&h=")
		const limit = 64
		left := limit - s.Len()
		if len(lastMoves) > left {
			lastMoves = lastMoves[len(lastMoves)-left:]
		}
		s.WriteString(lastMoves.ToBase64())
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
		switch mode {
		// case revgame.WithAI:
		// 	player = getPlayerFromString(q.Get("p"))
		case revgame.SinglePlayer, revgame.MultiPlayer: // OK
		case "":
			mode = revgame.MultiPlayer
		default:
			err = fmt.Errorf("unknown mode: [%v]", mode)
		}

		var board revgame.Board
		var disks int64
		var backSteps int
		if sBackSteps := q.Get("r"); sBackSteps != "" {
			if backSteps, err = strconv.Atoi(sBackSteps); err != nil {
				err = errors.WithMessage(err, "Parameters 'r' is epxected to be an integer")
				return
			}
		}

		transcript := revgame.NewTranscript(q.Get("h"))

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
			if len(transcript) > 0 {
				board.Last = transcript.LastMove().Address()
			} else {
				board.Last = revgame.EmptyAddress
			}
		}

		a := q.Get("a")
		if a == "~" {
			return aiAction(whc, callbackUrl, board, mode, transcript, backSteps)
		} else {
			switch a[0] {
			case '+', '-':
				return replayAction(whc, callbackUrl, board, mode, transcript, backSteps)
			default:
				return placeDiskAction(whc, callbackUrl, board, mode, transcript, backSteps)
			}
		}
	},
)

func aiAction(whc bots.WebhookContext, callbackUrl *url.URL, board revgame.Board, mode revgame.Mode, transcript revgame.Transcript, backSteps int) (m bots.MessageFromBot, err error) {
	// a = revgame.SimpleAI{}.GetMove(board, currentPlayer)
	// board, err = board.MakeMove(currentPlayer, a)
	return
}

func rewind(board revgame.Board, transcript revgame.Transcript, backSteps, replay int) (pastBoard revgame.Board, nextMove revgame.Address) {
	lastMoves := transcript
	stepsToRollback := backSteps - replay // replay is negative, so we need '-' to sum.
	pastBoard = board
	nextMove = revgame.EmptyAddress
	for stepsToRollback > 0 && len(lastMoves) > 0 {
		stepsToRollback--
		var lastMove revgame.Move
		lastMove, lastMoves = lastMoves.Pop()
		a := lastMove.Address()
		var prevMove revgame.Address
		if len(lastMoves) == 0 {
			prevMove = revgame.EmptyAddress
		} else {
			prevMove = lastMoves.LastMove().Address()
		}
		pastBoard = pastBoard.UndoMove(a, prevMove)
		nextMove = a
	}
	return
}

func replayAction(whc bots.WebhookContext, callbackUrl *url.URL, board revgame.Board, mode revgame.Mode, transcript revgame.Transcript, backSteps int) (m bots.MessageFromBot, err error) {
	q := callbackUrl.Query()
	var replay int
	if replay, err = strconv.Atoi(q.Get("a")); err != nil {
		return
	}

	pastBoard := board

	if replay == 0 {
		err = errors.New("Invalid 'a' e.g. 'replay' parameter, should be != 0")
		return
	} else {
		if replay < 0 {
			pastBoard, _ = rewind(board, transcript, backSteps, replay)
		} else if replay > 0 {
			var nextMove revgame.Address
			pastBoard, nextMove = rewind(board, transcript, backSteps, 0)
			nextPlayer := board.NextPlayer()
			board, err = board.MakeMove(nextPlayer, nextMove)
		}
	}

	return renderTelegramMessage(whc, callbackUrl, board, pastBoard, mode, transcript, backSteps,"")
}

func placeDiskAction(whc bots.WebhookContext, callbackUrl *url.URL, board revgame.Board, mode revgame.Mode, transcript revgame.Transcript, backSteps int) (m bots.MessageFromBot, err error) {
	c := whc.Context()
	q := callbackUrl.Query()
	a := revgame.CellAddressToRevAddress(turnbased.CellAddress(q.Get("a")))

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
	board, err = board.MakeMove(currentPlayer, a)
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
		lastMoves = revgame.NewTranscript(q.Get("h"))
		lastMoves = append(lastMoves, byte(a.Index()))
	}

	return renderTelegramMessage(whc, callbackUrl, board, board, mode, lastMoves, backSteps, possibleMove)
}

func renderTelegramMessage(whc bots.WebhookContext, callbackUrl *url.URL, board, pastBoard revgame.Board, mode revgame.Mode, lastMoves revgame.Transcript, backSteps int, possibleMove string) (m bots.MessageFromBot, err error) {
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
	m.Text = renderReversiBoardText(whc, pastBoard, mode, isCompleted, nil)
	m.Keyboard = renderReversiTgKeyboard(board, pastBoard, mode, isCompleted, lastMoves, backSteps, possibleMove, lang, tournament.ID)
	return
}

