package revmodels

import (
	"fmt"
	"github.com/prizarena/reversi/server-go/revgame"
	"github.com/strongo/dalgo/record"
)

const BoardKind = turnbased.BoardKind

type RevBoard struct {
	record.WithID[string]
	*RevBoardEntity
}

//var _ db.EntityHolder = (*RevBoard)(nil)

type RevBoardEntity struct {
	turnbased.BoardEntityBase
	BoardTurns   int    `datastore:"bt,noindex,omitempty"`
	BoardData    []byte `datastore:"bd,noindex,omitempty"`
	BoardHistory string `datastore:"bh,noindex,omitempty"`
}

func (entity *RevBoardEntity) SetBoardState(b revgame.Board) {
	entity.BoardTurns = b.Turns()
	entity.BoardData = b.ToBytes()
}

func (entity *RevBoardEntity) GetBoard() (board revgame.Board, err error) {
	switch len(entity.BoardData) {
	case 0:
		board = revgame.OthelloBoard
	case 17:
		board = revgame.NewBoardFromBytes(entity.BoardData)
	default:
		err = fmt.Errorf("len(*RevBoardEntity.BoardData) expected to be 0 or 17, got: %v", len(entity.BoardData))
	}
	return
}
