// Command shist prints your shell history with dates, filters and formatting.
package main

import (
	"os"

	"github.com/seanmizen/shist/internal/cli"
)

// version is stamped at build time via -ldflags "-X main.version=1.0.0".
var version = "dev"

func main() {
	os.Exit(cli.Main(version, cli.OSEnv()))
}
