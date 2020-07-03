package magic

import (
	"../bit"
	"../color"
	dir "../direction"
	"errors"
)

var (
	MASK [64]int64
	row = []int{
		0, 0, 0, 0, 0,
		1, 1, 1, 1, 1,
		2, 2, 2, 2, 2,
		3, 3, 3, 3, 3,
		4, 4, 4, 4, 4,
		5, 5, 5, 5, 5,
		6, 6, 6, 6, 6,
		7, 7, 7, 7, 7,
		8, 8, 8, 8, 8,
		9, 9, 9, 9, 9,
	}
	column = []int{
		1, 3, 5, 7, 9,
		0, 2, 4, 6, 8,
		1, 3, 5, 7, 9,
		0, 2, 4, 6, 8,
		1, 3, 5, 7, 9,
		0, 2, 4, 6, 8,
		1, 3, 5, 7, 9,
		0, 2, 4, 6, 8,
		1, 3, 5, 7, 9,
		0, 2, 4, 6, 8,
	}
	Index = [][]int{
		{-1, 0, -1, 1, -1, 2, -1, 3, -1, 4},
		{5, -1, 6, -1, 7, -1, 8, -1, 9, -1},
		{-1, 10, -1, 11, -1, 12, -1, 13, -1, 14},
		{15, -1, 16, -1, 17, -1, 18, -1, 19, -1},
		{-1, 20, -1, 21, -1, 22, -1, 23, -1, 24},
		{25, -1, 26, -1, 27, -1, 28, -1, 29, -1},
		{-1, 30, -1, 31, -1, 32, -1, 33, -1, 34},
		{35, -1, 36, -1, 37, -1, 38, -1, 39, -1},
		{-1, 40, -1, 41, -1, 42, -1, 43, -1, 44},
		{45, -1, 46, -1, 47, -1, 48, -1, 49, -1},
	}
	PawnNormalMove           [2][50]int64
	PawnOccupiedSq           [50]int64
	BesideSquares            [2][50]int64
	WallSquaresBeside        [50]int
	ForwardStormSquares      [2][50]int64
	ForwardStormSquaresCount [2][50]int
	ForwardRankThreeFiles    [2][50]int64
	DiagonalMask             [2][50]int64
	DirectionMask            [4][50]int64
	KingMoveBB               [50][2][1024]int64
	KingAttackBB             [50][2][1024]int64
	KingMoveMask             [50]int64
	Shift1Mask               [4][50]int64
	Shift2Mask               [4][50]int64
	Fill3Mask                [4][50]int64
	BetweenBB                [50][50]int64
	AfterBB                  [2][50]int64

	NE                      = 0
	NW                      = 1
	SE                      = 2
	SW                      = 3
	NDIR                    = 4
	LegalMask         int64 = 0x3ffffffffffff
	RANK1BB           int64 = 0x3e00000000000
	RANK2BB                 = RANK1BB >> 5
	RANK3BB                 = RANK1BB >> 10
	RANK4BB                 = RANK1BB >> 15
	RANK5BB                 = RANK1BB >> 20
	RANK6BB                 = RANK1BB >> 25
	RANK7BB                 = RANK1BB >> 30
	RANK8BB                 = RANK1BB >> 35
	RANK9BB                 = RANK1BB >> 40
	RANK10BB                = RANK1BB >> 45
	FILE1BB           int64 = 0x200802008020
	FILE2BB           int64 = 0x10040100401
	FILE3BB                 = FILE1BB << 1
	FILE4BB                 = FILE2BB << 1
	FILE5BB                 = FILE1BB << 2
	FILE6BB                 = FILE2BB << 2
	FILE7BB                 = FILE1BB << 3
	FILE8BB                 = FILE2BB << 3
	FILE9BB                 = FILE1BB << 4
	FILE10BB                = FILE2BB << 4
	RANKS                   = []int64{RANK1BB, RANK2BB, RANK3BB, RANK4BB, RANK5BB, RANK6BB, RANK7BB, RANK8BB, RANK9BB, RANK10BB}
	FILES                   = []int64{FILE1BB, FILE2BB, FILE3BB, FILE4BB, FILE5BB, FILE6BB, FILE7BB, FILE8BB, FILE9BB, FILE10BB}
	THREATING_SQUARES       = []int64{0x100c030, 0x300c02000000}
)

func GetKingMove(i int, occupied int64, capture bool) int64 {
	h := hash(i, occupied)
	if capture {
		return KingAttackBB[i][0][h[0]] | KingAttackBB[i][1][h[1]]
	} else {
		return KingMoveBB[i][0][h[0]] | KingMoveBB[i][1][h[1]]
	}
}

func init() {
	for i := 0; i < 64; i++ {
		MASK[i] = 1 << i
	}
	generateMask()
	generateMenNormalMove()
	generateKingMove()
}

// generator

func generateMenNormalMove() {
	for i := 0; i < 50; i++ {
		PawnNormalMove[0][i] |= bit.Shift(MASK[i], dir.NorthEast) | bit.Shift(MASK[i], dir.NorthWest)
		PawnNormalMove[1][i] |= bit.Shift(MASK[i], dir.SouthEast) | bit.Shift(MASK[i], dir.SouthWest)
		PawnOccupiedSq[i] = PawnNormalMove[0][i] | PawnNormalMove[1][i]
		BesideSquares[0][i] = bit.Shift(MASK[i], dir.NorthEast) | bit.Shift(MASK[i], dir.SouthWest)
		BesideSquares[1][i] = bit.Shift(MASK[i], dir.NorthWest) | bit.Shift(MASK[i], dir.SouthEast)
		WallSquaresBeside[i] = 4 - bit.BitCount(PawnOccupiedSq[i])

		tmp := DirectionMask[NW][i] | MASK[i]
		for tmp != 0 {
			index := bit.Index(tmp & -tmp)
			ForwardStormSquares[color.White][i] |= DirectionMask[NE][index] | MASK[index]
			tmp = tmp & (tmp - 1)
		}
		tmp = DirectionMask[NE][i] | MASK[i]
		for tmp != 0 {
			index := bit.Index(tmp & -tmp)
			ForwardStormSquares[color.White][i] |= DirectionMask[NW][index] | MASK[index]
			tmp = tmp & (tmp - 1)
		}
		ForwardStormSquares[color.White][i] ^= MASK[i]
		ForwardStormSquaresCount[color.White][i] = bit.BitCount(ForwardStormSquares[color.White][i])
		tmp = DirectionMask[SW][i] | MASK[i]
		for tmp != 0 {
			index := bit.Index(tmp & -tmp)
			ForwardStormSquares[color.Black][i] |= DirectionMask[SE][index] | MASK[index]
			tmp = tmp & (tmp - 1)
		}

		tmp = DirectionMask[SE][i] | MASK[i]
		for tmp != 0 {
			index := bit.Index(tmp & -tmp)
			ForwardStormSquares[color.Black][i] |= DirectionMask[SW][index] | MASK[index]
			tmp = tmp & (tmp - 1)
		}

		ForwardStormSquares[color.Black][i] ^= MASK[i]
		ForwardStormSquaresCount[color.Black][i] = bit.BitCount(ForwardStormSquares[color.Black][i])
	}
}

func generateKingMove() {
	var h []int
	for i := 0; i < 50; i++ {
		blockers := getBlockersList(KingMoveMask[i])
		for _, blocker := range blockers {
			h = hash(i, blocker)
			KingMoveBB[i][0][h[0]] = getBBFromBlocker(i, blocker, dir.NorthEast) | getBBFromBlocker(i, blocker, dir.SouthWest)
			KingMoveBB[i][1][h[1]] = getBBFromBlocker(i, blocker, dir.NorthWest) | getBBFromBlocker(i, blocker, dir.SouthEast)
			KingAttackBB[i][0][h[0]] = getCaptureBBFromBlocker(i, blocker, dir.NorthEast) | getCaptureBBFromBlocker(i, blocker, dir.SouthWest)
			KingAttackBB[i][1][h[1]] = getCaptureBBFromBlocker(i, blocker, dir.NorthWest) | getCaptureBBFromBlocker(i, blocker, dir.SouthEast)
		}
	}
}

func generateMask() {
	var mask int64
	for i := 0; i < 50; i++ {
		mask = 0x1 << i
		DirectionMask[NE][i] = getBBFromDirection(i, dir.NorthEast)
		DirectionMask[NW][i] = getBBFromDirection(i, dir.NorthWest)
		DirectionMask[SE][i] = getBBFromDirection(i, dir.SouthEast)
		DirectionMask[SW][i] = getBBFromDirection(i, dir.SouthWest)
		DiagonalMask[0][i] = DirectionMask[NE][i] | DirectionMask[SW][i]
		DiagonalMask[1][i] = DirectionMask[NW][i] | DirectionMask[SE][i]
		KingMoveMask[i] = DiagonalMask[0][i] | DiagonalMask[1][i]

		Shift1Mask[NE][i] = bit.Shift(mask, dir.NorthEast)
		Shift1Mask[NW][i] = bit.Shift(mask, dir.NorthWest)
		Shift1Mask[SE][i] = bit.Shift(mask, dir.SouthEast)
		Shift1Mask[SW][i] = bit.Shift(mask, dir.SouthWest)

		Shift2Mask[NE][i] = bit.Shift(Shift1Mask[NE][i], dir.NorthEast)
		Shift2Mask[NW][i] = bit.Shift(Shift1Mask[NW][i], dir.NorthWest)
		Shift2Mask[SE][i] = bit.Shift(Shift1Mask[SE][i], dir.SouthEast)
		Shift2Mask[SW][i] = bit.Shift(Shift1Mask[SW][i], dir.SouthWest)

		Fill3Mask[NE][i] = mask | Shift1Mask[NE][i] | Shift2Mask[NE][i]
		Fill3Mask[NW][i] = mask | Shift1Mask[NW][i] | Shift2Mask[NW][i]
		Fill3Mask[SE][i] = mask | Shift1Mask[SE][i] | Shift2Mask[SE][i]
		Fill3Mask[SW][i] = mask | Shift1Mask[SW][i] | Shift2Mask[SW][i]

		count := 1
		tmp := bit.Shift(mask, dir.North) | bit.Shift(mask, dir.NorthWest) | bit.Shift(mask, dir.NorthEast)
		for tmp != 0 {
			ForwardRankThreeFiles[color.White][i] |= tmp
			tmp = bit.Shift(tmp, dir.North)
			count++
			if count == 4 {
				break
			}
		}

		count = 1
		tmp = bit.Shift(0x1<<i, dir.South) | bit.Shift(0x1<<i, dir.SouthWest) | bit.Shift(0x1<<i, dir.SouthEast)
		for tmp != 0 {
			ForwardRankThreeFiles[color.Black][i] |= tmp
			tmp = bit.Shift(tmp, dir.South)
			count++
			if count == 4 {
				break
			}
		}

		AfterBB[color.White][i] = bit.Shift(mask, dir.SouthEast) | bit.Shift(mask, dir.SouthWest) | bit.Shift(mask, dir.South)
		AfterBB[color.Black][i] = bit.Shift(mask, dir.NorthEast) | bit.Shift(mask, dir.NorthWest) | bit.Shift(mask, dir.North)
	}

	for i := 0; i < 50; i++ {
		for j := 0; j < 50; j++ {
			if (KingMoveMask[i] & MASK[j]) == 0 {
				continue
			}
			if (DirectionMask[NE][i] & MASK[j]) != 0 {
				BetweenBB[i][j] = getBBFromBlocker(i, MASK[j], dir.NorthEast) ^ MASK[j]
			} else if (DirectionMask[NW][i] & MASK[j]) != 0 {
				BetweenBB[i][j] = getBBFromBlocker(i, MASK[j], dir.NorthWest) ^ MASK[j]
			} else if (DirectionMask[SE][i] & MASK[j]) != 0 {
				BetweenBB[i][j] = getBBFromBlocker(i, MASK[j], dir.SouthEast) ^ MASK[j]
			} else if (DirectionMask[SW][i] & MASK[j]) != 0 {
				BetweenBB[i][j] = getBBFromBlocker(i, MASK[j], dir.SouthWest) ^ MASK[j]
			} else {
				panic(errors.New("generate between bit board error"))
			}
		}
	}
}

// helper

func hashBB(bb int64) int {
	return int((((bb&0x1f07c1f07c1f)*0x210842108421)&0x1f0000000000)>>40 | (((bb&0x3e0f83e0f83e0)*0x210842108421)&0x3e00000000000)>>40)
}

func northeast(position int) int {
	if row[position] > 0 && column[position] < 9 {
		return Index[row[position]-1][column[position]+1]
	} else {
		return -1
	}
}

func northwest(position int) int {
	if row[position] > 0 && column[position] > 0 {
		return Index[row[position]-1][column[position]-1]
	} else {
		return -1
	}
}

func southeast(position int) int {
	if row[position] < 9 && column[position] < 9 {
		return Index[row[position]+1][column[position]+1]
	} else {
		return -1
	}
}

func southwest(position int) int {
	if row[position] < 9 && column[position] > 0 {
		return Index[row[position]+1][column[position]-1]
	} else {
		return -1
	}
}

func hash(index int, occupied int64) []int {
	return []int{
		hashBB(DiagonalMask[0][index] & occupied),
		hashBB(DiagonalMask[1][index] & occupied),
	}
}

func getBBFromArray(arrayList *[]int) int64 {
	var result int64
	for _, i := range *arrayList {
		result |= 0x1 << i
	}
	return result
}

func getBBFromBlocker(pos int, blocker int64, direction int) int64 {
	var result int64
	if direction == dir.NorthEast {
		for northeast(pos) != -1 && (0x1<<northeast(pos)&blocker) == 0 {
			result |= 0x1 << northeast(pos)
			pos = northeast(pos)
		}
	} else if direction == dir.NorthWest {
		for northwest(pos) != -1 && (0x1<<northwest(pos)&blocker) == 0 {
			result |= 0x1 << northwest(pos)
			pos = northwest(pos)
		}
	} else if direction == dir.SouthEast {
		for southeast(pos) != -1 && (0x1<<southeast(pos)&blocker) == 0 {
			result |= 0x1 << southeast(pos)
			pos = southeast(pos)
		}
	} else if direction == dir.SouthWest {
		for southwest(pos) != -1 && (0x1<<southwest(pos)&blocker) == 0 {
			result |= 0x1 << southwest(pos)
			pos = southwest(pos)
		}
	}

	return result
}

func getBBFromDirection(pos, direction int) int64 {
	return getBBFromBlocker(pos, 0, direction)
}

func getSubBB(start, max int, currentIndex, indexes *[]int, out *[]int64) {
	if start >= len(*indexes) && max != 0 {
		panic(errors.New("get_sub_bb exception!"))
	}

	if max == 0 {
		*out = append(*out, getBBFromArray(currentIndex))
	} else {
		for i := start; i < len(*indexes)-max+1; i++ {
			tmp := make([]int, len(*currentIndex))
			copy(tmp, *currentIndex)
			tmp = append(tmp, (*indexes)[i])
			getSubBB(i+1, max-1, &tmp, indexes, out)
		}
	}
}

func getBlockersList(bb int64) []int64 {
	indexes := bit.GetIndex(bb)
	blockers := make([]int64, 0)
	for i := 0; i <= len(indexes); i++ {
		tmp := make([]int, 0)
		getSubBB(0, i, &tmp, &indexes, &blockers)
	}
	return blockers
}

func getCaptureBBFromBlocker(pos int, blocker int64, direction int) int64 {
	var result int64
	if direction == dir.NorthEast {
		for northeast(pos) != -1 && (0x1<<northeast(pos)&blocker) == 0 {
			pos = northeast(pos)
		}
		if northeast(pos) != -1 {
			result |= 0x1 << northeast(pos)
			result |= getBBFromBlocker(northeast(pos), blocker, direction)
		}
	} else if direction == dir.NorthWest {
		for northwest(pos) != -1 && (0x1<<northwest(pos)&blocker) == 0 {
			pos = northwest(pos)
		}
		if northwest(pos) != -1 {
			result |= 0x1 << northwest(pos)
			result |= getBBFromBlocker(northwest(pos), blocker, direction)
		}
	} else if direction == dir.SouthEast {
		for southeast(pos) != -1 && (0x1<<southeast(pos)&blocker) == 0 {
			pos = southeast(pos)
		}
		if southeast(pos) != -1 {
			result |= 0x1 << southeast(pos)
			result |= getBBFromBlocker(southeast(pos), blocker, direction)
		}
	} else if direction == dir.SouthWest {
		for southwest(pos) != -1 && (0x1<<southwest(pos)&blocker) == 0 {
			pos = southwest(pos)
		}
		if southwest(pos) != -1 {
			result |= 0x1 << southwest(pos)
			result |= getBBFromBlocker(southwest(pos), blocker, direction)
		}
	}
	return result
}
