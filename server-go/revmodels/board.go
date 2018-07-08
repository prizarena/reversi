package revmodels

import (
	"github.com/strongo/db"
	"github.com/prizarena/turn-based"
	"github.com/prizarena/reversi/server-go/revgame"
	"fmt"
)

const BoardKind = turnbased.BoardKind

type RevBoard struct {
	db.StringID
	*RevBoardEntity
	turnbased.BoardEntityBase
}

var _ db.EntityHolder = (*RevBoard)(nil)

func (RevBoard) Kind() string {
	return BoardKind
}

func (RevBoard) NewEntity() interface{} {
	return &RevBoardEntity{}
}

func (b RevBoard) Entity() interface{} {
	return b.RevBoardEntity
}

func (b *RevBoard) SetEntity(v interface{}) {
	if v == nil {
		b.RevBoardEntity = nil
	} else {
		b.RevBoardEntity = (v).(*RevBoardEntity)
	}
}

type RevBoardEntity struct {
	BoardTurns int    `datastore:"bt,noindex,omitempty"`
	BoardData  []byte `datastore:"bd,noindex,omitempty"`
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
