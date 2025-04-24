package ui

import (
	"os"
	"runtime"
	"strings"

	"golang.org/x/term"
)

const (
	green  = "\033[0;32m"
	yellow = "\033[1;33m"
	reset  = "\033[0m"
)

var disabled bool

func Disable() { disabled = true }

func Colorize(s, color string) string {
	if disabled || !supportsColor() {
		return s
	}
	return color + s + reset
}

func Green(s string) string  { return Colorize(s, green) }
func Yellow(s string) string { return Colorize(s, yellow) }

func supportsColor() bool {
	if !term.IsTerminal(int(os.Stderr.Fd())) {
		return false
	}
	if runtime.GOOS != "windows" {
		return true
	}
	// heuristics for modern Windows terminals
	if os.Getenv("WT_SESSION") != "" ||
		strings.Contains(strings.ToLower(os.Getenv("TERM_PROGRAM")), "vscode") ||
		os.Getenv("ANSICON") != "" ||
		os.Getenv("ConEmuANSI") == "ON" {
		return true
	}
	return false
}
