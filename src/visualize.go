package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/guptarohit/asciigraph"
)

// CSVRecord represents a single row from the stats CSV file
type CSVRecord struct {
	Timestamp int64
	WPM       int
	CPM       int
	Accuracy  float64
	File      string
	N         int
}

// DailyStats holds aggregated statistics for a single day
type DailyStats struct {
	Date      time.Time
	MinWPM    float64
	MeanWPM   float64
	MaxWPM    float64
	TestCount int
}

// runVisualize is the main entry point for the visualize subcommand
func runVisualize(csvPath string) error {
	// If path is just a filename (no directory separators), prepend results directory
	if !strings.Contains(csvPath, string(filepath.Separator)) && len(csvPath) > 0 && csvPath[0] != '~' && csvPath[0] != '/' {
		csvPath = filepath.Join(RESULTS_DIR, csvPath)
	}

	// Expand home directory if needed
	if len(csvPath) > 0 && csvPath[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("unable to determine home directory: %w", err)
		}
		csvPath = filepath.Join(home, csvPath[1:])
	}

	// Check if file exists
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		return fmt.Errorf("CSV file not found at %s\nRun tests with the -csv flag first to generate data", csvPath)
	}

	// Read CSV file
	records, err := readCSVFile(csvPath)
	if err != nil {
		return fmt.Errorf("error reading CSV file: %w", err)
	}

	if len(records) == 0 {
		return fmt.Errorf("no test data found in CSV file\nRun some tests with -csv flag to collect data")
	}

	if len(records) < 2 {
		return fmt.Errorf("need at least 2 tests to visualize progress\nCurrent: %d test", len(records))
	}

	// Aggregate data by date (last 30 days)
	dailyStats := aggregateByDate(records, 30)

	if len(dailyStats) == 0 {
		return fmt.Errorf("no data found in the last 30 days")
	}

	// Generate and display the plot
	if err := displayPlot(dailyStats, csvPath); err != nil {
		return fmt.Errorf("error generating plot: %w", err)
	}

	return nil
}

// readCSVFile reads and parses a CSV stats file
func readCSVFile(path string) ([]CSVRecord, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return nil, nil
	}

	// Skip header row
	var records []CSVRecord
	for i, row := range rows {
		if i == 0 {
			continue
		}

		// Parse fields with error handling
		if len(row) < 4 {
			fmt.Fprintf(os.Stderr, "Warning: Skipping invalid record at line %d (insufficient fields)\n", i+1)
			continue
		}

		record := CSVRecord{}

		// Parse timestamp
		timestamp, err := strconv.ParseInt(row[0], 10, 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Skipping invalid record at line %d (bad timestamp)\n", i+1)
			continue
		}
		record.Timestamp = timestamp

		// Parse WPM
		wpm, err := strconv.Atoi(row[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Skipping invalid record at line %d (bad WPM)\n", i+1)
			continue
		}
		record.WPM = wpm

		// Parse CPM
		cpm, err := strconv.Atoi(row[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Skipping invalid record at line %d (bad CPM)\n", i+1)
			continue
		}
		record.CPM = cpm

		// Parse accuracy
		accuracy, err := strconv.ParseFloat(row[3], 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Skipping invalid record at line %d (bad accuracy)\n", i+1)
			continue
		}
		record.Accuracy = accuracy

		// Parse optional fields
		if len(row) > 4 {
			record.File = row[4]
		}
		if len(row) > 5 && row[5] != "" {
			n, err := strconv.Atoi(row[5])
			if err == nil {
				record.N = n
			}
		}

		records = append(records, record)
	}

	return records, nil
}

// aggregateByDate groups records by date and calculates daily statistics
func aggregateByDate(records []CSVRecord, daysBack int) []DailyStats {
	// Calculate cutoff date
	now := time.Now()
	cutoff := now.AddDate(0, 0, -daysBack)

	// Group records by date
	dateMap := make(map[string][]int)
	for _, record := range records {
		timestamp := time.Unix(record.Timestamp, 0)
		if timestamp.Before(cutoff) {
			continue
		}

		dateStr := timestamp.Format("2006-01-02")
		dateMap[dateStr] = append(dateMap[dateStr], record.WPM)
	}

	// Calculate statistics for each date
	var stats []DailyStats
	for dateStr, wpmValues := range dateMap {
		date, _ := time.Parse("2006-01-02", dateStr)

		minWPM := float64(wpmValues[0])
		maxWPM := float64(wpmValues[0])
		sum := 0

		for _, wpm := range wpmValues {
			if float64(wpm) < minWPM {
				minWPM = float64(wpm)
			}
			if float64(wpm) > maxWPM {
				maxWPM = float64(wpm)
			}
			sum += wpm
		}

		meanWPM := float64(sum) / float64(len(wpmValues))

		stats = append(stats, DailyStats{
			Date:      date,
			MinWPM:    minWPM,
			MeanWPM:   meanWPM,
			MaxWPM:    maxWPM,
			TestCount: len(wpmValues),
		})
	}

	// Sort by date ascending
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Date.Before(stats[j].Date)
	})

	return stats
}

// displayPlot generates and displays the ASCII graph
func displayPlot(stats []DailyStats, csvPath string) error {
	// Extract series data
	minSeries := make([]float64, len(stats))
	meanSeries := make([]float64, len(stats))
	maxSeries := make([]float64, len(stats))

	totalTests := 0
	for i, stat := range stats {
		minSeries[i] = stat.MinWPM
		meanSeries[i] = stat.MeanWPM
		maxSeries[i] = stat.MaxWPM
		totalTests += stat.TestCount
	}

	// Format dates for display
	var dateLabels []string
	if len(stats) <= 5 {
		// Show all dates if few data points
		for _, stat := range stats {
			dateLabels = append(dateLabels, stat.Date.Format("Jan 02"))
		}
	} else {
		// Show first, middle, and last dates
		first := stats[0].Date.Format("Jan 02")
		middle := stats[len(stats)/2].Date.Format("Jan 02")
		last := stats[len(stats)-1].Date.Format("Jan 02")
		dateLabels = []string{first, middle, last}
	}

	// Generate caption
	dateRange := fmt.Sprintf("%s - %s",
		stats[0].Date.Format("Jan 02"),
		stats[len(stats)-1].Date.Format("Jan 02"))

	// Create the plot
	graph := asciigraph.PlotMany(
		[][]float64{minSeries, meanSeries, maxSeries},
		asciigraph.Height(15),
		asciigraph.Width(60),
		asciigraph.Caption(fmt.Sprintf("Typing Speed Progress (%s)", dateRange)),
		asciigraph.SeriesColors(
			asciigraph.Blue,  // Min WPM
			asciigraph.Green, // Mean WPM
			asciigraph.Red,   // Max WPM
		),
	)

	// Display the plot
	fmt.Println()
	fmt.Println(graph)
	fmt.Println()

	// Display legend
	fmt.Println("   \033[34m━━━\033[0m Min WPM    \033[32m━━━\033[0m Mean WPM    \033[31m━━━\033[0m Max WPM")
	fmt.Println()

	// Display summary
	fmt.Printf("%d tests completed over %d days\n", totalTests, len(stats))
	fmt.Println("Press Enter to exit")

	var input string
	fmt.Scanln(&input)

	return nil
}
