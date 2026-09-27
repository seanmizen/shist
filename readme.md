# shist - Sean's History Tool

https://github.com/user-attachments/assets/c2d6e4dc-fdf3-4b32-a70b-20f771d4eda3

Prints your shell history with real dates, filters and your own output format.
Reads zsh, bash, fish and PowerShell (PSReadLine) history. The format is
detected from the file itself, so `shist -file ~/.bash_history` works from any
shell.

```console
$ shist -n 3
2025-04-01 09:12 | 18231 | git pull
2025-04-01 09:13 | 18232 | make test
2025-04-01 09:20 | 18233 | git commit -m "fix date filter"
```

- [Install](#install)
- [Usage](#usage)
- [Format tokens and colours](#format-tokens-and-colours)
- [Environment variables](#environment-variables)
- [How it compares](#how-it-compares)
- [Development](#development)

## Install

### Homebrew

```bash
brew tap seanmizen/tap
brew trust seanmizen/tap   # Homebrew requires explicit trust for third-party taps
brew install shist
```

### Go

```bash
go install github.com/seanmizen/shist@latest
```

### Prebuilt binaries

Download the archive for your platform from the
[latest release](https://github.com/seanmizen/shist/releases/latest), then:

```bash
tar -xzf shist_*_darwin_arm64.tar.gz
sudo mv shist /usr/local/bin/
shist
```

Windows: unzip and put `shist.exe` somewhere on your `PATH`. It reads your
PowerShell (PSReadLine) history.

Every release ships a `checksums.txt` if you want to verify the download:

```bash
shasum -a 256 -c checksums.txt --ignore-missing
```

### From source

```bash
git clone https://github.com/seanmizen/shist
cd shist
make install
```

## Usage

```
shist [options]
```

| Option | Meaning |
|---|---|
| `-n N` | show the newest N matches (`-1`, the default, shows all) |
| `-g`, `-grep PATTERN` | only commands matching this [regex](https://pkg.go.dev/regexp/syntax) |
| `-min-date DATE` | only entries on or after DATE |
| `-max-date DATE` | only entries on or before DATE; a bare date includes that whole day |
| `-min-index N` | only entries with index ≥ N |
| `-max-index N` | only entries with index ≤ N |
| `-file PATH` | history file to read (default: your shell's) |
| `-format TEMPLATE` | output template, see [below](#format-tokens-and-colours) |
| `-date-format LAYOUT` | [Go time layout](https://pkg.go.dev/time#pkg-constants) for `%d` (default `2006-01-02 15:04`) |
| `-c`, `-concat-multiline` | print multi-line commands on one line |
| `-no-color` | disable colour |
| `-v`, `-version` | print the version |
| `-h`, `-help` | show help |

- Flags take one or two dashes: `-n 20` and `--n 20` are the same.
- `DATE` is `YYYY-MM-DD`, `"YYYY-MM-DD HH:MM"` or UNIX seconds, in local time.
- All filters apply first; `-n` then keeps the newest N of what's left, so
  `-g docker -n 20` gives you 20 docker commands.
- The index (`%i`) is the entry's position in the history file, oldest first.

### Examples

```bash
shist -n 20                                  # last 20 commands
shist -g brew -format '%c' -c                # every brew command, one per line
shist -min-date 2025-04-01 -max-date 2025-04-07
shist -min-date "2025-04-01 09:00" -date-format "15:04" -format "%d %c"
shist -format '%c' -c | fzf --tac            # fuzzy-search your history
shist -file ~/.bash_history -n 10            # read another shell's history
shist -format '%C(dim)%i%C(reset) %C(bold)%c%C(reset)'
```

Exit status: `0` success, `1` couldn't read history or write output, `2` bad
usage (unknown flag, bad regex, bad date, bad env var).

## Format tokens and colours

| Token | Output |
|---|---|
| `%d` | date, per `-date-format` |
| `%t` | UNIX timestamp |
| `%i` | index (1 = oldest) |
| `%e` | elapsed seconds (zsh with `EXTENDED_HISTORY` only) |
| `%c` | the command |
| `%%` | a literal `%` |
| `%C(spec)` | colour, git log style |

The default format is `%C(green)%d%C(reset) | %C(yellow)%i%C(reset) | %c`.

A colour `spec` is space-separated words:

- **colours**: `black red green yellow blue magenta cyan white default`,
  bright versions (`brightred`), a 256-colour number (`0`–`255`), or 24-bit
  hex (`#fed7b0` or `fed7b0`). The first colour is the foreground and the
  second, if any, is the background.
- **attributes**: `bold dim italic ul blink reverse strike`.
- **reset** clears all colour and attributes.

For example `%C(bold red)`, `%C(214)`, `%C(#fed7b0 black)`, `%C(reset)`.

Colour is on when stdout is a terminal. Piped or redirected output has no
colour unless `FORCE_COLOR` is set.

History without timestamps (plain bash, PowerShell, zsh without
`EXTENDED_HISTORY`) prints an empty `%d` and `%t`.

## Environment variables

Set these in your shell profile to change the defaults. Flags still override them.

| Variable | Default for |
|---|---|
| `SHIST_DEFAULT_NUMBER_OF_ITEMS` | `-n`, e.g. `50` |
| `SHIST_DEFAULT_FILE` | `-file` |
| `SHIST_DEFAULT_MIN_DATE` | `-min-date` |
| `SHIST_DEFAULT_MAX_DATE` | `-max-date` |
| `SHIST_DEFAULT_MIN_INDEX` | `-min-index` |
| `SHIST_DEFAULT_MAX_INDEX` | `-max-index` |
| `SHIST_DEFAULT_CONCAT` | `-c` (`true` or `1`) |
| `SHIST_DEFAULT_GREP` | `-grep` |
| `SHIST_DEFAULT_NO_COLOR` | `-no-color` (`true` or `1`); `SHIST_NO_COLOR` also works |
| `NO_COLOR` | any value disables colour ([no-color.org](https://no-color.org)) |
| `FORCE_COLOR` | any value except `0` forces colour, even when piped |

## How it compares

| | shist | `history` / `fc` | [atuin](https://github.com/atuinsh/atuin) | [mcfly](https://github.com/cantino/mcfly) |
|---|---|---|---|---|
| What it is | reads your history file and prints it | your shell's built-in | replaces shell history with a database, plus sync | replaces Ctrl-R with a ranked search |
| Setup | none, one binary | none | shell hook + database | shell hook + database |
| Changes your shell | no | no | yes (Ctrl-R, up arrow) | yes (Ctrl-R) |
| Works on existing history | yes | yes | after import | after import |
| Same output across shells | yes | no, each shell differs | yes | n/a (interactive) |
| Dates | yes, any format | needs `HISTTIMEFORMAT` / `history -E` | yes | recorded, used for ranking |
| Scriptable output | yes, templates | limited | yes (`atuin history list`) | not its focus |
| Interactive search | no; pipe to `fzf` | no | yes | yes |
| Sync across machines | no | no | yes | no |

Use shist when you want to *look at* or *script over* the history you already
have ("what did I run last Tuesday?", "every `kubectl` command this month")
without installing hooks or changing how your shell records history. If you
want a better Ctrl-R or history sync between machines, atuin or mcfly are the
better fit, and shist works fine alongside them.

## Development

```bash
make build     # bin/shist-<os>-<arch>
make test      # go vet + go test
make bench     # benchmarks with -benchmem
make fuzz      # fuzz the parsers and templates (FUZZTIME=30s each)
```

Code layout:

- `main.go`: entry point
- `internal/history`: one parser per shell, plus format detection
- `internal/format`: output templates and colours
- `internal/cli`: flags, env defaults, filters, errors

CLI output is checked against golden files in `internal/cli/testdata/golden`.
If you change the output on purpose, regenerate them with
`go test ./internal/cli -update` and review the diff.

### Releasing

Tag and push. CI builds all six platform archives and publishes them:

```bash
git tag v1.1.0
git push origin v1.1.0
```

`make release` does the same build locally, into `dist/`. See
[CHANGELOG.md](CHANGELOG.md) for what changed in each version.

## Licence

MIT - see [LICENSE](LICENSE).
