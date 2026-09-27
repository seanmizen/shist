// Package history parses shell history files (zsh, bash, fish, PowerShell)
// into a common list of entries.
package history

// Entry is one command from a history file.
type Entry struct {
	Index     int      // 1-based position in the file
	Timestamp int64    // UNIX seconds; 0 if the shell didn't record one
	Elapsed   int64    // seconds the command ran (zsh only)
	Command   string   // the command, with multi-line commands joined by spaces
	Lines     []string // the raw lines; only set for multi-line commands
}
