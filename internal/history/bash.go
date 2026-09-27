package history

import (
	"os"
	"path/filepath"
)

type bashFormat struct{}

func (bashFormat) Name() string { return "bash" }

func (bashFormat) DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".bash_history")
}

// Parse handles plain history (one command per line, "\" continues) and
// HISTTIMEFORMAT history, where "#<unix seconds>" precedes each command and
// everything up to the next timestamp belongs to it.
func (bashFormat) Parse(data string) []Entry {
	b := newBuilder(data)
	timed := false
	for line := range lines(data) {
		if ts, ok := parseBashTimestamp(line); ok {
			b.flush()
			b.ts = ts
			timed = true
			continue
		}
		b.add(line)
		if !timed && !endsWithBackslash(line) {
			b.flush()
		}
	}
	return b.done()
}

func isBashTimestamp(line string) bool {
	_, ok := parseBashTimestamp(line)
	return ok
}

// parseBashTimestamp reads "#1700000000" (also tolerating "# 1700000000").
func parseBashTimestamp(line string) (int64, bool) {
	if len(line) < 10 || line[0] != '#' {
		return 0, false
	}
	s := line[1:]
	if s[0] == ' ' {
		s = s[1:]
	}
	ts, end, ok := digits(s, 0)
	return ts, ok && end == len(s) && end >= 9
}
