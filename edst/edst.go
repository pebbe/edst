package edst

import (
	"fmt"
	"http"
	"io/ioutil"
	"strings"
)

func init() {
	http.HandleFunc("/", index)
	http.HandleFunc("/upload", upload)
}

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `<html>
  <head>
    <title>Edit distance</title>
  </head>
  <body>
    <b>To do:</b>
    <ul>
      <li>edit distance on tokenised strings, with proper weights
          (now: plain edit distance on raw strings)
      <li>handle multiple strings per cell (now: treated as single string)
      <li>print list of tokens with classification
      <li>handle errors
    </ul>
    <form action="/upload" method="post" enctype="multipart/form-data">
      <fieldset>
        <legend>Data</legend>
        <a href="examples/example1.txt">example datafile</a><br>&nbsp;<br>
        Data file:<br>
        <input type="file" name="data" size="40">
      </fieldset>
      <input type="submit" value="Upload data file">
    </form>
  </body>
</html>
`)
}

func upload(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-type", "text/plain; charset=utf-8")

	f, _, e := r.FormFile("data")
	if e != nil {
		fmt.Fprintln(w, e)
		return
	}

	fmt.Fprintf(w, "%c", 0xfeff)

	d, e := ioutil.ReadAll(f)
	if e != nil {
		fmt.Fprintln(w, e)
		return
	}

	data, _ := decode(d)

	lines := strings.SplitAfter(data, "\n")

	idx := 0
	for {
		if strings.TrimSpace(lines[idx]) == "" {
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
			fmt.Fprintf(w, "\t%v__%v", strings.TrimSpace(xlabs[i]), strings.TrimSpace(xlabs[j]))
		}
	}
	fmt.Fprintln(w)

	for idx++; idx < len(lines); idx++ {
		if strings.TrimSpace(lines[idx]) == "" {
			continue
		}

		items := strings.Split(lines[idx], "\t")
		for i := 0; i < len(items); i++ {
			items[i] = strings.TrimSpace(items[i])
		}

		sep := ""
		for _, i := range items {
			fmt.Fprintf(w, "%v%v", sep, i)
			sep = "\t"
		}


		words := make([][]token, len(items)-1)
		for i := 0; i < len(items)-1; i++ {
			words[i] = make([]token, 0, len(items[i+1]))
			// TODO: replace with real tokeniser
			for _, c := range items[i+1] {
				t := token{head: c}
				words[i] = append(words[i], t)
			}

		}

		for i := 0; i < len(words)-1; i++ {
			for j := i + 1; j < len(words); j++ {
				if len(words[i]) == 0 || len(words[j]) == 0 {
					fmt.Fprintf(w, "\tNA")
				} else {
					fmt.Fprintf(w, "\t%.7f", Levenshtein(words[i], words[j]))
					// fmt.Fprintf(w, "\n\t%v\n\t%v", words[i], words[j])
				}
			}
		}
		fmt.Fprintln(w)
	}
}
