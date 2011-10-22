package edst

import (
	"fmt"
	"html"
	"http"
	"io/ioutil"
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

	reset()

	if deferror == nil {
		setup(deflines)
	}

	switch choice {
	case "edst":
		doEdst(w, datalines, dataerror)
	case "alig":
		doAlign(w, datalines, dataerror)
	} 
}

func gettextfile(r *http.Request, key string) ([]string, os.Error) {
	// get data as lines of string, properly decoded
	f, _, e := r.FormFile(key)
	if e != nil {
		return nil, e
	}
	d, e := ioutil.ReadAll(f)
	if e != nil {
		return nil, e
	}
	s, _ := decode(d)   // from []byte to string
	return strings.SplitAfter(s, "\n"), nil
}

func doEdst(w http.ResponseWriter, lines []string, error os.Error) {
	w.Header().Add("Content-type", "text/plain; charset=utf-8")
	
	// output BOM for UTF-8
	fmt.Fprintf(w, "%c", 0xfeff)

	if error != nil {
		fmt.Fprintf(w, "Error data file\n%v\n", error)
		return
	}

	// do the data header
	idx := 0
	for {
		if s := strings.TrimSpace(lines[idx]); s == "" || s[:1] == "#" {
			fmt.Fprintf(w, lines[idx])
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
		fmt.Fprintf(w, "\t%v", strings.TrimSpace(xlab))
	}
	for i := 0; i < len(xlabs)-1; i++ {
		for j := i + 1; j < len(xlabs); j++ {
			fmt.Fprintf(w, "\t%v:%v", strings.TrimSpace(xlabs[i]), strings.TrimSpace(xlabs[j]))
		}
	}
	fmt.Fprintln(w)


	// do the rest of the data
	for idx++; idx < len(lines); idx++ {
		if s := strings.TrimSpace(lines[idx]); s == "" || s[:1] == "#" {
			fmt.Fprintf(w, lines[idx])
			continue
		}

		cells := strings.Split(lines[idx], "\t")

		sep := ""
		for _, i := range cells {
			fmt.Fprintf(w, "%v%v", sep, strings.TrimSpace(i))
			sep = "\t"
		}

		items := make([]item, len(cells)-1)
		for i := 0; i < len(cells)-1; i++ {
			items[i] = itemize(cells[i+1]) // skip the first cell, that's a label
		}
		for i := 0; i < len(items)-1; i++ {
			for j := i + 1; j < len(items); j++ {
				fmt.Fprintf(w, "\t%.7f", editDistance(items[i], items[j]))
				// fmt.Fprintf(w, "\n\t%v\n\t%v", items[i], items[j])
			}
		}

		fmt.Fprintln(w)
	}

}

func doAlign(w http.ResponseWriter, lines []string, error os.Error) {
	fmt.Fprint(w, `<html>
  <head>
    <title>Alignments</title>
    <link rel="stylesheet" type="text/css" href="style.css">
  </head>
  <body>
`)

	if error != nil {
		fmt.Fprintf(w, "<h1>Error data file</h1>\n<div>%s</div>\n</body>\n</html>\n", html.EscapeString(error.String()))
		return
	}

	/*
	fmt.Fprintf(w, "<pre>\nequi:\n%v\n</pre>\n", equi)
	fmt.Fprintf(w, "<pre>\nparen:\n%v\n%v\n</pre>\n", paren, paren2)
	fmt.Fprintf(w, "<pre>\nmods:\n%v\n</pre>\n", mods)
	fmt.Fprintf(w, "<pre>\nindel:\n%v\n</pre>\n", indelSets)
	fmt.Fprintf(w, "<pre>\nsubst:\n%v\n</pre>\n", substSets)
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

		fmt.Fprintf(w, "<h2>%s</h2>\n", html.EscapeString(ylab))

		items := make([]item, len(cells))
		for i := 0; i < len(cells); i++ {
			items[i] = itemize(cells[i])
		}
		for i := 0; i < len(items)-1; i++ {
			for j := i + 1; j < len(items); j++ {
				fmt.Fprintf(w, "<div>%s &mdash; %s</div>\n", xlabs[i], xlabs[j])
				for _, iti := range items[i].w {
					for _, itj := range items[j].w {
						LevenshteinAlignment(w, iti, itj)
					}
				}
			}
		}

	}

	fmt.Fprintln(w, "</body>\n</html>")

}
