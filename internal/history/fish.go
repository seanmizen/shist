package history

import (
	"os"
	"path/filepath"
	"strings"
)

type fishFormat struct{}

func (fishFormat) Name() string { return "fish" }

func (fishFormat) DefaultPath() string {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".local/share/fish/fish_history") // ≥ 2.3.0
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return filepath.Join(home, ".config/fish/fish_history") // < 2.3.0
}

// Parse reads fish_history. It looks like YAML but isn't (e.g.
// "- cmd: echo a: b" is invalid YAML), so only the keys fish actually writes
// are read, by hand.
func (fishFormat) Parse(data string) []Entry {
	b := newBuilder(data)
	for line := range lines(data) {
		if cmd, ok := strings.CutPrefix(line, "- cmd:"); ok {
			b.flush()
			b.ts = 0
			cmd = fishUnescape(strings.TrimPrefix(cmd, " "))
			for l := range lines(cmd) {
				b.add(l)
			}
		} else if when, ok := strings.CutPrefix(line, "  when:"); ok {
			b.ts, _, _ = digits(strings.TrimSpace(when), 0)
		}
	}
	return b.done()
}

// fishUnescape reverses fish's history escaping: "\\" -> "\" and "\n" -> newline.
func fishUnescape(s string) string {
	if !strings.Contains(s, `\`) {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case '\\':
				b.WriteByte('\\')
				i++
				continue
			case 'n':
				b.WriteByte('\n')
				i++
				continue
			}
		}
		b.WriteByte(s[i])
	}
	return b.String()
}
