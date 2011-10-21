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
      <li><s>handle multiple strings per cell</s>
      <li>print list of tokens with classification
      <li>handle errors
    </ul>
    <form action="/upload" method="post" enctype="multipart/form-data">
      <fieldset>
        <legend>Data</legend>
        <a href="examples/example1.txt">example datafile 1</a><br>
        <a href="examples/example2.txt">example datafile 2</a><br>&nbsp;<br>
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


	// get data as lines of string, properly decoded
	f, _, e := r.FormFile("data")
	if e != nil {
		fmt.Fprintln(w, e)
		return
	}
	d, e := ioutil.ReadAll(f)
	if e != nil {
		fmt.Fprintln(w, e)
		return
	}
	data, _ := decode(d)   // from []byte to string
	lines := strings.SplitAfter(data, "\n")


	// output BOM for UTF-8
	fmt.Fprintf(w, "%c", 0xfeff)



	// do the data header
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


	// do the rest of the data
	for idx++; idx < len(lines); idx++ {
		if strings.TrimSpace(lines[idx]) == "" {
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
			items[i] = tokenize(cells[i+1]) // skip the first cell, that's a label
		}
		for i := 0; i < len(items)-1; i++ {
			for j := i + 1; j < len(items); j++ {
				fmt.Fprintf(w, "\t%.7f", editDistance(items[i], items[j]))
				//fmt.Fprintf(w, "\n\t%v\n\t%v", items[i], items[j])
			}
		}

		fmt.Fprintln(w)
	}
}
