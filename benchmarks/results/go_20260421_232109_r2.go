```go
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type FileStats struct {
	Filename    string
	ErrorCount  int
	WarnCount   int
	InfoCount   int
	DebugCount  int
	ParseErrors int
}

type ParseResult struct {
	Level   string
	Message string
	Valid   bool
}

func parseLine(line string) ParseResult {
	line = strings.TrimSpace(line)
	if line == "" {
		return ParseResult{Valid: false}
	}

	if !strings.HasPrefix(line, "[") {
		return ParseResult{Valid: false}
	}

	closeBracket := strings.Index(line, "]")
	if closeBracket < 0 {
		return ParseResult{Valid: false}
	}

	level := line[1:closeBracket]
	message := strings.TrimSpace(line[closeBracket+1:])

	switch level {
	case "ERROR", "WARN", "INFO", "DEBUG":
		return ParseResult{Level: level, Message: message, Valid: true}
	default:
		return ParseResult{Valid: false}
	}
}

func aggregateFileStats(filename string, lines []string) FileStats {
	stats := FileStats{Filename: filename}

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		result := parseLine(line)
		if !result.Valid {
			stats.ParseErrors++
			continue
		}

		switch result.Level {
		case "ERROR":
			stats.ErrorCount++
		case "WARN":
			stats.WarnCount++
		case "INFO":
			stats.InfoCount++
		case "DEBUG":
			stats.DebugCount++
		}
	}

	return stats
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: logbatch <directory>")
		os.Exit(1)
	}

	dirPath := os.Args[1]

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading directory: %v\n", err)
		os.Exit(1)
	}

	var logFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".log") {
			logFiles = append(logFiles, entry.Name())
		}
	}

	sort.Strings(logFiles)

	var allStats []FileStats
	totalLines := 0
	totalParseErrors := 0

	for _, filename := range logFiles {
		fullPath := filepath.Join(dirPath, filename)

		f, err := os.Open(fullPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file %s: %v\n", filename, err)
			continue
		}

		var lines []string
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		f.Close()

		stats := aggregateFileStats(filename, lines)
		allStats = append(allStats, stats)

		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				totalLines++
			}
		}
		totalParseErrors += stats.ParseErrors
	}

	for _, stats := range allStats {
		fmt.Printf("%s: ERROR=%d WARN=%d INFO=%d DEBUG=%d parse_errors=%d\n",
			stats.Filename,
			stats.ErrorCount,
			stats.WarnCount,
			stats.InfoCount,
			stats.DebugCount,
			stats.ParseErrors,
		)
	}

	fmt.Printf("TOTAL: files=%d lines=%d parse_errors=%d\n",
		len(allStats),
		totalLines,
		totalParseErrors,
	)

	os.Exit(0)
}
```