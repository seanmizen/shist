package cli

import (
	"regexp"
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
