package history

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type powershellFormat struct{}

func (powershellFormat) Name() string { return "powershell" }

func (powershellFormat) DefaultPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"),
			"Microsoft", "Windows", "PowerShell", "PSReadLine", "ConsoleHost_history.txt")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local/share/powershell/PSReadLine/ConsoleHost_history.txt")
}

// Parse reads PSReadLine history: one command per line, with a trailing
// backtick on every line of a multi-line command except the last.
// There are no timestamps.
func (powershellFormat) Parse(data string) []Entry {
	b := newBuilder(data)
	for line := range lines(data) {
		if cont, ok := strings.CutSuffix(line, "`"); ok {
			b.add(cont)
			continue
		}
		b.add(line)
		b.flush()
	}
	return b.done()
}
