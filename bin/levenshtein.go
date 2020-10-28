package main

import (
	"fmt"
	"html"
	"math"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

//. types

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

type cell struct {
	f                      float32
	above, left, aboveleft bool
}

type token struct {
	head string
	mods string // sorted for easy comparison
	str  string
}

type item struct {
	n int       // number of words in item: "a / b / c" = 3
	w [][]token // a list of n words, each a list of tokens
}

//. code

func maxint(i1, i2 int) int {
	if i1 > i2 {
		return i1
	}
	return i2
}

func kgv(l1, l2 int) int {
	ll1 := l1
	ll2 := l2
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

func diff(q *Context, i1, i2 token) float32 {
	if i1.head == "" || i2.head == "" {
		var i string
		if i1.head != "" {
			i = i1.head
		} else {
			i = i2.head
		}
		for _, j := range q.indelSets {
			if j.s[i] {
				return j.f
			}
		}
		return q.indelvalue
	}
	if i1.head == i2.head {
		if i1.mods == i2.mods {
			return 0.0
		}
		return q.modvalue
	}
	for _, i := range q.substSets {
		if i.s[i1.head] && i.s[i2.head] {
			return i.f
		}
	}
	return q.substvalue
}

func Levenshtein(q *Context, s1, s2 []token, wantAlign bool) float32 {
	var l1, l2, x, y int
	var f, aboveleft, above, left float32
	var xc, yc token

	l1 = len(s1)
	l2 = len(s2)

	if m := maxint(l1, l2); m > q.size {
		q.size = m * 2
		q.dst = make([][]cell, q.size+1)
		for i := 0; i <= q.size; i++ {
			q.dst[i] = make([]cell, q.size+1)
		}
	}

	nul := token{head: ""}

	for x = 0; x <= l2; x++ {
		for y = 0; y <= l1; y++ {
			q.dst[x][y].above, q.dst[x][y].left, q.dst[x][y].aboveleft = false, false, false
		}
	}

	q.dst[0][0].f = 0
	x = 0
	for _, xc = range s2 {
		x++
		q.dst[x][0].f = q.dst[x-1][0].f + diff(q, nul, xc)
		q.dst[x][0].left = true
	}
	y = 0
	for _, yc = range s1 {
		y++
		q.dst[0][y].f = q.dst[0][y-1].f + diff(q, nul, yc)
		q.dst[0][y].above = true
	}

	x = 0
	for _, xc = range s2 {
		x++
		y = 0
		for _, yc = range s1 {
			y++
			aboveleft = q.dst[x-1][y-1].f + diff(q, yc, xc)
			above = q.dst[x][y-1].f + diff(q, yc, nul)
			left = q.dst[x-1][y].f + diff(q, nul, xc)
			if aboveleft <= above && aboveleft <= left {
				q.dst[x][y].f = aboveleft
				q.dst[x][y].aboveleft = true
				continue
			}
			if above <= left {
				q.dst[x][y].f = above
				q.dst[x][y].above = true
				continue
			}
			q.dst[x][y].f = left
			q.dst[x][y].left = true
		}
	}

	if !wantAlign {
		return q.dst[l2][l1].f / float32(l1+l2)
	}

	line1 := make([]string, l1+l2)
	line2 := make([]string, l1+l2)
	line3 := make([]float32, l1+l2)
	ln := 0
	var F func(int, int)
	F = func(x, y int) {
		if x == 0 && y == 0 {
			return
		}
		line3[ln] = q.dst[x][y].f
		if q.dst[x][y].aboveleft {
			line1[ln] = html.EscapeString(s1[y-1].str)
			line2[ln] = html.EscapeString(s2[x-1].str)
			ln++
			F(x-1, y-1)
			return
		}
		if q.dst[x][y].above {
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

	q.Printf("<table class=\"align\"><tr class=\"txt\">\n")
	for i := ln - 1; i >= 0; i-- {
		if line1[i] == " " {
			line1[i] = "<span class=\"space\">SP</span>"
		}
		q.Printf("<td>&nbsp;%s&nbsp;</td>\n", line1[i])
	}
	q.Printf("<td class=\"white\">&nbsp;</td>\n</tr>\n<tr class=\"txt\">\n")
	for i := ln - 1; i >= 0; i-- {
		if line2[i] == " " {
			line2[i] = "<span class=\"space\">SP</span>"
		}
		q.Printf("<td>&nbsp;%s&nbsp;</td>\n", line2[i])
	}
	f = 0.0
	q.Printf("<td class=\"white\">&nbsp;</td>\n</tr>\n<tr>\n")
	for i := ln - 1; i >= 0; i-- {
		if line3[i] != f {
			q.Printf("<td>%g</td>\n", line3[i]-f)
			f = line3[i]
		} else {
			q.Printf("<td>&nbsp;</td>\n")
		}

	}
	q.Printf("<td class=\"total\">%g / %d = %.4f</td></tr>\n</table>\n", q.dst[l2][l1].f, l1+l2, q.dst[l2][l1].f/float32(l1+l2))

	return q.dst[l2][l1].f / float32(l1+l2)
}

func editDistance(q *Context, i1, i2 item) float32 {
	// 0 * n
	if i1.n == 0 || i2.n == 0 {
		return float32(math.NaN())
	}

	// 1 * 1
	if i1.n == 1 && i2.n == 1 {
		return Levenshtein(q, i1.w[0], i2.w[0], false)
	}

	// 1 * n
	if i2.n == 1 {
		i1, i2 = i2, i1
	}
	if i1.n == 1 {
		var sum float32
		sum = 0.0
		for i := 0; i < i2.n; i++ {
			sum += Levenshtein(q, i1.w[0], i2.w[i], false)
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
			d[i][j] = Levenshtein(q, i1.w[i], i2.w[j], false)
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

func itemize(q *Context, s string) item {
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
		it.w = append(it.w, tokenize(q, stringlist[i]))
	}
	return it
}

func tokenize(q *Context, s string) []token {
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
		if q.equi[c] != "" {
			cc = q.equi[c]
		}
		if q.paren[cc] != "" {
			finish()
			parlist = append(parlist, cc)
			str += c
		} else if q.paren2[cc] != "" {
			if len(parlist) > 0 && parlist[len(parlist)-1] == q.paren2[cc] {

				// pop parlist
				parlist = parlist[:len(parlist)-1]

			}
			str += c
		} else if q.mods[cc] {
			modlist = append(modlist, cc)
			str += c
		} else {
			finish()

			head = cc

			// copy parlist into modlist
			modlist = modlist[:0]
			for _, i := range parlist {
				modlist = append(modlist, i)
			}

			state = 1
			str += c
		}
	}
	finish()

	return tokens
}

func setup(q *Context, lines []string, e error) (err bool) {

	if e != nil {
		return false
	}

	var f float32
	err = false
	items := make([]string, 0, 300)
	state := NULL

	printfError := func(lineno int, format string, a ...interface{}) {
		q.isText = true
		if lineno < 0 {
			q.Print("Definition file: ")
		} else {
			q.Printf("Definition file, line %d: ", lineno+1)
		}
		q.Printf(format+"\n", a...)
		err = true
	}

	finish := func() {

		if state == EQUI {
			for _, c := range items {
				chrs := strings.Split(c, "")
				q.equi[chrs[0]] = chrs[1]
			}
		} else if state == PAREN {
			for _, c := range items {
				chrs := strings.Split(c, "")
				q.paren[chrs[0]] = chrs[1]
				q.paren2[chrs[1]] = chrs[0]
			}
		} else if state == MOD {
			for _, i := range items {
				q.mods[i] = true
			}
		} else if state == INDEL {
			m := make(map[string]bool)
			for _, i := range items {
				m[i] = true
			}
			q.indelSets = append(q.indelSets, set{f: f, s: m})
		} else if state == SUBST {
			m := make(map[string]bool)
			for _, i := range items {
				m[i] = true
			}
			q.substSets = append(q.substSets, set{f: f, s: m})
		}

		state = NULL
	}

	for lineno, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line[:1] == "#" {
			continue
		}
		if strings.HasPrefix(line, "DEFAULTS") {
			finish()
			a := strings.Fields(line)
			if len(a) != 4 {
				printfError(lineno, "Wrong number of arguments")
			} else {
				values := make([]float64, 3)
				var e error
				for i := 0; i < 3; i++ {
					values[i], e = strconv.ParseFloat(a[i+1], 32)
					if e != nil {
						printfError(lineno, "%v", e)
					}
				}
				q.substvalue = float32(values[0])
				q.indelvalue = float32(values[1])
				q.modvalue = float32(values[2])
			}
		} else if line == "EQUI" {
			finish()
			state = EQUI
			items = items[:0]
		} else if line == "PAREN" {
			finish()
			state = PAREN
			items = items[:0]
		} else if line == "MOD" {
			finish()
			state = MOD
			items = items[:0]
		} else if strings.HasPrefix(line, "INDEL") || strings.HasPrefix(line, "SUBST") {
			finish()
			a := strings.Fields(line)
			f = 0.0
			if len(a) != 2 {
				printfError(lineno, "Wrong number of arguments")
			} else {
				ff, e := strconv.ParseFloat(a[1], 32)
				if e != nil {
					printfError(lineno, "%v", e)
				} else {
					f = float32(ff)
				}
			}
			if strings.HasPrefix(line, "INDEL") {
				state = INDEL
			} else {
				state = SUBST
			}
			items = items[:0]
		} else {
			if state == MOD || state == INDEL || state == SUBST {
				a := strings.Fields(line)
				for _, c := range a {
					if utf8.RuneCountInString(c) == 1 {
						items = append(items, c)
					} else if c[:2] == "0x" || c[:2] == "0X" || c[:2] == "U+" || c[:2] == "u+" {
						var cc int
						_, e := fmt.Sscanf(c[2:], "%x", &cc)
						if e != nil {
							printfError(lineno, "\"%s\" is not a valid character value", c)
						} else {
							items = append(items, string(cc))
						}
					} else {
						cc, e := strconv.Atoi(c)
						if e != nil {
							printfError(lineno, "\"%s\" is not a valid character value", c)
						} else {
							items = append(items, string(cc))
						}
					}
				}
			} else if state == EQUI || state == PAREN {
				a := strings.Fields(line)
				for _, c := range a {
					if utf8.RuneCountInString(c) != 2 {
						printfError(lineno, "Invalid pair \"%v\", should be two characters", c)
					} else {
						items = append(items, c)
					}
				}
			} else {
				printfError(lineno, "Invalid line")
			}
		}
	}
	finish()
	return
}
