// Copyright (c) 2026 Scott W. Corley
// SPDX-License-Identifier: Apache-2.0

package manifest_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/scottcorleyg1/candor/compiler/manifest"
)

func writeManifest(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "Candor.toml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadBasic(t *testing.T) {
	path := writeManifest(t, `
[package]
name    = "hello"
version = "0.1.0"
entry   = "src/main.cnd"
`)
	m, err := manifest.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if m.Name != "hello" {
		t.Errorf("Name = %q, want %q", m.Name, "hello")
	}
	if m.Version != "0.1.0" {
		t.Errorf("Version = %q, want %q", m.Version, "0.1.0")
	}
	if m.Entry != "src/main.cnd" {
		t.Errorf("Entry = %q, want %q", m.Entry, "src/main.cnd")
	}
}

func TestLoadBuildSection(t *testing.T) {
	path := writeManifest(t, `
[package]
name = "myapp"

[build]
output  = "bin/myapp"
sources = ["src/main.cnd", "src/lib.cnd"]
`)
	m, err := manifest.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if m.Output != "bin/myapp" {
		t.Errorf("Output = %q", m.Output)
	}
	if len(m.Sources) != 2 {
		t.Fatalf("Sources len = %d, want 2", len(m.Sources))
	}
	if m.Sources[0] != "src/main.cnd" {
		t.Errorf("Sources[0] = %q", m.Sources[0])
	}
}

func TestLoadMissingName(t *testing.T) {
	path := writeManifest(t, `
[package]
version = "0.1.0"
`)
	_, err := manifest.Load(path)
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestLoadWithComments(t *testing.T) {
	path := writeManifest(t, `
## Project manifest
[package]
name    = "proj"  # inline comment
version = "1.0.0"
`)
	m, err := manifest.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if m.Name != "proj" {
		t.Errorf("Name = %q, want %q", m.Name, "proj")
	}
}

func TestOutputPath(t *testing.T) {
	path := writeManifest(t, `[package]
name = "app"
`)
	m, err := manifest.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	out := m.OutputPath(false)
	if filepath.Base(out) != "app" {
		t.Errorf("OutputPath = %q, want basename 'app'", out)
	}
}

func TestFindManifest(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub", "deep")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	tomlPath := filepath.Join(dir, "Candor.toml")
	if err := os.WriteFile(tomlPath, []byte("[package]\nname=\"x\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	found, err := manifest.FindManifest(sub)
	if err != nil {
		t.Fatal(err)
	}
	if found != tomlPath {
		t.Errorf("FindManifest = %q, want %q", found, tomlPath)
	}
}
