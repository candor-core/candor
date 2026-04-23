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

func parseLine(line string) (string, error) {
	if !strings.HasPrefix(line, "[") {
		return "", fmt.Errorf("invalid format")
	}
	closeBracket := strings.Index(line, "]")
	if closeBracket < 0 {
		return "", fmt.Errorf("invalid format")
	}
	level := line[1:closeBracket]
	switch level {
	case "ERROR", "WARN", "INFO", "DEBUG":
		return level, nil
	default:
		return "", fmt.Errorf("invalid level: %s", level)
	}
}

func aggregateFile(path string) (FileStats, error) {
	stats := FileStats{
		Filename: filepath.Base(path),
	}

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

		level, err := parseLine(line)
		if err != nil {
			stats.ParseErrors++
			continue
		}

		switch level {
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
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <logfile> [logfile...]\n", os.Args[0])
		os.Exit(1)
	}

	paths := os.Args[1:]

	totalFiles := 0
	totalLines := 0
	totalParseErrors := 0

	for _, path := range paths {
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

		fileLines := stats.ErrorCount + stats.WarnCount + stats.InfoCount + stats.DebugCount + stats.ParseErrors
		totalFiles++
		totalLines += fileLines
		totalParseErrors += stats.ParseErrors
	}

	fmt.Printf("TOTAL: files=%d lines=%d parse_errors=%d\n", totalFiles, totalLines, totalParseErrors)
	os.Exit(0)
}
