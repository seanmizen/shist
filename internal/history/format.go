package history

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Format is one shell's history file format.
type Format interface {
	Name() string
	DefaultPath() string
	// Parse reads a whole history file. It never fails: lines it can't
	// make sense of are kept as plain commands.
	Parse(data string) []Entry
}

// Formats lists every supported format.
var Formats = []Format{Zsh, Bash, Fish, PowerShell}

var (
	Zsh        Format = zshFormat{}
	Bash       Format = bashFormat{}
	Fish       Format = fishFormat{}
	PowerShell Format = powershellFormat{}
)

// sniffBytes is how much of the file Detect looks at.
const sniffBytes = 64 * 1024

// Detect picks a format by sniffing the start of the history itself,
// falling back to the file name, then to $SHELL.
func Detect(data, path string) Format {
	head := data[:min(len(data), sniffBytes)]
	for line := range lines(head) {
		switch {
		case strings.HasPrefix(line, "- cmd:"):
			return Fish
		case isZshHeader(line):
			return Zsh
		case isBashTimestamp(line):
			return Bash
		}
	}
	if f := forName(filepath.Base(path)); f != nil {
		return f
	}
	return ForShell(os.Getenv("SHELL"))
}

func forName(name string) Format {
	name = strings.ToLower(name)
	switch {
	case strings.Contains(name, "zsh"):
		return Zsh
	case strings.Contains(name, "bash"):
		return Bash
	case strings.Contains(name, "fish"):
		return Fish
	case strings.Contains(name, "consolehost_history"):
		return PowerShell
	}
	return nil
}

// ForShell guesses from a $SHELL value. That's the login shell, not
// necessarily the current one, so it's only used for the default path and as
// a last resort.
func ForShell(shell string) Format {
	if runtime.GOOS == "windows" {
		return PowerShell
	}
	switch filepath.Base(shell) {
	case "bash":
		return Bash
	case "fish":
		return Fish
	case "pwsh", "powershell":
		return PowerShell
	default:
		return Zsh
	}
}

// Load reads and parses the history file at path.
func Load(path string) ([]Entry, Format, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("reading history: %w", err)
	}
	s := string(data)
	f := Detect(s, path)
	return f.Parse(s), f, nil
}
