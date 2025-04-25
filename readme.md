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

Show the last ten items
<img width="592" alt="default 10" src="https://github.com/user-attachments/assets/50bbf99a-e1cf-47e1-814c-0890f371349c" />

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
shist --min-date "2025-04-20" \
      --date-format "15:04" \
```
<img width="718" alt="Screenshot 2025-04-25 at 19 44 32" src="https://github.com/user-attachments/assets/9791844f-1207-4f2f-8100-7bd8b627fa1e" />

```bash
shist --min-date "2025-04-20" \
      --date-format "15:04" \
      --format "%d | %c"
```
<img width="718" alt="Screenshot 2025-04-25 at 19 45 07" src="https://github.com/user-attachments/assets/282ecda4-d241-41f2-b2f9-5980b181ab93" />


#### Don't process color directives, just raw text
```bash
shist --no-color
```
<img width="718" alt="Screenshot 2025-04-25 at 19 45 46" src="https://github.com/user-attachments/assets/416adb49-b7c8-437f-9e55-e789de128030" />

#### Show the last ten with no format

<img width="592" alt="ten unformatted" src="https://github.com/user-attachments/assets/e5b4be95-b356-4e65-83e4-86dbbc63cdfc" />

#### Default: Multiline commands print as multiline

<img width="886" alt="Screenshot 2025-04-25 at 19 41 13" src="https://github.com/user-attachments/assets/3b4f1189-014b-4675-ac3a-c9667bb1d01c" />

#### Or, you can choose to concatinate them with -c

<img width="886" alt="Screenshot 2025-04-25 at 19 41 20" src="https://github.com/user-attachments/assets/7b8c5668-2c63-47e2-b7a0-e84e6a9d18b1" />

#### We can also play with color
```bash
shist --n 20   \
    --format "%C(yellow)%d%C(reset)  %i  %C(#00c8ff)%c%C(reset)"
```
<img width="816" alt="Screenshot 2025-04-25 at 19 48 02" src="https://github.com/user-attachments/assets/f70818d0-17c7-4fe2-8329-613f367c7159" />
