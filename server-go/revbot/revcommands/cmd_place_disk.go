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

func getPlaceDiskSinglePlayerCallbackData(p placeDiskPayload, address turnbased.CellAddress, lang, tournamentID string) string {
	s := new(bytes.Buffer)

	s.Write([]byte(address))
	s.WriteRune('.')
	if p.board.DisksCount() > 4 { // Not optimal to count for every button
		s.WriteString(p.board.ToBase64())
	}

	if tournamentID != "" {
		s.WriteString(".t=" + tournamentID)
	}
	if p.mode == revgame.MultiPlayer && lang != "" {
		s.WriteString(".l=" + lang)
	}
	switch p.mode {
	case revgame.SinglePlayer:
		if len(p.transcript) == 0 {
			s.WriteString(".m=s")
		} else {
			if p.backSteps > 0 {
				s.WriteString(".r=" + strconv.FormatInt(int64(p.backSteps), 36))
			}
			s.WriteString(".h=")
			s.WriteString(p.transcript.ToBase64())
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
	CallbackAction: placeDiskCallbackAction,
}

type placeDiskPayload struct {
	board, currentBoard revgame.Board
	mode revgame.Mode
	backSteps int
	transcript revgame.Transcript
}

func placeDiskCallbackAction(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {

	data := whc.Input().(bots.WebhookCallbackQuery).GetData()
	items := strings.SplitN(data, ".", 3)

	if len(items) > 2 {
		callbackUrl.RawQuery = strings.Replace(items[2], ".", "&", -1)
	}

	var p placeDiskPayload

	q := callbackUrl.Query()
	p.mode = revgame.Mode(q.Get("m"))
	switch p.mode {
	// case revgame.WithAI:
	// 	player = getPlayerFromString(q.Get("p"))
	case revgame.SinglePlayer, revgame.MultiPlayer: // OK
	case "":
		if q.Get("h") != "" {
			p.mode = revgame.SinglePlayer
		} else {
			p.mode = revgame.MultiPlayer
		}
	default:
		err = fmt.Errorf("unknown mode: [%v]", p.mode)
	}

	p.transcript = revgame.NewTranscript(q.Get("h"))

	if sBackSteps := q.Get("r"); sBackSteps != "" {
		var iBackStep int64
		if iBackStep, err = strconv.ParseInt(sBackSteps, 36, 8); err != nil {
			err = errors.WithMessage(err, "Parameter 'r' is expected to be a base36 encoded integer")
			return
		}
		p.backSteps = int(iBackStep)
	}

	{ // Get board & current board
		var b string
		if len(items) > 1 {
			b = items[1]
		}
		if b == "" {
			p.board = revgame.OthelloBoard
		} else {
			if p.board, err = revgame.NewBoardFromBase64(b); err != nil {
				return
			}
			if err = revgame.VerifyBoardTranscript(p.board, p.transcript); err != nil {
				return
			}
		}
	}

	a := items[0]
	log.Debugf(whc.Context(), "request.Query[a]=[%v]", a)
	if a == "~" {
		return aiAction(whc, callbackUrl, p)
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
			return replayAction(whc, callbackUrl, p, step)
		default:
			address := revgame.CellAddressToRevAddress(turnbased.CellAddress(a))
			if !address.IsOnBoard() {
				panic(fmt.Sprintf("Invalid adddress parameter {%v}.IsOnBoard() => false: %v", address, a))
			}
			return placeDiskAction(whc, callbackUrl, p, address)
		}
	}
}

func aiAction(whc bots.WebhookContext, callbackUrl *url.URL, p placeDiskPayload) (m bots.MessageFromBot, err error) {
	// if backSteps > 0 {
	// 	currentBoard, _ = revgame.Rewind(currentBoard, transcript, backSteps)
	// }
	p.currentBoard = revgame.Replay(p.board, p.transcript, p.backSteps)
	currentPlayer := p.currentBoard.NextPlayer()
	a := revgame.SimpleAI{}.GetMove(p.currentBoard, currentPlayer)
	p.currentBoard, err = p.currentBoard.MakeMove(currentPlayer, a)
	p.transcript, p.backSteps = revgame.AddMove(p.transcript, p.backSteps, a)
	return renderTelegramMessage(whc, callbackUrl, p, "")
}

func replayAction(whc bots.WebhookContext, callbackUrl *url.URL, p placeDiskPayload, step int) (m bots.MessageFromBot, err error) {
	if step == 0 {
		err = errors.New("replayAction(step == 0)")
		return
	}
	p.currentBoard = revgame.Replay(p.board, p.transcript, p.backSteps-step)
	p.backSteps -= step
	return renderTelegramMessage(whc, callbackUrl, p, "")
}

func placeDiskAction(whc bots.WebhookContext, callbackUrl *url.URL, p placeDiskPayload, a revgame.Address) (m bots.MessageFromBot, err error) {
	c := whc.Context()

	p.currentBoard = revgame.Replay(p.board, p.transcript, p.backSteps)

	var currentPlayer revgame.Disk

	if currentPlayer = p.currentBoard.NextPlayer(); currentPlayer == revgame.Completed {
		m.BotMessage = telegram.CallbackAnswer(tgbotapi.AnswerCallbackQueryConfig{
			Text:      "This game has been completed",
			ShowAlert: true,
		})
		return
	}

	possibleMove := ""

	// -- Start[ Make move ]--
	p.currentBoard, err = p.currentBoard.MakeMove(currentPlayer, a)
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
		p.transcript, p.backSteps = revgame.AddMove(p.transcript, p.backSteps, a)
	}

	return renderTelegramMessage(whc, callbackUrl, p, possibleMove)
}

func renderTelegramMessage(whc bots.WebhookContext, callbackUrl *url.URL, p placeDiskPayload, possibleMove string) (m bots.MessageFromBot, err error) {
	q := callbackUrl.Query()
	lang := q.Get("l")
	if lang != "" {
		if err = whc.SetLocale(lang); err != nil {
			return
		}
	}
	var tournament pamodels.Tournament
	tournament.ID = q.Get("t")

	{ // Slide history window
		const historyLimit = 11
		historyLen := len(p.transcript)
		log.Debugf(whc.Context(), "p.mode: %v, historyLimit: %v, historyLen: %v", p.mode, historyLimit, historyLen)
		if slideSteps := historyLen - historyLimit; p.mode == revgame.SinglePlayer && slideSteps > 0 {
			for ; slideSteps > 0; slideSteps-- {
				var move revgame.Move
				move, p.transcript = p.transcript.NextMove()
				player := p.board.NextPlayer()
				if p.board, err = p.board.MakeMove(player, move.Address()); err != nil {
					return
				}
			}
			p.backSteps += slideSteps
		}
	}

	if err = revgame.VerifyBoardTranscript(p.board, p.transcript); err != nil { // better to cover by unit tests
		return
	}

	m.IsEdit = true
	m.Format = bots.MessageFormatHTML
	isCompleted := p.currentBoard.IsCompleted()
	m.Text = renderReversiBoardText(whc, p.currentBoard, p.mode, isCompleted, nil)
	m.Keyboard = renderReversiTgKeyboard(p, isCompleted, possibleMove, lang, tournament.ID)
	return
}

// func getPlayerFromString(s string) (player revgame.Disk) {
// 	p, _ := utf8.DecodeRuneInString(s)
// 	player = revgame.Disk(p)
// 	return
// }
//
