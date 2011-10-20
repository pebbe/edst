package edst

var (
	dst  [][]float32
	size int = 0
)

type token struct {
	head    int
	class   int
	accents []int
}

func max2int(i1, i2 int) int {
	if i1 > i2 {
		return i1
	}
	return i2
}

func min3float32(f1, f2, f3 float32) float32 {
	if f1 < f2 && f1 < f3 {
		return f1
	}
	if f2 < f3 {
		return f2
	}
	return f3
}

func diff(i1, i2 token) float32 {
	if i1.head == 0 || i2.head == 0 {
		return 1.0
	}
	// TODO: complex comparison of tokens
	if i1.head == i2.head {
		return 0.0
	}
	return 2.0
}

func Levenshtein(s1, s2 []token) float32 {
	var l1, l2, x, y int
	var aboveleft, above, left float32
	var xc, yc token

	l1 = len(s1)
	l2 = len(s2)

	if m := max2int(l1, l2); m > size {
		size = m * 2
		dst = make([][]float32, size+1)
		for i := 0; i <= size; i++ {
			dst[i] = make([]float32, size+1)
		}
	}

	nul := token{head: 0}

	dst[0][0] = 0
	x = 0
	for _, xc = range s2 {
		x++
		dst[x][0] = dst[x-1][0] + diff(nul, xc)
	}
	y = 0
	for _, yc = range s1 {
		y++
		dst[0][y] = dst[0][y-1] + diff(nul, yc)
	}

	x = 0
	for _, xc = range s2 {
		x++
		y = 0
		for _, yc = range s1 {
			y++
			aboveleft = dst[x-1][y-1] + diff(yc, xc)
			above = dst[x][y-1] + diff(yc, nul)
			left = dst[x-1][y] + diff(nul, xc)
			dst[x][y] = min3float32(aboveleft, above, left)
		}
	}

	return dst[l2][l1] / float32(l1+l2)
}
