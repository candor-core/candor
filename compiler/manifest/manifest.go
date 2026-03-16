// Copyright (c) 2026 Scott W. Corley
// SPDX-License-Identifier: Apache-2.0
// https://github.com/scottcorleyg/candor

// Package manifest parses Candor.toml project manifest files.
//
// Format:
//
//	[package]
//	name    = "myapp"
//	version = "0.1.0"
//	entry   = "src/main.cnd"
//
//	[build]
//	sources = ["src/lib.cnd", "src/util.cnd"]   # optional; auto-discovered if absent
//	output  = "bin/myapp"                          # optional; defaults to <name>
package manifest

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Manifest holds the parsed contents of a Candor.toml file.
type Manifest struct {
	// Package section
	Name    string // package name
	Version string // semver string, e.g. "0.1.0"
	Entry   string // path to main .cnd file, relative to manifest dir

	// Build section
	Sources []string // explicit source list; empty = auto-discover all *.cnd in src/
	Output  string   // output binary path; empty = <name> or <name>.exe

	// Dir is the directory containing Candor.toml (set by Load).
	Dir string
}

// Load reads and parses the Candor.toml at path.
func Load(path string) (*Manifest, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	m := &Manifest{Dir: filepath.Dir(path)}
	var section string
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		// Skip blank lines and comments.
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Section header: [package] or [build]
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.ToLower(line[1 : len(line)-1])
			continue
		}
		// Key = "value" or key = ["a", "b"]
		eqIdx := strings.IndexByte(line, '=')
		if eqIdx < 0 {
			return nil, fmt.Errorf("Candor.toml:%d: expected key = value", lineNum)
		}
		key := strings.TrimSpace(line[:eqIdx])
		val := strings.TrimSpace(line[eqIdx+1:])
		// Strip inline comment after value (# ...).
		if ci := strings.Index(val, " #"); ci >= 0 {
			val = strings.TrimSpace(val[:ci])
		}

		switch section {
		case "package":
			switch key {
			case "name":
				m.Name = unquote(val)
			case "version":
				m.Version = unquote(val)
			case "entry":
				m.Entry = unquote(val)
			}
		case "build":
			switch key {
			case "output":
				m.Output = unquote(val)
			case "sources":
				m.Sources = parseStringArray(val)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if m.Name == "" {
		return nil, fmt.Errorf("Candor.toml: [package] name is required")
	}
	return m, nil
}

// FindManifest walks up from dir until it finds a Candor.toml or reaches the
// filesystem root. Returns ("", nil) if no manifest is found.
func FindManifest(dir string) (string, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(dir, "Candor.toml")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", nil // reached root
		}
		dir = parent
	}
}

// SourceFiles returns the .cnd source files to compile.
// If Sources is explicit, those paths (resolved relative to Dir) are returned.
// Otherwise, all *.cnd files under Dir/src/ are returned.
// The entry file is always included exactly once, first if specified.
func (m *Manifest) SourceFiles() ([]string, error) {
	if len(m.Sources) > 0 {
		abs := make([]string, len(m.Sources))
		for i, s := range m.Sources {
			abs[i] = filepath.Join(m.Dir, s)
		}
		return abs, nil
	}

	// Auto-discover: walk src/ (or Dir if no src/ exists).
	srcDir := filepath.Join(m.Dir, "src")
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		srcDir = m.Dir
	}

	var files []string
	err := filepath.WalkDir(srcDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".cnd") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// If entry is specified, move it to front.
	if m.Entry != "" {
		entryAbs := filepath.Join(m.Dir, m.Entry)
		reordered := []string{entryAbs}
		for _, f := range files {
			abs, _ := filepath.Abs(f)
			if abs != entryAbs {
				reordered = append(reordered, f)
			}
		}
		return reordered, nil
	}
	return files, nil
}

// OutputPath returns the output binary path.
func (m *Manifest) OutputPath(windows bool) string {
	if m.Output != "" {
		p := filepath.Join(m.Dir, m.Output)
		if windows && !strings.HasSuffix(p, ".exe") {
			p += ".exe"
		}
		return p
	}
	p := filepath.Join(m.Dir, m.Name)
	if windows {
		p += ".exe"
	}
	return p
}

// unquote strips surrounding double-quotes if present.
func unquote(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

// parseStringArray parses a TOML-style inline array: ["a", "b", "c"]
func parseStringArray(s string) []string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "[") || !strings.HasSuffix(s, "]") {
		return []string{unquote(s)}
	}
	inner := s[1 : len(s)-1]
	parts := strings.Split(inner, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		result = append(result, unquote(p))
	}
	return result
}
