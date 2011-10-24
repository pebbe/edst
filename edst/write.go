package edst

import (
	"fmt"
)

func setTextPlain(q *Context) {
	if !q.isText {
		q.isText = true
		q.w.Header().Add("Content-type", "text/plain; charset=utf-8")
		// output BOM for UTF-8
		fmt.Fprintf(q.w, "%c", 0xfeff)
	}
}
