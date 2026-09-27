// Package format renders history entries through a user template such as
// "%C(green)%d%C(reset) | %c".
package format

import (
	"strconv"
	"strings"
	"time"

	"github.com/seanmizen/shist/internal/history"
)

type kind uint8

const (
	literal kind = iota
	date         // %d
	unix         // %t
	index        // %i
	elapsed      // %e
	command      // %c
)

type part struct {
	kind kind
	text string // for literal
}

// Template is a parsed output template. Parse it once, render many entries.
type Template struct {
	parts      []part
	dateLayout string
	concat     bool
}

// Options controls how a template renders.
type Options struct {
	DateLayout string // Go time layout for %d
	Color      bool   // expand %C(…) directives; otherwise drop them
	Concat     bool   // print multi-line commands on one line
}

// Compile parses tmpl. Tokens: %d date, %t UNIX timestamp, %i index,
// %e elapsed seconds, %c command, %% a literal %, and %C(spec) colours
// (see sgr). Anything else is printed as-is.
func Compile(tmpl string, opt Options) *Template {
	t := &Template{dateLayout: opt.DateLayout, concat: opt.Concat}
	var lit strings.Builder
	emit := func(k kind) {
		if lit.Len() > 0 {
			t.parts = append(t.parts, part{kind: literal, text: lit.String()})
			lit.Reset()
		}
		if k != literal {
			t.parts = append(t.parts, part{kind: k})
		}
	}
	for i := 0; i < len(tmpl); i++ {
		if tmpl[i] != '%' || i+1 == len(tmpl) {
			lit.WriteByte(tmpl[i])
			continue
		}
		switch tmpl[i+1] {
		case 'd':
			emit(date)
		case 't':
			emit(unix)
		case 'i':
			emit(index)
		case 'e':
			emit(elapsed)
		case 'c':
			emit(command)
		case '%':
			lit.WriteByte('%')
		case 'C':
			spec, ok := strings.CutPrefix(tmpl[i+2:], "(")
			end := strings.IndexByte(spec, ')')
			if !ok || end < 0 {
				lit.WriteByte('%')
				continue
			}
			if opt.Color {
				lit.WriteString(sgr(spec[:end]))
			}
			i += 2 + end + 1 // skip "C(spec)"; the loop's i++ skips ")"
			continue
		default:
			lit.WriteByte('%')
			continue
		}
		i++
	}
	emit(literal)
	return t
}

// Append renders e onto buf, followed by a newline.
func (t *Template) Append(buf []byte, e history.Entry) []byte {
	for _, p := range t.parts {
		switch p.kind {
		case literal:
			buf = append(buf, p.text...)
		case date:
			if e.Timestamp != 0 {
				buf = time.Unix(e.Timestamp, 0).AppendFormat(buf, t.dateLayout)
			}
		case unix:
			if e.Timestamp != 0 {
				buf = strconv.AppendInt(buf, e.Timestamp, 10)
			}
		case index:
			buf = strconv.AppendInt(buf, int64(e.Index), 10)
		case elapsed:
			buf = strconv.AppendInt(buf, e.Elapsed, 10)
		case command:
			if t.concat || len(e.Lines) < 2 {
				buf = append(buf, e.Command...)
			} else {
				buf = appendMultiline(buf, e.Lines)
			}
		}
	}
	return append(buf, '\n')
}

// appendMultiline prints each line on its own row, indenting follow-ups and
// ending all but the last with " \".
func appendMultiline(buf []byte, lines []string) []byte {
	for i, l := range lines {
		if i > 0 {
			buf = append(buf, "\n    "...)
		}
		buf = append(buf, strings.TrimSpace(strings.TrimRight(strings.TrimSpace(l), "\\"))...)
		if i < len(lines)-1 {
			buf = append(buf, " \\"...)
		}
	}
	return buf
}
