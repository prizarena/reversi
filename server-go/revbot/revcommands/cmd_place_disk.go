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

	s.Write([]byte(address))
	s.WriteRune('.')
	if board != revgame.OthelloBoard {
		s.WriteString(board.DisksToString())
	}

	if tournamentID != "" {
		s.WriteString(".t=" + tournamentID)
	}
	if mode == revgame.MultiPlayer && lang != "" {
		s.WriteString(".l=" + lang)
	}
	switch mode {
	case revgame.SinglePlayer:
		if len(lastMoves) == 0 {
			s.WriteString(".m=s")
		} else {
			if backSteps > 0 {
				s.WriteString(".r=" + strconv.Itoa(backSteps))
			}
			s.WriteString(".h=")
			// const limit = 64
			// left := limit - s.Len() // - strings.Count(s.String(), "&")*5 // \u0026 // TODO: Consider replacing '&' with '.' and then do manual reverse replace in callbackURL
			// if len(lastMoves) > left {
			// 	lastMoves = lastMoves[len(lastMoves)-left:]
			// }
			if len(lastMoves) > 11 {
				lastMoves = lastMoves[len(lastMoves)-1-11:]
			}
			s.WriteString(lastMoves.ToBase64())
		}
	}
	return s.String()
}

var placeDiskCommand = bots.Command{
	Code: placeDiskCommandCode,
	Matcher: func(command bots.Command, whc bots.WebhookContext) bool {
		c := whc.Context()
		if cmd, ok := whc.Input().(bots.WebhookCallbackQuery); ok {
			data := cmd.GetData()
			log.Debugf(c, "placeDiskCommand.Matcher(): data: %v", data)
			if len(data) == 0 {
				return false
			}
			f := data[0]
			result := f == '+' || f == '-' || f == '~' || (f >= 'A' && f <= 'H')
			log.Debugf(c, "placeDiskCommand.Matcher(): result: %v", result)
			return result
		} else {
			log.Debugf(c, "placeDiskCommand.Matcher(): not a WebhookCallbackQuery")
		}
		return false
	},
	CallbackAction: func(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {

		data := whc.Input().(bots.WebhookCallbackQuery).GetData()
		items := strings.SplitN(data, ".", 3)

		if len(items) > 2 {
			callbackUrl.RawQuery = strings.Replace(items[2], ".", "&", -1)
		}

		q := callbackUrl.Query()
		mode := revgame.Mode(q.Get("m"))
		switch mode {
		// case revgame.WithAI:
		// 	player = getPlayerFromString(q.Get("p"))
		case revgame.SinglePlayer, revgame.MultiPlayer: // OK
		case "":
			if q.Get("h") != "" {
				mode = revgame.SinglePlayer
			} else {
				mode = revgame.MultiPlayer
			}
		default:
			err = fmt.Errorf("unknown mode: [%v]", mode)
		}

		var board revgame.Board

		transcript := revgame.NewTranscript(q.Get("h"))

		{
			var b string
			if len(items) > 1 {
				b = items[1]
			}
			if b == "" {
				board = revgame.OthelloBoard
			} else {
				if board, err = revgame.NewBoardFromDisksString(b); err != nil {
					return
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
		}

		var backSteps int
		if sBackSteps := q.Get("r"); sBackSteps != "" {
			if backSteps, err = strconv.Atoi(sBackSteps); err != nil {
				err = errors.WithMessage(err, "Parameter 'r' is expected to be an integer")
				return
			}
		}

		a := items[0]
		log.Debugf(whc.Context(), "request.Query[a]=[%v]", a)
		if a == "~" {
			return aiAction(whc, callbackUrl, board, mode, transcript, backSteps)
		} else {
			switch a[0] {
			case '+', '-', ' ': // + is replaced with space by URL encoding
				var step int
				if a[0] == ' ' || a[0] == '+' {
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
				address := revgame.CellAddressToRevAddress(turnbased.CellAddress(a))
				if !address.IsOnBoard() {
					panic(fmt.Sprintf("Invalid adddress parameter {%v}.IsOnBoard() => false: %v", address, a))
				}
				return placeDiskAction(whc, callbackUrl, address, board, mode, transcript, backSteps)
			}
		}
	},
}

func aiAction(whc bots.WebhookContext, callbackUrl *url.URL, board revgame.Board, mode revgame.Mode, transcript revgame.Transcript, backSteps int) (m bots.MessageFromBot, err error) {
	currentBoard := board
	if backSteps > 0 {
		currentBoard, _ = revgame.Rewind(currentBoard, transcript, backSteps)
	}
	currentPlayer := currentBoard.NextPlayer()
	a := revgame.SimpleAI{}.GetMove(currentBoard, currentPlayer)
	currentBoard, err = currentBoard.MakeMove(currentPlayer, a)
	transcript, backSteps = revgame.AddMoveToTranscript(transcript, backSteps, a)
	return renderTelegramMessage(whc, callbackUrl, currentBoard, currentBoard, a, mode, transcript, backSteps, "")
}

func replayAction(whc bots.WebhookContext, callbackUrl *url.URL, board revgame.Board, mode revgame.Mode, transcript revgame.Transcript, backSteps, step int) (m bots.MessageFromBot, err error) {
	var currentBoard revgame.Board
	if step == 0 {
		err = errors.New("replayAction(step == 0)")
		return
	}
	// var a revgame.Address
	currentBoard, _ = revgame.Rewind(board, transcript, backSteps-step)
	backSteps -= step
	// if step < 0 {
	// 	a = revgame.EmptyAddress
	// }
	return renderTelegramMessage(whc, callbackUrl, board, currentBoard, revgame.EmptyAddress, mode, transcript, backSteps, "")
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
				Text:      strings.Title(cause.Error()) + ".",
				ShowAlert: cause == revgame.ErrAlreadyOccupied,
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
	} else {
		transcript, backSteps = revgame.AddMoveToTranscript(transcript, backSteps, a)
	}

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

// func getPlayerFromString(s string) (player revgame.Disk) {
// 	p, _ := utf8.DecodeRuneInString(s)
// 	player = revgame.Disk(p)
// 	return
// }
//
