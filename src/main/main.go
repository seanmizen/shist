package main

import "shist/src/nix"

// version is stamped at build time via -ldflags "-X main.version=1.0.0".
var version = "dev"

func main() {
	// all CLI handling lives in nix.Run
	nix.Run(version)
}
