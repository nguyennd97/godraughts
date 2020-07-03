package international

import (
	Color "./color"
	"./magic"
)
const (
	endShift           int   = 50
	pieceShift         int   = 56
	colorShift         int   = 57
	captureFlagShift   int   = 58
	promotionFlagShift int   = 59
	legalMoveFlagShift int   = 60
	Mask1Bit           int64 = 0x1
	Mask6Bit           int64 = 0x3F
)

func getRemoveMap(move int64) int64 {
	return move & magic.LegalMask
}

func getEndIndex(move int64) int {
	return int(move >> endShift & Mask6Bit)
}

func getPiece(move int64) int {
	return int(move >> pieceShift & Mask1Bit)
}

func getColor(move int64) int {
	return int(move >> colorShift & Mask1Bit)
}

func isCapture(move int64) bool {
	return (move >> captureFlagShift & Mask1Bit) == 1
}

func isPromotion(move int64) bool {
	return (move >> promotionFlagShift & Mask1Bit) == 1
}

func isLegalMove(move int64) bool {
	return (move >> legalMoveFlagShift & Mask1Bit) == 1
}

func createMove(removeMap int64, endPos, piece, color, capture int) int64 {
	promotion := 0
	if color == Color.White {
		if (magic.MASK[endPos] & 0x1f) != 0 {
			promotion = 1
		}
	} else if color == Color.Black {
		if (magic.MASK[endPos] & 0x3e00000000000) != 0 {
			promotion = 1
		}
	}

	return removeMap | int64(endPos) << endShift | int64(piece) << pieceShift | int64(color) << colorShift | int64(capture) << captureFlagShift | int64(promotion) << promotionFlagShift | magic.MASK[legalMoveFlagShift]
}

func removeLegal(move int64) int64 {
	return move ^ magic.MASK[legalMoveFlagShift]
}