// Targets Linux and Darwin (MacOS, BSD).
// Supports zsh, bash, and fish.
package nix

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"shist/src/model"
	"shist/src/ui"

	"gopkg.in/yaml.v3"
)

/* ---------- shared interface ---------- */

type reader interface {
	DefaultPath() string
	Read(path string) ([]model.Entry, error)
}

/* ---------- public entry-point ---------- */

func Run() {
	/* ---------- CLI flags ---------- */
	n := flag.Int("n", -1, "Number of history items to show (-1 for all)")
	file := flag.String("file", "", "History file to read (auto-detected if empty)")
	minDate := flag.String("min-date", "", "Minimum date (YYYY-MM-DD, YYYY-MM-DD HH:MM, or UNIX seconds)")
	maxDate := flag.String("max-date", "", "Maximum date (YYYY-MM-DD, YYYY-MM-DD HH:MM, or UNIX seconds)")
	minIndex := flag.Int("min-index", -1, "Minimum index (inclusive)")
	maxIndex := flag.Int("max-index", -1, "Maximum index (inclusive)")
	noColor := flag.Bool("no-color", false, "Disable coloured output. Overrides color directives.")
	var concatMultiline bool
	flag.BoolVar(&concatMultiline, "concat-multiline", false, "Concat multiline commands to one line")
	flag.BoolVar(&concatMultiline, "c", false, "")
	var grepPattern string
	flag.StringVar(&grepPattern, "grep", "", "Only show entries matching this pattern (before filters)")
	flag.StringVar(&grepPattern, "g", "", "") // this prints a whole line whereas BoolVar doesn't. Poor lib design.

	dateFmt := flag.String("date-format", "2006-01-02 15:04",
		"Go time layout for the timestamp.\nYou must use Golang's Magical Reference Date: Mon Jan 2 15:04:05 MST 2006")
	outFmt := flag.String("format", "%C(green)%d%C(reset) | %C(yellow)%i%C(reset) | %c",
		"Output template (%d=date, %t=timestamp, %i=index, %e=elapsed, %c=command)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `
shist - Sean's History Tool

Usage:
	shist [options]

Options:
`)
		flag.PrintDefaults()
		const examples = `
Examples:
	shist --n 20 --format "%i %d %es - %c"
	shist --min-date "2025-04-01" --date-format "15:04" --format "%d | %c"
	shist --no-color
	shist --format "%i:%c"
	shist --format "%C(red)%i:%c%C(reset)"
	shist -g "foo" -n 20	# this greps before filtering 20 items. It gives you 20 matches!
	shist -format "%c" -g brew	# print just the command, grep 'brew'
	shist -format "%c" -c -g brew	# print just the command, grep 'brew', concatenate multiline commands into one.

shist uses git log-style color directives . %C(red), %C(fed7b0), %C(reset)
named colors are: black, red, green, yellow, blue, magenta, cyan, white.

`
		fmt.Fprintf(os.Stderr, "%s", examples)
	}

	flag.Parse()

	/* ---------- shell detection ---------- */
	r := pickReader()
	if r == nil {
		fmt.Fprintln(os.Stderr, "unsupported shell / platform")
		os.Exit(1)
	}

	histFile := *file
	if histFile == "" {
		histFile = r.DefaultPath()
	}
	entries, err := r.Read(histFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read error: %v\n", err)
		os.Exit(1)
	}

	/* grep first (cheap-ens downstream work) */
	if grepPattern != "" {
		pat, err := regexp.Compile(grepPattern)
		must(err)
		entries = grepEntries(entries, pat)
	}

	/* newest-first → newest-N */
	if *n > 0 && *n < len(entries) {
		entries = entries[len(entries)-*n:]
	}

	/* ---------- optional filters ---------- */
	var minTime, maxTime time.Time
	if *minDate != "" {
		minTime, err = parseDate(*minDate)
		must(err)
	}
	if *maxDate != "" {
		maxTime, err = parseDate(*maxDate)
		must(err)
	}

	if *noColor {
		ui.Disable()
	}

	for _, e := range entries { // oldest → newest
		t := time.Unix(e.Timestamp, 0)
		if (!minTime.IsZero() && t.Before(minTime)) ||
			(!maxTime.IsZero() && t.After(maxTime)) ||
			(*minIndex > 0 && e.Index < *minIndex) ||
			(*maxIndex > 0 && e.Index > *maxIndex) {
			continue
		}
		printEntry(e, *dateFmt, *outFmt, concatMultiline)
	}
}

func grepEntries(entries []model.Entry, pattern *regexp.Regexp) []model.Entry {
	out := entries[:0]
	for _, e := range entries {
		if pattern.MatchString(e.Command) {
			out = append(out, e)
		}
	}
	return out
}

/* ---------- reader selection ---------- */

func pickReader() reader {
	if runtime.GOOS == "windows" {
		return nil
	}
	switch filepath.Base(os.Getenv("SHELL")) {
	case "bash":
		return bashReader{}
	case "fish":
		return fishReader{}
	case "zsh":
		fallthrough
	default:
		return zshReader{}
	}
}

/* ---------- printers ---------- */

func formatMultiline(lines []string) string {
	if len(lines) == 0 {
			return ""
	}
	out := make([]string, len(lines))
	for i, l := range lines {
			t := strings.TrimSpace(l)
			t = strings.TrimRight(t, "\\")   // drop all original backslashes
			if i == 0 {
					out[i] = t + " \\"           // first line
			} else {
					out[i] = "    " + t + " \\"  // indented follow-ups
			}
	}
	// last line: remove the trailing backslash we just added
	out[len(out)-1] = strings.TrimSuffix(out[len(out)-1], " \\")
	return strings.Join(out, "\n")
}

func printEntry(e model.Entry, dateFmt, outFmt string, concatMultiline bool) {
	dateStr, tsStr := "", ""
	if e.Timestamp != 0 {
		t := time.Unix(e.Timestamp, 0)
		dateStr = t.Format(dateFmt)
		tsStr = strconv.FormatInt(e.Timestamp, 10)
	}

	out := ui.ExpandColors(outFmt)

	cmd := e.Command
	if !concatMultiline && len(e.Lines) > 1 {
		cmd = formatMultiline(e.Lines)
	}

	out = strings.ReplaceAll(out, "%d", dateStr)
	out = strings.ReplaceAll(out, "%t", tsStr)
	out = strings.ReplaceAll(out, "%i", strconv.Itoa(e.Index))
	out = strings.ReplaceAll(out, "%e", strconv.FormatInt(e.Elapsed, 10))
	out = strings.ReplaceAll(out, "%c", cmd)

	fmt.Println(out)
}

/* ---------- helpers ---------- */

func parseDate(s string) (time.Time, error) {
	if ts, err := strconv.ParseInt(s, 10, 64); err == nil {
		return time.Unix(ts, 0), nil
	}
	for _, layout := range []string{"2006-01-02 15:04", "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date: %s", s)
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

/* ---------- ZSH reader ---------- */

type zshReader struct{}

func (zshReader) DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".zsh_history")
}

var zshRe = regexp.MustCompile(`^: (\d+):(\d+);(.*)$`)

func (zshReader) Read(path string) ([]model.Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var (
		entries      []model.Entry
		current      strings.Builder
		lines        []string
		ts, el int64
		idx     = 0
	)

	flush := func() {
		cmd := strings.TrimSpace(current.String())
		if cmd == "" {
			return
		}
		idx++
		entries = append(entries, model.Entry{
			Index:     idx,
			Timestamp: ts,
			Elapsed:   el,
			Command:   cmd,
			Lines:     append([]string{}, lines...),
		})
		current.Reset()
		lines = lines[:0]
	}

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if m := zshRe.FindStringSubmatch(line); m != nil {
			flush()
			ts, _ = strconv.ParseInt(m[1], 10, 64)
			el, _ = strconv.ParseInt(m[2], 10, 64)
			line = m[3]
		}

		lines = append(lines, line)

		trimmed := strings.TrimSpace(line)
		if strings.HasSuffix(trimmed, "\\") {
				current.WriteString(strings.TrimRight(trimmed, "\\"))
				current.WriteString(" ")
		} else {
				current.WriteString(trimmed)
				flush()
		}
			}
	flush()

	return entries, sc.Err()
}

/* ---------- Bash reader ---------- */

type bashReader struct{}

func (bashReader) DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".bash_history")
}

var bashTS = regexp.MustCompile(`^# (\d{10})$`)

func (bashReader) Read(path string) ([]model.Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var (
		entries      []model.Entry
		current      strings.Builder
		lines        []string
		ts     int64
		idx    = 0
	)

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if m := bashTS.FindStringSubmatch(line); m != nil {
			// timestamp line: next lines belong to new command
			if current.Len() != 0 {
				idx++
				entries = append(entries, model.Entry{
					Index:     idx,
					Timestamp: ts,
					Command:   strings.TrimSpace(current.String()),
					Lines:     append([]string{}, lines...),
				})
				current.Reset()
				lines = lines[:0]
			}
			ts, _ = strconv.ParseInt(m[1], 10, 64)
			continue
		}

		lines = append(lines, line)
		trimmed := strings.TrimSpace(line)
		if strings.HasSuffix(trimmed, "\\") {
			current.WriteString(strings.TrimSuffix(trimmed, "\\"))
			current.WriteString(" ")
		} else {
			current.WriteString(trimmed + " ")
		}
	}

	if current.Len() != 0 {
		idx++
		entries = append(entries, model.Entry{
			Index:     idx,
			Timestamp: ts,
			Command:   strings.TrimSpace(current.String()),
			Lines:     append([]string{}, lines...),
		})
	}

	return entries, sc.Err()
}

/* ---------- Fish reader ---------- */

type fishReader struct{}

func (fishReader) DefaultPath() string {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".local/share/fish/fish_history") // ≥ 2.3.0
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return filepath.Join(home, ".config/fish/fish_history") // < 2.3.0
}

type fishItem struct {
	Cmd  string `yaml:"cmd"`
	When int64  `yaml:"when"`
}

func (fishReader) Read(path string) ([]model.Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var items []fishItem
	dec := yaml.NewDecoder(f)
	for {
		var it fishItem
		if err := dec.Decode(&it); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		items = append(items, it)
	}

	entries := make([]model.Entry, len(items))
	for i, it := range items {
		entries[i] = model.Entry{
			Index:     i + 1,
			Timestamp: it.When,
			Command:   strings.TrimSpace(it.Cmd),
			Lines:     []string{strings.TrimSpace(it.Cmd)},
		}
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Index < entries[j].Index })
	return entries, nil
}
