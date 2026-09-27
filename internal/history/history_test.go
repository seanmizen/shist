package history

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// want is the part of an Entry a fixture test checks. Index is implied by
// position; lines is len(Entry.Lines).
type want struct {
	ts, el int64
	cmd    string
	lines  int
}

var long = `echo "` + strings.Repeat("x", 100_000) + `"`

var fixtures = []struct {
	file   string
	format Format
	want   []want
}{
	{"zsh_extended", Zsh, []want{
		{1700000000, 0, "git status", 0},
		{1700000005, 2, "make build", 0},
		{1700000010, 0, "for f in *.go; do gofmt -l $f done", 3},
		// ": 1700000020:0;   " is blank and dropped
		{1700000030, 0, long, 0},
		{1700000040, 1, `echo "héllo 😀 世界"`, 0},
		{1700000050, 0, `cd C:\\ && ls`, 0},
	}},
	{"zsh_plain", Zsh, []want{
		{0, 0, "ls -la", 0},
		{1700000000, 0, "git pull", 0},
		{0, 0, "pwd", 0}, // doesn't inherit git pull's timestamp
		{0, 0, "echo a b", 2},
	}},
	{"bash_plain", Bash, []want{
		{0, 0, "ls -la", 0},
		{0, 0, "cd /tmp", 0},
		{0, 0, "echo one two", 2},
		{0, 0, "git log --oneline", 0},
	}},
	{"bash_timestamps", Bash, []want{
		{0, 0, "ls -la", 0}, // before HISTTIMEFORMAT was set
		{1700000000, 0, "git status", 0},
		{1700000060, 0, "for i in 1 2; do echo $i done", 3},
		{1700000120, 0, "exit", 0}, // "# 1700000120" tolerated
	}},
	{"fish", Fish, []want{
		{1700000000, 0, "echo a: b", 0}, // not valid YAML
		{1700000001, 0, "cd /tmp", 0},
		{1700000002, 0, "for x in 1 2 echo $x end", 3},
		{1700000003, 0, `echo back\slash`, 0},
		{1700000004, 0, `echo "😀"`, 0},
	}},
	{"ConsoleHost_history.txt", PowerShell, []want{
		{0, 0, "Get-ChildItem", 0}, // BOM stripped
		{0, 0, "foreach ($x in 1..2) { Write-Output $x }", 3},
		{0, 0, `cd C:\`, 0}, // a real trailing backslash is kept
	}},
}

func TestFixtures(t *testing.T) {
	for _, fx := range fixtures {
		t.Run(fx.file, func(t *testing.T) {
			path := filepath.Join("testdata", fx.file)
			got, f, err := Load(path)
			if err != nil {
				t.Fatal(err)
			}
			if f != fx.format {
				t.Errorf("detected %s, want %s", f.Name(), fx.format.Name())
			}
			if len(got) != len(fx.want) {
				t.Fatalf("got %d entries, want %d:\n%q", len(got), len(fx.want), commands(got))
			}
			for i, w := range fx.want {
				e := got[i]
				if e.Index != i+1 || e.Timestamp != w.ts || e.Elapsed != w.el ||
					e.Command != w.cmd || len(e.Lines) != w.lines {
					t.Errorf("entry %d:\n got {%d %d %d %.60q lines=%d}\nwant {%d %d %d %.60q lines=%d}",
						i, e.Index, e.Timestamp, e.Elapsed, e.Command, len(e.Lines),
						i+1, w.ts, w.el, w.cmd, w.lines)
				}
			}
		})
	}
}

func commands(entries []Entry) []string {
	out := make([]string, len(entries))
	for i, e := range entries {
		out[i] = e.Command
	}
	return out
}

func TestDetect(t *testing.T) {
	t.Setenv("SHELL", "/bin/zsh")
	cases := []struct {
		name, path, content string
		want                Format
	}{
		{"fish content", "/x/history", "- cmd: ls\n  when: 1\n", Fish},
		{"zsh content beats name", "/x/.bash_history", ": 1700000000:0;ls\n", Zsh},
		{"bash content beats name", "/x/.zsh_history", "#1700000000\nls\n", Bash},
		{"bash by name", "/x/.bash_history", "ls\n", Bash},
		{"powershell by name", "/x/ConsoleHost_history.txt", "ls\n", PowerShell},
		{"fallback to $SHELL", "/x/history", "ls\n", Zsh},
		{"empty file", "/x/history", "", Zsh},
	}
	for _, c := range cases {
		if got := Detect(c.content, c.path); got != c.want {
			t.Errorf("%s: got %s, want %s", c.name, got.Name(), c.want.Name())
		}
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, _, err := Load(filepath.Join(t.TempDir(), "nope"))
	if !os.IsNotExist(err) && !strings.Contains(err.Error(), "no such file") {
		t.Fatalf("err = %v", err)
	}
}

func TestParseZshHeader(t *testing.T) {
	cases := map[string]bool{
		": 1700000000:0;ls": true,
		": 1700000000:12;":  true,
		": 1700000000;ls":   false,
		":1700000000:0;ls":  false,
		": abc:0;ls":        false,
		": 1700000000:0 ls": false,
		"ls":                false,
		": 1:2;x":           true,
	}
	for line, ok := range cases {
		if _, _, _, got := parseZshHeader(line); got != ok {
			t.Errorf("parseZshHeader(%q) ok = %v, want %v", line, got, ok)
		}
	}
}

func TestUnmetafy(t *testing.T) {
	// 😀 is F0 9F 98 80; zsh escapes 9F and 98 as 0x83 followed by byte^32
	if got := unmetafy("echo \xf0\x83\xbf\x83\xb8\x80"); got != "echo 😀" {
		t.Errorf("got %q", got)
	}
	if got := unmetafy("plain"); got != "plain" {
		t.Errorf("got %q", got)
	}
	if got := unmetafy("dangling\x83"); got != "dangling\x83" {
		t.Errorf("got %q", got)
	}
}
