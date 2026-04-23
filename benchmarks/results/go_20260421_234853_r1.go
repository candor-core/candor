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

func (s FileStats) TotalLines() int {
	return s.ErrorCount + s.WarnCount + s.InfoCount + s.DebugCount + s.ParseErrors
}

type ParseResult struct {
	Level   string
	Message string
	IsValid bool
}

func parseLine(line string) ParseResult {
	if !strings.HasPrefix(line, "[") {
		return ParseResult{IsValid: false}
	}
	closeBracket := strings.Index(line, "]")
	if closeBracket < 0 {
		return ParseResult{IsValid: false}
	}
	level := line[1:closeBracket]
	switch level {
	case "ERROR", "WARN", "INFO", "DEBUG":
	default:
		return ParseResult{IsValid: false}
	}
	rest := line[closeBracket+1:]
	if len(rest) == 0 || rest[0] != ' ' {
		return ParseResult{IsValid: false}
	}
	message := rest[1:]
	return ParseResult{Level: level, Message: message, IsValid: true}
}

func aggregateFile(path string) (FileStats, error) {
	f, err := os.Open(path)
	if err != nil {
		return FileStats{}, err
	}
	defer f.Close()

	stats := FileStats{
		Filename: filepath.Base(path),
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		result := parseLine(line)
		if !result.IsValid {
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
		return FileStats{}, err
	}

	return stats, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: logbatch <file1> [file2 ...]")
		os.Exit(1)
	}

	paths := os.Args[1:]
	var allStats []FileStats
	totalLines := 0
	totalParseErrors := 0

	for _, path := range paths {
		stats, err := aggregateFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", path, err)
			os.Exit(1)
		}
		allStats = append(allStats, stats)
		totalLines += stats.TotalLines()
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
