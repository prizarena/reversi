package revgame

// ValidMoves returns every legal move for the given player on this board.
// Exported wrapper over getValidMoves so callers in other packages (e.g. a bot
// play layer's Random opponent) get legal moves from the engine rather than
// re-implementing move generation.
func (b Board) ValidMoves(player Disk) []Address {
	return b.getValidMoves(player)
}

// HasValidMoves reports whether the given player has at least one legal move.
// Exported wrapper over hasValidMoves.
func (b Board) HasValidMoves(player Disk) bool {
	return b.hasValidMoves(player)
}
