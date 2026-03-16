// Copyright (c) 2026 Scott W. Corley
// SPDX-License-Identifier: Apache-2.0
// https://github.com/scottcorleyg/candor

// Package diagnostics provides rich compiler diagnostics with source
// snippets, carets, severity levels, and hint messages.
//
// Usage:
//
//	sm := diagnostics.NewSourceMap(map[string]string{"main.cnd": src})
//	for _, d := range diags {
//	    fmt.Fprintln(os.Stderr, d.Render(sm))
//	}
package diagnostics

import (
	"fmt"
	"strings"
)

// Severity classifies a diagnostic.
type Severity int

const (
	SeverityError   Severity = iota // hard error; compilation stops
	SeverityWarning                 // non-fatal; emitted before output
	SeverityNote                    // informational annotation
)

// Diag is a compiler diagnostic with source position, message, and optional hint.
type Diag struct {
	Severity Severity
	File     string
	Line     int    // 1-based
	Col      int    // 1-based
	Msg      string // primary message
	Hint     string // optional suggestion ("did you mean X?", "add `mut` here")
}

// Error satisfies the error interface (for error-type diagnostics).
func (d *Diag) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s", d.File, d.Line, d.Col, d.Msg)
}

// SourceMap maps file names to their source lines (1-indexed).
type SourceMap map[string][]string

// NewSourceMap builds a SourceMap from a filename → full-source-text map.
func NewSourceMap(files map[string]string) SourceMap {
	sm := make(SourceMap, len(files))
	for name, src := range files {
		sm[name] = strings.Split(src, "\n")
	}
	return sm
}

// Add inserts or replaces a single file's source text in the map.
func (sm SourceMap) Add(name, src string) {
	sm[name] = strings.Split(src, "\n")
}

// Render formats d into a human-readable string.
// If sm is non-nil and the source line is available, a code snippet with
// a caret is appended.  If d.Hint is non-empty, a hint line is appended.
//
// Example output:
//
//	main.cnd:12:5: error: undefined identifier "foo"
//	   12 |     let x = foo()
//	      |             ^
//	    hint: did you mean "fov"?
func (d *Diag) Render(sm SourceMap) string {
	sev := severityName(d.Severity)
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s:%d:%d: %s: %s", d.File, d.Line, d.Col, sev, d.Msg)

	if sm != nil {
		if lines, ok := sm[d.File]; ok && d.Line >= 1 && d.Line <= len(lines) {
			srcLine := lines[d.Line-1]
			lineStr := fmt.Sprintf("%d", d.Line)
			// right-align line number in a 4-wide field
			pad := 4 - len(lineStr)
			if pad < 0 {
				pad = 0
			}
			gutter := strings.Repeat(" ", pad) + lineStr + " | "
			blank := strings.Repeat(" ", pad+len(lineStr)) + " | "

			sb.WriteString("\n")
			sb.WriteString(gutter)
			sb.WriteString(srcLine)

			// caret under the token start
			col := d.Col - 1 // 0-based
			if col < 0 {
				col = 0
			}
			if col > len(srcLine) {
				col = len(srcLine)
			}
			sb.WriteString("\n")
			sb.WriteString(blank)
			sb.WriteString(strings.Repeat(" ", col))
			sb.WriteString("^")
		}
	}

	if d.Hint != "" {
		sb.WriteString("\n    hint: ")
		sb.WriteString(d.Hint)
	}

	return sb.String()
}

// RenderAll renders a slice of Diags joined by newlines.
func RenderAll(diags []Diag, sm SourceMap) string {
	parts := make([]string, len(diags))
	for i, d := range diags {
		parts[i] = d.Render(sm)
	}
	return strings.Join(parts, "\n")
}

// CountErrors returns the number of error-severity diagnostics in the slice.
func CountErrors(diags []Diag) int {
	n := 0
	for _, d := range diags {
		if d.Severity == SeverityError {
			n++
		}
	}
	return n
}

func severityName(s Severity) string {
	switch s {
	case SeverityError:
		return "error"
	case SeverityWarning:
		return "warning"
	case SeverityNote:
		return "note"
	default:
		return "diagnostic"
	}
}
