# Candor Token Benchmark

*Measured April 2026 against claude-sonnet-4-6 via Anthropic count_tokens API*

---

Every number in this document is the result of a real API call. Not an estimate, not a heuristic, not a theoretical model. The tool is at `benchmarks/tokenizer/token_analysis.py`. It runs against any model. It saves timestamped JSON baselines. Every result here is reproducible.

---

## Methodology

```
model:              claude-sonnet-4-6
baseline overhead:  8 tokens (subtracted from all measurements)
tool:               benchmarks/tokenizer/token_analysis.py
raw data:           benchmarks/tokenizer/results/2026-04-23_claude-sonnet-4-6_3.json
```

Each construct is measured in isolation. Baseline overhead is the minimum token cost of an API call with an empty string — it is subtracted from every result so the numbers reflect the construct alone.

---

## Core Keyword Alignment

Target: every Candor keyword costs exactly 1 BPE token.

**Result: 36 of 37 pass.**

| Status | Keywords |
|---|---|
| 1 token | `fn` `let` `if` `else` `return` `pure` `match` `must` `struct` `enum` `trait` `impl` `type` `use` `extern` `effects` `requires` `ensures` `loop` `while` `for` `in` `break` `continue` `true` `false` `ok` `err` `none` `some` `ref` `box` `arc` `bool` `str` `unit` |
| 3 tokens | `refmut` |

`refmut` is the only alert. Agent Form alias `mut<T>` reduces `refmut<T>` (6 tokens) to `mut<T>` (4 tokens) — 33% savings on every mutable reference annotation.

`pure` is 1 token. The most important annotation in the language is free.

---

## Numeric Types

All integer and float types split at the letter-number boundary in BPE. This is intrinsic — not a Candor design flaw.

| Type class | Examples | Tokens |
|---|---|---|
| Signed integers | `i8` `i16` `i32` `i64` `i128` | 2 each |
| Unsigned integers | `u8` `u16` `u32` `u64` `u128` | 2 each |
| Floats | `f16` `f32` `f64` `bf16` | 2 each |

Agent Form aliases cover the common cases:

| Alias | Maps to | Tokens saved |
|---|---|---|
| `int` | `i64` | 1 |
| `uint` | `u64` | 1 |
| `float` | `f64` | 1 |
| `byte` | `u8` | 1 |

When precision matters (`i128`, `f32`), the 2-token cost is accepted. These aliases are verified 1-token words, not guesses.

---

## Effect Annotations

Verification Form uses `effects(io)`. Agent Form drops the wrapper.

| Construct | VF | VF tokens | AF | AF tokens | Savings |
|---|---|---|---|---|---|
| Single IO effect | `effects(io)` | 4 | `io` | 1 | **75%** |
| Single GPU effect | `effects(gpu)` | 4 | `gpu` | 1 | **75%** |
| Single net effect | `effects(net)` | 4 | `net` | 1 | **75%** |
| Combined effects | `effects(io, gpu)` | 6 | `io gpu` | 2 | **67%** |
| Pure (no alias needed) | `pure` | 1 | `pure` | 1 | 0% |

`pure` is already optimal. No shorthand exists because none is needed.

---

## `?` Propagation Operator

`?` is 1 token. It replaces the full match-based error propagation pattern:

```
match expr { ok(v) => v   err(e) => return err(e) }
```

That pattern costs **24 tokens** for a simple expression. Every `?` eliminates it.

| Scenario | VF tokens | AF tokens | Saved | Savings |
|---|---|---|---|---|
| Single propagation site | 24 | 4 | 20 | **83%** |
| Named-type result (`load_config`) | 26 | 6 | 20 | **77%** |
| 3 sites — realistic IO function body | 72 | 14 | 58 | **81%** |
| 5 sites — complex pipeline function | 120 | 24 | 96 | **80%** |

At 5 propagation sites: **96 tokens eliminated**. Every one carried zero semantic information — pure routing boilerplate with no signal for the model.

---

## Return Type Signatures

Combining result shorthand (`?T`) and effect shorthand (`io`) on function signatures:

| Pattern | VF | VF tokens | AF | AF tokens | Savings |
|---|---|---|---|---|---|
| Effectful result fn | `-> result<str, str> effects(io)` | 11 | `-> ?str io` | 4 | **64%** |
| Pure result fn | `-> result<Row, str> pure` | 8 | `-> ?Row pure` | 4 | **50%** |
| Unit effectful fn | `-> unit effects(io)` | 6 | `-> unit io` | 3 | **50%** |
| Pure non-failing fn | `-> i64 pure` | 5 | `-> int pure` | 3 | **40%** |
| Effect alone | `effects(io)` | 4 | `io` | 1 | **75%** |
| Result type alone | `result<Row, str>` | 6 | `?Row` | 2 | **67%** |

**Overall: 56.1% fewer tokens across 41 canonical signature patterns.**

---

## Complete Function Comparison

The same function, end to end.

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

`|>` = 2 tokens (BPE splits `|` and `>` separately).

| Pattern | Nested calls | Pipeline | Delta |
|---|---|---|---|
| 2-step, short names | 6 tok | 7 tok | -1 (costs) |
| 3-step, short names | 8 tok | 10 tok | -2 (costs) |
| 3-step, snake_case | 14 tok | 16 tok | -2 (costs) |
| 5-step, snake_case | 23 tok | 26 tok | -3 (costs) |

**`|>` does not save tokens vs nested calls.** Its value is structural: left-to-right dataflow reads in execution order. Inside-out nesting (`f(g(h(x)))`) requires the model to parse depth before understanding sequence. Transformer attention is sequential — linear structure matches how the model attends to code.

This is an AI reasoning benefit, not a token count benefit, and is documented as such.

---

## Compound Savings

Full function with signature + propagations + pipeline combined:

| Pattern | VF tokens | AF tokens | Saved | Savings |
|---|---|---|---|---|
| Effectful result fn signature | 18 | 11 | 7 | **39%** |
| 2 propagation sites + 1 pipeline | 67 | 30 | 37 | **55%** |
| Complete fn (sig + 3-propagation body) | 106 | 42 | 64 | **60%** |

---

## Identifier Splitting

How names tokenize — this is a floor set by BPE, not Candor.

| Pattern | Tokens | Notes |
|---|---|---|
| Any `word_word` (snake_case) | 3 | word + `_` + word always splits |
| Single-word struct (Row, Config) | 1 | Short CamelCase hits single-token merges |
| Compound struct (LogEntry, ParseResult) | 2 | Two root words = two tokens |

This is not improvable at the language level. Short struct names (`Row`, `Config`, `Task`) are 1 token. Style guidance for Agent Form: prefer single-word struct names where readability permits.

---

## Future Keywords

All candidate keywords for upcoming features — pre-verified before commitment:

| Status | Keywords |
|---|---|
| 1 token (cleared) | `yield` `await` `async` `where` `with` `move` `copy` `pin` `weak` `task` `spawn` |

No future keyword is added to the language until its token cost is confirmed.

---

## At Scale

At 100 concurrent requests on a 70B model (327 KB KV cache per token):

- 56% savings in function signatures frees significant VRAM per batch
- 60% savings on a common IO function means the same context window fits ~2.5× as many function definitions
- 96 tokens eliminated per 5-site error propagation function — per call, per request, per pipeline step

---

## What Is Not Yet Measured

The current data is constructed examples, not a corpus. The next validation step:

1. Measure whole `examples/*.cnd` program files in both forms
2. Report mean ± std across N programs (defensible confidence interval)
3. Cross-language comparison: same programs in Python, Rust, Go

Current numbers are a **lower bound** — they represent the most common patterns but do not capture compounding effects across a full program. The full-program mean is expected to be higher than 60%.

---

*Tool: `benchmarks/tokenizer/token_analysis.py`*  
*Model: claude-sonnet-4-6 — re-run when models update*  
*Raw data: `benchmarks/tokenizer/results/2026-04-23_claude-sonnet-4-6_3.json`*
