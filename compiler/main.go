// Copyright (c) 2026 Scott W. Corley
// SPDX-License-Identifier: Apache-2.0
// https://github.com/scottcorleyg1/candor

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	emit_c "github.com/scottcorleyg1/candor/compiler/emit_c"
	"github.com/scottcorleyg1/candor/compiler/lexer"
	"github.com/scottcorleyg1/candor/compiler/parser"
	"github.com/scottcorleyg1/candor/compiler/typeck"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: candorc <file.cnd> [file.cnd ...]")
		os.Exit(1)
	}
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "candorc: %v\n", err)
		os.Exit(1)
	}
}

func run(srcPaths []string) error {
	// Parse all files, merge declarations.
	var allDecls []parser.Decl
	for _, srcPath := range srcPaths {
		src, err := os.ReadFile(srcPath)
		if err != nil {
			return err
		}
		tokens, err := lexer.Tokenize(srcPath, string(src))
		if err != nil {
			return err
		}
		file, err := parser.Parse(srcPath, tokens)
		if err != nil {
			return err
		}
		allDecls = append(allDecls, file.Decls...)
	}

	// Type-check the merged program.
	merged := &parser.File{Name: srcPaths[0], Decls: allDecls}
	res, err := typeck.Check(merged)
	if err != nil {
		return err
	}

	// Emit a single C file.
	cSrc, err := emit_c.Emit(merged, res)
	if err != nil {
		return err
	}

	// Write C file next to the first source file.
	base := strings.TrimSuffix(srcPaths[0], filepath.Ext(srcPaths[0]))
	cPath := base + ".c"
	if err := os.WriteFile(cPath, []byte(cSrc), 0o644); err != nil {
		return err
	}

	// Compile with system C compiler.
	outPath := base
	if isWindows() {
		outPath += ".exe"
	}
	cc := findCC()
	cmd := exec.Command(cc, "-o", outPath, cPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("C compiler %q failed: %w", cc, err)
	}
	fmt.Printf("candorc: wrote %s\n", outPath)
	return nil
}

// findCC returns the C compiler to use, preferring $CC, then gcc, then cc.
func findCC() string {
	if cc := os.Getenv("CC"); cc != "" {
		return cc
	}
	if path, err := exec.LookPath("gcc"); err == nil {
		return path
	}
	return "cc"
}

func isWindows() bool {
	return os.PathSeparator == '\\'
}
