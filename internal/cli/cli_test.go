package cli

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var update = flag.Bool("update", false, "rewrite testdata/golden files")

func TestMain(m *testing.M) {
	time.Local = time.UTC // dates in golden files must not depend on the machine
	os.Exit(m.Run())
}

func fixture(name string) string { return filepath.Join("..", "history", "testdata", name) }

// shist runs the CLI in-process with only the given env vars set.
func shist(t testing.TB, env map[string]string, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	var out, errOut bytes.Buffer
	code = Main("test", Env{
		Args:      args,
		Stdout:    &out,
		Stderr:    &errOut,
		LookupEnv: func(k string) (string, bool) { v, ok := env[k]; return v, ok },
	})
	return out.String(), errOut.String(), code
}

// TestGolden compares full output against testdata/golden/<name>.txt.
// Regenerate with: go test ./internal/cli -update
func TestGolden(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		args []string
	}{
		{"zsh_default", nil, []string{"-file", fixture("zsh_plain")}},
		{"zsh_multiline", nil, []string{"-file", fixture("zsh_extended"), "-max-index", "3"}},
		{"zsh_concat", nil, []string{"-file", fixture("zsh_extended"), "-max-index", "3", "-c"}},
		{"bash_timestamps", nil, []string{"-file", fixture("bash_timestamps"), "-format", "%i %t %c"}},
		{"fish", nil, []string{"-file", fixture("fish"), "-date-format", "15:04:05"}},
		{"powershell", nil, []string{"-file", fixture("ConsoleHost_history.txt"), "-format", "%i: %c"}},
		{"n_after_grep", nil, []string{"-file", fixture("bash_plain"), "-g", "^[lg]", "-n", "1"}},
		{"n_after_date", nil, []string{"-file", fixture("fish"), "-max-date", "2023-11-14", "-n", "2", "-format", "%c"}},
		{"force_color", map[string]string{"FORCE_COLOR": "1"}, []string{"-file", fixture("fish"), "-n", "1"}},
		{"custom_color", map[string]string{"FORCE_COLOR": "1"},
			[]string{"-file", fixture("fish"), "-n", "1", "-format", "%C(bold 214)%i%C(reset) %C(dim)%c%C(reset) 100%%"}},
		{"no_color_env", map[string]string{"FORCE_COLOR": "1", "NO_COLOR": "1"}, []string{"-file", fixture("fish"), "-n", "1"}},
		{"env_defaults", map[string]string{"SHIST_DEFAULT_NUMBER_OF_ITEMS": "2", "SHIST_DEFAULT_GREP": "echo"},
			[]string{"-file", fixture("fish"), "-format", "%c"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, errOut, code := shist(t, c.env, c.args...)
			if code != 0 {
				t.Fatalf("exit %d: %s", code, errOut)
			}
			golden := filepath.Join("testdata", "golden", c.name+".txt")
			if *update {
				if err := os.MkdirAll(filepath.Dir(golden), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(golden, []byte(out), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("%v (run with -update to create)", err)
			}
			if out != string(want) {
				t.Errorf("output differs from %s:\n got:\n%s\nwant:\n%s", golden, out, want)
			}
		})
	}
}

func TestErrors(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nope")
	cases := []struct {
		name string
		env  map[string]string
		args []string
		code int
		msg  string
	}{
		{"bad regex", nil, []string{"-file", fixture("fish"), "-g", "("}, exitUsage, `invalid --grep pattern "(": missing closing )`},
		{"bad date", nil, []string{"-file", fixture("fish"), "-min-date", "yesterday"}, exitUsage, `invalid --min-date: "yesterday": want YYYY-MM-DD`},
		{"bad flag", nil, []string{"-nope"}, exitUsage, "flag provided but not defined: -nope"},
		{"stray arg", nil, []string{"20"}, exitUsage, `unexpected argument "20"`},
		{"bad env int", map[string]string{"SHIST_DEFAULT_NUMBER_OF_ITEMS": "lots"}, nil, exitUsage,
			`SHIST_DEFAULT_NUMBER_OF_ITEMS="lots" is not a whole number`},
		{"missing file", nil, []string{"-file", missing}, exitError, "reading history: open " + missing},
		{"missing default file", map[string]string{"SHELL": "/bin/zsh", "HOME": t.TempDir()}, nil, exitError, "use --file"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.env["HOME"] != "" {
				t.Setenv("HOME", c.env["HOME"]) // os.UserHomeDir reads the real env
			}
			out, errOut, code := shist(t, c.env, c.args...)
			if code != c.code {
				t.Errorf("exit %d, want %d", code, c.code)
			}
			if !strings.HasPrefix(errOut, "shist: ") || !strings.Contains(errOut, c.msg) {
				t.Errorf("stderr = %q, want it to contain %q", errOut, c.msg)
			}
			if out != "" {
				t.Errorf("stdout = %q, want nothing", out)
			}
		})
	}
}

func TestHelpAndVersion(t *testing.T) {
	out, _, code := shist(t, nil, "-h")
	if code != 0 || !strings.HasPrefix(out, "shist test - ") {
		t.Errorf("-h: exit %d, out %.40q", code, out)
	}
	if strings.Contains(out, "%!") {
		t.Errorf("help has a Printf formatting error:\n%s", out)
	}
	out, _, code = shist(t, nil, "--version")
	if code != 0 || out != "shist test\n" {
		t.Errorf("--version: exit %d, out %q", code, out)
	}
}

// TestHelpMentionsEveryFlag keeps the hand-written help in sync with the flags.
func TestHelpMentionsEveryFlag(t *testing.T) {
	var c config
	var bad []string
	flags := newFlagSet(Env{LookupEnv: func(string) (string, bool) { return "", false }}, &c, &bad)
	flags.VisitAll(func(f *flag.Flag) {
		if !strings.Contains(helpText, "-"+f.Name+" ") && !strings.Contains(helpText, "-"+f.Name+",") {
			t.Errorf("help text doesn't mention -%s", f.Name)
		}
	})
}

// Run with: go test -bench=. -benchmem ./internal/cli
func BenchmarkMain(b *testing.B) {
	var buf bytes.Buffer
	for i := range 100_000 {
		fmt.Fprintf(&buf, ": %d:0;git commit -m \"change %d\"\n", 1700000000+i, i)
	}
	path := filepath.Join(b.TempDir(), "zsh_history")
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		b.Fatal(err)
	}
	for name, args := range map[string][]string{
		"all":     nil,
		"n20":     {"-n", "20"},
		"grep":    {"-g", "commit"},
		"regex":   {"-g", "change [0-9]+5$"},
		"colored": {"-format", "%C(green)%d%C(reset) %c"},
	} {
		args = append([]string{"-file", path}, args...)
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if _, errOut, code := shist(b, nil, args...); code != 0 {
					b.Fatal(errOut)
				}
			}
		})
	}
}
