package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type LineResult struct {
	Level      string
	Message    string
	ParseError bool
}

type FileStats struct {
	Filename    string
	Counts      map[string]int
	ParseErrors int
	TotalLines  int
}

func parseLine(line string) LineResult {
	validLevels := map[string]bool{
		"ERROR": true,
		"WARN":  true,
		"INFO":  true,
		"DEBUG": true,
	}

	line = strings.TrimSpace(line)
	if line == "" {
		return LineResult{ParseError: true}
	}

	if !strings.HasPrefix(line, "[") {
		return LineResult{ParseError: true}
	}

	closeBracket := strings.Index(line, "]")
	if closeBracket == -1 {
		return LineResult{ParseError: true}
	}

	level := line[1:closeBracket]
	if !validLevels[level] {
		return LineResult{ParseError: true}
	}

	message := ""
	if closeBracket+1 < len(line) {
		message = strings.TrimSpace(line[closeBracket+1:])
	}

	return LineResult{
		Level:      level,
		Message:    message,
		ParseError: false,
	}
}

func aggregateFileStats(filename string, lines []string) FileStats {
	stats := FileStats{
		Filename: filename,
		Counts: map[string]int{
			"ERROR": 0,
			"WARN":  0,
			"INFO":  0,
			"DEBUG": 0,
		},
	}

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		stats.TotalLines++
		result := parseLine(line)
		if result.ParseError {
			stats.ParseErrors++
		} else {
			stats.Counts[result.Level]++
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

	for _, path := range args {
		file, err := os.Open(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: cannot read file %s: %v\n", path, err)
			os.Exit(1)
		}

		var lines []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		file.Close()

		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", path, err)
			os.Exit(1)
		}

		filename := filepath.Base(path)
		stats := aggregateFileStats(filename, lines)

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
