// Package revplay is the host-agnostic Reversi "play layer". It drives the
// existing revgame engine (all rules + AI come from there) and turns a game
// into (a) a compact snapshot that round-trips through Telegram callback data
// and (b) an inline keyboard built with bots-go-core/botkb. It deliberately
// depends only on revgame and botkb — never on a messenger API — so any Sneat
// bot can mount it.
package revplay

import (
	"errors"
	"fmt"
	"strings"

	"github.com/prizarena/reversi/server-go/revgame"
)

// Opponent identifies which built-in robotic player takes White's moves.
type Opponent string

const (
	// OpponentAI plays revgame.SimpleAI's heuristic move (greedy score-max with
	// a corner preference; non-deterministic tie-break, as-is from the engine).
	OpponentAI Opponent = "a"
	// OpponentRandom plays a uniformly random legal move.
	OpponentRandom Opponent = "r"
)

// valid reports whether o is a recognised opponent mode.
func (o Opponent) valid() bool {
	switch o {
	case OpponentAI, OpponentRandom:
		return true
	default:
		return false
	}
}

// Snapshot is the complete, self-contained game state carried in a button's
// callback data: the engine Board (both bitboards + last move) plus the chosen
// opponent mode. There is no server-side game storage — the message is the game.
type Snapshot struct {
	Board    revgame.Board
	Opponent Opponent
}

// NewGame returns the Othello starting position against the given opponent,
// with the human (Black) to move first.
func NewGame(opp Opponent) Snapshot {
	return Snapshot{Board: revgame.OthelloBoard, Opponent: opp}
}

// snapshotSeparator separates the opponent mode from the board. base64 RawURL
// never emits '.', so it is an unambiguous, callback-safe delimiter.
const snapshotSeparator = "."

// Encode serialises the snapshot into a compact, callback-safe string of the
// form "<opponent>.<board-base64>". The board uses the engine's 23-char base64
// encoder, so the whole string is 25 chars — comfortably under Telegram's
// 64-byte callback limit even once a host command prefix and a target cell are
// appended.
func (s Snapshot) Encode() string {
	return string(s.Opponent) + snapshotSeparator + s.Board.ToBase64()
}

// ErrInvalidSnapshot is returned when a string cannot be decoded into a Snapshot.
var ErrInvalidSnapshot = errors.New("invalid reversi snapshot")

// DecodeSnapshot is the exact inverse of Encode. It fails with ErrInvalidSnapshot
// if the delimiter is missing, the opponent mode is unknown, or the board
// segment is not valid engine base64.
func DecodeSnapshot(s string) (Snapshot, error) {
	i := strings.Index(s, snapshotSeparator)
	if i < 0 {
		return Snapshot{}, ErrInvalidSnapshot
	}
	opp := Opponent(s[:i])
	if !opp.valid() {
		return Snapshot{}, fmt.Errorf("%w: unknown opponent %q", ErrInvalidSnapshot, s[:i])
	}
	board, err := revgame.NewBoardFromBase64(s[i+1:])
	if err != nil {
		return Snapshot{}, fmt.Errorf("%w: %v", ErrInvalidSnapshot, err)
	}
	return Snapshot{Board: board, Opponent: opp}, nil
}
