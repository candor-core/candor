# LAYER 7: EFFECTS
**Status:** STABLE
**Depends on:** L6-FUNCTIONS.md, L0-AXIOMS.md

---

## PURPOSE

The effects system enforces AXIOM-002: every observable side effect of a function
must be declared in its annotation. The compiler verifies that effects used inside
a function body are a subset of the effects declared on that function.

Effects are purely additive annotations — they constrain what a function may do,
not how it does it. The effects system operates entirely at compile time.

---

### RULE EFF-001: Effect System Overview

**Status:** STABLE
**Layer:** EFF
**Depends on:** AXIOM-002

Every function falls into one of three effect modes:

1. **Unannotated (trusted):** No effects annotation present. The compiler performs
   no effects checking on calls made from this function. This mode exists for
   gradual adoption — legacy or trusted code that has not yet been annotated.

2. **`pure`:** The function declares it has no side effects. May only call other
   `pure` functions or functions that are themselves unannotated (trusted). Any
   call to a function with a declared effects annotation is a compile error.

3. **`effects(e1, e2, ...)`:** The function declares a specific set of permitted
   effects. May only call functions whose declared effects are a subset of the
   declared set. Unannotated callees are trusted and pass without restriction.

**Invariant:**
For every function F with annotation A and every function G called from F:
- If A is unannotated: no check performed.
- If A is `pure`: G must be `pure` or unannotated.
- If A is `effects(S)`: declared_effects(G) ⊆ S, or G is unannotated.

**Compliance Test:**
PASS:
```candor
fn add(a: i64, b: i64) -> i64 pure { return a + b }
```
FAIL:
```candor
fn bad() -> unit pure { print("hello") }
```

**On Violation:**
`pure function cannot call "X" which has effects [Y]`

---

### RULE EFF-002: Known Effect Names

**Status:** STABLE
**Layer:** EFF
**Depends on:** EFF-001

The following effect names are recognized by the compiler. Using an unrecognized
effect name in an `effects(...)` annotation produces a compile-time **warning**
(not an error), guarding against typos while allowing forward compatibility.

| Effect name | Meaning |
|-------------|---------|
| `io`        | File I/O, stdin, stdout, stderr; directory operations |
| `sys`       | OS-level calls: process arguments, environment variables, exit |
| `time`      | Wall clock, monotonic clock, sleep |
| `rand`      | Non-deterministic random number generation |
| `async`     | Suspendable / coroutine-style execution |
| `gpu`       | CUDA/VRAM access; GPU compute workers |
| `net`       | Network transfers (NIXL, InfiniBand, RoCE, NVLink) |
| `storage`   | SSD / object store access (S3, VAST); memory-mapped files |
| `mem`       | CPU RAM management; KV block manager, eviction logic |
| `simd`      | SIMD width-dependent operations (vec_dot, vec_l2, tensor_matmul) |

**Invariant:**
`used_effect ∈ KnownEffects ∨ compiler_emits_warning`

**Compliance Test:**
PASS: `fn f() -> unit effects(io) { print("ok") }` — `io` is known
PASS (with warning): `fn f() -> unit effects(zork) { }` — `zork` is unknown, warning emitted
FAIL: no outright rejection for unknown effects; a warning suffices

**On Violation:**
`unknown effect "X"; known effects: io, sys, time, rand, async, gpu, net, storage, mem, simd`
(warning, not error)

---

### RULE EFF-005: `pure` Annotation

**Status:** STABLE
**Layer:** EFF
**Depends on:** EFF-001

The keyword `pure` placed after the return type declares that a function has no
observable side effects.

A `pure` function:
- May not call any function that has a declared effects annotation (`effects(...)`)
- May call functions that are themselves `pure`
- May call unannotated (trusted) functions without restriction
- May not call any builtin that carries an effects annotation (e.g., `print`, `read_file`)

`pure` is syntactic sugar for `effects []` (see EFF-010). Both forms produce
the same internal representation (`EffectsPure`).

**Invariant:**
`pure_fn calls G → (G is pure) ∨ (G is unannotated)`

**Compliance Test:**
PASS:
```candor
fn double(x: i64) -> i64 pure { return x * 2 }
fn quad(x: i64) -> i64 pure { return double(double(x)) }
```
FAIL:
```candor
fn bad() -> unit pure { print("hello") }
```

**On Violation:**
`pure function cannot call "X" which has effects [Y]`

---

### RULE EFF-010: Effects Declaration Syntax

**Status:** STABLE
**Layer:** EFF
**Depends on:** EFF-001

An effects annotation appears after the return type and before any contract
clauses (`requires`/`ensures`). The annotation is optional; its absence means
the function is unannotated (trusted).

There are three syntactic forms for effects annotations:

1. `pure` — keyword; declares no effects (EffectsPure).
2. `effects []` — bracket form; syntactic sugar for `pure` (EffectsPure). The
   empty bracket form `effects []` is the canonical way to write "zero effects"
   using the `effects` keyword.
3. `effects(e1, e2, ...)` — declares one or more named effects (EffectsDecl).
   At least one effect name is required. **`effects()` with empty parentheses
   is a parse error** — use `effects []` instead.

A fourth form, `cap(X)`, is used for capability-gated functions (see EFF-060).

Order of effect names within `effects(...)` does not matter. Trailing commas
are permitted.

**Grammar:**
```ebnf
EffectsAnnotation ::= 'pure'
                    | 'effects' '[' ']'
                    | 'effects' '(' Ident (',' Ident)* ','? ')'
                    | 'cap' '(' Ident ')'
```

**Invariant:**
- `effects()` with zero names is a parse error.
- `effects []` and `pure` are identical after parsing.
- Effect name order is not significant for checking.

**Compliance Test:**
PASS: `fn f(x: i64) -> i64 effects(io, sys) { ... }`
PASS: `fn f(x: i64) -> i64 pure { ... }`
PASS: `fn f(x: i64) -> i64 effects [] { ... }`
FAIL: `fn f(x: i64) -> i64 effects() { ... }` — empty parens is a parse error

**On Violation:**
`effects() requires at least one effect name; use effects [] for pure`

---

### RULE EFF-020: Effects Are Not Part of the Function Type

**Status:** DRAFT
**Layer:** EFF
**Depends on:** EFF-001

Effects annotations are **not** part of the function type. Two functions with
identical parameter and return types but different effects annotations have the
same function type and are interchangeable in type position.

This is a deliberate simplification. Higher-order effects tracking (where a
function type carries its effects set) is deferred to a future layer.

**Invariant:**
`type_of(fn(A) -> B effects(X)) = type_of(fn(A) -> B effects(Y))` for all X, Y.

**Compliance Test:**
PASS:
```candor
fn apply(f: fn(i64) -> i64, x: i64) -> i64 { return f(x) }
fn inc(x: i64) -> i64 pure { return x + 1 }
let result = apply(inc, 5)
```

**On Violation:**
N/A — this rule currently imposes no compiler rejection. It documents the
absence of effects in function types.

---

### RULE EFF-030: Effects Propagation Check

**Status:** STABLE
**Layer:** EFF
**Depends on:** EFF-001, EFF-005

When a function F with declared effects set S calls a function G with declared
effects set T, the compiler checks that T ⊆ S.

Formally:
- If F is unannotated (`curEffects == nil`): no check — any callee is permitted.
- If F is `pure`: G must be `pure` or unannotated (see EFF-031).
- If F is `effects(S)` and G is unannotated: permitted (G is trusted).
- If F is `effects(S)` and G is `pure`: permitted (pure ⊆ everything).
- If F is `effects(S)` and G is `effects(T)`: each effect in T must appear in S.

The check applies only to direct named calls where the callee name is statically
known. Calls through function-value variables are not effects-checked at this
layer (see EFF-020).

**Invariant:**
`F:effects(S) calls G:effects(T) → T ⊆ S`

**Compliance Test:**
PASS:
```candor
fn logger() -> unit effects(io) { print("log") }
fn main() -> unit effects(io, sys) { logger() }
```
FAIL:
```candor
fn logger() -> unit effects(io) { print("log") }
fn wrong() -> unit effects(sys) { logger() }
```

**On Violation:**
`function with effects([S]) cannot call 'G' which requires effect 'T'`

---

### RULE EFF-031: `pure` Call Restriction

**Status:** STABLE
**Layer:** EFF
**Depends on:** EFF-005, EFF-030

A `pure` function may not call any function that has a declared effects
annotation, regardless of which effects that annotation names.

Permitted callees from a `pure` function:
- Another function annotated `pure` (or `effects []`).
- An unannotated (trusted) function — the compiler treats it as having no
  known effects.

Forbidden callees from a `pure` function:
- Any function annotated `effects(X, ...)` where the `Names` list is non-empty.

**Invariant:**
`pure_fn calls G:effects(T) → T is empty (i.e., G is also pure or unannotated)`

**Compliance Test:**
PASS:
```candor
fn helper(x: i64) -> i64 pure { return x + 1 }
fn compute(x: i64) -> i64 pure { return helper(x) }
```
FAIL:
```candor
fn effectful() -> unit effects(io) { print("x") }
fn bad() -> unit pure { effectful() }
```

**On Violation:**
`pure function cannot call "X" which has effects [Y]`

---

### RULE EFF-040: Builtin Effect Assignments

**Status:** STABLE
**Layer:** EFF
**Depends on:** EFF-001, EFF-002

Built-in functions carry pre-assigned effects annotations. These assignments are
fixed by the compiler and cannot be overridden by user code.

**I/O builtins (`io`):**

| Builtin | Effect |
|---------|--------|
| `print` | `io` |
| `print_int` | `io` |
| `print_bool` | `io` |
| `print_u32` | `io` |
| `print_f64` | `io` |
| `print_char` | `io` |
| `print_err` | `io` |
| `read_line` | `io` |
| `read_int` | `io` |
| `read_f64` | `io` |
| `try_read_line` | `io` |
| `try_read_int` | `io` |
| `try_read_f64` | `io` |
| `read_file` | `io` |
| `write_file` | `io` |
| `append_file` | `io` |
| `read_all_lines` | `io` |
| `read_csv_line` | `io` |
| `flush_stdout` | `io` |
| `path_list_dir` | `io` |
| `path_mkdir` | `io` |
| `path_remove` | `io` |

**OS/system builtins (`sys`):**

| Builtin | Effect |
|---------|--------|
| `os_args` | `sys` |
| `os_getenv` | `sys` |
| `os_exit` | `sys` |
| `os_cwd` | `sys` |

**Time builtins (`time`):**

| Builtin | Effect |
|---------|--------|
| `time_now_ms` | `time` |
| `time_now_mono_ns` | `time` |
| `time_sleep_ms` | `time` |

**Randomness builtins (`rand`):**

| Builtin | Effect |
|---------|--------|
| `rand_u64` | `rand` |
| `rand_f64` | `rand` |
| `rand_range` | `rand` |
| `rand_set_seed` | `rand` |

A user-defined function calling any of the above must declare the corresponding
effect (or be unannotated/trusted). Calling `print()` from a function declared
`effects(sys)` is a compile error because `io` is not in `{sys}`.

**Invariant:**
For every builtin B with assigned effect E:
`F:effects(S) calls B → E ∈ S`

**Compliance Test:**
PASS: `fn log() -> unit effects(io) { print("hello") }`
FAIL: `fn bad() -> unit effects(sys) { print("hello") }`

**On Violation:**
`function with effects([sys]) cannot call 'print' which requires effect 'io'`

---

### RULE EFF-050: Effects Are Not Inferred

**Status:** DRAFT
**Layer:** EFF
**Depends on:** EFF-001

Effects are **not inferred** by the compiler. Every function that wishes to
participate in effects checking must explicitly declare its annotation.

A function body may call effectful functions without a compile error **only if**
the function has no annotation (unannotated / trusted mode). In trusted mode,
the compiler performs no effects checking on calls made within the body.

Automatic effects inference — where the compiler derives the minimum necessary
`effects(...)` set by analysing the call graph — is explicitly out of scope
for STABLE. It may be introduced in a future DRAFT amendment.

**Invariant:**
The compiler does not add, remove, or modify effects annotations. It only
checks annotations that are explicitly present.

**Compliance Test:**
PASS (no annotation, no error):
```candor
fn unchecked() -> unit { print("ok") }
```
PASS (explicit annotation):
```candor
fn checked() -> unit effects(io) { print("ok") }
```
FAIL (annotation present but incorrect):
```candor
fn bad() -> unit effects(sys) { print("ok") }
```

**On Violation:**
`function with effects([sys]) cannot call 'print' which requires effect 'io'`

---

### RULE EFF-060: Capability-Gated Functions (`cap`)

**Status:** DRAFT
**Layer:** EFF
**Depends on:** EFF-001

A function may be annotated `cap(X)` to declare that callers must hold a
capability token of type `cap<X>` in scope. This is a separate mechanism from
`effects(...)` and serves access-control purposes rather than effect classification.

A caller may invoke a `cap(X)` function if and only if one of the following holds:
1. The calling function is itself annotated `cap(X)`, OR
2. A value of type `cap<X>` is visible in the current scope (e.g., passed as a
   parameter).

Capability tokens are not consumed on use. They are structural permissions
threaded through the call chain. The `cap(X)` annotation does not interact with
`effects(...)` checking — a function may have both.

**Grammar:**
```ebnf
CapAnnotation ::= 'cap' '(' Ident ')'
```

**Invariant:**
`F calls G:cap(X) → (F is cap(X)) ∨ (∃ v: cap<X> in scope(F))`

**Compliance Test:**
PASS:
```candor
fn privileged(token: cap<Admin>) -> unit cap(Admin) { ... }
fn caller(tok: cap<Admin>) -> unit { privileged(tok) }
```
FAIL:
```candor
fn privileged(token: cap<Admin>) -> unit cap(Admin) { ... }
fn caller() -> unit { privileged() }
```

**On Violation:**
`call to cap(X) function "F" requires a cap<X> token in scope; pass one as a parameter`

---

### RULE EFF-070: `secret<T>` and Purity Enforcement

**Status:** STABLE
**Layer:** EFF
**Depends on:** EFF-005

A value of type `secret<T>` may only be passed as an argument to a `pure`
function. Passing a `secret<T>` to any non-pure function is a compile error,
regardless of the callee's specific effects.

This rule prevents accidental leakage of secret values through effectful
channels (e.g., printing a secret via `print`).

To pass a secret value to a non-pure function, the caller must first explicitly
unwrap it using `reveal()`.

**Invariant:**
`arg: secret<T> passed to G → G is pure`

**Compliance Test:**
PASS:
```candor
fn hash(x: secret<i64>) -> i64 pure { return x * 2654435761 }
let s = secret(42)
let h = hash(s)
```
FAIL:
```candor
fn sink(x: i64) -> unit effects(io) { print_int(x) }
let s = secret(99)
sink(s)
```

**On Violation:**
`secret<T> value cannot be passed to a non-pure function; use reveal() to unwrap explicitly`

---

*End of L7-EFFECTS.md*
