package revplay

import (
	"math/rand"

	"github.com/prizarena/reversi/server-go/revgame"
)

// HumanMove applies the human's Black move at cell, then lets the configured
// opponent play White replies until it is the human's turn again (or the game
// is over). Every rule comes from the revgame engine — this layer only
// sequences engine calls.
//
// The opponent loop runs while revgame.Board.NextPlayer reports White. Because
// NextPlayer already resolves passes, this naturally covers the case where
// Black must pass after the reply: White simply moves again.
//
// An illegal tap — it is not Black's turn, or cell is occupied, or the move
// flanks nothing — is a no-op: HumanMove returns the ORIGINAL snapshot with
// applied=false and a nil error, so a stray tap never crashes or advances the
// game. The rnd source is used only by the Random opponent (a nil rnd falls
// back to the math/rand global source); the AI opponent ignores it.
func (s Snapshot) HumanMove(cell revgame.Address, rnd *rand.Rand) (next Snapshot, applied bool, err error) {
	if s.Board.NextPlayer() != revgame.Black {
		return s, false, nil
	}

	board, err := s.Board.MakeMove(revgame.Black, cell)
	if err != nil {
		// Illegal / occupied tap: a no-op, not an error surfaced to the caller.
		return s, false, nil
	}

	for board.NextPlayer() == revgame.White {
		move, ok := opponentMove(s.Opponent, board, revgame.White, rnd)
		if !ok {
			break // defensive: NextPlayer==White implies a legal move exists
		}
		if board, err = board.MakeMove(revgame.White, move); err != nil {
			// The opponent picked an illegal move — should never happen.
			return Snapshot{Board: board, Opponent: s.Opponent}, true, err
		}
	}

	return Snapshot{Board: board, Opponent: s.Opponent}, true, nil
}

// opponentMove returns the move the given opponent plays for player on board,
// and ok=false if player has no legal move. AI routes to revgame.SimpleAI;
// Random (and any unrecognised mode, defensively) picks a uniformly random
// legal move using rnd.
func opponentMove(opp Opponent, board revgame.Board, player revgame.Disk, rnd *rand.Rand) (move revgame.Address, ok bool) {
	if opp == OpponentAI {
		if !board.HasValidMoves(player) {
			return revgame.Address{}, false
		}
		return revgame.SimpleAI{}.GetMove(board, player), true
	}
	moves := board.ValidMoves(player)
	if len(moves) == 0 {
		return revgame.Address{}, false
	}
	return moves[intn(rnd, len(moves))], true
}

// intn returns a uniform int in [0,n); a nil rnd uses the math/rand global.
func intn(rnd *rand.Rand, n int) int {
	if rnd == nil {
		return rand.Intn(n)
	}
	return rnd.Intn(n)
}
