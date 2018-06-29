package revmodels

import (
	"github.com/strongo/db"
	"github.com/prizarena/turn-based"
	"github.com/prizarena/reversi/server-go/revgame"
)

const BoardKind = turnbased.BoardKind

type Board struct {
	db.StringID
	*BoardEntity
	turnbased.BoardEntityBase
}

var _ db.EntityHolder = (*Board)(nil)

func (Board) Kind() string {
	return BoardKind
}

func (Board) NewEntity() interface{} {
	return &BoardEntity{}
}

func (b Board) Entity() interface{} {
	return b.BoardEntity
}

func (b *Board) SetEntity(v interface{})  {
	if v == nil {
		b.BoardEntity = nil
	} else {
		b.BoardEntity = (v).(*BoardEntity)
	}
}

type BoardEntity struct {
	BoardBlacks int64 `datastore:"bb,noindex,omitempty"` // users[0]
	BoardWhites int64 `datastore:"bw,noindex,omitempty"` // users[1]
}

func (entity *BoardEntity) SetBoardState(b revgame.Board) {
	entity.BoardBlacks = int64(b.Blacks)
	entity.BoardWhites = int64(b.Whites)
}