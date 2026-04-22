% typr(1)

# NAME

typr - A terminal based typing test

# SYNOPSIS

usage: typr \[OPTION\]... \[FILE\]

typr visualize <FILE>

# SUBCOMMANDS

**visualize** <*FILE*>

:   Display an ASCII graph of typing speed progress over time.
    Reads data from CSV stats files generated with the `-csv` flag.
    Shows min, mean, and max WPM aggregated by day over the last 30 days.

    FILE is required. If FILE is a simple filename without directory separators,
    it will be looked up in the default results directory (~/.local/share/typr/results/).
    Otherwise, the path is used as-is (supports relative and absolute paths).

    Examples: `quotes-stats.csv`, `words-stats.csv`

# DESCRIPTION

  By default typr creates a test consisting of 50 randomly generated words from
  the top 1000 words in the English language. If provided with a path, typr will
  use the given file as input treating each paragraph as a separate segment of
  the test. The program will automatically keep track of your position in the
  file so subsequent invocations on the same path will place you at the most
  recent paragraph (-start 0 can be used to reset your position).

  Arbitrary text can also be piped directly into the program to create a custom
  test. Each paragraph of the input is treated as a segment unless '-multi' is
  supplied in which case each paragraph is treated as a separate test.

# OPTIONS

## Modes

-words  *WORDFILE*

: Specifies the file from which words are randomly drawn (default: 1000en).

-quotes

: Starts quote mode using the ZenQuotes API. Quotes are automatically cached
  to `~/.local/share/typr/quotes/zenlog.json` for offline fallback. If the API
  is unavailable, typr falls back to cached quotes or a built-in default.

-quotefile *QUOTEFILE*

: Override the quote source. Use the special value `zen` to use only locally
  cached quotes from previous ZenQuotes API sessions (no network required).
  Other values should be a JSON file of the form:

    [{"text": "foo", "attribution": "bar"}]

## Word Mode

-n *GROUPSZ*

: Sets the number of words which constitute a group.

-g *NGROUPS*

: Sets the number of groups which constitute a test.

## File Mode

-start *PARAGRAPH*

: The offset of the starting paragraph, set this to 0 to reset progress on a given file.

## Aesthetics

-showwpm

: Display WPM whilst typing.

-theme *THEMEFILE*

: The theme to use.

-notheme

: Attempt to use the default terminal theme. This may produce odd results depending on the theme colours.

-blockcursor

: Use the default cursor style.

-bold

: Embolden typed text.

-w

: The maximum line length in characters. This option is ignored if -raw is present.

## Test Parameters

-t *SECONDS*

: Terminate the test after the given number of seconds.

-noskip

: Disable word skipping when space is pressed.

-nohighlight

: Disable highlighting.

-highlight1

: Only highlight the current word.

-highlight2

: Only highlight the next word.

## Scripting

-oneshot

: Automatically exit after a single run.

-noreport

: Don't show a report at the end of a test.

-csv

: Write test results to CSV files.

    Stats: `~/.local/share/typr/results/{mode}-stats.csv` (timestamp,wpm,cpm,accuracy,file,n)\
    Errors: `~/.local/share/typr/results/{mode}-errors.csv` (timestamp,word,error)

    Enabled by default via config.yaml.

    Configure output directory via `csvdir` in `$XDG_CONFIG_HOME/typr/config.yaml` or `~/.config/typr/config.yaml`.

-json

: Print the test output in JSON.

-raw

: Don't reflow STDIN text or show one paragraph at a time. Note that line breaks
are determined exclusively by the input.

-multi

: Treat each input paragraph as a self contained test.

## Misc

**-list** *TYPE*\

    Lists internal resources of the given type. TYPE=[themes|quotes|words].

**-V**, **--version**\

    Print the current version.

# EXAMPLES

Fetch a random quote from the ZenQuotes API (cached locally):
```
typr -quotes
```

Use only locally cached quotes (offline):
```
typr -quotefile zen
```

Creates a series of tests each consisting of a random quote drawn from the
builtin quote file 'en':
```
typr -quotes en
```

Creates a series of tests each consisting of 10 random words drawn from
words.txt:
```
typr -words words.txt -n 10
```

Starts a sequence of tests in which each test consists of a paragraph from
War and Peace starting with paragraph 1:
```
typr ~/war_and_peace.txt -start 1
```

Produces a test consisting of 40 random words drawn from the system dictionary:
```
shuf -n 40 /usr/share/dict/words | typr
```

Runs a single timed test and saves results to CSV:
```
typr -t 60 -csv -oneshot
```

# PATHS

  Some options like **-words** and **-theme** accept a path. If the given path does
  not exist, the following directories are searched for a file with the given
  name before falling back to internal resources:

  ~/.config/typr/words\
  ~/.config/typr/themes\
  ~/.config/typr/quotes\
  /etc/typr/words\
  /etc/typr/themes\
  /etc/typr/quotes

# KEYS

  **esc: ** Exits the test\
  **tab: ** Restarts the test or starts new test after completion\
  **C-backspace: ** Deletes the previous word\
  **C-w: ** Deletes the previous word during typing

# AUTHOR

Originally by Aetnaeus <aetnaeus@protonmail.com>\
Extended and maintained by Jerid Francom

# SEE ALSO

## Project Page

    https://github.com/francojc/typr

# LICENSE

MIT — see LICENSE and NOTICE for details.
