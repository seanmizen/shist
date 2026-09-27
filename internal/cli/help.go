package cli

// helpText is printed for -h / --help. %s is the version.
// Every flag must appear here (TestHelpMentionsEveryFlag checks).
const helpText = `shist %s - Sean's History Tool

Prints your shell history (zsh, bash, fish, PowerShell) with dates, filters
and a custom format. The file format is detected automatically.

Usage:
  shist [options]

Options:
  -n N                  show the newest N matches (-1: all)
  -g, -grep PATTERN     only entries whose command matches this regex
  -min-date DATE        only entries on or after DATE
  -max-date DATE        only entries on or before DATE (a date means the whole day)
  -min-index N          only entries with index >= N
  -max-index N          only entries with index <= N
  -file PATH            history file to read (default: your shell's)
  -format TEMPLATE      output template (see below)
                        default: "%%C(green)%%d%%C(reset) | %%C(yellow)%%i%%C(reset) | %%c"
  -date-format LAYOUT   Go time layout for %%d (default: "2006-01-02 15:04")
                        uses Go's reference date: Mon Jan 2 15:04:05 MST 2006
  -c, -concat-multiline print multi-line commands on one line
  -no-color             disable colour
  -v, -version          print the version
  -h, -help             show this help

  Flags take one or two dashes: -n 20 and --n 20 are the same.
  DATE is YYYY-MM-DD, "YYYY-MM-DD HH:MM" or UNIX seconds, in local time.
  Filters apply first; -n then keeps the newest N of what's left.

Format tokens:
  %%d  date (per -date-format)      %%i  index (1 = oldest)
  %%t  UNIX timestamp               %%e  elapsed seconds (zsh only)
  %%c  command                      %%%%  a literal %%
  %%C(spec)  colour, git log style. spec is space-separated words:
       colours     red green yellow blue magenta cyan white black default,
                   brightred etc., 0-255, or #rrggbb / rrggbb
                   (first colour is foreground, second is background)
       attributes  bold dim italic ul blink reverse strike
       reset       clear all colour and attributes
     e.g. %%C(bold red)  %%C(214)  %%C(#fed7b0 black)  %%C(reset)

Examples:
  shist -n 20 -format "%%i %%d %%es - %%c"
  shist -min-date 2025-04-01 -date-format "15:04" -format "%%d | %%c"
  shist -g foo -n 20              # the newest 20 commands matching foo
  shist -format "%%c" -c -g brew   # just commands containing brew, one per line
  shist -format "%%C(dim)%%i%%C(reset) %%C(bold)%%c%%C(reset)"

Environment:
  SHIST_DEFAULT_NUMBER_OF_ITEMS   default for -n, e.g. 50
  SHIST_DEFAULT_FILE              default for -file
  SHIST_DEFAULT_MIN_DATE          default for -min-date
  SHIST_DEFAULT_MAX_DATE          default for -max-date
  SHIST_DEFAULT_MIN_INDEX         default for -min-index
  SHIST_DEFAULT_MAX_INDEX         default for -max-index
  SHIST_DEFAULT_CONCAT            "true" or "1" to default -c on
  SHIST_DEFAULT_GREP              default for -grep
  SHIST_DEFAULT_NO_COLOR          "true" or "1" to default -no-color on
  NO_COLOR                        any value disables colour (no-color.org)
  FORCE_COLOR                     any value except 0 forces colour, even when piped

Exit status: 0 ok, 1 couldn't read history or write output, 2 bad usage.
`
