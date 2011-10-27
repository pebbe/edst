package edst

import (
	"html"
	"http"
	"os"
	"strings"
)

func init() {
	http.HandleFunc("/submit", submit)
}

func submit(w http.ResponseWriter, r *http.Request) {
	datalines, dataerror := gettextfile(r, "data")
	deflines, deferror := gettextfile(r, "def")
	choice := r.FormValue("choice")

	q := NewContext(w)

	err1 := checkData(q, datalines, dataerror)
	err2 := setup(q, deflines, deferror)
	if err1 || err2 {
		return
	}

	switch choice {
	case "edst":
		doEdst(q, datalines)
	case "alig":
		doAlign(q, datalines)
	}
}

func checkData(q *Context, datalines []string, dataerror os.Error) (err bool) {
	err = false

	printfError := func(lineno int, format string, a ...interface{}) {
		setTextPlain(q)
		if lineno < 0 {
			q.Print("Data file: ")
		} else {
			q.Printf("Data file, line %d: ", lineno+1)
		}
		q.Printf(format + "\n", a...)
		err = true
	}

	if dataerror != nil {
		printfError(-1, "%v", dataerror)
		return
	}

	idx := 0
	nlines := len(datalines)
	for {
		if idx == nlines {
			printfError(idx, "No data in file")
			return
		}
		if s := strings.TrimSpace(datalines[idx]); s == "" || s[:1] == "#" {
			idx++
		} else {
			break
		}
	}
	a := strings.Split(datalines[idx], "\t")
	nCols := len(a)
	if strings.TrimSpace(a[0]) == "" {
		nCols--
	}
	if nCols < 2 {
		printfError(idx, "No column labels found")
		return
	}

	for idx++; idx < nlines; idx++ {
		if s := strings.TrimSpace(datalines[idx]); s == "" || s[:1] == "#" {
			continue
		}
		cells := strings.Split(datalines[idx], "\t")
		if n := len(cells) - 1; n != nCols {
			printfError(idx, "Found %d data cells, should be %d", n, nCols)
		}
	}

	return
}

func doEdst(q *Context, lines []string) {
	setTextPlain(q)

	// do the data header
	idx := 0
	for {
		if s := strings.TrimSpace(lines[idx]); s == "" || s[:1] == "#" {
			q.Print(lines[idx])
			idx++
		} else {
			break
		}
	}
	xlabs := strings.Split(lines[idx], "\t")
	if strings.TrimSpace(xlabs[0]) == "" {
		xlabs = xlabs[1:]
	}
	for _, xlab := range xlabs {
		q.Printf("\t%v", strings.TrimSpace(xlab))
	}
	for i := 0; i < len(xlabs)-1; i++ {
		for j := i + 1; j < len(xlabs); j++ {
			q.Printf("\t%v:%v", strings.TrimSpace(xlabs[i]), strings.TrimSpace(xlabs[j]))
		}
	}
	q.Println()

	// do the rest of the data
	for idx++; idx < len(lines); idx++ {
		if s := strings.TrimSpace(lines[idx]); s == "" || s[:1] == "#" {
			q.Print(lines[idx])
			continue
		}

		cells := strings.Split(lines[idx], "\t")

		sep := ""
		for _, i := range cells {
			q.Printf("%v%v", sep, strings.TrimSpace(i))
			sep = "\t"
		}

		items := make([]item, len(cells)-1)
		for i := 0; i < len(cells)-1; i++ {
			items[i] = itemize(q, cells[i+1]) // skip the first cell, that's a label
		}
		for i := 0; i < len(items)-1; i++ {
			for j := i + 1; j < len(items); j++ {
				if items[i].n == 0 || items[j].n == 0 {
					q.Print("\t")
				} else {
					q.Printf("\t%.7f", editDistance(q, items[i], items[j]))
					// q.Printf("\n\t%v\n\t%v", items[i], items[j])
				}
			}
		}

		q.Println()
	}

}

func doAlign(q *Context, lines []string) {
	q.Print(`<html>
  <head>
    <title>Alignments</title>
    <link rel="stylesheet" type="text/css" href="style.css">
  </head>
  <body>
`)

	/*
		q.Printf("<pre>\nequi:\n%v\n</pre>\n", q.equi)
		q.Printf("<pre>\nparen:\n%v\n%v\n</pre>\n", q.paren, q.paren2)
		q.Printf("<pre>\nmods:\n%v\n</pre>\n", q.mods)
		q.Printf("<pre>\nindel:\n%v\n</pre>\n", q.indelSets)
		q.Printf("<pre>\nsubst:\n%v\n</pre>\n", q.substSets)
	*/

	// do the data header
	idx := 0
	for {
		if s := strings.TrimSpace(lines[idx]); s == "" || s[:1] == "#" {
			idx++
		} else {
			break
		}
	}
	xlabs := strings.Split(lines[idx], "\t")
	if strings.TrimSpace(xlabs[0]) == "" {
		xlabs = xlabs[1:]
	}
	for i, xlab := range xlabs {
		xlabs[i] = html.EscapeString(strings.TrimSpace(xlab))
	}

	// do the rest of the data
	for idx++; idx < len(lines); idx++ {
		if s := strings.TrimSpace(lines[idx]); s == "" || s[:1] == "#" {
			continue
		}

		cells := strings.Split(lines[idx], "\t")

		ylab := cells[0]
		cells = cells[1:]

		q.Printf("<h2>%s</h2>\n", html.EscapeString(ylab))

		items := make([]item, len(cells))
		for i := 0; i < len(cells); i++ {
			items[i] = itemize(q, cells[i])
		}
		for i := 0; i < len(items)-1; i++ {
			for j := i + 1; j < len(items); j++ {
				q.Printf("<div>%s &mdash; %s</div>\n", xlabs[i], xlabs[j])
				for _, iti := range items[i].w {
					for _, itj := range items[j].w {
						Levenshtein(q, iti, itj, true)
					}
				}
			}
		}

	}

	q.Println("</body>\n</html>")

}
