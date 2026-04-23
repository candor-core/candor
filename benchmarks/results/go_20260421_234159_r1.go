package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type LogEntry struct {
	Level   string
	Message string
	Valid   bool
}

type FileStats struct {
	Filename    string
	Counts      map[string]int
	ParseErrors int
	TotalLines  int
}

func parseLine(line string) LogEntry {
	line = strings.TrimSpace(line)
	if line == "" {
		return LogEntry{Valid: false}
	}

	if !strings.HasPrefix(line, "[") {
		return LogEntry{Valid: false}
	}

	closeBracket := strings.Index(line, "]")
	if closeBracket == -1 {
		return LogEntry{Valid: false}
	}

	level := line[1:closeBracket]
	validLevels := map[string]bool{
		"ERROR": true,
		"WARN":  true,
		"INFO":  true,
		"DEBUG": true,
	}

	if !validLevels[level] {
		return LogEntry{Valid: false}
	}

	message := strings.TrimSpace(line[closeBracket+1:])
	return LogEntry{
		Level:   level,
		Message: message,
		Valid:   true,
	}
}

func aggregateFileStats(filename string, lines []string) FileStats {
	stats := FileStats{
		Filename: filepath.Base(filename),
		Counts: map[string]int{
			"ERROR": 0,
			"WARN":  0,
			"INFO":  0,
			"DEBUG": 0,
		},
		ParseErrors: 0,
		TotalLines:  0,
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		stats.TotalLines++
		entry := parseLine(trimmed)
		if entry.Valid {
			stats.Counts[entry.Level]++
		} else {
			stats.ParseErrors++
		}
	}

	return stats
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: logbatch <file1> [file2 ...]")
		os.Exit(1)
	}

	totalFiles := 0
	totalLines := 0
	totalParseErrors := 0

	for _, filePath := range args {
		f, err := os.Open(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file %s: %v\n", filePath, err)
			os.Exit(1)
		}

		var lines []string
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		f.Close()

		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filePath, err)
			os.Exit(1)
		}

		stats := aggregateFileStats(filePath, lines)

		fmt.Printf("%s: ERROR=%d WARN=%d INFO=%d DEBUG=%d parse_errors=%d\n",
			stats.Filename,
			stats.Counts["ERROR"],
			stats.Counts["WARN"],
			stats.Counts["INFO"],
			stats.Counts["DEBUG"],
			stats.ParseErrors,
		)

		totalFiles++
		totalLines += stats.TotalLines
		totalParseErrors += stats.ParseErrors
	}

	fmt.Printf("TOTAL: files=%d lines=%d parse_errors=%d\n", totalFiles, totalLines, totalParseErrors)
	os.Exit(0)
}
