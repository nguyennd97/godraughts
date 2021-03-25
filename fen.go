package godraughts

import (
	"./color"
	"log"
	"strconv"
	"strings"
)

type FEN struct {
	fen string
}

func NewFEN(fen string) *FEN {
	return &FEN{fen:fen}
}

func (f *FEN) getBoard() *BitBoard {
	var Kings, WhitePieces, BlackPieces int64
	tmp := strings.Split(strings.ReplaceAll(f.fen, " ", ""), ":")
	white := strings.Split(strings.ReplaceAll(tmp[1][1:], " ", ""), ",")
	black := strings.Split(strings.ReplaceAll(tmp[2][1:], " ", ""), ",")
	for _, i := range white {
		if i[0] == 'K' {
			Kings |= 0x1 << (getInt(i[1:]) - 1)
			WhitePieces |= 0x1 << (getInt(i[1:]) - 1)
		} else {
			WhitePieces |= 0x1 << (getInt(i) - 1)
		}
	}

	for _, i := range black {
		if i[0] == 'K' {
			Kings |= 0x1 << (getInt(i[1:]) - 1)
			BlackPieces |= 0x1 << (getInt(i[1:]) - 1)
		} else {
			BlackPieces |= 0x1 << (getInt(i) - 1)
		}
	}

	colorToMove := color.White
	switch strings.ReplaceAll(f.fen, " ", "")[0] {
	case 'W':
		colorToMove = color.White
		break
	case 'B':
		colorToMove = color.Black
		break
	}
	return NewCustomBitBoard(WhitePieces, BlackPieces, Kings, colorToMove)
}

func GetFenFromBB(board *BitBoard) string {
	var fen string
	if board.colorToMove == 0 {
		fen = "W"
	} else {
		fen = "B"
	}

	w := ":W"
	b := ":B"
	for i := 0; i < 50; i++ {
		mask := int64(0x1 << i)
		if (mask & board.WhitePieces) != 0 {
			if (mask & board.Kings) != 0 {
				w += "K"
			}
			w += strconv.FormatInt(int64(i + 1), 10)
			w += ","
		} else if (mask & board.BlackPieces) != 0 {
			if (mask & board.Kings) != 0 {
				b += "K"
			}
			b += strconv.FormatInt(int64(i + 1), 10)
			b += ","
		}
	}

	if w[len(w) - 1] == ',' {
		w = w[:len(w) - 1]
	}
	if b[len(b) - 1] == ',' {
		b = b[:len(b) - 1]
	}

	fen += w
	fen += b
	return fen
}

func getInt(s string) int {
	i, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		log.Println(err)
	}
	return int(i)
}