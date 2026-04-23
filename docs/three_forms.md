# The Three Forms of Candor

*Architecture specification — April 2026*

---

## The Brief

Every Candor program exists in three lossless projections.

**Agent Form** is what AI writes — dense, BPE-aligned, 1-token keywords, shorthands for
common patterns. **Verification Form** is what the compiler checks — full syntax,
human-auditable, canonical. **Machine Form** is LLVM IR — where purity is not a
comment but a `memory(none)` attribute that LLVM's own verifier will reject if violated.

Same program. Three projections. Every transformation is deterministic and reversible.

This is the first programming language designed to be auditable at all three levels
simultaneously: token efficiency for AI, readability for humans, machine-verifiable
guarantees for hardware.

---

## The Three Forms

### Agent Form (AF) — What AI writes

Optimized for token efficiency across AI code generation. Every design choice is
measured against real BPE tokenization (Claude `count_tokens` API, not approximation).

**Properties:**
- All core keywords are 1 BPE token — verified against `claude-sonnet-4-6`
- Result type shorthand: `?T` expands to `result<T, str>`
- Effect shorthand: bare `io`, `net`, `gpu` expand to `effects(io)` etc.
- `?` propagation: postfix operator, 1 token, replaces 15–25 token match boilerplate
- `|>` pipeline: left-to-right function composition
- `mut<T>` shorthand for `refmut<T>` (3-token keyword reduced to 1)

**Example:**
```candor
fn process(path: str) -> ?str io {
    let f = open(path)?
    let s = read(f)?
    let r = parse(s)?
    return ok(r)
}
```
*42 tokens.*

### Verification Form (VF) — What the compiler checks

The canonical form. This is what gets stored in source control, reviewed by humans,
and checked by the compiler. Full syntax, no shorthands, every annotation explicit.

**Properties:**
- `result<T, str>` everywhere — error type visible
- `effects(io)` annotation — effect named explicitly
- `match expr { ok(v) => v   err(e) => return err(e) }` — propagation spelled out
- Compiler enforces EFFECTS-001: pure callers may not call effectful functions
- Compiler enforces AXIOM-003: result and option values cannot be silently discarded

**Example:**
```candor
fn process(path: str) -> result<str, str> effects(io) {
    let f = match open(path) { ok(v) => v   err(e) => return err(e) }
    let s = match read(f) { ok(v) => v   err(e) => return err(e) }
    let r = match parse(s) { ok(v) => v   err(e) => return err(e) }
    return ok(r)
}
```
*106 tokens.*

**Transformation from AF to VF is mechanical, one-pass, deterministic:**

| Agent Form | Verification Form |
|---|---|
| `-> ?T` | `-> result<T, str>` |
| `io` (effect) | `effects(io)` |
| `expr?` | `match expr { ok(v) => v   err(e) => return err(e) }` |
| `x \|> f` | `f(x)` |
| `mut<T>` | `refmut<T>` |

No inference. No context-dependence. Every rule is a substitution.

### Machine Form (MF) — What runs on hardware

LLVM IR produced by the Candor compiler. This is where annotations become
machine-verifiable constraints — not comments, not documentation, not trust.

**Properties:**
- `pure` functions emit `memory(none) nounwind` — LLVM's verifier rejects any
  `memory(none)` function that contains a `load` or `store` instruction
- Effect annotations determine optimizer permissions — pure functions can be
  freely reordered, hoisted, memoized, and eliminated by LLVM passes
- `result<T, E>` maps to a typed aggregate: `{ i1, T, E }` — error paths visible in IR

**Example:**
```llvm
define ptr @process(ptr %path.in) {          ; effectful — no memory attr
entry:
  ...

define i64 @add(i64 %a.in, i64 %b.in) memory(none) nounwind {   ; pure — verified
entry:
  ...
```

**The transparency chain:**

```
Candor source:  fn add(a: i64, b: i64) -> i64 pure { return a + b }
     |
     v EFFECTS-001 (typeck rejects pure callers of effectful code)
     |
     v emit_llvm
     |
LLVM IR:  define i64 @add(i64 %a.in, i64 %b.in) memory(none) nounwind {
     |
     v LLVM verifier
     |
Hardware: guaranteed — no memory side effects
```

No step in this chain requires trust. Each layer is independently auditable.

---

## Measured Token Data

All measurements from Claude `count_tokens` API against `claude-sonnet-4-6`,
2026-04-23. Baseline overhead subtracted. Results in
`benchmarks/tokenizer/results/`.

### Core keyword alignment

36/37 core keywords = **1 BPE token**. One exception: `refmut` = 3 tokens.
Agent Form alias `mut<T>` reduces `refmut<T>` (6 tokens) to `mut<T>` (4 tokens).

### `?` propagation — per site savings

| Pattern | VF tokens | AF tokens | Savings |
|---|---|---|---|
| Single propagation site | 24 | 4 | **83%** |
| 3 sites (realistic IO function) | 72 | 14 | **81%** |
| 5 sites (complex pipeline) | 120 | 24 | **80%** |
| Complete function (sig + 3-site body) | 106 | 42 | **60%** |

### Return type + effect annotation savings

| Pattern | VF tokens | AF tokens | Savings |
|---|---|---|---|
| `-> result<str, str> effects(io)` | 11 | 4 | **64%** |
| `-> result<Row, str> pure` | 8 | 4 | **50%** |
| `effects(io)` alone | 4 | 1 | **75%** |
| `result<Row, str>` alone | 6 | 2 | **67%** |

### Overall

**56.1% fewer tokens in function signatures**, measured across canonical patterns.
At 100 concurrent requests on a 70B model, each token saved frees ~327 KB of VRAM.

---

## Why This Matters

The cost function for programming language design has a new variable: **token cost**.

Before AI as a force multiplier, languages optimized for human writing time, human
reading time, and machine execution. Token cost was zero.

At AI scale — agentic coding, multi-agent pipelines, continuous AI-assisted
development — tokens represent real compute, real energy, real infrastructure cost.
A language that reduces tokens by 56–83% in the most common patterns does not just
improve developer experience. It changes the economics of AI-assisted software.

Candor is the first language designed to optimize this cost function deliberately,
measurably, and provably — at the token level, the type level, and the IR level.

---

## Status (April 2026)

| Layer | Status |
|---|---|
| Agent Form → Verification Form transformation rules | Defined |
| AF keyword token alignment | Measured and verified |
| `?` and `\|>` operators | Implemented (M14) |
| VF → MF: pure → `memory(none) nounwind` | Implemented and tested |
| VF → MF: `result<T,E>` as named IR struct | Planned (M14.5) |
| VF → MF: `?` as `extractvalue + br + ret` | Planned (M14.5) |
| VF → MF: contracts as `llvm.assume` | Planned (M15) |

---

*Candor is developed by Scott W. Corley. Apache-2.0.*
*Source: https://github.com/candor-core/candor*
