package edst

import (
	"http"
)

// The context acts as global store for a single request

type Context struct {
	w      http.ResponseWriter
	isText bool

	dst  [][]cell
	size int

	substvalue float32
	indelvalue float32
	modvalue   float32

	paren     map[string]string
	paren2    map[string]string
	equi      map[string]string
	mods      map[string]bool
	indelSets []set
	substSets []set
}

func NewContext(w http.ResponseWriter) *Context {
	return &Context{
		w:      w,
		isText: false,

		dst:  nil,
		size: -1,

		substvalue: 2.0,
		indelvalue: 1.0,
		modvalue:   0.5,

		paren:     make(map[string]string),
		paren2:    make(map[string]string),
		equi:      make(map[string]string),
		mods:      make(map[string]bool),
		indelSets: make([]set, 0, 50),
		substSets: make([]set, 0, 50)}
}
