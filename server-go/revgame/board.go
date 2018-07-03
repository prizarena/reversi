package revgame

import (
	"math/bits"
	"bytes"
	"strings"
	"fmt"
	"github.com/pkg/errors"
	"github.com/prizarena/turn-based"
)

type Disk rune

var (
	Black     Disk = 'b'
	White     Disk = 'w'
	Empty     Disk = ' '
	Completed Disk = '!'
)

const (
	BoardSize = 8
	BoardCellsCount = BoardSize*BoardSize
)

func OtherPlayer(player Disk) Disk {
	switch player {
	case White:
		return Black
	case Black:
		return White
	}
	panic(fmt.Sprintf("unknown player: [%v]", player))
}

type Board struct {
	Blacks Disks
	Whites Disks
	Last   Disk
}

func (b Board) Turns() int {
	return bits.OnesCount64(uint64(b.Blacks)) + bits.OnesCount64(uint64(b.Whites)) - 4
}

var OthelloBoard = Board{
	Last:   White,
	Whites: (1 << (3*8 + 3)) | (1 << (4*8 + 4)),
	Blacks: (1 << (3*8 + 4)) | (1 << (4*8 + 3)),
}

func (b Board) flip(a address) (board Board) {
	if b.Whites&b.Blacks != 0 {
		panic("b.White&b.Black != 0")
	}

	board = b

	whiteIsPlaced := b.Whites.isPlaced(a)
	blackIsPlaced := b.Blacks.isPlaced(a)

	if whiteIsPlaced && blackIsPlaced {
		panic("whiteIsPlaced && blackIsPlaced")
	}

	doFlip := func(adding, removing Disks) (Disks, Disks) {
		var err error
		removing = removing.remove(a);
		if adding, err = adding.add(a); err != nil {
			panic(err)
		}
		return adding, removing
	}
	switch true {
	case whiteIsPlaced:
		board.Blacks, board.Whites = doFlip(b.Blacks, b.Whites)
	case blackIsPlaced:
		board.Whites, board.Blacks = doFlip(b.Whites, b.Blacks)
	default:
		panic("!whiteIsPlaced && !blackIsPlaced")
	}

	return
}

func (b Board) NextPlayer() Disk {
	switch b.Last {
	case White: // Blacks are making 1st move
		if b.hasValidMoves(Black) {
			return Black
		} else if b.hasValidMoves(White) {
			return White
		}
		return Completed
	case Black:
		if b.hasValidMoves(White) {
			return White
		} else if b.hasValidMoves(Black) {
			return Black
		}
		return Completed
	case Empty:
		turns := b.Turns()
		if turns < 10 {
			switch turns % 2 {
			case 0:
				return White
			case 1:
				return Black
			}
		}
		panic(fmt.Sprintf("can't detect last player as there are more then %v turns", turns))
	default:
		panic(fmt.Sprintf("unexpected b.Last: [%v]", b.Last))
	}
}

func (b Board) Rows(black, white, possibleMove, empty string) (rows [8][8]string) {
	// rows = make([][]string, 8)
	var validMoves []address
	if possibleMove != "" {
		player := b.NextPlayer()
		validMoves = b.getValidMoves(player)
	}

	for y := int8(0); y < 8; y++ {
		for x := int8(0); x < 8; x++ {
			bit := Disks(1 << (uint(y)*8 + uint(x)))
			if bit&b.Whites != 0 {
				rows[y][x] = white
			} else if bit&b.Blacks != 0 {
				rows[y][x] = black
			} else {
				for _, validMove := range validMoves {
					if validMove.X == x && validMove.Y == y {
						rows[y][x] = possibleMove
						break
					}
				}
				if rows[y][x] == "" {
					rows[y][x] = empty
				}
			}
		}
	}

	return
}

func (b Board) DrawBoard(black, white, possibleMove string, colSeparator, rowSeparator string) string {
	s := new(bytes.Buffer)
	s.WriteRune('\n')
	rows := b.Rows(black, white, possibleMove, " ")
	s.WriteRune(' ')
	for x := 0; x < 8; x++ {
		s.WriteRune('A' + rune(x))
	}
	s.WriteRune('\n')
	for y, row := range rows {
		s.WriteRune('1' + rune(y))
		s.WriteString(strings.Join(row[:], colSeparator))
		s.WriteString(rowSeparator)
	}
	return s.String()
}

type address struct {
	X, Y int8
}

func (a address) Index() int8 {
	return a.Y*BoardSize + a.X
}

var EmptyAddress = address{-127, -127}

func (a address) IsOnBoard() bool {
	return a.X >= 0 && a.X < BoardCellsCount && a.Y >= 0 && a.Y < BoardCellsCount
}

func (a address) move(d direction) address {
	a.X += d.x
	a.Y += d.y
	return a
}

func (a address) String() string {
	return fmt.Sprintf("{x: %v, y: %v}", a.X, a.Y)
}

func isOnBoard(a address) bool {
	return a.X >= 0 && a.X <= 7 && a.Y >= 0 && a.Y <= 7
}

func (b Board) disk(a address) Disk {
	if b.Whites.isPlaced(a) {
		return White
	} else if b.Blacks.isPlaced(a) {
		return Black
	}
	return Empty
}

func (b Board) UndoMove(a, prevMove address) (board Board) {
	board = b
	switch b.disk(a) {
	case Black:
		board.Blacks, board.Whites = board.undoMove(a, board.Blacks, board.Whites)
	case White:
		board.Whites, board.Blacks = board.undoMove(a, board.Whites, board.Blacks)
	}
	board.Last = board.disk(prevMove)
	if board.Last == Empty {
		board.Last = board.NextPlayer()
	}
	return
}

func (b Board) undoMove(disk address, removing, adding Disks) (Disks, Disks) {
	removing = removing.remove(disk)

	for _, direction := range directions {
		a := disk
		for {
			a = a.move(direction)
			next := a.move(direction)
			if !isOnBoard(next) || !removing.isPlaced(next) {
				break
			}
			removing = removing.remove(a)
			var err error
			if adding, err = adding.add(a); err != nil {
				panic(err)
			}
		}
	}
	return removing, adding
}

func (b Board) MakeMove(player Disk, a address) (board Board, err error) {
	var disksToFlip []address
	board = b

	if disksToFlip, err = b.getDisksToFlip(a, player); err != nil {
		return
	}
	if len(disksToFlip) == 0 {
		err = errors.WithMessage(ErrNotValidMove, a.String())
		return
	}

	switch player {
	case Black:
		board.Blacks, err = board.Blacks.add(a)
	case White:
		board.Whites, err = board.Whites.add(a)
	default:
		err = fmt.Errorf("unknown player: %v", player)
		return
	}
	if err != nil {
		return
	}
	for _, diskToFlip := range disksToFlip {
		board = board.flip(diskToFlip)
	}
	board.Last = player
	return
}

type direction struct {
	x, y int8
}

var directions = []direction{
	{0, 1}, {1, 1}, {1, 0}, {1, -1}, {0, -1}, {-1, -1}, {-1, 0}, {-1, 1},
}

func (b Board) getDisksToFlip(start address, player Disk) (disksToFlip []address, err error) {
	if !isOnBoard(start) {
		panic(fmt.Sprintf("address is outside of board: %v", start))
	}
	if b.Whites.isPlaced(start) || b.Blacks.isPlaced(start) {
		err = ErrAlreadyOccupied
		return
	}

	board := b

	switch player { // temporarily set the tile on the board.
	case Black:
		board.Blacks = board.Blacks.mustAdd(start)
	case White:
		board.Whites = board.Whites.mustAdd(start)
	default:
		panic(fmt.Sprintf("unknown plaeyr: %v", player))
	}

	otherDisk := OtherPlayer(player)

	disksToFlip = make([]address, 0, 5*3+4) // Theoretical maximum we can flip with 1 move

	for _, direction := range directions {
		a := start
		a.X += direction.x // first step in the direction
		a.Y += direction.y // first step in the direction
		if isOnBoard(a) && board.disk(a) == otherDisk { // There is a piece belonging to the other player next to our piece.
			a.X += direction.x
			a.Y += direction.y
			if !isOnBoard(a) {
				continue
			}

			for board.disk(a) == otherDisk {
				a.X += direction.x
				a.Y += direction.y
				if !isOnBoard(a) {
					break
				}
			}

			if !isOnBoard(a) {
				continue
			}

			if board.disk(a) == player {
				// There are disks to flip over. Go in the reverse direction until we reach the original space, noting all the tiles along the way.
				for {
					a.X -= direction.x
					a.Y -= direction.y
					if a == start {
						break
					}
					disksToFlip = append(disksToFlip, a)
				}
			}
		}
	}

	switch player { // temporarily set the tile on the board.
	case Black:
		board.Blacks = board.Blacks.remove(start)
	case White:
		board.Whites = board.Whites.remove(start)
	}
	return
}

func (b Board) freeCellsCount() int {
	return 64 - bits.OnesCount64(uint64(b.Blacks)) - bits.OnesCount64(uint64(b.Whites))
}

func (b Board) getValidMoves(player Disk) (validMoves []address) {
	//  Returns a list of [x,y] lists of valid moves for the given player on the given board.
	validMoves = make([]address, 0, b.freeCellsCount())
	for x := int8(0); x < 8; x++ {
		for y := int8(0); y < 8; y++ {
			disksToFlip, err := b.getDisksToFlip(address{x, y}, player)
			if err == nil && len(disksToFlip) > 0 {
				validMoves = append(validMoves, address{x, y})
			}
		}
	}
	return
}

func (b Board) hasValidMoves(player Disk) bool {
	for x := int8(0); x < 8; x++ {
		for y := int8(0); y < 8; y++ {
			disksToFlip, err := b.getDisksToFlip(address{x, y}, player) // TODO: no need slice of disksToFlip
			if err == nil && len(disksToFlip) > 0 {
				return true
			}
		}
	}
	return false
}

func (b Board) IsCompleted() bool {
	return b.NextPlayer() == Completed
}

func (b Board) Scores() (black, white int) {
	return bits.OnesCount64(uint64(b.Blacks)), bits.OnesCount64(uint64(b.Whites))
}

func (b Board) Score(player Disk) int {
	switch player {
	case Black:
		return bits.OnesCount64(uint64(b.Blacks))
	case White:
		return bits.OnesCount64(uint64(b.Whites))
	default:
		panic(fmt.Sprintf("unknown player: %v", player))
	}
}

func CellAddressToRevAddress(ca turnbased.CellAddress) address {
	if ca == "" {
		return EmptyAddress
	}
	x, y := ca.XY()
	return address{X: int8(x), Y: int8(y)}
}