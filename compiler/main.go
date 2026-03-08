// Copyright (c) 2026 Scott W. Corley
// SPDX-License-Identifier: Apache-2.0
// https://github.com/scottcorleyg1/candor

package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: candorc <file.cnd>")
		os.Exit(1)
	}
	filename := os.Args[1]
	fmt.Printf("candorc: compiling %s\n", filename)
}
