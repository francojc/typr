package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"os"
	"time"
)

// ZenQuote represents a single quote from the ZenQuotes API
type ZenQuote struct {
	Text   string `json:"q"`
	Author string `json:"a"`
}

// quoteCache holds the last N quotes fetched from the API
var quoteCache []segment
var cacheMaxSize = 10

// fallbackQuote is used when API fails and cache is empty
var fallbackQuote = segment{
	Text:        "The only way to do great work is to love what you do.",
	Attribution: "Steve Jobs",
}

// loadZenlog reads the zenlog file and returns stored quotes.
// Returns empty slice if file doesn't exist or is invalid.
func loadZenlog() []segment {
	data, err := os.ReadFile(ZENLOG_FILE)
	if err != nil {
		return []segment{}
	}

	var quotes []segment
	if err := json.Unmarshal(data, &quotes); err != nil {
		return []segment{}
	}

	return quotes
}

// quoteExists checks if a quote with same text exists (deduplication).
func quoteExists(quotes []segment, newQuote segment) bool {
	for _, q := range quotes {
		if q.Text == newQuote.Text {
			return true
		}
	}
	return false
}

// appendToZenlog adds a new quote if it doesn't already exist.
// Returns true if appended, false if duplicate or error.
func appendToZenlog(quote segment) bool {
	quotes := loadZenlog()

	if quoteExists(quotes, quote) {
		return false
	}

	quotes = append(quotes, quote)

	data, err := json.MarshalIndent(quotes, "", "  ")
	if err != nil {
		return false
	}

	if err := os.WriteFile(ZENLOG_FILE, data, 0644); err != nil {
		return false
	}

	return true
}

// fetchZenQuote fetches a random quote from the ZenQuotes API
func fetchZenQuote() (segment, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get("https://zenquotes.io/api/random")
	if err != nil {
		return segment{}, err
	}
	defer resp.Body.Close()

	var quotes []ZenQuote
	if err := json.NewDecoder(resp.Body).Decode(&quotes); err != nil {
		return segment{}, err
	}

	if len(quotes) == 0 {
		return segment{}, err
	}

	// Convert to segment format
	result := segment{
		Text:        quotes[0].Text,
		Attribution: quotes[0].Author,
	}

	// Add to cache
	quoteCache = append(quoteCache, result)
	if len(quoteCache) > cacheMaxSize {
		quoteCache = quoteCache[1:]
	}

	// Log to zenlog file (errors silently ignored)
	appendToZenlog(result)

	return result, nil
}

// getQuoteWithFallback attempts to fetch from API, falls back to cache,
// then falls back to zenlog file, then falls back to default quote
func getQuoteWithFallback() segment {
	// Try to fetch from API
	quote, err := fetchZenQuote()
	if err == nil {
		return quote
	}

	// Fall back to cache
	if len(quoteCache) > 0 {
		idx := rand.Intn(len(quoteCache))
		return quoteCache[idx]
	}

	// Fall back to zenlog file
	zenlogQuotes := loadZenlog()
	if len(zenlogQuotes) > 0 {
		idx := rand.Intn(len(zenlogQuotes))
		return zenlogQuotes[idx]
	}

	// Fall back to hardcoded quote
	return fallbackQuote
}

// generateZenQuotesTest returns a function that fetches from the API
// (with zenlog fallback) each time it's called.
func generateZenQuotesTest() func() []segment {
	return func() []segment {
		return []segment{getQuoteWithFallback()}
	}
}

// generateZenlogTest returns a function that picks a random quote from
// the local zenlog file (previously cached API quotes). Dies if empty.
func generateZenlogTest() func() []segment {
	quotes := loadZenlog()
	if len(quotes) == 0 {
		die("No locally logged quotes found. Run typr -quotes first to cache quotes from the ZenQuotes API.")
	}
	return func() []segment {
		idx := rand.Intn(len(quotes))
		return []segment{quotes[idx]}
	}
}
