package history

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// FuzzParse feeds arbitrary bytes to every parser and checks the invariants
// the rest of shist relies on. Run with: go test -fuzz=FuzzParse ./internal/history
func FuzzParse(f *testing.F) {
	for _, fx := range fixtures {
		data, err := os.ReadFile(filepath.Join("testdata", fx.file))
		if err != nil {
			f.Fatal(err)
		}
		f.Add(string(data))
	}
	f.Add(": 1:2;a \\\n\\\n")
	f.Add("- cmd: \\\n  when: 99999999999999999999\n")

	f.Fuzz(func(t *testing.T, data string) {
		for _, format := range Formats {
			for i, e := range format.Parse(data) {
				if e.Index != i+1 {
					t.Fatalf("%s: entry %d has index %d", format.Name(), i, e.Index)
				}
				if e.Command == "" || e.Command != strings.TrimSpace(e.Command) {
					t.Fatalf("%s: bad command %q", format.Name(), e.Command)
				}
				if len(e.Lines) == 1 {
					t.Fatalf("%s: Lines set for a single-line command", format.Name())
				}
			}
		}
		Detect(data, "history") // must not panic
	})
}
