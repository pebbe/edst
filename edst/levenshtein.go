package edst

import (
	"fmt"
	"html"
	"http"
	"math"
	"sort"
	"strconv"
	"strings"
	"utf8"
)

var example = `

# EXAMPLE definition file

# Default values for: substitution indel modifier
DEFAULTS 2.0 1.0 0.5

# First letter of each pair is replaced by second letter
EQUI
Aa Bb Cc Dd Ee Ff Gg Hh
Ii Jj Kk Ll Mm Nn Oo Pp
Qq Rr Ss Tt Uu Vv Ww Xx
Yy Zz

# This: a{bc}d
# will tokenize as: a b{ c{ d
# Mismatch of paren will cause error
PAREN
{} [] ()

# Characters in all following definitions can be written by itself or numerically
# These are the same: a 97 0x61 U+0061

# This: ab~c
# will tokenize as: a b~ c
# Modifiers not after head will cause error
MOD
~ ^

# Indels with non-default values
# Characters can be in only one set
INDEL 0.0
32

INDEL 0.5
- "

# Substitution sets with non-default values
# Characters can be in multiple sets
# Order is importent, first matching set is used
SUBST 0.5
v f

SUBST 0.5
v w

SUBST 1.0
b d c f g h j
k l m n p q r
s t v w x y z

SUBST 1.0
a e i o u

# all remaining characters are treated as heads with default subst value 

# END

# Possible future extensions
# - Multi-character heads, such as: ij ch

`

type StateType int
const (
	NULL StateType = iota
	EQUI
	PAREN
	MOD
	INDEL
	SUBST
)

type set struct {
	f float32
	s map[string]bool
}

var (
	dst   [][]float32
	size  int
	adst  [][]cell
	asize int

	substvalue float32
	indelvalue float32
	modvalue   float32

	paren  map[string]string
	paren2  map[string]string
	equi  map[string]string
	mods  map[string]bool
	indelSets []set
	substSets []set
)

func reset() {
	size = 0
	asize = 0
	paren = make(map[string]string)
	paren2 = make(map[string]string)
	equi = make(map[string]string)
	mods = make(map[string]bool)
	substvalue = 2.0
	indelvalue = 1.0
	modvalue   = 0.5
	indelSets = make([]set, 0)
	substSets = make([]set, 0)
}


type cell struct {
	f float32
	above, left, aboveleft bool
}

type token struct {
	head string
	mods string   // sorted for easy comparison
	str  string
}

type item struct {
	n int       // number of words in item: "a / b / c" = 3
	w [][]token // a list of n words, each a list of tokens
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
	if i1.head == "" || i2.head == "" {
		var i string
		if i1.head != "" {
			i = i1.head
		} else {
			i = i2.head
		}
		for _, j := range indelSets {		
			if j.s[i] {
				return j.f
			}
		}
		return indelvalue
	}
	// TODO: complex comparison of tokens
	if i1.head == i2.head {
		if i1.mods == i2.mods {
			return 0.0
		}
		return modvalue
	}
	for _, i := range substSets {		
		if i.s[i1.head] && i.s[i2.head] {
			return i.f
		}
	}
	return substvalue
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

	nul := token{head: ""}

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

func LevenshteinAlignment(w http.ResponseWriter, s1, s2 []token) {
	var l1, l2, x, y int
	var f, aboveleft, above, left float32
	var xc, yc token

	l1 = len(s1)
	l2 = len(s2)

	if m := max2int(l1, l2); m > asize {
		asize = m * 2
		adst = make([][]cell, asize+1)
		for i := 0; i <= asize; i++ {
			adst[i] = make([]cell, asize+1)
		}
	}

	nul := token{head: ""}

	for x = 0; x <= l2; x++ {
		for y = 0; y <= l1; y++ {
			adst[x][y].above, adst[x][y].left, adst[x][y].aboveleft = false, false, false
		}
	}

	adst[0][0].f = 0
	x = 0
	for _, xc = range s2 {
		x++
		adst[x][0].f = adst[x-1][0].f + diff(nul, xc)
		adst[x][0].left = true
	}
	y = 0
	for _, yc = range s1 {
		y++
		adst[0][y].f = adst[0][y-1].f + diff(nul, yc)
		adst[0][y].above = true 
	}

	x = 0
	for _, xc = range s2 {
		x++
		y = 0
		for _, yc = range s1 {
			y++
			aboveleft = adst[x-1][y-1].f + diff(yc, xc)
			above = adst[x][y-1].f + diff(yc, nul)
			left = adst[x-1][y].f + diff(nul, xc)
			if aboveleft <= above && aboveleft <= left {
				adst[x][y].f = aboveleft
				adst[x][y].aboveleft = true
				continue
			}
			if  above <= left {
				adst[x][y].f = above
				adst[x][y].above = true
				continue
			}
			adst[x][y].f = left
			adst[x][y].left = true
		}
	}

	line1 := make([]string, l1 + l2)
	line2 := make([]string, l1 + l2)
	line3 := make([]float32, l1 + l2)
	ln := 0
	var F func(int, int)
	F = func(x, y int) {
		if x == 0 && y == 0 {
			return
		}
		line3[ln] = adst[x][y].f
		if adst[x][y].aboveleft {
			line1[ln] = html.EscapeString(s1[y-1].str)
			line2[ln] = html.EscapeString(s2[x-1].str)
			ln++
			F(x-1, y-1)
			return
		}
		if adst[x][y].above {
			line1[ln] = html.EscapeString(s1[y-1].str)
			line2[ln] = ""
			ln++
			F(x, y-1)
			return
		}
		line1[ln] = ""
		line2[ln] = html.EscapeString(s2[x-1].str)
		ln++
		F(x-1, y)
		
	}
	F(l2, l1)

	fmt.Fprintf(w, "<table class=\"align\"><tr class=\"txt\">\n")
	for i := ln - 1; i >= 0; i-- {
		if line1[i] == " " {
			line1[i] = "<span class=\"space\">SP</span>"
		}
		fmt.Fprintf(w, "<td>&nbsp;%s&nbsp;</td>\n", line1[i])
	}
	fmt.Fprintf(w, "<td class=\"white\">&nbsp;</td>\n</tr>\n<tr class=\"txt\">\n")
	for i := ln - 1; i >= 0; i-- {
		if line2[i] == " " {
			line2[i] = "<span class=\"space\">SP</span>"
		}
		fmt.Fprintf(w, "<td>&nbsp;%s&nbsp;</td>\n", line2[i])
	}
	f = 0.0
	fmt.Fprintf(w, "<td class=\"white\">&nbsp;</td>\n</tr>\n<tr>\n")
	for i := ln - 1; i >= 0; i-- {
		if line3[i] != f {
			fmt.Fprintf(w, "<td>%g</td>\n", line3[i] - f)
			f = line3[i]
		} else {
			fmt.Fprintf(w, "<td>&nbsp;</td>\n")
		}

	}
	fmt.Fprintf(w, "<td class=\"total\">%g / %d = %.4f</td></tr>\n</table>\n", adst[l2][l1].f, l1+l2, adst[l2][l1].f / float32(l1 + l2))

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
			sum += Levenshtein(i1.w[0], i2.w[i])
		}
		return sum / float32(i2.n)
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
				if d[d1[i]][d2[i]]+d[d1[j]][d2[j]] > d[d1[i]][d2[j]]+d[d1[j]][d2[i]] {
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
	stringlist := make([]string, 0, strings.Count(s, " / ")+1)
	for _, i := range strings.Split(s, " / ") {
		i := strings.TrimSpace(i)
		if i != "" {
			stringlist = append(stringlist, i)
		}
	}
	n := len(stringlist)
	it := item{n: n, w: make([][]token, 0, n)}
	for i := 0; i < n; i++ {
		it.w = append(it.w, tokenize(stringlist[i]))
	}
	return it
}

func tokenize(s string) []token {
	max := len(s)
	tokens := make([]token, 0, max)
	parlist := make([]string, 0, max)
	var modlist sort.StringSlice
	head := ""
	str := ""
	state := 0
	finish := func() {
		if state > 0 {

			sort.Sort(modlist)
			s := ""
			for _, c := range modlist {
				s = s + c
			}

			t := token{head: head, mods: s, str: str}
			tokens = append(tokens, t)

			// reset modlist
			modlist = modlist[:0]

			head = ""
			str = ""
			state = 0
		}
	}
	for _, c := range strings.Split(s, "") {
		cc := c
		if equi[c] != "" {
			cc = equi[c]
		}
		if paren[cc] != "" {
			finish()
			parlist = append(parlist, cc)
			str = str + c
		} else if paren2[cc] != "" {
			if len(parlist) > 0 && parlist[len(parlist)-1] == paren2[cc] {

				// pop parlist
				parlist = parlist[:len(parlist)-1]

			}
			str = str + c
		} else if mods[cc] {
			modlist = append(modlist, cc)
			str = str + c
		} else {
			finish()

			head = cc

			// copy parlist into modlist
			modlist = parlist[0:len(parlist)]

			state = 1
			str = str + c
		}
	}
	finish()

	return tokens
}

func setup(lines []string) {

	var f float32
	var items []string

	substvalue = 2.0
	indelvalue = 1.0
	modvalue = 0.5

	state := NULL

	finish := func() {

		if state == EQUI {
			for _, c := range items {
				chrs := strings.Split(c, "")
				equi[chrs[0]] = chrs[1]
			}
		} else if state == PAREN {
			for _, c := range items {
				chrs := strings.Split(c, "")
				paren[chrs[0]] = chrs[1]
				paren2[chrs[1]] = chrs[0]
			}
		} else if state == MOD {
			for _,i := range(items) {
				mods[i] = true
			}
		} else if state == INDEL {
			m := make(map[string]bool)
			for _,i := range(items) {
				m[i] = true
			}
			indelSets = append(indelSets, set{f:f, s:m})		
		} else if state == SUBST {
			m := make(map[string]bool)
			for _,i := range(items) {
				m[i] = true
			}
			substSets = append(substSets, set{f:f, s:m})		
		}


		state = NULL
	}

	for _, line:= range lines {
		line = strings.TrimSpace(line)
		if line == "" || line[:1] == "#" {
			continue
		}
		if strings.HasPrefix(line, "DEFAULTS") {
			finish()
			a := strings.Fields(line)
			substvalue, _ = strconv.Atof32(a[1])
			indelvalue,_ = strconv.Atof32(a[2])
			modvalue,_ = strconv.Atof32(a[3])
		} else if line == "EQUI" {
			finish()
			state = EQUI
			items = make([]string, 0, 300)
		} else if line == "PAREN" {
			finish()
			state = PAREN
			items = make([]string, 0, 300)
		} else if line == "MOD" {
			finish()
			state = MOD
			items = make([]string, 0, 300)
		} else if strings.HasPrefix(line, "INDEL") {
			finish()
			a := strings.Fields(line)
			f,_ = strconv.Atof32(a[1])
			state = INDEL
			items = make([]string, 0, 300)
		} else if strings.HasPrefix(line, "SUBST") {
			finish()
			a := strings.Fields(line)
			f,_ = strconv.Atof32(a[1])
			state = SUBST
			items = make([]string, 0, 300)
		} else {
			if state == MOD || state == INDEL || state == SUBST {
				a := strings.Fields(line)
				for _, c := range a {
					if utf8.RuneCountInString(c) == 1 {
						items = append(items, c)
					} else if c[:2] == "0x" || c[:2] == "0X" || c[:2] == "U+" || c[:2] == "u+" {
						var cc int
						fmt.Sscanf(c[2:], "%x", &cc)
						items = append(items, fmt.Sprintf("%c", cc))
					} else {
						cc, _ := strconv.Atoi(c)
						items = append(items, fmt.Sprintf("%c", cc))
					}
				}
			} else if state == EQUI || state == PAREN {
				a := strings.Fields(line)
				for _, c := range a {
					items = append(items, c)
				}
			}
		}
	}
	finish()

}
