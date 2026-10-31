// Package cli is shist's command line: flags, env defaults, filtering and output.
package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/seanmizen/shist/internal/format"
	"github.com/seanmizen/shist/internal/history"
)

// Exit codes.
const (
	exitOK    = 0
	exitError = 1 // couldn't read or write history
	exitUsage = 2 // bad flag, env var, pattern or date
)

// usageError is a mistake in how shist was invoked.
type usageError struct{ error }

func usagef(f string, a ...any) error { return usageError{fmt.Errorf(f, a...)} }

// Env is everything shist needs from the outside world, so tests can fake it.
type Env struct {
	Args      []string // without the program name
	Stdout    io.Writer
	Stderr    io.Writer
	LookupEnv func(string) (string, bool)
}

// OSEnv is the real process environment.
func OSEnv() Env {
	return Env{Args: os.Args[1:], Stdout: os.Stdout, Stderr: os.Stderr, LookupEnv: os.LookupEnv}
}

// Main runs shist and returns the process exit code.
func Main(version string, env Env) int {
	err := run(version, env)
	switch {
	case err == nil, errors.Is(err, flag.ErrHelp):
		return exitOK
	case errors.As(err, new(usageError)):
		fmt.Fprintf(env.Stderr, "shist: %v\nRun 'shist -h' for usage.\n", err)
		return exitUsage
	default:
		fmt.Fprintf(env.Stderr, "shist: %v\n", err)
		return exitError
	}
}

type config struct {
	n                  int
	file               string
	minDate, maxDate   string
	minIndex, maxIndex int
	noColor            bool
	concat             bool
	grep               string
	context            int
	before, after      int
	contextFlag        bool // -context, -before or -after was on the command line
	dateFormat         string
	format             string
	version            bool
	helpTopic          string // the word after -h, if any
}

// newFlagSet defines every flag on c, with defaults from the environment.
// Env values that don't parse are recorded in bad.
func newFlagSet(env Env, c *config, bad *[]string) *flag.FlagSet {
	envStr := func(key, def string) string {
		if v, ok := env.LookupEnv(key); ok {
			return v
		}
		return def
	}
	envInt := func(key string, def int) int {
		if v, ok := env.LookupEnv(key); ok {
			i, err := strconv.Atoi(v)
			if err != nil {
				*bad = append(*bad, fmt.Sprintf("%s=%q is not a whole number", key, v))
				return def
			}
			return i
		}
		return def
	}
	envBool := func(key string, def bool) bool {
		if v, ok := env.LookupEnv(key); ok {
			return v == "1" || strings.EqualFold(v, "true")
		}
		return def
	}
	noColorEnv, _ := env.LookupEnv("NO_COLOR") // https://no-color.org

	flags := flag.NewFlagSet("shist", flag.ContinueOnError)
	flags.SetOutput(io.Discard) // we print our own help and errors
	flags.IntVar(&c.n, "n", envInt("SHIST_DEFAULT_NUMBER_OF_ITEMS", -1), "")
	flags.StringVar(&c.file, "file", envStr("SHIST_DEFAULT_FILE", ""), "")
	flags.StringVar(&c.minDate, "min-date", envStr("SHIST_DEFAULT_MIN_DATE", ""), "")
	flags.StringVar(&c.maxDate, "max-date", envStr("SHIST_DEFAULT_MAX_DATE", ""), "")
	flags.IntVar(&c.minIndex, "min-index", envInt("SHIST_DEFAULT_MIN_INDEX", -1), "")
	flags.IntVar(&c.maxIndex, "max-index", envInt("SHIST_DEFAULT_MAX_INDEX", -1), "")
	// SHIST_NO_COLOR is the pre-1.1 name, kept working.
	flags.BoolVar(&c.noColor, "no-color",
		envBool("SHIST_DEFAULT_NO_COLOR", envBool("SHIST_NO_COLOR", noColorEnv != "")), "")
	flags.BoolVar(&c.concat, "concat-multiline", envBool("SHIST_DEFAULT_CONCAT", false), "")
	flags.BoolVar(&c.concat, "c", c.concat, "")
	flags.StringVar(&c.grep, "grep", envStr("SHIST_DEFAULT_GREP", ""), "")
	flags.StringVar(&c.grep, "g", c.grep, "")
	flags.IntVar(&c.context, "context", envInt("SHIST_DEFAULT_CONTEXT", 0), "")
	flags.IntVar(&c.context, "C", c.context, "")
	flags.IntVar(&c.before, "before", envInt("SHIST_DEFAULT_BEFORE", 0), "")
	flags.IntVar(&c.before, "B", c.before, "")
	flags.IntVar(&c.after, "after", envInt("SHIST_DEFAULT_AFTER", 0), "")
	flags.IntVar(&c.after, "A", c.after, "")
	flags.StringVar(&c.dateFormat, "date-format", "2006-01-02 15:04", "")
	flags.StringVar(&c.format, "format", "%C(green)%d%C(reset) | %C(yellow)%i%C(reset) | %c", "")
	flags.BoolVar(&c.version, "version", false, "")
	flags.BoolVar(&c.version, "v", false, "")

	return flags
}

func parseArgs(env Env) (config, error) {
	var c config
	var bad []string // env vars that didn't parse
	flags := newFlagSet(env, &c, &bad)
	if err := flags.Parse(env.Args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			c.helpTopic = flags.Arg(0)
			return c, err
		}
		return c, usageError{err}
	}
	flags.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "context", "C", "before", "B", "after", "A":
			c.contextFlag = true
		}
	})
	if flags.NArg() > 0 {
		return c, usagef("unexpected argument %q", flags.Arg(0))
	}
	if len(bad) > 0 {
		return c, usagef("%s", strings.Join(bad, "; "))
	}
	return c, nil
}

func run(version string, env Env) error {
	c, err := parseArgs(env)
	if errors.Is(err, flag.ErrHelp) {
		page, ok := helpPages[c.helpTopic]
		if !ok {
			return usagef("no help topic %q (try format or env)", c.helpTopic)
		}
		color := format.ColorEnabled(env.Stdout, c.noColor, env.LookupEnv)
		fmt.Fprint(env.Stdout, renderHelp(page, version, color))
		return err
	}
	if err != nil {
		return err
	}
	if c.version {
		fmt.Fprintf(env.Stdout, "shist %s\n", version)
		return nil
	}

	/* ---------- filters (checked before touching the file) ---------- */
	f := filters{n: c.n, minIndex: c.minIndex, maxIndex: c.maxIndex}
	if c.context < 0 || c.before < 0 || c.after < 0 {
		return usagef("-context, -before and -after can't be negative")
	}
	if c.grep != "" { // context is only for grep; env defaults alone don't complain
		f.before, f.after = c.before, c.after
		if f.before == 0 {
			f.before = c.context
		}
		if f.after == 0 {
			f.after = c.context
		}
	} else if c.contextFlag {
		return usagef("-context, -before and -after need -grep")
	}
	if c.grep != "" {
		re, err := regexp.Compile(c.grep)
		if err != nil {
			return usagef("invalid --grep pattern %q: %v", c.grep, regexpErr(err))
		}
		f.match = matcher(re)
	}
	if c.minDate != "" {
		start, _, err := parseDate(c.minDate)
		if err != nil {
			return usagef("invalid --min-date: %v", err)
		}
		f.minUnix, f.hasMin = start.Unix(), true
	}
	if c.maxDate != "" {
		_, end, err := parseDate(c.maxDate)
		if err != nil {
			return usagef("invalid --max-date: %v", err)
		}
		f.maxUnix, f.hasMax = end.Unix(), true
	}

	/* ---------- read ---------- */
	path := c.file
	if path == "" {
		shell, _ := env.LookupEnv("SHELL")
		path = history.ForShell(shell).DefaultPath()
	}
	entries, _, err := history.Load(path)
	if err != nil {
		if c.file == "" && errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("%w (use --file to pick a history file)", err)
		}
		return err
	}

	/* ---------- write ---------- */
	opt := format.Options{
		DateLayout: c.dateFormat,
		Color:      format.ColorEnabled(env.Stdout, c.noColor, env.LookupEnv),
		Concat:     c.concat,
	}
	tmpl := format.Compile(c.format, opt)
	const flushAt = 64 * 1024
	buf := make([]byte, 0, flushAt+4096)
	flush := func(min int) error {
		if len(buf) < min {
			return nil
		}
		if _, err := env.Stdout.Write(buf); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
		buf = buf[:0]
		return nil
	}
	if f.before == 0 && f.after == 0 {
		for _, e := range f.apply(entries) { // oldest → newest
			buf = tmpl.Append(buf, e)
			if err := flush(flushAt); err != nil {
				return err
			}
		}
		return flush(0)
	}

	opt.Dim = true
	dim := format.Compile(c.format, opt) // neighbours of a match
	sep := "--\n"                        // between groups that aren't next to each other
	if opt.Color {
		sep = format.SGR("dim") + "--" + format.SGR("reset") + "\n"
	}
	for _, r := range f.withContext(entries) {
		if r.gap {
			buf = append(buf, sep...)
		}
		if r.context {
			buf = dim.Append(buf, r.Entry)
		} else {
			buf = tmpl.Append(buf, r.Entry)
		}
		if err := flush(flushAt); err != nil {
			return err
		}
	}
	return flush(0)
}

// regexpErr drops Go's "error parsing regexp: " prefix; the caller adds its own context.
func regexpErr(err error) string {
	return strings.TrimPrefix(err.Error(), "error parsing regexp: ")
}
