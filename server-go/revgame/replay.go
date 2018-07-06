package revgame

import (
	"fmt"
	"bytes"
	"github.com/pkg/errors"
)

func Replay(board Board, transcript Transcript, backSteps int) (Board) {
	var err error
	for len(transcript) > backSteps {
		var move Move
		move, transcript = transcript.NextMove()
		if board, err = board.MakeMove(board.NextPlayer(), move.Address()); err != nil {
			panic(err)
		}
	}
	return board
}

// func Rewind(board Board, transcript Transcript, backSteps int) (currentBoard Board, nextMove Address) {
// 	lastMoves := transcript
// 	//stepsToRollback := backSteps - replay // replay is negative, so we need '-' to sum.
// 	currentBoard = board
// 	nextMove = EmptyAddress
// 	for backSteps > 0 && (len(lastMoves) > 0 || board.Turns() < 5) {
// 		backSteps--
// 		var lastMove Move
// 		lastMove, lastMoves = lastMoves.Pop()
// 		a := lastMove.Address()
// 		var prevMove Address
// 		if len(lastMoves) == 0 {
// 			prevMove = EmptyAddress
// 		} else {
// 			prevMove = lastMoves.LastMove().Address()
// 		}
// 		currentBoard = currentBoard.UndoMove(a, prevMove)
// 		nextMove = a
// 	}
// 	return
// }

func AddMove(transcript Transcript, backSteps int, a Address) (Transcript, int) {
	if backSteps > 0 {
		transcript = transcript[:len(transcript)-backSteps]
		backSteps = 0
	}
	b := byte(a.Index())
	for _, v := range transcript {
		if v == b {
			panic(fmt.Sprintf("Duplicate move: %v, transcript: %v", a, transcript.String()))
		}
	}
	return append(transcript, b), backSteps
}

func VerifyBoardTranscript(b Board, t Transcript) error {
	s := new(bytes.Buffer)
	for _, m := range t {
		a := Move(m).Address()
		if b.Blacks.isPlaced(a) || b.Whites.isPlaced(a) {
			s.WriteString(string(a.ToCellAddress()))
		}
	}
	if s.Len() > 0 {
		return errors.New("transcript references occupied cells: " + s.String())
	}
	return nil
}

