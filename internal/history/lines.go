package history

import (
	"iter"
	"strings"
)

// lines yields each line of data without its line ending. Every line is a
// substring of data, so iterating allocates nothing and has no length cap.
func lines(data string) iter.Seq[string] {
	return func(yield func(string) bool) {
		data := strings.TrimPrefix(data, "\ufeff") // UTF-8 BOM
		for len(data) > 0 {
			var line string
			if i := strings.IndexByte(data, '\n'); i >= 0 {
				line, data = data[:i], data[i+1:]
			} else {
				line, data = data, ""
			}
			if !yield(strings.TrimSuffix(line, "\r")) {
				return
			}
		}
	}
}

// builder gathers the raw lines of one command and turns them into entries.
type builder struct {
	entries []Entry
	lines   []string
	ts, el  int64
}

func newBuilder(data string) *builder {
	// one entry per line is a good upper bound; avoids regrowing the slice
	return &builder{entries: make([]Entry, 0, strings.Count(data, "\n")+1)}
}

func (b *builder) add(line string) { b.lines = append(b.lines, line) }

func (b *builder) flush() {
	switch len(b.lines) {
	case 0:
		return
	case 1:
		if cmd := strings.TrimSpace(b.lines[0]); cmd != "" {
			b.push(cmd, nil)
		}
	default:
		parts := make([]string, 0, len(b.lines))
		for i, l := range b.lines {
			t := strings.TrimSpace(l)
			if i < len(b.lines)-1 { // continuation backslash; keep a real one like `cd C:\`
				t = strings.TrimSpace(strings.TrimRight(t, "\\"))
			}
			if t != "" {
				parts = append(parts, t)
			}
		}
		if len(parts) > 0 {
			b.push(strings.Join(parts, " "), append([]string{}, b.lines...))
		}
	}
	b.lines = b.lines[:0]
}

func (b *builder) push(cmd string, raw []string) {
	b.entries = append(b.entries, Entry{
		Index:     len(b.entries) + 1,
		Timestamp: b.ts,
		Elapsed:   b.el,
		Command:   cmd,
		Lines:     raw,
	})
}

func (b *builder) done() []Entry {
	b.flush()
	return b.entries
}

func endsWithBackslash(line string) bool {
	return strings.HasSuffix(strings.TrimSpace(line), "\\")
}

// digits parses the run of ASCII digits starting at s[i].
func digits(s string, i int) (n int64, end int, ok bool) {
	start := i
	for i < len(s) && s[i] >= '0' && s[i] <= '9' && i-start < 18 {
		n = n*10 + int64(s[i]-'0')
		i++
	}
	return n, i, i > start
}
