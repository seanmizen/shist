# SHIST
Sean's History Tool. Est 2025.

`history` too long on your shell?
Try this:

```bash
# coming soon!
brew install shist
shist
```

```cmd
:: coming soon!
choco install shist
shist
```

## Build From Source
```bash
make build
./bin/shist
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

