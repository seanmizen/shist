package ui

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"golang.org/x/term"
)

/* ---------- public API ---------- */

// Disable turns colour off globally (used by the --no-color flag).
func Disable() { disabled = true }

// ExpandColors replaces every %C(<spec>) token in s with an ANSI
// sequence (if colours are enabled) or removes it (if not).
// <spec> may be a named colour (red, green, …) or a hex literal #rrggbb.
func ExpandColors(s string) string {
	return colorRe.ReplaceAllStringFunc(s, func(m string) string {
		if disabled || !supports() {
			return ""
		}
		spec := colorRe.FindStringSubmatch(m)[1]
		if spec == "reset" {
			return "\033[0m"
		}
		return ansi(spec)
	})
}

/* ---------- internals ---------- */

var (
	disabled bool

	colorRe = regexp.MustCompile(`%C\(([^)]+)\)`) // %C(foo)
	hexRe   = regexp.MustCompile(`^#?([0-9a-fA-F]{6})$`)

	named = map[string]string{ // 8-colour palette
		"black": "30", "red": "31", "green": "32", "yellow": "33",
		"blue": "34", "magenta": "35", "cyan": "36", "white": "37",
	}
)

// ansi converts a colour spec to "CSI … m".
func ansi(spec string) string {
	spec = strings.ToLower(strings.TrimSpace(spec))

	// named 8-colour
	if code, ok := named[spec]; ok {
		return "\033[" + code + "m"
	}

	// 24-bit #rrggbb
	if m := hexRe.FindStringSubmatch(spec); m != nil {
		r, _ := strconv.ParseUint(m[1][0:2], 16, 8)
		g, _ := strconv.ParseUint(m[1][2:4], 16, 8)
		b, _ := strconv.ParseUint(m[1][4:6], 16, 8)
		return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
	}

	return "" // unknown → no colour
}

func supports() bool {
	if !term.IsTerminal(int(os.Stderr.Fd())) {
		return false
	}
	if runtime.GOOS != "windows" {
		return true
	}
	// heuristics for modern Windows terminals
	return os.Getenv("WT_SESSION") != "" ||
		strings.Contains(strings.ToLower(os.Getenv("TERM_PROGRAM")), "vscode") ||
		os.Getenv("ANSICON") != "" ||
		os.Getenv("ConEmuANSI") == "ON"
}
