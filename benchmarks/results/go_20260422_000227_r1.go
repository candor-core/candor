package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
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

func (s *FileStats) TotalLines() int {
	return s.ErrorCount + s.WarnCount + s.InfoCount + s.DebugCount + s.ParseErrors
}

type ParseResult struct {
	Level   string
	Message string
	Valid   bool
}

func parseLine(line string) ParseResult {
	if !strings.HasPrefix(line, "[") {
		return ParseResult{Valid: false}
	}
	closeBracket := strings.Index(line, "]")
	if closeBracket < 0 {
		return ParseResult{Valid: false}
	}
	level := line[1:closeBracket]
	rest := line[closeBracket+1:]
	if !strings.HasPrefix(rest, " ") {
		return ParseResult{Valid: false}
	}
	message := rest[1:]
	switch level {
	case "ERROR", "WARN", "INFO", "DEBUG":
		return ParseResult{Level: level, Message: message, Valid: true}
	default:
		return ParseResult{Valid: false}
	}
}

func aggregateFile(path string) (FileStats, error) {
	filename := filepath.Base(path)
	stats := FileStats{Filename: filename}

	f, err := os.Open(path)
	if err != nil {
		return stats, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
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

	if err := scanner.Err(); err != nil {
		return stats, err
	}

	return stats, nil
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
		stats, err := aggregateFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", path, err)
			os.Exit(1)
		}

		fmt.Printf("%s: ERROR=%d WARN=%d INFO=%d DEBUG=%d parse_errors=%d\n",
			stats.Filename,
			stats.ErrorCount,
			stats.WarnCount,
			stats.InfoCount,
			stats.DebugCount,
			stats.ParseErrors,
		)

		totalFiles++
		totalLines += stats.TotalLines()
		totalParseErrors += stats.ParseErrors
	}

	fmt.Printf("TOTAL: files=%d lines=%d parse_errors=%d\n", totalFiles, totalLines, totalParseErrors)
	os.Exit(0)
}
