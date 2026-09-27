package history

import (
	"os"
	"path/filepath"
	"strings"
)

type zshFormat struct{}

func (zshFormat) Name() string { return "zsh" }

func (zshFormat) DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".zsh_history")
}

// Parse handles both EXTENDED_HISTORY (": <ts>:<elapsed>;<cmd>") and plain
// lines. Multi-line commands continue while a line ends in a backslash.
func (zshFormat) Parse(data string) []Entry {
	b := newBuilder(data)
	continuing := false
	for line := range lines(data) {
		line = unmetafy(line)
		if ts, el, cmd, ok := parseZshHeader(line); ok {
			b.flush()
			b.ts, b.el = ts, el
			line = cmd
		} else if !continuing {
			// plain line (no EXTENDED_HISTORY): don't inherit the last timestamp
			b.flush()
			b.ts, b.el = 0, 0
		}
		b.add(line)
		continuing = endsWithBackslash(line)
		if !continuing {
			b.flush()
		}
	}
	return b.done()
}

func isZshHeader(line string) bool {
	_, _, _, ok := parseZshHeader(line)
	return ok
}

// parseZshHeader splits ": 1700000000:0;cmd" into its parts.
func parseZshHeader(line string) (ts, el int64, cmd string, ok bool) {
	if len(line) < 6 || line[0] != ':' || line[1] != ' ' {
		return 0, 0, "", false
	}
	ts, i, ok := digits(line, 2)
	if !ok || i >= len(line) || line[i] != ':' {
		return 0, 0, "", false
	}
	el, i, ok = digits(line, i+1)
	if !ok || i >= len(line) || line[i] != ';' {
		return 0, 0, "", false
	}
	return ts, el, line[i+1:], true
}

// zshMeta marks a byte zsh has escaped in its history file; the byte after
// it is the original XOR 32. Non-ASCII text (e.g. emoji) is stored this way.
const zshMeta = 0x83

func unmetafy(s string) string {
	i := strings.IndexByte(s, zshMeta)
	if i < 0 {
		return s
	}
	b := make([]byte, 0, len(s))
	b = append(b, s[:i]...)
	for ; i < len(s); i++ {
		c := s[i]
		if c == zshMeta && i+1 < len(s) {
			i++
			c = s[i] ^ 32
		}
		b = append(b, c)
	}
	return string(b)
}
