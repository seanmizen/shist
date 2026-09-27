package history

import (
	"fmt"
	"strings"
	"testing"
)

// synthetic returns n commands in format's on-disk layout, every 20th one
// multi-line, so benchmarks exercise both paths.
func synthetic(format Format, n int) string {
	var b strings.Builder
	for i := range n {
		ts := 1700000000 + i
		multi := i%20 == 0
		switch format {
		case Zsh:
			if multi {
				fmt.Fprintf(&b, ": %d:0;for f in *; do \\\n  echo $f %d \\\ndone\n", ts, i)
			} else {
				fmt.Fprintf(&b, ": %d:0;git commit -m \"change %d\"\n", ts, i)
			}
		case Bash:
			if multi {
				fmt.Fprintf(&b, "#%d\nfor f in *; do\n  echo $f %d\ndone\n", ts, i)
			} else {
				fmt.Fprintf(&b, "#%d\ngit commit -m \"change %d\"\n", ts, i)
			}
		case Fish:
			if multi {
				fmt.Fprintf(&b, "- cmd: for f in *\\n  echo $f %d\\nend\n  when: %d\n", i, ts)
			} else {
				fmt.Fprintf(&b, "- cmd: git commit -m \"change %d\"\n  when: %d\n", i, ts)
			}
		case PowerShell:
			if multi {
				fmt.Fprintf(&b, "foreach ($f in ls) {`\n  echo $f %d`\n}\n", i)
			} else {
				fmt.Fprintf(&b, "git commit -m \"change %d\"\n", i)
			}
		}
	}
	return b.String()
}

// Run with: go test -bench=. -benchmem ./internal/history
func BenchmarkParse(b *testing.B) {
	for _, format := range Formats {
		data := synthetic(format, 100_000)
		b.Run(format.Name(), func(b *testing.B) {
			b.SetBytes(int64(len(data)))
			b.ReportAllocs()
			for b.Loop() {
				format.Parse(data)
			}
		})
	}
}

func BenchmarkDetect(b *testing.B) {
	data := synthetic(Bash, 10_000)
	for b.Loop() {
		Detect(data, "history")
	}
}
