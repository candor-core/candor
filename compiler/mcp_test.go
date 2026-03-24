// Copyright (c) 2026 Scott W. Corley
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ── candorTypeToJsonSchema unit tests ──────────────────────────────────────────

func TestCandorTypeToJsonSchema_Primitives(t *testing.T) {
	cases := []struct{ in, want string }{
		{"str", "string"},
		{"bool", "boolean"},
		{"f64", "number"},
		{"f32", "number"},
		{"f16", "number"},
		{"bf16", "number"},
		{"i8", "integer"},
		{"i16", "integer"},
		{"i32", "integer"},
		{"i64", "integer"},
		{"u8", "integer"},
		{"u16", "integer"},
		{"u32", "integer"},
		{"u64", "integer"},
	}
	for _, c := range cases {
		got := candorTypeToJsonSchema(c.in)
		if got != c.want {
			t.Errorf("candorTypeToJsonSchema(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestCandorTypeToJsonSchema_StructDefault(t *testing.T) {
	// Unknown / struct types should map to "object", not "integer".
	for _, t2 := range []string{"Point", "MyStruct", "Token", "unit"} {
		got := candorTypeToJsonSchema(t2)
		if got != "object" {
			t.Errorf("candorTypeToJsonSchema(%q) = %q, want %q", t2, got, "object")
		}
	}
}

// ── cmdMcp integration test ────────────────────────────────────────────────────

func TestCmdMcpJsonSchema(t *testing.T) {
	// Write a minimal .cnd file with one #mcp_tool function.
	src := `
#mcp_tool "Search documents by query and limit"
fn search(query: str, limit: i64, tags: vec<str>, maybe_score: option<f64>) -> vec<str> {
    return vec_new()
}
`
	dir := t.TempDir()
	cndPath := filepath.Join(dir, "test.cnd")
	if err := os.WriteFile(cndPath, []byte(src), 0o644); err != nil {
		t.Fatalf("write cnd: %v", err)
	}

	// cmdMcp writes tools.json to the current directory; change to a temp dir.
	origWD, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer os.Chdir(origWD)

	if err := cmdMcp([]string{cndPath}); err != nil {
		t.Fatalf("cmdMcp: %v", err)
	}

	raw, err := os.ReadFile("tools.json")
	if err != nil {
		t.Fatalf("read tools.json: %v", err)
	}

	var manifest struct {
		Tools []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			InputSchema struct {
				Type       string `json:"type"`
				Properties map[string]struct {
					Type     string `json:"type"`
					Nullable bool   `json:"nullable"`
					Items    *struct {
						Type string `json:"type"`
					} `json:"items"`
				} `json:"properties"`
				Required []string `json:"required"`
			} `json:"inputSchema"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatalf("unmarshal tools.json: %v\nraw: %s", err, raw)
	}

	if len(manifest.Tools) != 1 {
		t.Fatalf("expected 1 tool, got %d\nraw: %s", len(manifest.Tools), raw)
	}
	tool := manifest.Tools[0]

	if tool.Name != "search" {
		t.Errorf("tool name: got %q, want %q", tool.Name, "search")
	}
	if !strings.Contains(tool.Description, "Search documents") {
		t.Errorf("description: got %q", tool.Description)
	}
	if tool.InputSchema.Type != "object" {
		t.Errorf("inputSchema.type: got %q, want %q", tool.InputSchema.Type, "object")
	}

	// query → string
	if p := tool.InputSchema.Properties["query"]; p.Type != "string" {
		t.Errorf("query.type: got %q, want string", p.Type)
	}
	// limit → integer
	if p := tool.InputSchema.Properties["limit"]; p.Type != "integer" {
		t.Errorf("limit.type: got %q, want integer", p.Type)
	}
	// tags → array with items.type = string
	if p := tool.InputSchema.Properties["tags"]; p.Type != "array" {
		t.Errorf("tags.type: got %q, want array", p.Type)
	} else if p.Items == nil || p.Items.Type != "string" {
		t.Errorf("tags.items.type: got %v, want string", p.Items)
	}
	// maybe_score → number + nullable
	if p := tool.InputSchema.Properties["maybe_score"]; p.Type != "number" {
		t.Errorf("maybe_score.type: got %q, want number", p.Type)
	} else if !p.Nullable {
		t.Errorf("maybe_score.nullable: got false, want true")
	}

	// All params should be required.
	reqSet := make(map[string]bool)
	for _, r := range tool.InputSchema.Required {
		reqSet[r] = true
	}
	for _, param := range []string{"query", "limit", "tags", "maybe_score"} {
		if !reqSet[param] {
			t.Errorf("expected %q in required list", param)
		}
	}
}
