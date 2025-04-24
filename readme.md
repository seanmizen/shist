# shist
Shell history tool. sean's history tool? not sure. your pick.

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
shist
  -n           Show last N entries (default 100, -1 = all)
  --file       Path to history file (default: ~/.zsh_history)
  --minDate    Filter entries after date (YYYY-MM-DD or YYYY-MM-DD HH:MM)
  --maxDate    Filter entries before date
  --minIndex   Filter entries with index >= N
  --maxIndex   Filter entries with index <= N
  --no-color   Disable colored output
  --help       Show this help message
```

e.g. `shist --minDate
### Examples

# yellow date, cyan command, custom layout
```bash
shist --n 20 \
      --format "%C(yellow)%d%C(reset)  %i  %C(#00c8ff)%c%C(reset)"
```

# only entries from April 2025, compact time format
note the use of Golang's Magical Reference Date [1](https://pkg.go.dev/time#pkg-constants) [2](https://devrants.blog/2021/10/04/golang-magical-reference-date/) 
```bash
shist --min-date "2025-04-01" \
      --date-format "15:04" \
      --format "%d | %c"
```
# Don't process color directives, just raw text
```bash
shist --no-color
```