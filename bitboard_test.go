package godraughts

import (
	"testing"
)

func TestNewBitBoard(t *testing.T) {
	board := NewBitBoard()
	util := NewMoveGenUtil()
	util.GenerateMoves(board)
	board.doMove(util.moves[0])
	util.GenerateMoves(board)
	board.doMove(util.moves[0])
	util.GenerateMoves(board)
	board.doMove(util.moves[0])
	util.GenerateMoves(board)
	board.doMove(util.moves[0])
	t.Log(board)
}
