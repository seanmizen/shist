package format

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/seanmizen/shist/internal/history"
)

func TestMain(m *testing.M) {
	time.Local = time.UTC // %d output must not depend on the machine
	os.Exit(m.Run())
}

var entry = history.Entry{Index: 7, Timestamp: 1700000000, Elapsed: 3, Command: "ls -la"}

func render(tmpl string, opt Options, e history.Entry) string {
	if opt.DateLayout == "" {
		opt.DateLayout = "2006-01-02 15:04"
	}
	return string(Compile(tmpl, opt).Append(nil, e))
}

func TestTokens(t *testing.T) {
	cases := []struct{ tmpl, want string }{
		{"%d | %i | %c", "2023-11-14 22:13 | 7 | ls -la"},
		{"%t %e", "1700000000 3"},
		{"100%% %c", "100% ls -la"},
		{"%x %", "%x %"}, // unknown and trailing % print as-is
		{"%C(red", "%C(red"},
		{"%C(red)%c%C(reset)", "ls -la"}, // colour off: directives dropped
		{"", ""},
	}
	for _, c := range cases {
		if got := render(c.tmpl, Options{}, entry); got != c.want+"\n" {
			t.Errorf("%q: got %q, want %q", c.tmpl, got, c.want)
		}
	}
}

func TestCommandIsNotReinterpreted(t *testing.T) {
	e := entry
	e.Command = "printf '%d %i %C(red)'"
	if got := render("%c", Options{Color: true}, e); got != e.Command+"\n" {
		t.Errorf("got %q", got)
	}
}

func TestNoTimestamp(t *testing.T) {
	e := entry
	e.Timestamp = 0
	if got := render("[%d][%t]%c", Options{}, e); got != "[][]ls -la\n" {
		t.Errorf("got %q", got)
	}
}

func TestMultiline(t *testing.T) {
	e := history.Entry{Index: 1, Command: "for x in 1 2 do echo $x done",
		Lines: []string{"for x in 1 2 \\", "do echo $x \\", "done"}}
	want := "for x in 1 2 \\\n    do echo $x \\\n    done\n"
	if got := render("%c", Options{}, e); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got := render("%c", Options{Concat: true}, e); got != e.Command+"\n" {
		t.Errorf("concat: got %q", got)
	}
}

func TestSGR(t *testing.T) {
	cases := map[string]string{
		"red":            "\033[31m",
		"RED":            "\033[31m",
		"reset":          "\033[0m",
		"bold":           "\033[1m",
		"bold red":       "\033[1;31m",
		"dim":            "\033[2m",
		"red blue":       "\033[31;44m",
		"brightred":      "\033[91m",
		"default":        "\033[39m",
		"214":            "\033[38;5;214m",
		"0":              "\033[38;5;0m",
		"#fed7b0":        "\033[38;2;254;215;176m",
		"fed7b0":         "\033[38;2;254;215;176m",
		"ul #000000 214": "\033[4;38;2;0;0;0;48;5;214m",
		"256":            "", // out of range
		"red blue green": "", // three colours
		"bogus":          "", // unknown word: no colour, as in 1.0
		"":               "",
	}
	for spec, want := range cases {
		if got := sgr(spec); got != want {
			t.Errorf("sgr(%q) = %q, want %q", spec, got, want)
		}
	}
}

func TestColorEnabled(t *testing.T) {
	env := func(kv ...string) func(string) (string, bool) {
		m := map[string]string{}
		for i := 0; i < len(kv); i += 2 {
			m[kv[i]] = kv[i+1]
		}
		return func(k string) (string, bool) { v, ok := m[k]; return v, ok }
	}
	var buf bytes.Buffer // not a terminal
	cases := []struct {
		name    string
		noColor bool
		env     func(string) (string, bool)
		want    bool
	}{
		{"pipe", false, env(), false},
		{"FORCE_COLOR", false, env("FORCE_COLOR", "1"), true},
		{"FORCE_COLOR=0", false, env("FORCE_COLOR", "0"), false},
		{"FORCE_COLOR empty", false, env("FORCE_COLOR", ""), false},
		{"no-color beats FORCE_COLOR", true, env("FORCE_COLOR", "1"), false},
	}
	for _, c := range cases {
		if got := ColorEnabled(&buf, c.noColor, c.env); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

func FuzzCompile(f *testing.F) {
	f.Add("%C(green)%d%C(reset) | %C(yellow)%i%C(reset) | %c", true)
	f.Add("%%%C(%C()%", false)
	f.Fuzz(func(t *testing.T, tmpl string, color bool) {
		out := render(tmpl, Options{Color: color}, entry)
		if !strings.HasSuffix(out, "\n") {
			t.Fatalf("no trailing newline: %q", out)
		}
		if !strings.Contains(tmpl, "%") && out != tmpl+"\n" {
			t.Fatalf("template without %% changed: %q -> %q", tmpl, out)
		}
	})
}

func FuzzSGR(f *testing.F) {
	f.Add("bold #fed7b0 214")
	f.Fuzz(func(t *testing.T, spec string) {
		if s := sgr(spec); s != "" && (!strings.HasPrefix(s, "\033[") || !strings.HasSuffix(s, "m")) {
			t.Fatalf("sgr(%q) = %q", spec, s)
		}
	})
}

// Run with: go test -bench=. -benchmem ./internal/format
func BenchmarkAppend(b *testing.B) {
	tmpl := Compile("%C(green)%d%C(reset) | %C(yellow)%i%C(reset) | %c",
		Options{DateLayout: "2006-01-02 15:04", Color: true})
	buf := make([]byte, 0, 256)
	b.ReportAllocs()
	for b.Loop() {
		buf = tmpl.Append(buf[:0], entry)
	}
}
