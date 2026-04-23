# Token Density: Measured Data

*Methodology and results — April 2026*

---

## Methodology

All measurements use the Anthropic `count_tokens` API against `claude-sonnet-4-6`.
This is not an approximation. These are the actual token counts Claude uses when
processing Candor code.

- Baseline overhead (8 tokens) subtracted from all measurements
- Each construct measured in isolation
- Results reproducible: see `benchmarks/tokenizer/token_analysis.py`
- Raw results: `benchmarks/tokenizer/results/2026-04-23_claude-sonnet-4-6_3.json`

The tool re-runs against any model. Measurements will shift as models are updated —
that is a feature, not a bug. Candor's design is monitored against real tokenizers,
not assumed to be stable.

---

## Core Keyword Alignment

All 37 core Candor keywords measured. Target: 1 token each.

**Result: 36/37 pass.** One alert.

| Status | Keywords |
|---|---|
| ✓ 1 token | `fn` `let` `if` `else` `return` `pure` `match` `must` `struct` `enum` `trait` `impl` `type` `use` `extern` `effects` `requires` `ensures` `loop` `while` `for` `in` `break` `continue` `true` `false` `ok` `err` `none` `some` `ref` `box` `arc` `bool` `str` `unit` |
| ⚠ 3 tokens | `refmut` |

**`refmut` fix:** Agent Form alias `mut<T>` = 4 tokens vs `refmut<T>` = 6 tokens.
33% savings. The 3-token keyword is a known design constraint; the alias compensates.

**Numeric types:** All `i8`–`i128`, `u8`–`u128`, `f16`–`f64` = **2 tokens** each
(letter + number always splits in BPE). This is intrinsic to how tokenizers work,
not a Candor design flaw. Agent Form aliases `int`, `float`, `byte` = 1 token each.

---

## `?` Propagation Operator

The `?` operator replaces the full match-based error propagation pattern.

**Operator cost:** `?` = **1 token**.

**Boilerplate it replaces** (Verification Form):
```
match expr { ok(v) => v   err(e) => return err(e) }
```
= **24 tokens** for a simple expression.

| Scenario | Verification Form | Agent Form | Savings |
|---|---|---|---|
| Single propagation site | 24 tok | 4 tok | **83%** |
| Named-type result (`load_config`) | 26 tok | 6 tok | **77%** |
| 3 sites — realistic IO function body | 72 tok | 14 tok | **81%** |
| 5 sites — complex pipeline function | 120 tok | 24 tok | **80%** |

At 5 propagation sites: **96 tokens eliminated** that carry zero semantic information —
pure routing boilerplate with no signal for the AI.

---

## Return Type + Effect Annotation

The most common signature suffixes in Candor programs:

| Pattern | Verification Form | Agent Form | Savings |
|---|---|---|---|
| Single IO effect | `effects(io)` (4 tok) | `io` (1 tok) | **75%** |
| Result type | `result<Row, str>` (6 tok) | `?Row` (2 tok) | **67%** |
| Effectful result sig | `-> result<str, str> effects(io)` (11 tok) | `-> ?str io` (4 tok) | **64%** |
| Pure result sig | `-> result<Row, str> pure` (8 tok) | `-> ?Row pure` (4 tok) | **50%** |
| Unit effectful | `-> unit effects(io)` (6 tok) | `-> unit io` (3 tok) | **50%** |

---

## Complete Function Comparison

The same function in Verification Form vs Agent Form, full body:

**Verification Form — 106 tokens:**
```candor
fn process(path: str) -> result<str, str> effects(io) {
    let f = match open(path) { ok(v) => v   err(e) => return err(e) }
    let s = match read(f)  { ok(v) => v   err(e) => return err(e) }
    let r = match parse(s) { ok(v) => v   err(e) => return err(e) }
    return ok(r)
}
```

**Agent Form — 42 tokens:**
```candor
fn process(path: str) -> ?str io {
    let f = open(path)?
    let s = read(f)?
    let r = parse(s)?
    return ok(r)
}
```

**60% fewer tokens. Same program. Same semantics. Same compiled output.**

---

## `|>` Pipeline Operator — Honest Assessment

`|>` = **2 tokens** (BPE splits `|` and `>` separately).

| Pattern | Nested calls | Pipeline | Difference |
|---|---|---|---|
| 2-step, short names | 6 tok | 7 tok | -1 (costs) |
| 3-step, snake_case | 14 tok | 16 tok | -2 (costs) |
| 5-step, snake_case | 23 tok | 26 tok | -3 (costs) |

**`|>` does not save tokens vs nested calls.** Its value is cognitive: left-to-right
dataflow matches how transformers attend sequentially. Inside-out nesting requires
the model to track depth; pipeline does not. This is an AI reasoning benefit,
not a token count benefit, and should be claimed as such.

---

## Overall Signature Savings

Measured across 41 canonical Verification Form signature tokens:

**Agent Form: 18 tokens. Savings: 56.1%.**

At scale: 100 concurrent requests on a 70B model (327 KB KV cache per token),
56% savings in signature-heavy coding contexts frees significant VRAM per request.

---

## What Is Not Yet Measured

Current data is constructed examples, not a corpus. The next validation step:

1. Measure whole `examples/*.cnd` program files in both forms
2. Report mean ± std across N programs (defensible confidence interval)
3. Cross-language comparison: same programs in Python, Rust, Go

The current numbers are a lower bound — they represent the most common patterns
but do not capture compounding effects across a full program.

---

*Tool: `benchmarks/tokenizer/token_analysis.py`*
*Model: claude-sonnet-4-6 — re-run when models update*
*Data: `benchmarks/tokenizer/results/`*
*Candor source: https://github.com/candor-core/candor*
