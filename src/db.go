package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

var FILE_STATE_DB string
var MISTAKE_DB string
var RESULTS_DIR string
var CONFIG_FILE string
var ZENLOG_FILE string

func init() {
	var ok bool
	var data string
	var home string

	if home, ok = os.LookupEnv("HOME"); !ok {
		die("Could not resolve home directory.")
	}

	if data, ok = os.LookupEnv("XDG_DATA_HOME"); ok {
		data = filepath.Join(data, "/typr")
	} else {
		data = filepath.Join(home, "/.local/share/typr")
	}

	os.MkdirAll(data, 0700)

	FILE_STATE_DB = filepath.Join(data, ".db")
	MISTAKE_DB = filepath.Join(data, ".errors")

	// Set up quotes directory for zenlog
	quotesDir := filepath.Join(data, "quotes")
	os.MkdirAll(quotesDir, 0700)
	ZENLOG_FILE = filepath.Join(quotesDir, "zenlog.json")

	// Set up config directory
	var configDir string
	if configDir, ok = os.LookupEnv("XDG_CONFIG_HOME"); ok {
		configDir = filepath.Join(configDir, "typr")
	} else {
		configDir = filepath.Join(home, ".config/typr")
	}

	os.MkdirAll(configDir, 0700)
	CONFIG_FILE = filepath.Join(configDir, "config.json")
	YAML_CONFIG_FILE = filepath.Join(configDir, "config.yaml")

	// Ensure YAML config file exists (create with defaults if missing)
	if err := ensureConfigExists(YAML_CONFIG_FILE); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to initialize config file: %v\n", err)
	}

	// Load CSV directory from config, default to data/results
	RESULTS_DIR = loadCSVDir(data)
	os.MkdirAll(RESULTS_DIR, 0700)
}

func readValue(path string, o interface{}) error {
	b, err := os.ReadFile(path)

	if err != nil {
		return err
	}

	return json.Unmarshal(b, o)
}

func writeValue(path string, o interface{}) {
	b, err := json.Marshal(o)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(path, b, 0600)
	if err != nil {
		panic(err)
	}
}

type Config struct {
	CSVDir string `json:"csvdir"`
}

func loadCSVDir(defaultData string) string {
	var config Config

	// Try to read config file
	if err := readValue(CONFIG_FILE, &config); err != nil {
		// Config doesn't exist or invalid, use default
		return filepath.Join(defaultData, "results")
	}

	// If csvdir is set in config, use it
	if config.CSVDir != "" {
		// Expand ~ to home directory if present
		if config.CSVDir[0] == '~' {
			home, _ := os.LookupEnv("HOME")
			return filepath.Join(home, config.CSVDir[1:])
		}
		return config.CSVDir
	}

	// Config exists but no csvdir set, use default
	return filepath.Join(defaultData, "results")
}

func writeCSVStats(testType string, timestamp int64, wpm, cpm int, accuracy float64, file string, n int) error {
	filename := filepath.Join(RESULTS_DIR, testType+"-stats.csv")

	// Check if file exists to determine if header is needed
	needsHeader := false
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		needsHeader = true
	}

	// Open file in append mode
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write header if new file
	if needsHeader {
		if _, err := fmt.Fprintf(f, "timestamp,wpm,cpm,accuracy,file,n\n"); err != nil {
			return err
		}
	}

	// Write stats row
	_, err = fmt.Fprintf(f, "%d,%d,%d,%.2f,%s,%d\n", timestamp, wpm, cpm, accuracy, file, n)
	return err
}

func writeCSVErrors(testType string, timestamp int64, mistakes []mistake) error {
	if len(mistakes) == 0 {
		return nil // No errors to write
	}

	filename := filepath.Join(RESULTS_DIR, testType+"-errors.csv")

	// Check if file exists to determine if header is needed
	needsHeader := false
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		needsHeader = true
	}

	// Open file in append mode
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write header if new file
	if needsHeader {
		if _, err := fmt.Fprintf(f, "timestamp,word,error\n"); err != nil {
			return err
		}
	}

	// Write error rows (one per mistake)
	for _, m := range mistakes {
		if _, err := fmt.Fprintf(f, "%d,%s,%s\n", timestamp, m.Word, m.Typed); err != nil {
			return err
		}
	}

	return nil
}
