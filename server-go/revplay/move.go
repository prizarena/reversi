package revplay

import (
	"math/rand"

	"github.com/prizarena/reversi/server-go/revgame"
)

// ApplyHumanMove applies ONLY the human's move (their colour) at cell. It does
// NOT play the opponent — the host is expected to render the board and then call
// PlayOpponent so the human sees their move land before the reply arrives.
//
// An illegal tap — it is not the human's turn, or cell is occupied, or the move
// flanks nothing — is a no-op: ApplyHumanMove returns the ORIGINAL snapshot with
// applied=false and a nil error, so a stray tap never crashes or advances the
// game.
//
// If the human's move leaves the opponent with no reply, revgame resolves the
// pass automatically: IsHumanTurn stays true and the human simply plays again
// (a subsequent PlayOpponent would be a no-op).
func (s Snapshot) ApplyHumanMove(cell revgame.Address) (next Snapshot, applied bool, err error) {
	if !s.IsHumanTurn() {
		return s, false, nil
	}
	board, err := s.Board.MakeMove(s.Human, cell)
	if err != nil {
		// Illegal / occupied tap: a no-op, not an error surfaced to the caller.
		return s, false, nil
	}
	return Snapshot{Board: board, Opponent: s.Opponent, Human: s.Human}, true, nil
}

// PlayOpponent plays the opponent's reply move(s) while it is the opponent's
// turn, continuing automatically if the human must pass, and stopping when it
// becomes the human's turn or the game is over. moved reports whether >=1
// opponent move was made (false if it was not the opponent's turn to begin
// with). The rnd source is used only by the Random opponent (a nil rnd falls
// back to the math/rand global source); the AI opponent ignores it.
//
// Because revgame.Board.NextPlayer already resolves passes, the loop naturally
// covers the case where the human has no move after the opponent's reply: the
// opponent simply moves again, and PlayOpponent never stops on a turn the human
// cannot play.
func (s Snapshot) PlayOpponent(rnd *rand.Rand) (next Snapshot, moved bool, err error) {
	opp := s.OpponentColor()
	board := s.Board

	for !board.IsCompleted() && board.NextPlayer() == opp {
		move, ok := opponentMove(s.Opponent, board, opp, rnd)
		if !ok {
			break // defensive: NextPlayer==opp implies a legal move exists
		}
		newBoard, moveErr := board.MakeMove(opp, move)
		if moveErr != nil {
			// The opponent picked an illegal move — should never happen.
			return Snapshot{Board: board, Opponent: s.Opponent, Human: s.Human}, moved, moveErr
		}
		board = newBoard
		moved = true
	}

	return Snapshot{Board: board, Opponent: s.Opponent, Human: s.Human}, moved, nil
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
