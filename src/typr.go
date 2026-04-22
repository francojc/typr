package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gdamore/tcell"
	"github.com/mattn/go-isatty"
)

var scr tcell.Screen
var csvMode bool
var jsonMode bool
var currentTestType string
var currentTestFile string
var currentTestN int

type result struct {
	Wpm       int       `json:"wpm"`
	Cpm       int       `json:"cpm"`
	Accuracy  float64   `json:"accuracy"`
	Timestamp int64     `json:"timestamp"`
	Mistakes  []mistake `json:"mistakes"`
	File      string    `json:"file"`
	N         int       `json:"n"`
}

func die(format string, args ...interface{}) {
	if scr != nil {
		scr.Fini()
	}
	fmt.Fprintf(os.Stderr, "ERROR: ")
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}

var results []result

func parseConfig(b []byte) map[string]string {
	if b == nil {
		return nil
	}

	cfg := map[string]string{}
	for _, ln := range bytes.Split(b, []byte("\n")) {
		a := strings.SplitN(string(ln), ":", 2)
		if len(a) == 2 {
			cfg[a[0]] = strings.Trim(a[1], " ")
		}
	}

	return cfg
}

func exit(rc int) {
	scr.Fini()

	if jsonMode {
		//Avoid null in serialized JSON.
		for i := range results {
			if results[i].Mistakes == nil {
				results[i].Mistakes = []mistake{}
			}
		}

		b, err := json.Marshal(results)
		if err != nil {
			panic(err)
		}
		os.Stdout.Write(b)
	}

	if csvMode {
		for _, r := range results {
			// Write stats to file
			if err := writeCSVStats(currentTestType, r.Timestamp, r.Wpm, r.Cpm, r.Accuracy, r.File, r.N); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to write stats CSV: %v\n", err)
			}

			// Write errors to file
			if err := writeCSVErrors(currentTestType, r.Timestamp, r.Mistakes); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to write errors CSV: %v\n", err)
			}
		}
	}

	os.Exit(rc)
}

func showReport(scr tcell.Screen, cpm, wpm int, accuracy float64, attribution string, mistakes []mistake) {
	mistakeStr := ""
	if attribution != "" {
		attribution = "\n\nAttribution: " + attribution
	}

	if len(mistakes) > 0 {
		mistakeStr = "\nMistakes:    "
		for i, m := range mistakes {
			mistakeStr += m.Word
			if i != len(mistakes)-1 {
				mistakeStr += ", "
			}
		}
	}

	report := fmt.Sprintf("WPM:         %d\nCPM:         %d\nAccuracy:    %.2f%%%s%s", wpm, cpm, accuracy, mistakeStr, attribution)

	scr.Clear()
	drawStringAtCenter(scr, report, tcell.StyleDefault)
	scr.HideCursor()
	scr.Show()

	for {
		if key, ok := scr.PollEvent().(*tcell.EventKey); ok && key.Key() == tcell.KeyEscape {
			exit(0)
		} else if ok && key.Key() == tcell.KeyTab {
			return
		}
	}
}

func createDefaultTyper(scr tcell.Screen) *typer {
	return NewTyper(scr, true, tcell.ColorDefault,
		tcell.ColorDefault,
		tcell.ColorWhite,
		tcell.ColorGreen,
		tcell.ColorGreen,
		tcell.ColorMaroon)
}

func createTyper(scr tcell.Screen, bold bool, themeName string) *typer {
	var theme map[string]string

	if b := readResource("themes", themeName); b == nil {
		die("%s does not appear to be a valid theme, try '-list themes' for a list of built in thems.", themeName)
	} else {
		theme = parseConfig(b)
	}

	var bgcol, fgcol, hicol, hicol2, hicol3, errcol tcell.Color
	var err error

	if bgcol, err = newTcellColor(theme["bgcol"]); err != nil {
		die("bgcol is not defined and/or a valid hex colour.")
	}
	if fgcol, err = newTcellColor(theme["fgcol"]); err != nil {
		die("fgcol is not defined and/or a valid hex colour.")
	}
	if hicol, err = newTcellColor(theme["hicol"]); err != nil {
		die("hicol is not defined and/or a valid hex colour.")
	}
	if hicol2, err = newTcellColor(theme["hicol2"]); err != nil {
		die("hicol2 is not defined and/or a valid hex colour.")
	}
	if hicol3, err = newTcellColor(theme["hicol3"]); err != nil {
		die("hicol3 is not defined and/or a valid hex colour.")
	}
	if errcol, err = newTcellColor(theme["errcol"]); err != nil {
		die("errcol is not defined and/or a valid hex colour.")
	}

	return NewTyper(scr, bold, fgcol, bgcol, hicol, hicol2, hicol3, errcol)
}

var usage = `usage: typr [options] [file]
       typr visualize <file>

Subcommands
    visualize <file>    Display typing speed progress graph.
                        Simple filenames are looked up in the results directory.
                        Examples: quotes-stats.csv, words-stats.csv

Modes
    -words              Start word mode using the default word list from config
                        (default: 1000en).
    -wordfile WORDFILE  Override word list with specific file.
    -quotes             Start quote mode using the ZenQuotes API (cached locally
                        for offline fallback).
    -quotefile QUOTEFILE Override quote file. Use 'zen' for locally logged quotes
                        from previous ZenQuotes API sessions.
                        Quote files should be JSON encoded:
                        [{"text": "foo", "attribution": "bar"}]

Word Mode
    -n GROUPSZ          Sets the number of words which constitute a group.
    -g NGROUPS          Sets the number of groups which constitute a test.

File Mode
    -start PARAGRAPH    The offset of the starting paragraph, set this to 0 to
                        reset progress on a given file.
Aesthetics
    -showwpm            Display WPM whilst typing.
    -theme THEMEFILE    The theme to use.
    -w                  The maximum line length in characters.
    -notheme            Attempt to use the default terminal theme.
                        This may produce odd results depending
                        on the theme colours.
    -blockcursor        Use the default cursor style.
    -bold               Embolden typed text.
                        ignored if -raw is present.
Test Parameters
    -t SECONDS          Terminate the test after the given number of seconds.
    -noskip             Disable word skipping when space is pressed.
    -nobackspace        Disable the backspace key.
    -nohighlight        Disable current and next word highlighting.
    -highlight1         Only highlight the current word.
    -highlight2         Only highlight the next word.

Scripting
    -oneshot            Automatically exit after a single run.
    -noreport           Don't show a report at the end of a test.
    -csv                Write test results to CSV files in configured directory.
                        Enabled by default via config.yaml.
                        Stats: {csvdir}/{mode}-stats.csv (timestamp,wpm,cpm,accuracy)
                        Errors: {csvdir}/{mode}-errors.csv (timestamp,word,error)
                        Default dir: $XDG_DATA_HOME/typr/results or ~/.local/share/typr/results
                        Configure via: $XDG_CONFIG_HOME/typr/config.yaml or ~/.config/typr/config.yaml
    -json               Print the test output in JSON.
    -raw                Don't reflow STDIN text or show one paragraph at a time.
                        Note that line breaks are determined exclusively by the
                        input.
    -multi              Treat each input paragraph as a self contained test.

Misc
    -list TYPE          Lists internal resources of the given type.
                        TYPE=[themes|quotes|words]

Version
    -V, --version       Print the current version.
`

func saveMistakes(mistakes []mistake) {
	var db []mistake

	if err := readValue(MISTAKE_DB, &db); err != nil {
		db = nil
	}

	db = append(db, mistakes...)
	writeValue(MISTAKE_DB, db)
}

func main() {
	// Check for subcommands before processing flags
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "visualize", "viz":
			// Require file argument
			if len(os.Args) < 3 {
				fmt.Fprintf(os.Stderr, "Error: file argument required\n\n")
				fmt.Fprintf(os.Stderr, "Usage: typr visualize <file>\n")
				fmt.Fprintf(os.Stderr, "       typr visualize quotes-stats.csv\n")
				fmt.Fprintf(os.Stderr, "       typr visualize words-stats.csv\n\n")
				fmt.Fprintf(os.Stderr, "Simple filenames are looked up in: %s\n", RESULTS_DIR)
				os.Exit(1)
			}
			csvPath := os.Args[2]
			if err := runVisualize(csvPath); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		}
	}

	var n int
	var g int

	var rawMode bool
	var oneShotMode bool
	var noHighlightCurrent bool
	var noHighlightNext bool
	var noHighlight bool
	var maxLineLen int
	var noSkip bool
	var noBackspace bool
	var noReport bool
	var noTheme bool
	var normalCursor bool
	var timeout int
	var startParagraph int

	var listFlag string
	var wordFile string
	var quoteFile string
	var wordsMode bool
	var quotesMode bool
	var wordFileOverride string
	var quoteFileOverride string

	var themeName string
	var showWpm bool
	var multiMode bool
	var versionFlag bool
	var boldFlag bool

	var err error
	var testFn func() []segment

	// Load config file to set defaults
	cfg, err := loadConfig(YAML_CONFIG_FILE)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to load config file: %v\n", err)
		cfg = &AppConfig{}
		*cfg = getDefaultConfig()
	}

	// Set flag defaults from config
	flag.IntVar(&n, "n", cfg.N, "")
	flag.IntVar(&g, "g", cfg.G, "")
	flag.IntVar(&startParagraph, "start", cfg.Start, "")

	flag.IntVar(&maxLineLen, "w", cfg.W, "")
	flag.IntVar(&timeout, "t", cfg.T, "")

	flag.BoolVar(&versionFlag, "version", false, "")
	flag.BoolVar(&versionFlag, "V", false, "")

	flag.BoolVar(&wordsMode, "words", false, "")
	flag.StringVar(&wordFileOverride, "wordfile", "", "")
	flag.BoolVar(&quotesMode, "quotes", false, "")
	flag.StringVar(&quoteFileOverride, "quotefile", "", "")

	flag.BoolVar(&showWpm, "showwpm", cfg.ShowWpm, "")
	flag.BoolVar(&noSkip, "noskip", cfg.NoSkip, "")
	flag.BoolVar(&normalCursor, "blockcursor", cfg.BlockCursor, "")
	flag.BoolVar(&noBackspace, "nobackspace", cfg.NoBackspace, "")
	flag.BoolVar(&noTheme, "notheme", cfg.NoTheme, "")
	flag.BoolVar(&oneShotMode, "oneshot", cfg.OneShot, "")
	flag.BoolVar(&noHighlight, "nohighlight", cfg.NoHighlight, "")
	flag.BoolVar(&noHighlightCurrent, "highlight2", cfg.Highlight2, "")
	flag.BoolVar(&noHighlightNext, "highlight1", cfg.Highlight1, "")
	flag.BoolVar(&noReport, "noreport", cfg.NoReport, "")
	flag.BoolVar(&boldFlag, "bold", cfg.Bold, "")
	flag.BoolVar(&csvMode, "csv", cfg.Csv, "")
	flag.BoolVar(&jsonMode, "json", cfg.Json, "")
	flag.BoolVar(&rawMode, "raw", cfg.Raw, "")
	flag.BoolVar(&multiMode, "multi", cfg.Multi, "")
	flag.StringVar(&themeName, "theme", cfg.Theme, "")
	flag.StringVar(&listFlag, "list", "", "")

	flag.Usage = func() { os.Stdout.Write([]byte(usage)) }
	flag.Parse()

	// Detect which mode flags were explicitly provided by the user
	var wordsExplicit, quotesExplicit bool
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "words" || f.Name == "wordfile" {
			wordsExplicit = true
		}
		if f.Name == "quotes" || f.Name == "quotefile" {
			quotesExplicit = true
		}
	})

	// Resolve actual values with precedence: override > config > hardcoded default
	if wordFileOverride != "" {
		wordFile = wordFileOverride
	} else if wordsExplicit || cfg.Words != "" {
		wordFile = cfg.Words
	}
	if wordFile == "" {
		wordFile = "1000en" // Hardcoded fallback
	}

	if quoteFileOverride != "" {
		quoteFile = quoteFileOverride
	} else if quotesExplicit || cfg.Quotes != "" {
		quoteFile = cfg.Quotes
	}
	// Note: quoteFile can remain empty (defaults to words mode)

	// Store the n value for CSV logging (already resolved from flag or config)
	currentTestN = n

	if listFlag != "" {
		prefix := listFlag + "/"
		for path, _ := range packedFiles {
			if strings.Index(path, prefix) == 0 {
				_, f := filepath.Split(path)
				fmt.Println(f)
			}
		}

		// Special case: 'zen' uses local zenlog, not embedded
		if listFlag == "quotes" {
			fmt.Println("zen")
		}

		os.Exit(0)
	}

	if versionFlag {
		fmt.Fprintf(os.Stderr, "typr version 1.0.0\n")
		os.Exit(1)
	}

	if noTheme {
		os.Setenv("TCELL_TRUECOLOR", "disable")
	}

	reflow := func(s string) string {
		sw, _ := scr.Size()

		wsz := maxLineLen
		if wsz > sw {
			wsz = sw - 8
		}

		s = regexp.MustCompile("\\s+").ReplaceAllString(s, " ")
		return strings.Replace(
			wordWrap(strings.Trim(s, " "), wsz),
			"\n", " \n", -1)
	}

	switch {
	case wordsExplicit:
		// User explicitly provided -words flag
		testFn = generateWordTest(wordFile, n, g)
		currentTestType = "words"
		currentTestFile = wordFile
	case quotesExplicit:
		// User explicitly provided -quotes or -quotefile flag
		switch quoteFile {
		case "": // bare -quotes: use API with zenlog fallback
			testFn = generateZenQuotesTest()
			currentTestType = "quotes"
			currentTestFile = "zen-api"
		case "zen": // -quotefile zen: use local zenlog only
			testFn = generateZenlogTest()
			currentTestType = "quotes"
			currentTestFile = "zen"
		default:
			testFn = generateQuoteTest(quoteFile)
			currentTestType = "quotes"
			currentTestFile = quoteFile
		}
	case !isatty.IsTerminal(os.Stdin.Fd()):
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			panic(err)
		}

		testFn = generateTestFromData(b, rawMode, multiMode)
		currentTestType = "stdin"
		currentTestFile = "stdin"
	case len(flag.Args()) > 0:
		path := flag.Args()[0]
		testFn = generateTestFromFile(path, startParagraph)
		currentTestType = "file"
		currentTestFile = filepath.Base(path)
	default:
		// Use config defaults: check if quotes are configured
		if quoteFile != "" {
			switch quoteFile {
			case "zen": // config: quotes: zen → local zenlog
				testFn = generateZenlogTest()
				currentTestType = "quotes"
				currentTestFile = "zen"
			default:
				testFn = generateQuoteTest(quoteFile)
				currentTestType = "quotes"
				currentTestFile = quoteFile
			}
		} else {
			// Fall back to words mode with default
			testFn = generateWordTest(wordFile, n, g)
			currentTestType = "words"
			currentTestFile = wordFile
		}
	}

	scr, err = tcell.NewScreen()
	if err != nil {
		panic(err)
	}

	if err := scr.Init(); err != nil {
		panic(err)
	}

	defer func() {
		if r := recover(); r != nil {
			scr.Fini()
			panic(r)
		}
	}()

	var typer *typer
	if noTheme {
		typer = createDefaultTyper(scr)
	} else {
		typer = createTyper(scr, boldFlag, themeName)
	}

	if noHighlightNext || noHighlight {
		typer.currentWordStyle = typer.nextWordStyle
		typer.nextWordStyle = typer.defaultStyle
	}

	if noHighlightCurrent || noHighlight {
		typer.currentWordStyle = typer.defaultStyle
	}

	typer.SkipWord = !noSkip
	typer.DisableBackspace = noBackspace
	typer.BlockCursor = normalCursor
	typer.ShowWpm = showWpm

	if timeout != -1 {
		timeout *= 1E9
	}

	var tests [][]segment
	var idx = 0

	for {
		if idx >= len(tests) {
			tests = append(tests, testFn())
		}

		if tests[idx] == nil {
			exit(0)
		}

		if !rawMode {
			for i, _ := range tests[idx] {
				tests[idx][i].Text = reflow(tests[idx][i].Text)
			}
		}

		nerrs, ncorrect, t, rc, mistakes := typer.Start(tests[idx], time.Duration(timeout))
		saveMistakes(mistakes)

		switch rc {
		case TyperNext:
			idx++
		case TyperPrevious:
			if idx > 0 {
				idx--
			}
		case TyperComplete:
			cpm := int(float64(ncorrect) / (float64(t) / 60E9))
			wpm := cpm / 5
			accuracy := float64(ncorrect) / float64(nerrs+ncorrect) * 100

			results = append(results, result{wpm, cpm, accuracy, time.Now().Unix(), mistakes, currentTestFile, currentTestN})
			if !noReport {
				attribution := ""
				if len(tests[idx]) == 1 {
					attribution = tests[idx][0].Attribution
				}
				showReport(scr, cpm, wpm, accuracy, attribution, mistakes)
			}
			if oneShotMode {
				exit(0)
			}

			idx++
		case TyperSigInt:
			exit(1)

		case TyperQuit:
			exit(0)

		case TyperResize:
			//Resize events restart the test, this shouldn't be a problem in the vast majority of cases
			//and allows us to avoid baking rewrapping logic into the typer.

			//TODO: implement state-preserving resize (maybe)
		}
	}
}
