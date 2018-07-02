package revgame

import (
	"math/rand"
	"time"
)

func isInCorner(a address) bool {
	return (a.X == 0 || a.X == 7) && (a.Y == 0 || a.Y == 7)
}

type SimpleAI struct {
}

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

func (SimpleAI) GetMove(board Board, player Disk) (move address) {
	validMoves := board.getValidMoves(player)
	// if len(validMoves) == 0 {
	// 	return
	// }

	bestMove := func(moves []address) address {
		if len(moves) == 1 {
			return moves[0]
		}
		var bestScore int
		bestMoves := make([]address, 0, len(moves))
		for _, m := range moves {
			b, err := board.MakeMove(player, m.X, m.Y)
			if err != nil {
				panic(err)
			}
			score := b.Score(player)
			if score == bestScore {
				bestMoves = append(bestMoves, m)
			} else if score > bestScore {
				bestScore = score
				bestMoves = bestMoves[0:1]
				bestMoves[0] = m
			}
		}
		if len(bestMoves) == 1 {
			return bestMoves[0]
		}
		return bestMoves[rnd.Intn(len(bestMoves))]
	}

	{
		cornerMoves := make([]address, 0, 4)
		for _, m := range validMoves {
			if isInCorner(m) {
				cornerMoves = append(cornerMoves, m)
			}
		}
		if len(cornerMoves) > 0 {
			return bestMove(cornerMoves)
		}
	}

	return bestMove(validMoves)
}
