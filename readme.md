# shist
Shell history tool. Sean's history tool? Not sure. Your pick.

`history` too long on your shell?
Try this:

```bash
brew install shist  # coming soon!
shist
```

```powershell
choco install shist # also coming soon!
shist
```

## Build From Source
```bash
make build
OS=windows ARCH=amd64 make build
# etc
```
## Install From Source
This should detect your hardware + OS and install accordingly
```bash
make install 
shist
```

## Usage
### CLI Flags
```
> shist --help

shist - Sean's History Tool
Usage:
  shist [options]

Options:
  -date-format string
        Go time layout for the timestamp.
        You must use Golang's Magical Reference Date: Mon Jan 2 15:04:05 MST 2006 (default "2006-01-02 15:04")
  -file string
        History file to read (auto-detected if empty)
  -format string
        Output template (%d=date, %t=timestamp, %i=index, %e=elapsed, %c=command) (default "%C(green)%d%C(reset) | %C(yellow)%i%C(reset) | %c")
  -max-date string
        Maximum date (YYYY-MM-DD, YYYY-MM-DD HH:MM, or UNIX seconds)
  -max-index int
        Maximum index (inclusive) (default -1)
  -min-date string
        Minimum date (YYYY-MM-DD, YYYY-MM-DD HH:MM, or UNIX seconds)
  -min-index int
        Minimum index (inclusive) (default -1)
  -n int
        Number of history items to show (-1 for all) (default -1)
  -no-color
        Disable coloured output. Overrides color directives.

```

### Examples

#### Yellow date, cyan command, custom layout
```bash
shist --n 20 \
      --format "%C(yellow)%d%C(reset)  %i  %C(#00c8ff)%c%C(reset)"

2022-08-14 12:56  5272  npm ci
```
Note: Color doesn't show up on markdown. You'd better google this.

#### Only entries from April 2025, compact time format
note the use of Golang's Magical Reference Date [[1]](https://pkg.go.dev/time#pkg-constants) [[2]](https://devrants.blog/2021/10/04/golang-magical-reference-date/) 
```bash
shist --min-date "2025-04-01" \
      --date-format "15:04" \
      --format "%d | %c"

11:02 | corepack use yarn@^4.0.0
```
#### Don't process color directives, just raw text
```bash
shist --no-color

2022-08-14 12:56 | 5266 | npm ci
```
