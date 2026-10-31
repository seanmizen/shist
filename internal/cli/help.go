package cli

import (
	"regexp"
	"strings"

	"github.com/seanmizen/shist/internal/format"
)

// Help pages, by the topic after -h. {version} is the version, and {spec}
// tags are colours: a %C spec (e.g. {dim}) or a name from helpStyles; {/}
// resets. Every flag must appear on the main page (TestHelpMentionsEveryFlag)
// and every env var on the env page (TestHelpMentionsEveryEnvVar).
var helpPages = map[string]string{"": helpMain, "format": helpFormat, "env": helpEnv}

var helpStyles = map[string]string{"h": "bold yellow", "f": "green", "a": "cyan", "/": "reset"}

const helpMain = `{bold}shist {version}{/} {dim}·{/} your shell history, with dates

{h}USAGE{/}
  shist [flags]

{h}FILTER{/}
  {f}-g, -grep{/} {a}PATTERN{/}       commands matching a regex
  {f}-C, -context{/} {a}N{/}          also show N commands either side of each match
  {f}-B, -before{/} {a}N{/}           {dim}…{/}only before
  {f}-A, -after{/} {a}N{/}            {dim}…{/}only after
  {f}-n{/} {a}N{/}                    only the newest N matches
  {f}-min-date{/} {a}DATE{/}          on or after DATE
  {f}-max-date{/} {a}DATE{/}          on or before DATE {dim}(a bare date is the whole day){/}
  {f}-min-index{/} {a}N{/}            index >= N
  {f}-max-index{/} {a}N{/}            index <= N

{h}OUTPUT{/}
  {f}-format{/} {a}TEMPLATE{/}        e.g. '%d %c' {dim}· see{/} shist -h format
  {f}-date-format{/} {a}LAYOUT{/}     Go layout for %d {dim}(default "2006-01-02 15:04"){/}
  {f}-c, -concat-multiline{/}   multi-line commands on one line
  {f}-no-color{/}               no colour

{h}OTHER{/}
  {f}-file{/} {a}PATH{/}              read this history file {dim}(default: your shell's){/}
  {f}-v, -version{/}            print the version
  {f}-h, -help{/} {a}[TOPIC]{/}       this help {dim}· topics:{/} format{dim},{/} env

{h}EXAMPLES{/}
  shist -g docker -n 20                 {dim}# newest 20 docker commands{/}
  shist -g 'git push' -C 3              {dim}# each push, 3 commands either side{/}
  shist -min-date 2025-04-01 -format '%c' | fzf

{dim}DATE is YYYY-MM-DD, "YYYY-MM-DD HH:MM" or UNIX seconds. -n and --n both work.{/}
`

const helpFormat = `{h}FORMAT{/}  {f}-format{/} {a}TEMPLATE{/}
  default '{green}%C(green)%d%C(reset){/} | {yellow}%C(yellow)%i%C(reset){/} | %c'

{h}TOKENS{/}
  {f}%d{/}  date {dim}(per -date-format){/}      {f}%i{/}  index {dim}(1 = oldest){/}
  {f}%t{/}  UNIX timestamp               {f}%e{/}  elapsed seconds {dim}(zsh only){/}
  {f}%c{/}  command                      {f}%%{/}  a literal %
  {f}%C({/}{a}spec{/}{f}){/}  colour, git log style

{h}COLOURS{/}  {dim}spec is space-separated words; 1st colour = text, 2nd = background{/}
  colours     {red}red{/} {green}green{/} {yellow}yellow{/} {blue}blue{/} {magenta}magenta{/} {cyan}cyan{/} {white}white{/} {black white}black{/} default
              {brightred}brightred{/} …, {214}0-255{/}, {#fed7b0}#rrggbb{/}
  attributes  {bold}bold{/} {dim}dim{/} {italic}italic{/} {ul}ul{/} {blink}blink{/} {reverse}reverse{/} {strike}strike{/}
  reset       clear all colour and attributes

{h}DATES{/}  {f}-date-format{/} {a}LAYOUT{/}
  Go layout, written as the reference time {bold}Mon Jan 2 15:04:05 MST 2006{/}
  e.g. "15:04"  "Jan 2"  "2006-01-02T15:04:05"

{h}EXAMPLES{/}
  shist -format '%i %d %es - %c'
  shist -format '%C(dim)%i%C(reset) %C(bold)%c%C(reset)'
  shist -format '%c' -c -g brew         {dim}# bare commands, one per line{/}
`

const helpEnv = `{h}ENVIRONMENT{/}  {dim}set in your shell profile; flags still win{/}
  {a}SHIST_DEFAULT_NUMBER_OF_ITEMS{/}   {f}-n{/}, e.g. 50
  {a}SHIST_DEFAULT_FILE{/}              {f}-file{/}
  {a}SHIST_DEFAULT_MIN_DATE{/}          {f}-min-date{/}
  {a}SHIST_DEFAULT_MAX_DATE{/}          {f}-max-date{/}
  {a}SHIST_DEFAULT_MIN_INDEX{/}         {f}-min-index{/}
  {a}SHIST_DEFAULT_MAX_INDEX{/}         {f}-max-index{/}
  {a}SHIST_DEFAULT_GREP{/}              {f}-grep{/}
  {a}SHIST_DEFAULT_CONTEXT{/}           {f}-context{/}  {dim}(only used with a grep){/}
  {a}SHIST_DEFAULT_BEFORE{/}            {f}-before{/}   {dim}(only used with a grep){/}
  {a}SHIST_DEFAULT_AFTER{/}             {f}-after{/}    {dim}(only used with a grep){/}
  {a}SHIST_DEFAULT_CONCAT{/}            {f}-c{/}  {dim}("true" or "1"){/}
  {a}SHIST_DEFAULT_NO_COLOR{/}          {f}-no-color{/}  {dim}("true" or "1"){/}
  {a}NO_COLOR{/}                        any value disables colour {dim}(no-color.org){/}
  {a}FORCE_COLOR{/}                     any value but 0 forces colour, even when piped

{h}EXIT STATUS{/}
  {f}0{/} ok  {dim}·{/}  {f}1{/} couldn't read history or write output  {dim}·{/}  {f}2{/} bad usage
`

var helpTag = regexp.MustCompile(`\{[a-z0-9#/ ]+\}`)

// renderHelp fills in the version and turns {spec} tags into colour, or
// drops them when color is false.
func renderHelp(page, version string, color bool) string {
	page = strings.ReplaceAll(page, "{version}", version)
	return helpTag.ReplaceAllStringFunc(page, func(tag string) string {
		if !color {
			return ""
		}
		spec := tag[1 : len(tag)-1]
		if s, ok := helpStyles[spec]; ok {
			spec = s
		}
		return format.SGR(spec)
	})
}
