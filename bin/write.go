package main

import (
	"fmt"
	"io"
)

func (q *Context) Printf(format string, a ...interface{}) (n int, errno error) {
	return fmt.Fprintf(&q.buf, format, a...)
}

func (q *Context) Print(a ...interface{}) (n int, errno error) {
	return fmt.Fprint(&q.buf, a...)
}

func (q *Context) Println(a ...interface{}) (n int, errno error) {
	return fmt.Fprintln(&q.buf, a...)
}

func (q *Context) output(w io.Writer) {
	if q.isText {
		fmt.Fprint(w, "Content-type: text/plain; charset=utf-8\n\n")
	} else {
		fmt.Fprint(w, "Content-type: text/html; charset=utf-8\n\n")
	}
	fmt.Fprint(w, q.buf.String())
}
