package cli

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/seanmizen/shist/internal/history"
)

func TestParseDate(t *testing.T) {
	day := time.Date(2025, 4, 1, 0, 0, 0, 0, time.Local)
	cases := []struct {
		in         string
		start, end time.Time
	}{
		{"2025-04-01", day, day.AddDate(0, 0, 1).Add(-time.Nanosecond)},
		{"2025-04-01 15:04", day.Add(15*time.Hour + 4*time.Minute), day.Add(15*time.Hour + 5*time.Minute - time.Nanosecond)},
		{"1700000000", time.Unix(1700000000, 0), time.Unix(1700000000, 0)},
	}
	for _, c := range cases {
		start, end, err := parseDate(c.in)
		if err != nil || !start.Equal(c.start) || !end.Equal(c.end) {
			t.Errorf("parseDate(%q) = %v, %v, %v; want %v, %v", c.in, start, end, err, c.start, c.end)
		}
	}
	if _, _, err := parseDate("04/01/2025"); err == nil {
		t.Error("want an error for 04/01/2025")
	}
}

func TestFiltersApply(t *testing.T) {
	var entries []history.Entry
	for i := 1; i <= 10; i++ {
		entries = append(entries, history.Entry{Index: i, Timestamp: int64(100 + i), Command: "cmd"})
	}
	got := filters{n: 3, maxIndex: 5}.apply(append([]history.Entry{}, entries...))
	if len(got) != 3 || got[0].Index != 3 || got[2].Index != 5 {
		t.Errorf("n after max-index: got %v", got)
	}
	got = filters{minUnix: 105, hasMin: true, maxUnix: 107, hasMax: true}.apply(append([]history.Entry{}, entries...))
	if len(got) != 3 || got[0].Index != 5 {
		t.Errorf("date range: got %v", got)
	}
}

func TestFiltersWithContext(t *testing.T) {
	var entries []history.Entry
	for i, c := range "..x....x.x....x" { // matches at indexes 3, 8, 10, 15
		entries = append(entries, history.Entry{Index: i + 1, Command: string(c)})
	}
	match := func(s string) bool { return s == "x" }
	// show renders rows as "1 2 [3] | 7 ...": brackets mark matches, | a gap
	show := func(rows []row) string {
		var b strings.Builder
		for _, r := range rows {
			if r.gap {
				b.WriteString("| ")
			}
			if r.context {
				fmt.Fprintf(&b, "%d ", r.Index)
			} else {
				fmt.Fprintf(&b, "[%d] ", r.Index)
			}
		}
		return strings.TrimSpace(b.String())
	}
	cases := []struct {
		f    filters
		want string
	}{
		{filters{before: 1, after: 1}, "2 [3] 4 | 7 [8] 9 [10] 11 | 14 [15]"},
		{filters{before: 2}, "1 2 [3] | 6 7 [8] 9 [10] | 13 14 [15]"},
		{filters{after: 3}, "[3] 4 5 6 | [8] 9 [10] 11 12 13 | [15]"},
		{filters{before: 4, after: 4}, "1 2 [3] 4 5 6 7 [8] 9 [10] 11 12 13 14 [15]"}, // touching windows merge
		{filters{before: 2, after: 2, n: 2}, "8 9 [10] 11 12 13 14 [15]"},             // -n counts matches
		{filters{before: 3, n: 1, minIndex: 15}, "12 13 14 [15]"},                     // neighbours ignore other filters
	}
	for _, c := range cases {
		c.f.match = match
		if got := show(c.f.withContext(entries)); got != c.want {
			t.Errorf("%+v:\n got %s\nwant %s", c.f, got, c.want)
		}
	}
}

func TestMatcher(t *testing.T) {
	cases := []struct {
		pattern, s string
		want       bool
	}{
		{"brew", "brew install go", true}, // literal fast path
		{"brew", "git status", false},
		{"^git (status|log)$", "git log", true}, // real regex
		{"^git (status|log)$", "git push", false},
		{"a.c", "abc", true}, // metacharacter: must not use the literal path
		{"a.c", "a.c", true},
	}
	for _, c := range cases {
		if got := matcher(regexp.MustCompile(c.pattern))(c.s); got != c.want {
			t.Errorf("%q on %q = %v, want %v", c.pattern, c.s, got, c.want)
		}
	}
}
