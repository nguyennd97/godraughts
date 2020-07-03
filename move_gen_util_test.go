package international

import (
	"fmt"
	"testing"
)

func TestMoveGenUtil_GenerateMoves(t *testing.T) {
	board := NewBitBoard()
	fmt.Println(board)
	moveGenUtil := NewMoveGenUtil()
	moveGenUtil.GenerateMoves(board)
	board.doMove(moveGenUtil.moves[0])
	fmt.Println(board)
	moveGenUtil.GenerateMoves(board)
	board.doMove(moveGenUtil.moves[0])
	fmt.Println(board)
	moveGenUtil.GenerateMoves(board)
	t.Log(len(moveGenUtil.moves))
}
