package international

import (
	"strconv"
	"testing"
)

func TestNewFEN(t *testing.T) {
	board := NewFEN(GetFenFromBB(NewBitBoard())).getBoard()
	t.Log(strconv.FormatInt(board.WhitePieces, 16))
	t.Log(strconv.FormatInt(board.BlackPieces, 16))
	t.Log(strconv.FormatInt(board.Kings, 16))

}

func TestGetFenFromBB(t *testing.T) {
	t.Log(GetFenFromBB(NewBitBoard()))
}
