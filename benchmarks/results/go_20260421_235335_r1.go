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

func aggregateStats(filename string, lines []string) FileStats {
	stats := FileStats{
		Filename: filepath.Base(filename),
	}
	for _, line := range lines {
		if line == "" {
			continue
		}
		level, ok := parseLine(line)
		if !ok {
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
	return stats
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: logbatch <file1> [file2] ...")
		os.Exit(1)
	}

	files := os.Args[1:]
	totalFiles := len(files)
	totalLines := 0
	totalParseErrors := 0

	for _, filePath := range files {
		f, err := os.Open(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file %s: %v\n", filePath, err)
			os.Exit(1)
		}

		var lines []string
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			lines = append(lines, line)
		}
		f.Close()

		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filePath, err)
			os.Exit(1)
		}

		stats := aggregateStats(filePath, lines)

		fileLines := stats.ErrorCount + stats.WarnCount + stats.InfoCount + stats.DebugCount + stats.ParseErrors
		totalLines += fileLines
		totalParseErrors += stats.ParseErrors

		fmt.Printf("%s: ERROR=%d WARN=%d INFO=%d DEBUG=%d parse_errors=%d\n",
			stats.Filename,
			stats.ErrorCount,
			stats.WarnCount,
			stats.InfoCount,
			stats.DebugCount,
			stats.ParseErrors,
		)
	}

	fmt.Printf("TOTAL: files=%d lines=%d parse_errors=%d\n", totalFiles, totalLines, totalParseErrors)
	os.Exit(0)
}
