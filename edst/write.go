package edst

import (
	"fmt"
	"os"
)

func (q *Context) Printf(format string, a ...interface{}) (n int, errno os.Error) {
	return fmt.Fprintf(q.w, format, a...)
}

func (q *Context) Print(a ...interface{}) (n int, errno os.Error) {
	return fmt.Fprint(q.w, a...)
}

func (q *Context) Println(a ...interface{}) (n int, errno os.Error) {
	return fmt.Fprintln(q.w, a...)
}

func setTextPlain(q *Context) {
	if !q.isText {
		q.isText = true
		q.w.Header().Add("Content-type", "text/plain; charset=utf-8")
		// output BOM for UTF-8
		fmt.Fprintf(q.w, "%c", 0xfeff)
	}
}
