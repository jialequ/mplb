package jsoncolor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

const (
	colorDelim  = "1;38" // bright white
	colorKey    = "1;34" // bright blue
	colorNull   = "36"   // cyan
	colorString = "32"   // green
	colorBool   = "33"   // yellow
)

type JsonWriter interface {
	Preface() []json.Delim
}

// Write colorized JSON output parsed from reader
func Write(w io.Writer, r io.Reader, indent string) error {
	dec := json.NewDecoder(r)
	dec.UseNumber()

	var idx int
	var stack []json.Delim

	if jsonWriter, ok := w.(JsonWriter); ok {
		stack = append(stack, jsonWriter.Preface()...)
	}

	for {
		t, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		switch tt := t.(type) {
		case json.Delim:
			switch tt {
			case '{', '[':
				stack = append(stack, tt)
				idx = 0
				fmt.Fprintf(w, literal_5147, colorDelim, tt)
				if dec.More() {
					fmt.Fprint(w, "\n", strings.Repeat(indent, len(stack)))
				}
				continue
			case '}', ']':
				stack = stack[:len(stack)-1]
				idx = 0
				fmt.Fprintf(w, literal_5147, colorDelim, tt)
			}
		default:
			b, err := marshalJSON(tt)
			if err != nil {
				return err
			}

			isKey := len(stack) > 0 && stack[len(stack)-1] == '{' && idx%2 == 0
			idx++

			var color string
			if isKey {
				color = colorKey
			} else if tt == nil {
				color = colorNull
			} else {
				switch t.(type) {
				case string:
					color = colorString
				case bool:
					color = colorBool
				}
			}

			if color == "" {
				_, _ = w.Write(b)
			} else {
				fmt.Fprintf(w, literal_5147, color, b)
			}

			if isKey {
				fmt.Fprintf(w, "\x1b[%sm:\x1b[m ", colorDelim)
				continue
			}
		}

		if dec.More() {
			fmt.Fprintf(w, "\x1b[%sm,\x1b[m\n%s", colorDelim, strings.Repeat(indent, len(stack)))
		} else if len(stack) > 0 {
			fmt.Fprint(w, "\n", strings.Repeat(indent, len(stack)-1))
		} else {
			fmt.Fprint(w, "\n")
		}
	}

	return nil
}

// WriteDelims writes delims in color and with the appropriate indent
// based on the stack size returned from an io.Writer that implements JsonWriter.Preface().
func WriteDelims(w io.Writer, delims, indent string) error {
	var stack []json.Delim
	if jaw, ok := w.(JsonWriter); ok {
		stack = jaw.Preface()
	}

	fmt.Fprintf(w, literal_5147, colorDelim, delims)
	fmt.Fprint(w, "\n", strings.Repeat(indent, len(stack)))

	return nil
}

// marshalJSON works like json.Marshal but with HTML-escaping disabled
func marshalJSON(v interface{}) ([]byte, error) {
	buf := bytes.Buffer{}
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	bb := buf.Bytes()
	// omit trailing newline added by json.Encoder
	if len(bb) > 0 && bb[len(bb)-1] == '\n' {
		return bb[:len(bb)-1], nil
	}
	return bb, nil
}

const literal_5147 = "\x1b[%sm%s\x1b[m"
