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
)

const placeDiskCommandCode = "p"

func getPlaceDiskSinglePlayerCallbackData(board revgame.Board, mode revgame.Mode, address turnbased.CellAddress, lastMoves revgame.Transcript, backSteps int, lang, tournamentID string) string {
	s := new(bytes.Buffer)
	s.WriteString(placeDiskCommandCode + "?a=" + string(address))
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
			if err = revgame.VerifyBoardTranscript(board, transcript); err != nil {
				return
			}
		}

		var backSteps int
		if sBackSteps := q.Get("r"); sBackSteps != "" {
			if backSteps, err = strconv.Atoi(sBackSteps); err != nil {
				err = errors.WithMessage(err, "Parameter 'r' is expected to be an integer")
				return
			}
		}

		a := q.Get("a")
		log.Debugf(whc.Context(), "request.Query[a]=[%v]", a)
		if a == "~" {
			return aiAction(whc, callbackUrl, board, mode, transcript, backSteps)
		} else {
			switch a[0] {
			case ' ', '-': // + is replaced with space by URL encoding
				var step int
				if a[0] == ' ' {
					a = a[1:]
				}
				if step, err = strconv.Atoi(a); err != nil {
					return
				} else if step == 0 {
					err = errors.New("Invalid 'a' e.g. 'replay' parameter, should be != 0")
					return
				}
				return replayAction(whc, callbackUrl, board, mode, transcript, backSteps, step)
			default:
				a := revgame.CellAddressToRevAddress(turnbased.CellAddress(a))
				return placeDiskAction(whc, callbackUrl, a, board, mode, transcript, backSteps)
			}
		}
	},
)

func aiAction(whc bots.WebhookContext, callbackUrl *url.URL, board revgame.Board, mode revgame.Mode, transcript revgame.Transcript, backSteps int) (m bots.MessageFromBot, err error) {
	currentBoard := board
	if backSteps > 0 {
		currentBoard, _ = revgame.Rewind(currentBoard, transcript, backSteps)
	}
	currentPlayer := currentBoard.NextPlayer()
	a := revgame.SimpleAI{}.GetMove(board, currentPlayer)
	currentBoard, err = board.MakeMove(currentPlayer, a)
	transcript, backSteps = revgame.AddMoveToTranscript(transcript, backSteps, a)
	return renderTelegramMessage(whc, callbackUrl, currentBoard, currentBoard, a, mode, transcript, backSteps, "")
}

func replayAction(whc bots.WebhookContext, callbackUrl *url.URL, board revgame.Board, mode revgame.Mode, transcript revgame.Transcript, backSteps, step int) (m bots.MessageFromBot, err error) {
	var currentBoard revgame.Board
	if step == 0 {
		err = errors.New("replayAction(step == 0)")
		return
	}
	var a revgame.Address
	currentBoard, a = revgame.Rewind(board, transcript, backSteps-step)
	backSteps -= step
	if step < 0 {
		a = revgame.EmptyAddress
	}
	return renderTelegramMessage(whc, callbackUrl, board, currentBoard, a, mode, transcript, backSteps, "")
}

func placeDiskAction(whc bots.WebhookContext, callbackUrl *url.URL, a revgame.Address, board revgame.Board, mode revgame.Mode, transcript revgame.Transcript, backSteps int) (m bots.MessageFromBot, err error) {
	c := whc.Context()

	currentBoard := board

	var currentPlayer revgame.Disk
	if backSteps > 0 {
		currentBoard, _ = revgame.Rewind(currentBoard, transcript, backSteps)
	}

	if currentPlayer = currentBoard.NextPlayer(); currentPlayer == revgame.Completed {
		m.BotMessage = telegram.CallbackAnswer(tgbotapi.AnswerCallbackQueryConfig{
			Text:      "This game has been completed",
			ShowAlert: true,
		})
		return
	}

	possibleMove := ""

	// -- Start[ Make move ]--
	currentBoard, err = currentBoard.MakeMove(currentPlayer, a)
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

	transcript, backSteps = revgame.AddMoveToTranscript(transcript, backSteps, a)

	return renderTelegramMessage(whc, callbackUrl, currentBoard, currentBoard, a, mode, transcript, backSteps, possibleMove)
}

func renderTelegramMessage(whc bots.WebhookContext, callbackUrl *url.URL, board, currentBoard revgame.Board, a revgame.Address, mode revgame.Mode, lastMoves revgame.Transcript, backSteps int, possibleMove string) (m bots.MessageFromBot, err error) {
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
	m.Text = renderReversiBoardText(whc, currentBoard, mode, isCompleted, nil)
	m.Keyboard = renderReversiTgKeyboard(board, currentBoard, a, mode, isCompleted, lastMoves, backSteps, possibleMove, lang, tournament.ID)
	return
}

//func getPlayerFromString(s string) (player revgame.Disk) {
//	p, _ := utf8.DecodeRuneInString(s)
//	player = revgame.Disk(p)
//	return
//}
//
