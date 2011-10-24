package edst

import (
	"fmt"
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

	writeError := func(lineno int, msg string) {
		setTextPlain(q)
		if lineno < 0 {
			fmt.Fprintln(q.w, "Data file:", msg)
		} else {
			fmt.Fprintf(q.w, "Data file, line %d: %v\n", lineno+1, msg)
		}
		err = true
	}

	if dataerror != nil {
		writeError(-1, dataerror.String())
		return
	}

	idx := 0
	nlines := len(datalines)
	for {
		if idx == nlines {
			writeError(idx, "No data in file")
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
		writeError(idx, "No column labels found")
		return
	}

	for idx++; idx < nlines; idx++ {
		if s := strings.TrimSpace(datalines[idx]); s == "" || s[:1] == "#" {
			continue
		}
		cells := strings.Split(datalines[idx], "\t")
		if n := len(cells) - 1; n != nCols {
			writeError(idx, fmt.Sprintf("Found %d data cells, should be %d", n, nCols))
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
			fmt.Fprintf(q.w, lines[idx])
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
		fmt.Fprintf(q.w, "\t%v", strings.TrimSpace(xlab))
	}
	for i := 0; i < len(xlabs)-1; i++ {
		for j := i + 1; j < len(xlabs); j++ {
			fmt.Fprintf(q.w, "\t%v:%v", strings.TrimSpace(xlabs[i]), strings.TrimSpace(xlabs[j]))
		}
	}
	fmt.Fprintln(q.w)

	// do the rest of the data
	for idx++; idx < len(lines); idx++ {
		if s := strings.TrimSpace(lines[idx]); s == "" || s[:1] == "#" {
			fmt.Fprintf(q.w, lines[idx])
			continue
		}

		cells := strings.Split(lines[idx], "\t")

		sep := ""
		for _, i := range cells {
			fmt.Fprintf(q.w, "%v%v", sep, strings.TrimSpace(i))
			sep = "\t"
		}

		items := make([]item, len(cells)-1)
		for i := 0; i < len(cells)-1; i++ {
			items[i] = itemize(q, cells[i+1]) // skip the first cell, that's a label
		}
		for i := 0; i < len(items)-1; i++ {
			for j := i + 1; j < len(items); j++ {
				fmt.Fprintf(q.w, "\t%.7f", editDistance(q, items[i], items[j]))
				// fmt.Fprintf(q.w, "\n\t%v\n\t%v", items[i], items[j])
			}
		}

		fmt.Fprintln(q.w)
	}

}

func doAlign(q *Context, lines []string) {
	fmt.Fprint(q.w, `<html>
  <head>
    <title>Alignments</title>
    <link rel="stylesheet" type="text/css" href="style.css">
  </head>
  <body>
`)

	/*
		fmt.Fprintf(q.w, "<pre>\nequi:\n%v\n</pre>\n", q.equi)
		fmt.Fprintf(q.w, "<pre>\nparen:\n%v\n%v\n</pre>\n", q.paren, q.paren2)
		fmt.Fprintf(q.w, "<pre>\nmods:\n%v\n</pre>\n", q.mods)
		fmt.Fprintf(q.w, "<pre>\nindel:\n%v\n</pre>\n", q.indelSets)
		fmt.Fprintf(q.w, "<pre>\nsubst:\n%v\n</pre>\n", q.substSets)
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

		fmt.Fprintf(q.w, "<h2>%s</h2>\n", html.EscapeString(ylab))

		items := make([]item, len(cells))
		for i := 0; i < len(cells); i++ {
			items[i] = itemize(q, cells[i])
		}
		for i := 0; i < len(items)-1; i++ {
			for j := i + 1; j < len(items); j++ {
				fmt.Fprintf(q.w, "<div>%s &mdash; %s</div>\n", xlabs[i], xlabs[j])
				for _, iti := range items[i].w {
					for _, itj := range items[j].w {
						LevenshteinAlignment(q, iti, itj)
					}
				}
			}
		}

	}

	fmt.Fprintln(q.w, "</body>\n</html>")

}
