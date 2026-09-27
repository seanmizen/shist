package cli

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/seanmizen/shist/internal/history"
)

type filters struct {
	match              func(string) bool
	minUnix, maxUnix   int64
	hasMin, hasMax     bool
	minIndex, maxIndex int
	n                  int
}

func (f filters) keep(e history.Entry) bool {
	return (f.match == nil || f.match(e.Command)) &&
		(!f.hasMin || e.Timestamp >= f.minUnix) &&
		(!f.hasMax || e.Timestamp <= f.maxUnix) &&
		(f.minIndex <= 0 || e.Index >= f.minIndex) &&
		(f.maxIndex <= 0 || e.Index <= f.maxIndex)
}

// apply runs every filter first, then keeps the newest n, so "-n 20" always
// means 20 matches. It filters entries in place.
func (f filters) apply(entries []history.Entry) []history.Entry {
	out := entries[:0]
	for _, e := range entries {
		if f.keep(e) {
			out = append(out, e)
		}
	}
	if f.n > 0 && f.n < len(out) {
		out = out[len(out)-f.n:]
	}
	return out
}

// matcher returns a fast path for plain-text patterns, which are most of them.
func matcher(re *regexp.Regexp) func(string) bool {
	if lit, complete := re.LiteralPrefix(); complete {
		return func(s string) bool { return strings.Contains(s, lit) }
	}
	return re.MatchString
}

// parseDate reads s in local time and returns the span it covers:
// a whole day for YYYY-MM-DD, a minute for YYYY-MM-DD HH:MM, and a single
// instant for UNIX seconds. Min filters use start, max filters use end.
func parseDate(s string) (start, end time.Time, err error) {
	if ts, err := strconv.ParseInt(s, 10, 64); err == nil {
		t := time.Unix(ts, 0)
		return t, t, nil
	}
	if t, err := time.ParseInLocation("2006-01-02 15:04", s, time.Local); err == nil {
		return t, t.Add(time.Minute - time.Nanosecond), nil
	}
	if t, err := time.ParseInLocation("2006-01-02", s, time.Local); err == nil {
		return t, t.AddDate(0, 0, 1).Add(-time.Nanosecond), nil
	}
	return time.Time{}, time.Time{}, fmt.Errorf("%q: want YYYY-MM-DD, YYYY-MM-DD HH:MM, or UNIX seconds", s)
}
