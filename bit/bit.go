package bit

import dir "github.com/dangnguyendota/godraughts/direction"

var bitIndex = []int{
	-1, 0, 1, 39, 2, 15, 40, 23, 3, 12, 16,
	-1, 41, 19, 24, -1, 4, -1, 13, 10, 17,
	-1, -1, 28, 42, 30, 20, -1, 25, 44, -1,
	47, 5, 32, -1, 38, 14, 22, 11, -1, 18, -1, -1,
	9, -1, 27, 29, -1, 43, 46, 31, 37, 21, -1, -1, 8, 26, 49, 45,
	36, -1, 7, 48, 35, 6, 34, 33,
}

func BitCount(i int64) int {
	i = i - ((i >> 1) & 0x5555555555555555)
	i = (i & 0x3333333333333333) + ((i >> 2) & 0x3333333333333333)
	i = (i + (i >> 4)) & 0x0f0f0f0f0f0f0f0f
	i = i + (i >> 8)
	i = i + (i >> 16)
	i = i + (i >> 32)
	return int(i & 0x7f)
}

func Index(i int64) int {
	return bitIndex[int(i%67)]
}

func GetIndex(bit int64) []int {
	result := make([]int, 0)
	for bit != 0 {
		result = append(result, Index(bit & -bit))
		bit = bit & (bit - 1)
	}
	return result
}

func Shift(bb int64, direction int) int64 {
	if direction == dir.NorthEast {
		return (bb & 0xf03c0f03c00) >> 4 | (bb & 0x3e0f83e0f83e0) >> 5
	} else if direction == dir.NorthWest {
		return (bb & 0x1f07c1f07c00) >> 5 | (bb & 0x3c0f03c0f03c0) >> 6
	} else if direction == dir.SouthEast {
		return ((bb & 0xf03c0f03c0f) << 6 | (bb & 0xf83e0f83e0) << 5) & 0x3ffffffffffff
	} else if direction == dir.SouthWest {
		return ((bb & 0x1f07c1f07c1f) << 5 | (bb & 0xf03c0f03c0) << 4) & 0x3ffffffffffff
	} else if direction == dir.North {
		return bb >> 10
	} else if direction == dir.South {
		return bb << 10 & 0x3ffffffffffff
	} else if direction == dir.West {
		return bb >> 1 & 0x1ef7bdef7bdef
	} else if direction == dir.East {
		return bb << 1 & 0x3def7bdef7bde
	}
	return 0
}

func Reverse(bb int64) int64 {
	return ^bb & 0x3ffffffffffff
}

func MoreThanOne(bb int64) bool {
	return (bb & (bb - 1)) != 0
}