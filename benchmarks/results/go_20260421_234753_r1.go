package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type FileStats struct {
	ErrorCount      int
	WarnCount       int
	InfoCount       int
	DebugCount      int
	ParseErrorCount int
}

func parseLine(line string) (string, bool) {
	if !strings.HasPrefix(line, "[") {
		return "", false
	}
	closeBracket := strings.Index(line, "]")
	if closeBracket < 0 {
		return "", false
	}
	level := line[1:closeBracket]
	switch level {
	case "ERROR", "WARN", "INFO", "DEBUG":
		return level, true
	default:
		return "", false
	}
}

func aggregateStats(lines []string) FileStats {
	var stats FileStats
	for _, line := range lines {
		level, ok := parseLine(line)
		if !ok {
			stats.ParseErrorCount++
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
	return stats
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "No log files specified")
		os.Exit(1)
	}

	totalFiles := len(args)
	totalLines := 0
	totalParseErrors := 0

	for _, filePath := range args {
		f, err := os.Open(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot read file: %s\n", filePath)
			os.Exit(1)
		}

		var nonBlankLines []string
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.TrimSpace(line) == "" {
				continue
			}
			nonBlankLines = append(nonBlankLines, line)
		}
		f.Close()

		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %s\n", filePath)
			os.Exit(1)
		}

		stats := aggregateStats(nonBlankLines)
		filename := filepath.Base(filePath)

		fmt.Printf("%s: ERROR=%d WARN=%d INFO=%d DEBUG=%d parse_errors=%d\n",
			filename,
			stats.ErrorCount,
			stats.WarnCount,
			stats.InfoCount,
			stats.DebugCount,
			stats.ParseErrorCount,
		)

		totalLines += len(nonBlankLines)
		totalParseErrors += stats.ParseErrorCount
	}

	fmt.Printf("TOTAL: files=%d lines=%d parse_errors=%d\n", totalFiles, totalLines, totalParseErrors)
	os.Exit(0)
}
