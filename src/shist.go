package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	green  = "\033[0;32m"
	yellow = "\033[1;33m"
	reset  = "\033[0m"
)

type entry struct {
	index     int
	timestamp int64
	elapsed   int64
	command   string
	raw       string
}

func main() {
	// flags
	n := flag.Int("n", -1, "Number of history items to show (-1 for all)")
	file := flag.String("file", defaultHistoryFile(), "History file to read")
	minDate := flag.String("min-date", "", "Minimum date (YYYY-MM-DD, YYYY-MM-DD HH:MM, or 10-digit UNIX timestamp)")
	maxDate := flag.String("max-date", "", "Maximum date (YYYY-MM-DD, YYYY-MM-DD HH:MM, or 10-digit UNIX timestamp)")
	minIndex := flag.Int("min-index", -1, "Minimum index (inclusive)")
	maxIndex := flag.Int("max-index", -1, "Maximum index (inclusive)")
	noColor := flag.Bool("no-color", false, "Disable coloured output")

	dateFormat := flag.String("date-format", "2006-01-02 15:04", "Go time layout for the timestamp.\nYou must use Golang's Magical Reference Date: Mon Jan 2 15:04:05 MST 2006")
	outputFormat := flag.String("format", "%d | %i | %c", "Output template (%d=date, %t=timestamp, %i=index, %e=elapsed, %c=command)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Shist - Sean's History Tool

Usage:
	shist [options]

Options:

`)
		flag.PrintDefaults()

		fmt.Fprintln(os.Stderr, `
Examples:
	shist --n 20 --format "%i %d %es - %c"
	shist --min-date "2025-04-01" --date-format "15:04" --format "%d | %c"
	shist --no-color --format "%i:%c"

Sample output (default format):

` + green + `2025-04-22 21:57` + reset + ` | ` + yellow + `123` + reset + ` | git status
` + green + `2025-04-22 21:58` + reset + ` | ` + yellow + `124` + reset + ` | shist --help
	`)
	}
	flag.Parse()

	lines, err := readLines(*file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	total := len(lines)
	if *n < 0 || *n > total {
		*n = total
	}
	start := total - *n
	entries := parseEntries(lines[start:], start)

	// optional time filters
	var minTime, maxTime time.Time
	if *minDate != "" {
		minTime, err = parseDate(*minDate)
		if err != nil {
			die("invalid --min-date", err)
		}
	}
	if *maxDate != "" {
		maxTime, err = parseDate(*maxDate)
		if err != nil {
			die("invalid --max-date", err)
		}
	}

	color := func(s, c string) string {
		if *noColor {
			return s
		}
		return c + s + reset
	}

	for _, e := range entries {
		tm := time.Unix(e.timestamp, 0)

		// filters
		if !minTime.IsZero() && tm.Before(minTime) {
			continue
		}
		if !maxTime.IsZero() && tm.After(maxTime) {
			continue
		}
		if *minIndex > 0 && e.index < *minIndex {
			continue
		}
		if *maxIndex > 0 && e.index > *maxIndex {
			continue
		}

		// render
		out := *outputFormat
		dateStr := ""
		tsStr := ""
		if e.timestamp != 0 {
			dateStr = tm.Format(*dateFormat)
			tsStr = strconv.FormatInt(e.timestamp, 10)
		}
		out = strings.ReplaceAll(out, "%d", color(dateStr, green))
		out = strings.ReplaceAll(out, "%t", color(tsStr, green))
		out = strings.ReplaceAll(out, "%i", color(strconv.Itoa(e.index), yellow))
		out = strings.ReplaceAll(out, "%e", strconv.FormatInt(e.elapsed, 10))
		out = strings.ReplaceAll(out, "%c", e.command)

		fmt.Println(out)
	}
}

func defaultHistoryFile() string {
	u, err := user.Current()
	if err != nil {
		return os.Getenv("HOME") + "/.zsh_history"
	}
	switch filepath.Base(os.Getenv("SHELL")) {
	case "bash":
		return filepath.Join(u.HomeDir, ".bash_history")
	case "zsh":
		return filepath.Join(u.HomeDir, ".zsh_history")
	default:
		return filepath.Join(u.HomeDir, ".zsh_history")
	}
}

func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines, sc.Err()
}

func parseEntries(lines []string, offset int) []entry {
	// matches both zsh extended_history and raw bash lines
	re := regexp.MustCompile(`^: (\d+):(\d+);(.*)$`)
	var entries []entry

	for i := len(lines) - 1; i >= 0; i-- {
		ix := offset + i + 1
		line := lines[i]

		if m := re.FindStringSubmatch(line); m != nil {
			ts, _ := strconv.ParseInt(m[1], 10, 64)
			elapsed, _ := strconv.ParseInt(m[2], 10, 64)
			cmd := m[3]
			entries = append(entries, entry{index: ix, timestamp: ts, elapsed: elapsed, command: cmd})
		} else {
			entries = append(entries, entry{index: ix, raw: line, command: line})
		}
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].index < entries[j].index })
	return entries
}

func parseDate(input string) (time.Time, error) {
	if ts, err := strconv.ParseInt(input, 10, 64); err == nil {
		return time.Unix(ts, 0), nil
	}
	layouts := []string{"2006-01-02 15:04", "2006-01-02"}
	for _, l := range layouts {
		if t, err := time.Parse(l, input); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date format: %s", input)
}

func die(msg string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", msg, err)
	os.Exit(1)
}
