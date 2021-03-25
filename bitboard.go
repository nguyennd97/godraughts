package godraughts

import (
	"./bit"
	c "./color"
	"./magic"
	p "./piece"
)

type BitBoard struct {
	WhitePieces int64
	BlackPieces int64
	Kings       int64
	Occupied    int64
	Empty       int64
	King        [2]int64
	Pawn        [2]int64
	Pieces      [2]int64
	colorToMove int
}

func NewBitBoard() *BitBoard {
	b := &BitBoard{
		WhitePieces: 0x3ffffc0000000,
		BlackPieces: 0xfffff,
	}

	b.setupValues()
	return b
}

func NewCustomBitBoard(WP, BP, K int64, colorToMove int) *BitBoard {
	b := &BitBoard{
		WhitePieces: WP,
		BlackPieces: BP,
		Kings:       K,
		colorToMove: colorToMove,
	}

	b.setupValues()
	return b
}

func (b *BitBoard) setupValues() {
	b.King[c.White] = b.WhitePieces & b.Kings
	b.King[c.Black] = b.BlackPieces & b.Kings
	b.Pawn[c.White] = b.WhitePieces ^ b.King[c.White]
	b.Pawn[c.Black] = b.BlackPieces ^ b.King[c.Black]
	b.Occupied = b.WhitePieces | b.BlackPieces
	b.Pieces[c.White] = b.WhitePieces
	b.Pieces[c.Black] = b.BlackPieces
	b.Empty = bit.Reverse(b.Occupied)
}

func (b *BitBoard) doMove(move int64) {
	removed := getRemoveMap(move)
	end := getEndIndex(move)
	nonRemoved := bit.Reverse(removed)
	piece := getPiece(move)
	color := getColor(move)

	b.WhitePieces &= nonRemoved
	b.BlackPieces &= nonRemoved
	b.Kings &= nonRemoved
	b.colorToMove = 1 - b.colorToMove

	if piece == p.Pawn {
		if color == c.White {
			if (magic.MASK[end] & 0x1f) != 0 {
				b.Kings |= magic.MASK[end]
			}
		} else {
			if (magic.MASK[end] & 0x3e00000000000) != 0 {
				b.Kings |= magic.MASK[end]
			}
		}
	} else if piece == p.King {
		b.Kings |= magic.MASK[end]
	}

	if color == c.White {
		b.WhitePieces |= magic.MASK[end]
	} else {
		b.BlackPieces |= magic.MASK[end]
	}

	b.setupValues()
}

func (b *BitBoard) isFirstForBlack() bool {
	return b.BlackPieces == 0xfffff && b.Kings == 0
}

func (b *BitBoard) isFirstForWhite() bool {
	return b.WhitePieces == 0x3ffffc0000000 && b.Kings == 0
}

func (b *BitBoard) isFirst() bool {
	return b.isFirstForWhite() || b.isFirstForBlack()
}

func (b *BitBoard) String() string {
	wm := b.WhitePieces & b.Kings ^ b.WhitePieces
	wk := b.WhitePieces & b.Kings
	bm := b.BlackPieces & b.Kings ^ b.BlackPieces
	bk := b.BlackPieces & b.Kings

	board := "\n +---+---+---+---+---+---+---+---+---+---+\n |"
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			if (i + j) % 2 == 0 {
				board += "   |"
			} else if (1 << magic.Index[i][j] & wm) != 0 {
				board += " M |"
			} else if (1 << magic.Index[i][j] & wk) != 0 {
				board += " K |"
			} else if (1 << magic.Index[i][j] & bm) != 0 {
				board += " m |"
			} else if (1 << magic.Index[i][j] & bk) != 0 {
				board += " k |"
			} else {
				board += "///|"
			}
		}

		if i == 9 {
			board += "\n +---+---+---+---+---+---+---+---+---+---+\n"
		} else {
			board += "\n +---+---+---+---+---+---+---+---+---+---+\n |"
		}
	}

	return board
}