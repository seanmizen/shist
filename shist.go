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
	command   string
	raw       string
}

func main() {
	n := flag.Int("n", -1, "number of lines to show (-1 for all)")
	file := flag.String("file", defaultHistoryFile(), "history file to read")
	minDate := flag.String("minDate", "", "minimum date (YYYY-MM-DD, YYYY-MM-DD HH:MM, or 10-digit UNIX timestamp)")
	maxDate := flag.String("maxDate", "", "maximum date (YYYY-MM-DD, YYYY-MM-DD HH:MM, or 10-digit UNIX timestamp)")
	minIndex := flag.Int("minIndex", -1, "minimum index (inclusive)")
	maxIndex := flag.Int("maxIndex", -1, "maximum index (inclusive)")
	noColor := flag.Bool("no-color", false, "disable colored output")

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
	selected := lines[start:]

	entries := parseEntries(selected, start)

	// ⛔️ Filtering stubs (not implemented yet)
	var minTime, maxTime time.Time
	if *minDate != "" {
		t, err := parseDate(*minDate)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid --minDate: %v\n", err)
			os.Exit(1)
		}
		minTime = t
	}
	if *maxDate != "" {
		t, err := parseDate(*maxDate)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid --maxDate: %v\n", err)
			os.Exit(1)
		}
		maxTime = t
	}

	for _, e := range entries {
		// ⛏️ Apply filters (to be filled in)
		if !minTime.IsZero() && time.Unix(e.timestamp, 0).Before(minTime) {
			continue
		}
		if !maxTime.IsZero() && time.Unix(e.timestamp, 0).After(maxTime) {
			continue
		}
		if *minIndex > 0 && e.index < *minIndex {
			continue
		}
		if *maxIndex > 0 && e.index > *maxIndex {
			continue
		}

		// 🖨️ Output
		color := func(s, c string) string {
			if *noColor {
				return s
			}
			return c + s + reset
		}

		if e.timestamp != 0 {
			dt := time.Unix(e.timestamp, 0).Format("2006-01-02 15:04")
			fmt.Printf("%s | %s\t| %s\n",
				color(dt, green),
				color(fmt.Sprintf("%d", e.index), yellow),
				e.command,
			)
		} else {
			fmt.Printf("\t | %s\t| %s\n", color(fmt.Sprintf("%d", e.index), yellow), e.raw)
		}
	}
}

func defaultHistoryFile() string {
  u, err := user.Current()
  if err != nil {
		return os.Getenv("HOME") + "/.zsh_history"
	}
	shell := filepath.Base(os.Getenv("SHELL"))
	switch shell {
	case "bash":
		return filepath.Join(u.HomeDir, ".bash_history")
	case "zsh":
		return filepath.Join(u.HomeDir, ".zsh_history")
	default:
		return filepath.Join(u.HomeDir, ".zsh_history")
	}
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	sc := bufio.NewScanner(file)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines, sc.Err()
}

func parseEntries(lines []string, offset int) []entry {
	re := regexp.MustCompile(`^: (\d+):0;(.*)$`)
	var entries []entry
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		ix := offset + i + 1
		if matches := re.FindStringSubmatch(line); matches != nil {
			ts, _ := strconv.ParseInt(matches[1], 10, 64)
			cmd := matches[2]
			entries = append(entries, entry{index: ix, timestamp: ts, command: cmd})
		} else {
			entries = append(entries, entry{index: ix, raw: line})
		}
	}
	// Oldest first
	sort.Slice(entries, func(i, j int) bool { return entries[i].index < entries[j].index })
	return entries
}

func parseDate(input string) (time.Time, error) {
	// attempt 1: UNIX timestamp
	if ts, err := strconv.ParseInt(input, 10, 64); err == nil {
		return time.Unix(ts, 0), nil
	}

	// attempt 2: ISO format
	layouts := []string{"2006-01-02 15:04", "2006-01-02"}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, input); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid date format: %s", input)
}
