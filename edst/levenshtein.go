package edst

/*

 Planning voor levenshtein met tokens

 INDEL: 1

 SUBST:

     als head gelijk en accenten gelijk: 0.0

     als head gelijk en accenten ongelijk: 0.5

     als head ongelijk en bitwise-or klassen niet nul: 1.0

     anders: 2.0

 KLASSEN

     0x0000 matcht niks
     0xffff matcht alles
     etc

     accenten: klasse -1, alleen in definitie, niet gebruikt als tokenklasse

     onbekende tekens: klasse 0

     haakjes: { [ ( gelden als accenten van volgende tokens tot na matching ) ] } 

     errors: mismatch van haakjes

 MISSCHIEN

     errors: accenten alleen na sommige klassen?

     pre-modifyers als accenten?

     subklassen?

*/

import (
	"math"
	"strings"
)

var (
	dst  [][]float32
	size int = 0
)

type token struct {
	head    int
	// class   int
	// accents []int
}

type item struct {
	n int          // number of words in item: "a / b / c" = 3
	w [][]token    // a list of n words, each a list of tokens
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

func kgv(l1, l2 int) int {
    var ll1, ll2 int
    ll1 = l1
    ll2 = l2
    for {
        if ll1 == ll2 {
			return ll1
		}
        for ll1 < ll2 {
            ll1 += l1
		}
        for ll2 < ll1 {
            ll2 += l2
		}
    }
	return 0
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

func editDistance(i1, i2 item) float32 {
	// 0 * n
	if i1.n == 0 || i2.n == 0 {
		return float32(math.NaN())
	}

	// 1 * 1
	if i1.n == 1 && i2.n == 1 {
		return Levenshtein(i1.w[0], i2.w[0])
	}

	// 1 * n
	if i2.n == 1 {
		i1, i2 = i2, i1
	}
	if i1.n == 1 {
		var sum float32
		sum = 0.0
		for i := 0; i < i2.n; i++ {
			sum +=  Levenshtein(i1.w[0], i2.w[i])
		}
		return sum / float32(i2.n);
	}

	// n * n
	n1 := i1.n
	n2 := i2.n
	d := make([][]float32, n1)
	for i := 0; i < n1; i++ {
		d[i] = make([]float32, n2)
	}
	for i := 0; i < n1; i++ {
		for j := 0; j < n2; j++ {
			d[i][j] = Levenshtein(i1.w[i], i2.w[j])
		}
	}
	l := kgv(n1, n2)
	d1 := make([]int, l)
	d2 := make([]int, l)
	for i := 0; i < l; i++ {
		d1[i] = i % n1
		d2[i] = i % n2
	}
    for found := true; found; {
        found = false
        for i := 0; i < l; i++ {
            for j := i + 1; j < l; j++ {
                if d[d1[i]][d2[i]] + d[d1[j]][d2[j]] > d [d1[i]][d2[j]] + d[d1[j]][d2[i]] {
                    d2[i], d2[j] = d2[j], d2[i]
                    found = true
                }
            }
		}
	}
	var sum float32
	sum = 0.0
    for i := 0; i < l; i++ {
        sum += d[d1[i]][d2[i]]
	}
    return sum / float32(l)

}

func itemize(s string) item {
	stringlist := make([]string, 0, strings.Count(s, " / ") + 1)
	for _, i := range strings.Split(s, " / ") {
		i := strings.TrimSpace(i)
		if i != "" {
			stringlist = append(stringlist, i)
		}
	}
	n := len(stringlist)
	it := item{n:n, w:make([][]token, 0, n)}
	for i := 0; i < n; i++ {
		it.w = append(it.w, tokenize(stringlist[i]))
	}
	return it
}

func tokenize(s string) []token {
	tokens := make([]token, 0, len(s))

	// TODO: replace with real tokeniser
	for _, c := range s {
		t := token{head: c}
		tokens = append(tokens, t)
	}

	return tokens
}
