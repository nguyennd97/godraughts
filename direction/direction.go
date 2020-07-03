package direction

const (
	NorthEast int = -7
	NorthWest int = -9
	SouthEast int = 9
	SouthWest int = 7
	North     int = -8
	South     int = 8
	East      int = 1
	West      int = -1
	NonDir    int = 0
)

var (
	Dir = []int{NorthEast, NorthWest, SouthEast, SouthWest, NonDir}
)
