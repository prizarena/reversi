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
	"strings"
	"github.com/strongo/log"
	"fmt"
	"github.com/prizarena/reversi/server-go/revdal"
	"context"
	"github.com/strongo/db"
	"github.com/prizarena/reversi/server-go/revmodels"
	"github.com/strongo/bots-framework/platforms/telegram"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/emoji/go/emoji"
	"github.com/strongo/slices"
	"time"
	"regexp"
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
	userID, userName    string
	userNames           []string
	isNewUser           bool
	board, currentBoard revgame.Board
	mode                revgame.Mode
	backSteps           int
	transcript          revgame.Transcript
}

var reTranscript = regexp.MustCompile(`Transcript: (([A-H][1-8])+(-\d+)?)`)

func placeDiskCallbackAction(whc bots.WebhookContext, callbackUrl *url.URL) (m bots.MessageFromBot, err error) {

	data := whc.Input().(bots.WebhookCallbackQuery).GetData()
	items := strings.SplitN(data, ".", 3)

	if len(items) > 2 {
		callbackUrl.RawQuery = strings.Replace(items[2], ".", "&", -1)
	}

	var p placeDiskPayload

	q := callbackUrl.Query()
	p.mode = revgame.Mode(q.Get("m"))
	p.userID = whc.AppUserStrID()
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

	if p.mode == revgame.SinglePlayer { // Get board & current board
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

		// Migrating to shorten callback data
		if update := whc.Input().(telegram.TgWebhookInput).TgUpdate(); update.CallbackQuery.Message == nil {
			log.Warningf(whc.Context(), "update.CallbackQuery.Message == nil")
		} else {
			groups := reTranscript.FindStringSubmatch(update.CallbackQuery.Message.Text)
			if len(groups) > 1 {
				log.Debugf(whc.Context(), "reTranscript groups: %v", groups)
				items := strings.Split(groups[1], "-")
				transcript := revgame.NewTranscriptFromHumanReadable(items[0])
				if !transcript.Equal(p.transcript) {
					log.Debugf(whc.Context(), "transcript != p.transcript: %v != %v", transcript, p.transcript)
				}
				var backSteps int
				if len(items) > 1 {
					if backSteps, err  = strconv.Atoi(items[1]); err != nil {
						return
					}
				}
				if backSteps != p.backSteps {
					log.Debugf(whc.Context(), "backSteps != p.backSteps: %v != %v", backSteps, p.backSteps)
				}
			} else {
				log.Debugf(whc.Context(), "Transcript not found in message")
			}
			const startTag = "</b>: <i>"
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
	switch p.mode {
	case revgame.SinglePlayer:
		return placeDiskSinglePlayer(whc, callbackUrl, p, a)
	case revgame.MultiPlayer:
		return placeDiskMultiPlayer(whc, callbackUrl, p, a)
	default:
		panic("unknown mode")
	}
}

func placeDiskSinglePlayer(whc bots.WebhookContext, callbackUrl *url.URL, p placeDiskPayload, a revgame.Address) (m bots.MessageFromBot, err error) {
	p.currentBoard = revgame.Replay(p.board, p.transcript, p.backSteps)
	if _, m, err = placeDiskToBoard(whc, callbackUrl, &p, a); err != nil {
		return
	}
	return
}

func placeDiskMultiPlayer(whc bots.WebhookContext, callbackUrl *url.URL, p placeDiskPayload, a revgame.Address) (m bots.MessageFromBot, err error) {
	c := whc.Context()
	var boardEH revmodels.RevBoard
	if boardEH.ID, err = turnbased.GetBoardID(whc.Input(), callbackUrl.Query().Get("b")); err != nil {
		return
	}

	err = revdal.DB.RunInTransaction(c, func(c context.Context) (err error) {
		if err = revdal.DB.Get(c, &boardEH); err != nil {
			if db.IsNotFound(err) {
				log.Debugf(c, "New board entity")
				boardEH.RevBoardEntity = &revmodels.RevBoardEntity{
					BoardEntityBase: turnbased.BoardEntityBase{
						Created: time.Now(),
						Lang:    whc.Locale().Code5,
						Round:   1,
					},
				}
				p.currentBoard = revgame.OthelloBoard
			} else {
				return
			}
		} else if p.currentBoard, err = boardEH.GetBoard(); err != nil {
			return
		}
		p.board = p.currentBoard // In multi-player we do not have rollback steps, so always "board == currentBoard"

		p.isNewUser = !slices.IsInStringSlice(p.userID, boardEH.UserIDs)
		if p.isNewUser {
			if len(boardEH.UserIDs) >= 2 { // Attempt to join as 3d user
				m.BotMessage = telegram.CallbackAnswer(tgbotapi.AnswerCallbackQueryConfig{
					Text: "This board already have 2 players",
				})
				return
			}
			if appUserEntity, err := whc.GetAppUser(); err != nil {
				return err
			} else {
				p.userName = appUserEntity.(*revmodels.UserEntity).GetFullName()
				boardEH.UserIDs = append(boardEH.UserIDs, p.userID)
				boardEH.UserNames = append(boardEH.UserNames, p.userName)
			}
		} else { // Existing user already belongs to the board

			// Check it is user's turn
			var isUserTurn bool
			switch p.currentBoard.NextPlayer() {
			case revgame.Black:
				isUserTurn = len(boardEH.UserIDs) > 0 && boardEH.UserIDs[0] == p.userID
			case revgame.White:
				isUserTurn = len(boardEH.UserIDs) > 1 && boardEH.UserIDs[1] == p.userID
			}
			if !isUserTurn {
				m.BotMessage = telegram.CallbackAnswer(tgbotapi.AnswerCallbackQueryConfig{
					Text: "You already made your move, wait for other player to respond.",
				})
				return
			}

			p.userName = boardEH.GetUserName(p.userID)
		}

		p.userNames = boardEH.UserNames

		var isPlaced bool
		prevBoardState := p.currentBoard
		if isPlaced, m, err = placeDiskToBoard(whc, callbackUrl, &p, a); err != nil {
			return
		}

		if isPlaced && prevBoardState == p.currentBoard {
			err = errors.New("game logic failure: isPlaced && prevBoardState == p.currentBoard")
			return
		}

		if isPlaced {
			boardEH.SetBoardState(p.currentBoard)
			transcript, _ := revgame.AddMove(revgame.NewTranscript(boardEH.BoardHistory), 0, a)
			boardEH.BoardHistory = transcript.ToBase64()

			if err = revdal.DB.Update(c, &boardEH); err != nil {
				return
			}
		}
		return
	}, db.SingleGroupTransaction)
	return
}

func placeDiskToBoard(whc bots.WebhookContext, callbackUrl *url.URL, p *placeDiskPayload, a revgame.Address) (isPlaced bool, m bots.MessageFromBot, err error) {
	c := whc.Context()

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
				// possibleMove = emoji.SoccerBall
				switch p.currentBoard.NextPlayer() {
				case revgame.White:
					possibleMove = emoji.WhiteSmallSquare
				case revgame.Black:
					possibleMove = emoji.BlackSmallSquare
				default:
					panic(fmt.Sprintf("unknown player: %v", p.board.NextPlayer()))
				}
			}
			m.BotMessage = nil
		} else {
			return
		}
	} else {
		isPlaced = true
		p.transcript, p.backSteps = revgame.AddMove(p.transcript, p.backSteps, a)
	}

	m, err = renderTelegramMessage(whc, callbackUrl, *p, possibleMove)
	return
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
	m.Text = renderReversiBoardText(whc, p, isCompleted, possibleMove)
	m.Keyboard = renderReversiTgKeyboard(whc, p, isCompleted, possibleMove, lang, tournament.ID)
	return
}

// func getPlayerFromString(s string) (player revgame.Disk) {
// 	p, _ := utf8.DecodeRuneInString(s)
// 	player = revgame.Disk(p)
// 	return
// }
//
