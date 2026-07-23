// Package revplay is the host-agnostic Reversi "play layer". It drives the
// existing revgame engine (all rules + AI come from there) and turns a game
// into (a) a compact snapshot that round-trips through Telegram callback data
// and (b) an inline keyboard built with bots-go-core/botkb. It deliberately
// depends only on revgame and botkb — never on a messenger API — so any Sneat
// bot can mount it.
//
// The human may play either colour. Move application is split into two steps —
// ApplyHumanMove then PlayOpponent — so a host bot can render the board once the
// human has moved and again after the opponent replies, giving the player a
// visible "you moved, now the opponent thinks" beat.
package revplay

import (
	"errors"
	"fmt"
	"strings"

	"github.com/prizarena/reversi/server-go/revgame"
)

// Opponent identifies which built-in robotic player takes the opponent's moves.
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
// callback data: the engine Board (both bitboards + last move), the chosen
// opponent mode, and the colour the human plays. There is no server-side game
// storage — the message is the game.
type Snapshot struct {
	Board    revgame.Board
	Opponent Opponent
	// Human is the colour the human plays: revgame.Black or revgame.White.
	Human revgame.Disk
}

// NewGame returns the Othello starting position for the given opponent and human
// colour. Black always moves first (standard Othello): if human is White, the
// opponent (Black) is to move first.
func NewGame(opp Opponent, human revgame.Disk) Snapshot {
	return Snapshot{Board: revgame.OthelloBoard, Opponent: opp, Human: human}
}

// HumanColor returns the colour the human plays.
func (s Snapshot) HumanColor() revgame.Disk { return s.Human }

// OpponentColor returns the colour the robotic opponent plays.
func (s Snapshot) OpponentColor() revgame.Disk { return revgame.OtherPlayer(s.Human) }

// IsGameOver reports whether the game has finished (neither side can move).
// Hosts use it to decide whether to offer a "start a new game" affordance.
func (s Snapshot) IsGameOver() bool { return s.Board.IsCompleted() }

// IsHumanTurn reports whether the game is live and it is the human's move.
func (s Snapshot) IsHumanTurn() bool {
	return !s.IsGameOver() && s.Board.NextPlayer() == s.Human
}

// IsOpponentTurn reports whether the game is live and it is the opponent's move.
func (s Snapshot) IsOpponentTurn() bool {
	return !s.IsGameOver() && s.Board.NextPlayer() == s.OpponentColor()
}

// snapshotSeparator separates the opponent+human prefix from the board. base64
// RawURL never emits '.', so it is an unambiguous, callback-safe delimiter.
const snapshotSeparator = "."

// Encode serialises the snapshot into a compact, callback-safe string of the
// form "<opponent><humanChar>.<board-base64>", where humanChar is 'b' or 'w'.
// The board uses the engine's 23-char base64 encoder, so the whole string is 26
// chars — comfortably under Telegram's 64-byte callback limit even once a host
// command prefix and a target cell are appended.
func (s Snapshot) Encode() string {
	return string(s.Opponent) + string(rune(s.Human)) + snapshotSeparator + s.Board.ToBase64()
}

// ErrInvalidSnapshot is returned when a string cannot be decoded into a Snapshot.
var ErrInvalidSnapshot = errors.New("invalid reversi snapshot")

// DecodeSnapshot is the exact inverse of Encode. It fails with ErrInvalidSnapshot
// if the delimiter is missing, the prefix is not two chars, the opponent mode is
// unknown, the human colour char is not 'b'/'w', or the board segment is not
// valid engine base64.
func DecodeSnapshot(s string) (Snapshot, error) {
	i := strings.Index(s, snapshotSeparator)
	if i < 0 {
		return Snapshot{}, ErrInvalidSnapshot
	}
	prefix := s[:i]
	if len(prefix) != 2 {
		return Snapshot{}, fmt.Errorf("%w: bad prefix %q", ErrInvalidSnapshot, prefix)
	}
	opp := Opponent(prefix[0:1])
	if !opp.valid() {
		return Snapshot{}, fmt.Errorf("%w: unknown opponent %q", ErrInvalidSnapshot, prefix[0:1])
	}
	human := revgame.Disk(prefix[1])
	if human != revgame.Black && human != revgame.White {
		return Snapshot{}, fmt.Errorf("%w: unknown human colour %q", ErrInvalidSnapshot, prefix[1:2])
	}
	board, err := revgame.NewBoardFromBase64(s[i+1:])
	if err != nil {
		return Snapshot{}, fmt.Errorf("%w: %v", ErrInvalidSnapshot, err)
	}
	return Snapshot{Board: board, Opponent: opp, Human: human}, nil
}
