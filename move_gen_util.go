package godraughts

import (
	"github.com/dangnguyendota/godraughts/bit"
	Direction "github.com/dangnguyendota/godraughts/direction"
	"github.com/dangnguyendota/godraughts/magic"
	Piece "github.com/dangnguyendota/godraughts/piece"
)
type MoveGenUtil struct {
	moves []int64
	captureCount int
}

func NewMoveGenUtil() *MoveGenUtil {
	return &MoveGenUtil{
		moves:        make([]int64, 0),
		captureCount: 0,
	}
}

func (m *MoveGenUtil) GenerateMoves(board *BitBoard) {
	m.reset()
	capture := canTake(board, board.colorToMove)
	if capture {
		m.addCapture(board, board.colorToMove)
	} else {
		m.addMove(board, board.colorToMove)
	}

	var tmp []int64
	for _, move := range m.moves {
		if isLegalMove(move) {
			tmp = append(tmp, move)
		}
	}
	m.moves = tmp
}

func (m *MoveGenUtil) HasMove(move int64) bool {
	if !isLegalMove(move) {
		return false
	}
	for i := 0; i < len(m.moves); i++ {
		if m.moves[i] == move {
			return true
		}
	}
	return false
}

func (m *MoveGenUtil) IsEndGame() bool {
	if len(m.moves) <= 0 {
		return true
	}

	for i := 0; i < len(m.moves); i++ {
		if !isLegalMove(m.moves[i]) {
			continue
		}
		return false
	}
	return true
}

func (m *MoveGenUtil) reset() {
	m.moves = make([]int64, 0)
	m.captureCount = 0
}

func (m *MoveGenUtil) addManMove(men int64, color int, occupied int64) {
	var from, to, move int64
	for men != 0 {
		from = men & -men
		move = magic.PawnNormalMove[color][bit.Index(from)] & bit.Reverse(occupied)
		for move != 0 {
			to = move & -move
			m.moves = append(m.moves, createMove(from, bit.Index(to), Piece.Pawn, color, 0))
			move = move & (move - 1)
		}
		men = men & (men - 1)
	}
}

func (m *MoveGenUtil) addKingMove(kings int64, color int, occupied int64) {
	var from, to, move int64
	for kings != 0 {
		from = kings & -kings
		move = magic.GetKingMove(bit.Index(from), occupied, false)
		for move != 0 {
			to = move & -move
			m.moves = append(m.moves, createMove(from, bit.Index(to), Piece.King, color, 0))
			move = move & (move - 1)
		}
		kings = kings & (kings - 1)
	}
}

func (m *MoveGenUtil) addCaptureMove(removed, end int64, piece, color int) {
	removeCount := bit.BitCount(removed) - 1
	if len(m.moves) > 0 {
		// nếu mà nước ăn quân này ăn nhiều hơn
		// các nước khác thì xóa hết các nước đã lưu đi
		// luật: phải đi nước ăn quân nhiều nhất
		if removeCount > m.captureCount {
			m.moves = make([]int64, 0)
		} else if removeCount < m.captureCount {
			return
		}
	}
	m.captureCount = removeCount
	m.moves = append(m.moves, createMove(removed, bit.Index(end), piece, color, 1))
}

func (m *MoveGenUtil) addMapCapture(enemy, empty, end, removed int64, color int) {
	i := bit.Index(end)
	captured := false
	for dir := magic.NE; dir <= magic.SW; dir++ {
		if (magic.Shift1Mask[dir][i] & enemy) != 0 && (magic.Shift2Mask[dir][i] & empty) != 0 {
			m.addMapCapture(enemy ^ magic.Shift1Mask[dir][i], empty ^ magic.Fill3Mask[dir][i],
				magic.Shift2Mask[dir][i], removed | magic.Shift1Mask[dir][i], color)
			captured = true
		}
	}

	if !captured && bit.MoreThanOne(removed) {
		m.addCaptureMove(removed, end, Piece.Pawn, color)
	}
}

func (m *MoveGenUtil) addKingCapture(enemy, empty, end, removed int64, color, comeFrom int) {
	i := bit.Index(end)
	occupied := bit.Reverse(empty)
	captured := false
	move := magic.GetKingMove(i, occupied, true)
	var capPiece, dirMove, pop int64

	if move != 0 {
		for dir := magic.NE; dir <= magic.SW; dir++ {
			if Direction.Dir[dir] + Direction.Dir[comeFrom] == 0 {
				continue
			}
			capPiece = magic.DirectionMask[dir][i] & move & enemy
			if capPiece != 0 {
				dirMove = (magic.DirectionMask[dir][i] & move) ^ capPiece
				captured = dirMove != 0
				for dirMove != 0 {
					pop = dirMove & -dirMove
					m.addKingCapture(enemy ^ capPiece, empty ^ (end | pop | capPiece), pop, removed | capPiece, color, dir)
					dirMove = dirMove & (dirMove - 1)
				}
			}
		}
	}

	if !captured && bit.MoreThanOne(removed) {
		m.addCaptureMove(removed, end, Piece.King, color)
	}
}

func (m *MoveGenUtil) addCapture(board *BitBoard, color int) {
	kings := board.King[color]
	men := board.Pawn[color]
	var pop int64
	Them := 1 - color
	for kings != 0 {
		pop = kings & -kings
		m.addKingCapture(board.Pieces[Them], board.Empty, pop, pop, color, magic.NDIR)
		kings = kings & (kings - 1)
	}

	for men != 0 {
		pop = men & -men
		m.addMapCapture(board.Pieces[Them], board.Empty, pop, pop, color)
		men = men & (men - 1)
	}

	var remove int64
	var end int
	for i := 0; i < len(m.moves); i++ {
		if !isLegalMove(m.moves[i]) {
			continue
		}
		remove = getRemoveMap(m.moves[i])
		end = getEndIndex(m.moves[i])
		for j := i + 1; j < len(m.moves); j++ {
			if !isLegalMove(m.moves[j])  {
				continue
			}
			if getRemoveMap(m.moves[j]) == remove && getEndIndex(m.moves[j]) == end {
				m.moves[j] = removeLegal(m.moves[j])
			}
		}
	}
}

func (m *MoveGenUtil) addMove(board *BitBoard, color int) {
	m.addKingMove(board.King[color], color, board.Occupied)
	m.addManMove(board.Pawn[color], color, board.Occupied)
}

func (m *MoveGenUtil) relativeRank(s, color int) int {
	if color == 0 {
		return s
	} else {
		return 9 - s
	}
}

func canTake(board *BitBoard, color int) bool {
	Them := 1 - color
	occupiedNe := bit.Shift(board.Pieces[color], Direction.NorthEast) & board.Pieces[Them]
	if (bit.Shift(occupiedNe, Direction.NorthEast) & board.Empty) != 0 {
		return true
	}

	occupiedNw := bit.Shift(board.Pieces[color], Direction.NorthWest) & board.Pieces[Them]
	if (bit.Shift(occupiedNw, Direction.NorthWest) & board.Empty) != 0 {
		return true
	}

	occupiedSe := bit.Shift(board.Pieces[color], Direction.SouthEast) & board.Pieces[Them]
	if (bit.Shift(occupiedSe, Direction.SouthEast) & board.Empty) != 0 {
		return true
	}

	occupiedSw := bit.Shift(board.Pieces[color], Direction.SouthWest) & board.Pieces[Them]
	if (bit.Shift(occupiedSw, Direction.SouthWest) & board.Empty) != 0 {
		return true
	}

	var kingAttack int64
	var index int
	for i := board.King[color]; i != 0; i = i & (i - 1) {
		index = bit.Index(i & -i)
		kingAttack = magic.GetKingMove(index, board.Occupied, true)
		for dir := magic.NE; dir <= magic.SW; dir++ {
			if (kingAttack & magic.DirectionMask[dir][index] & board.Pieces[Them]) != 0 && bit.MoreThanOne(kingAttack & magic.DirectionMask[dir][index]) {
				return true
			}
		}
	}

	return false
}