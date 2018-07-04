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

func getPlaceDiskSinglePlayerCallbackData(board revgame.Board, mode revgame.Mode, player revgame.Disk, address turnbased.CellAddress, lastMoves revgame.Transcript, backSteps int, lang, tournamentID string) string {
	s := new(bytes.Buffer)
	s.WriteString("place?a=" + string(address))
	if mode != revgame.MultiPlayer {
		s.WriteString("&m=" + string(mode))
		if turns := board.Turns(); turns > 0 {
			s.WriteString("&c=" + strconv.Itoa(turns))
		}
		if mode == revgame.WithAI {
			switch player {
			case revgame.Black, revgame.White:
				s.WriteString("&p=" + string(player))
			default:
				panic("mode=WithAI has unexpected player: " + string(player))
			}
		}
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
		switch a[0] {
		case '+', '-':
			return replayAction(whc, callbackUrl, board, mode, transcript, backSteps, player)
		default:
			return placeDiskAction(whc, callbackUrl, board, mode, transcript, backSteps, player)
		}
	},
)

func replayAction(whc bots.WebhookContext, callbackUrl *url.URL, board revgame.Board, mode revgame.Mode, transcript revgame.Transcript, backSteps int, player revgame.Disk) (m bots.MessageFromBot, err error) {
	q := callbackUrl.Query()
	var replay int
	if replay, err = strconv.Atoi(q.Get("a")); err != nil {
		return
	}

	if replay == 0 {
		err = errors.New("Invalid 'a' e.g. 'replay' parameter, should be != 0")
		return
	} else if replay < 0 {
		lastMoves := transcript
		for replay < 0 && len(lastMoves) > 0 {
			var lastMove revgame.Move
			lastMove, lastMoves = lastMoves.Pop()
			a := lastMove.Address()
			var prevMove revgame.Address
			if len(lastMoves) == 0 {
				prevMove = revgame.EmptyAddress
			} else {
				prevMove = lastMoves.LastMove().Address()
			}
			board = board.UndoMove(a, prevMove)
		}
	} else if replay > 0 {
		//
		// board, err = board.MakeMove(currentPlayer, x, y)
	}
	return renderTelegramMessage(whc, callbackUrl, board, mode, player, transcript, backSteps,"")
}

func placeDiskAction(whc bots.WebhookContext, callbackUrl *url.URL, board revgame.Board, mode revgame.Mode, transcript revgame.Transcript, backSteps int, player revgame.Disk) (m bots.MessageFromBot, err error) {
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
	if mode == revgame.WithAI && player != currentPlayer {
		a = revgame.SimpleAI{}.GetMove(board, currentPlayer)
		board, err = board.MakeMove(currentPlayer, a)
	} else {
		board, err = board.MakeMove(currentPlayer, a)
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
		lastMoves = revgame.NewTranscript(q.Get("h"))
		lastMoves = append(lastMoves, byte(a.Index()))
	}

	return renderTelegramMessage(whc, callbackUrl, board, mode, player, lastMoves, backSteps, possibleMove)
}

func renderTelegramMessage(whc bots.WebhookContext, callbackUrl *url.URL, board revgame.Board, mode revgame.Mode, player revgame.Disk, lastMoves revgame.Transcript, backSteps int, possibleMove string) (m bots.MessageFromBot, err error) {
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
	m.Keyboard = renderReversiTgKeyboard(board, mode, player, isCompleted, lastMoves, backSteps, possibleMove, lang, tournament.ID)
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
