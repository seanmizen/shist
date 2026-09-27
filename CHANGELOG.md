# Changelog

All notable changes to shist. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and shist uses
[semantic versioning](https://semver.org/).

## [1.1.0] - 2026-09-27

### Added

- **PowerShell support.** shist reads PSReadLine history
  (`ConsoleHost_history.txt`), so the Windows builds work now.
- **The history format is detected from the file itself**, then from its name,
  then from `$SHELL`. `shist -file ~/.bash_history` works from zsh.
- **More colours.** `%C(…)` takes git-style specs: `bold`, `dim`, `italic`,
  `ul`, `blink`, `reverse` and `strike`; `bright*` colours; 256-colour numbers
  (`%C(214)`); and a background colour (`%C(white red)`).
- `%%` in `-format` prints a literal `%`.
- `NO_COLOR` ([no-color.org](https://no-color.org)) and `FORCE_COLOR` are
  honoured. `SHIST_DEFAULT_NO_COLOR` now works, as the help always said.
- Exit codes: `0` success, `1` couldn't read history or write output, `2` bad
  usage.
- `go install github.com/seanmizen/shist@latest` works.
- Clearer errors, e.g. `shist: invalid --grep pattern "(": missing closing ): ...`,
  and a hint to use `-file` when the default history file is missing.
- The README now documents every flag, format token and env var, and compares
  shist with `history`, atuin and mcfly.

### Changed

- **`-n` applies after all other filters.** `-n 20 -min-date X` now gives the
  20 newest entries after X. Before, it took the newest 20 overall and then
  filtered them, which could return fewer.
- **Dates are local time.** `-min-date` and `-max-date` were parsed as UTC
  while output was printed in local time.
- **`-max-date` is inclusive.** `-max-date 2025-04-01` now includes all of
  April 1st, and `"2025-04-01 15:04"` includes that whole minute. Before, it
  stopped at midnight at the start of the day.
- Unexpected arguments (`shist 20`) are an error instead of being silently
  ignored. Use `-n 20`.
- Invalid numeric env vars (`SHIST_DEFAULT_NUMBER_OF_ITEMS=lots`) are an error
  instead of being silently ignored.
- `-h` prints to stdout and exits 0.
- Multi-line commands no longer have a double space before each ` \`.

### Fixed

- **fish history couldn't be read at all.** It failed with
  `cannot unmarshal !!seq`, and on any command containing `: `, because
  fish_history isn't valid YAML.
- **bash history without `HISTTIMEFORMAT` came out as one giant entry.**
- **Timestamped bash history** wasn't recognised: bash writes `#1700000000`,
  and shist expected `# 1700000000`.
- **A single history line over 64KB** (a large paste) broke the whole read
  with `bufio.Scanner: token too long`.
- **Colour was written into pipes and files.** shist checked whether stderr was
  a terminal instead of stdout, so `shist > file` got escape codes.
- zsh non-ASCII text (accents, emoji, CJK) was garbled. zsh escapes these bytes
  in its history file, and shist now decodes them.
- zsh lines without `EXTENDED_HISTORY` inherited the previous entry's
  timestamp.
- The help text named `SHIST_DEFAULT_NO_COLOR` while the code read
  `SHIST_NO_COLOR`. Both work now.

### Performance

About 20× faster on large histories (300,000 entries, 20MB, Apple Silicon):

| | 1.0.1 | 1.1.0 |
|---|---|---|
| all entries | 1.11 s | 53 ms |
| `-n 20` | 298 ms | 24 ms |
| `-g "msg 1234"` | 302 ms | 33 ms |

- The format template and colours are parsed once, not once per entry
  (1.0 made up to four terminal-check system calls per line).
- Output is buffered instead of one `write` per line.
- Parsers work on the whole file in memory without a regex per line, and
  single-line commands need no allocations. Parsing runs at about 1 GB/s.
- Plain-text `-g` patterns use a substring search instead of the regex engine.

### Internal

- Module path is `github.com/seanmizen/shist`. Code moved from
  `src/{main,nix,model,ui}` to `main.go` + `internal/{cli,history,format}`.
- `gopkg.in/yaml.v3` dependency removed.
- Tests added: fixture tests per shell, golden-file CLI tests, fuzz tests for
  the parsers and templates, and benchmarks (`make test`, `make bench`,
  `make fuzz`).

## [1.0.1] - 2026-09-20

### Added

- MIT licence, included in release archives.

## [1.0.0] - 2026-09-20

First tagged release: zsh, bash and fish history, format templates, date,
index and grep filters, prebuilt archives for six platforms.

[1.1.0]: https://github.com/seanmizen/shist/compare/v1.0.1...v1.1.0
[1.0.1]: https://github.com/seanmizen/shist/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/seanmizen/shist/releases/tag/v1.0.0
