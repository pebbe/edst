package edst

import (
	"io/ioutil"
	"net/http"
	"strings"
)

func gettextfile(r *http.Request, key string) ([]string, error) {
	// get data as lines of string, properly decoded
	f, _, e := r.FormFile(key)
	if e != nil {
		return nil, e
	}
	d, e := ioutil.ReadAll(f)
	if e != nil {
		return nil, e
	}
	s, _ := decode(d) // from []byte to string
	if strings.Index(s, "\n") < 0 && strings.Index(s, "\r") >= 0 {
		return strings.SplitAfter(s, "\r"), nil
	}
	return strings.SplitAfter(s, "\n"), nil
}

func decode(b []byte) (data string, charset string) {
	// BOM-UTF-8
	if b[0] == 0xef && b[1] == 0xbb && b[2] == 0xbf {
		return string(b[3:]), "UTF-8"
	}

	// BOM-UTF-16-BE
	if b[0] == 0xfe && b[1] == 0xff {
		ln := len(b)/2 - 1
		s := make([]int32, ln)
		for i := 0; i < ln; i++ {
			s[i] = int32(b[2*i+3]) + 256*int32(b[2*i+2])
		}
		return string(s), "UTF-16-BE"
	}

	// BOM-UTF-16-LE
	if b[0] == 0xff && b[1] == 0xfe {
		ln := len(b)/2 - 1
		s := make([]int32, ln)
		for i := 0; i < ln; i++ {
			s[i] = int32(b[2*i+2]) + 256*int32(b[2*i+3])
		}
		return string(s), "UTF-16-LE"
	}

	// UTF-8 or US-ASCII
	r8 := func(b []byte) bool {
		for i := 0; i < len(b); i++ {
			if b[i] < 0x80 || b[i] > 0xBF {
				return false
			}
		}
		return true
	}
	ascii := true
	utf8 := true
	for i := 0; i < len(b); i++ {
		if !utf8 {
			break
		}
		switch {
		case b[i] <= 0x7F:
		case b[i] >= 0xC0 && b[i] <= 0xDF && r8(b[i+1:i+2]):
			ascii = false
			i += 1
		case b[i] >= 0xE0 && b[i] <= 0xEF && r8(b[i+1:i+3]):
			ascii = false
			i += 2
		case b[i] >= 0xF0 && b[i] <= 0xF7 && r8(b[i+1:i+4]):
			ascii = false
			i += 3
		case b[i] >= 0xF8 && b[i] <= 0xFB && r8(b[i+1:i+5]):
			ascii = false
			i += 4
		case b[i] >= 0xFC && b[i] <= 0xFD && r8(b[i+1:i+6]):
			ascii = false
			i += 5
		default:
			ascii = false
			utf8 = false
		}
	}
	if ascii {
		return string(b), "US-ASCII"
	}
	if utf8 {
		return string(b), "UTF-8"
	}

	// default: ISO-8859-1 
	ln := len(b)
	s := make([]int32, ln)
	for i := 0; i < ln; i++ {
		s[i] = int32(b[i])
	}
	return string(s), "ISO-8859-1"
}
