---
layout: default
title: Candor
---

# Candor

**A programming language designed for AI code generation.**

60% fewer tokens on a common IO function. Same program. Same semantics. Same compiled output. Not a claim — a measurement.

---

## The Problem

Every token an AI writes costs compute. Every redundant token in a language's syntax is a tax on every AI-assisted edit, every agentic loop, every multi-model pipeline. Programming languages were designed for human writing speed and human reading time. Token cost was zero.

At AI scale, token cost is not zero.

---

## The Design

Candor has three lossless projections of every program:

**Agent Form** — what AI writes. BPE-aligned. Dense. Every core keyword is 1 token, verified against a real tokenizer. Shorthands for the most common patterns.

**Verification Form** — what gets stored in source control and reviewed by humans. Full syntax, every annotation explicit, compiler-checked.

**Machine Form** — LLVM IR. Where `pure` is not a comment but a `memory(none) nounwind` attribute that LLVM's own verifier will reject if violated.

Same program. Three projections. Every transformation is deterministic and reversible.

---

## Measured Numbers

All measurements use the Anthropic `count_tokens` API against `claude-sonnet-4-6`. Baseline overhead subtracted. Results are reproducible.

### Core keywords

36 of 37 core keywords = **1 BPE token**. The one exception (`refmut`) has an Agent Form alias that reduces it by 33%.

### `?` propagation operator

`?` replaces `match expr { ok(v) => v   err(e) => return err(e) }` — 24 tokens of routing boilerplate — with 1 token.

| Scenario | Boilerplate tokens | With `?` | Savings |
|---|---|---|---|
| Single propagation site | 24 | 4 | **83%** |
| 3 sites (typical IO function) | 72 | 14 | **81%** |
| 5 sites (complex pipeline) | 120 | 24 | **80%** |

### Function signatures

| Pattern | Verbose | Candor Agent Form | Savings |
|---|---|---|---|
| Effectful result fn | `-> result<str, str> effects(io)` (11 tok) | `-> ?str io` (4 tok) | **64%** |
| Pure result fn | `-> result<Row, str> pure` (8 tok) | `-> ?Row pure` (4 tok) | **50%** |
| Effect alone | `effects(io)` (4 tok) | `io` (1 tok) | **75%** |

**Overall: 56.1% fewer tokens across canonical signature patterns.**

### Complete function

```candor
// Verification Form — 106 tokens
fn process(path: str) -> result<str, str> effects(io) {
    let f = match open(path) { ok(v) => v   err(e) => return err(e) }
    let s = match read(f)  { ok(v) => v   err(e) => return err(e) }
    let r = match parse(s) { ok(v) => v   err(e) => return err(e) }
    return ok(r)
}

// Agent Form — 42 tokens
fn process(path: str) -> ?str io {
    let f = open(path)?
    let s = read(f)?
    let r = parse(s)?
    return ok(r)
}
```

**60% fewer tokens. Same program. Same semantics. Same compiled output.**

---

## Machine-Verifiable Guarantees

Candor's purity is not a type annotation the compiler trusts. It is a `memory(none) nounwind` attribute in LLVM IR. LLVM's own verifier rejects any `memory(none)` function that contains a load or store instruction. The guarantee is enforced at the IR level, independently of the Candor compiler.

```llvm
; pure function — machine-verifiable
define i64 @add(i64 %a.in, i64 %b.in) memory(none) nounwind {
  ...

; effectful function — bare define, optimizer treats conservatively
define ptr @process(ptr %path.in) {
  ...
```

No other active language encodes purity this way in LLVM IR.

---

## Current Status

| Feature | Status |
|---|---|
| Self-hosting bootstrap (stage4 == stage2) | Done |
| EFFECTS-001: pure callers cannot call effectful code | Enforced in type checker |
| AXIOM-003: result/option values cannot be silently discarded | Enforced |
| `?` propagation operator | Implemented |
| `\|>` pipeline operator | Implemented |
| Agent Form → Verification Form transformation | Defined |
| pure → `memory(none) nounwind` in LLVM IR | Implemented and tested |
| `result<T,E>` as named LLVM struct | Planned |
| `?` as `extractvalue + br + ret` in IR | Planned |

---

## Benchmark Data

- Full benchmark: [token\_benchmark.md](token_benchmark.md)
- Three Forms spec: [three\_forms.md](three_forms.md)
- Measurement tool: `benchmarks/tokenizer/token_analysis.py`
- Raw results: `benchmarks/tokenizer/results/`

---

## Source

[github.com/candor-core/candor](https://github.com/candor-core/candor)

Apache-2.0. Developed by Scott W. Corley.
