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
	noColor := flag.Bool("no-color", false, "Disable coloured output")

	dateFmt := flag.String("date-format", "2006-01-02 15:04",
		"Go time layout for the timestamp.")
	outFmt := flag.String("format", "%d | %i | %c",
		"Output template (%d=date, %t=timestamp, %i=index, %e=elapsed, %c=command)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `
shist - Sean's History Tool

Usage:
	shist [options]

Options:
`)
		flag.PrintDefaults()
		const examples = `Examples:
	shist --n 20 --format "%i %d %es - %c"
	shist --min-date "2025-04-01" --date-format "15:04" --format "%d | %c"
	shist --no-color --format "%i:%c"
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

	/* newest-first -> newest-N */
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

	for _, e := range entries {
		if !minTime.IsZero() && time.Unix(e.Timestamp, 0).Before(minTime) {
			continue
		}
		if !maxTime.IsZero() && time.Unix(e.Timestamp, 0).After(maxTime) {
			continue
		}
		if *minIndex > 0 && e.Index < *minIndex {
			continue
		}
		if *maxIndex > 0 && e.Index > *maxIndex {
			continue
		}
		printEntry(e, *dateFmt, *outFmt)
	}
}

/* ---------- reader selection ---------- */

func pickReader() reader {
	if runtime.GOOS == "windows" {
		return nil // future: powershell / cmd readers
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

/* ---------- printers / helpers ---------- */

func printEntry(e model.Entry, dateFmt, outFmt string) {
	dateStr, tsStr := "", ""
	if e.Timestamp != 0 {
		tm := time.Unix(e.Timestamp, 0)
		dateStr = tm.Format(dateFmt)
		tsStr = strconv.FormatInt(e.Timestamp, 10)
	}

	out := outFmt
	out = strings.ReplaceAll(out, "%d", ui.Green(dateStr))
	out = strings.ReplaceAll(out, "%t", ui.Green(tsStr))
	out = strings.ReplaceAll(out, "%i", ui.Yellow(strconv.Itoa(e.Index)))
	out = strings.ReplaceAll(out, "%e", strconv.FormatInt(e.Elapsed, 10))
	out = strings.ReplaceAll(out, "%c", e.Command)

	fmt.Println(out)
}

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

	var raw []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		raw = append(raw, sc.Text())
	}

	entries := make([]model.Entry, 0, len(raw))
	for i := len(raw) - 1; i >= 0; i-- {
		idx := len(raw) - i
		line := raw[i]
		if m := zshRe.FindStringSubmatch(line); m != nil {
			ts, _ := strconv.ParseInt(m[1], 10, 64)
			el, _ := strconv.ParseInt(m[2], 10, 64)
			entries = append(entries, model.Entry{
				Index:     idx,
				Timestamp: ts,
				Elapsed:   el,
				Command:   m[3],
			})
		} else {
			entries = append(entries, model.Entry{
				Index:   idx,
				Command: line,
				Raw:     line,
			})
		}
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Index < entries[j].Index })
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

	var entries []model.Entry
	sc := bufio.NewScanner(f)
	index := 0
	for sc.Scan() {
		line := sc.Text()
		if m := bashTS.FindStringSubmatch(line); m != nil {
			if !sc.Scan() {
				break
			}
			cmd := sc.Text()
			ts, _ := strconv.ParseInt(m[1], 10, 64)
			index++
			entries = append(entries, model.Entry{
				Index:     index,
				Timestamp: ts,
				Command:   cmd,
			})
		} else {
			index++
			entries = append(entries, model.Entry{
				Index:   index,
				Command: line,
				Raw:     line,
			})
		}
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Index < entries[j].Index })
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
			Command:   it.Cmd,
		}
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Index < entries[j].Index })
	return entries, nil
}
