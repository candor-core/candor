# Pure Function Arena Allocation

*Design specification — April 2026*

---

## Problem Statement

A `pure` function that performs non-trivial work — parsing, transforming, aggregating — may
create many intermediate heap allocations: strings, vecs, result values, temporary structs.
Under a general-purpose allocator, each allocation is a separate `malloc` call and each
intermediate that goes out of scope is a separate `free` call. For a `pure` function that
produces one return value from ten intermediate values, this is ten mallocs and nine frees
for allocations the compiler already knows are transient.

This is not a theoretical concern. The Candor compiler itself (`typeck.cnd`, `parser.cnd`)
contains hundreds of `pure` functions that allocate intermediate vecs and strings. Every
compilation is paying this tax.

---

## The Core Insight

`pure` carries a compiler-verified guarantee: **no allocation made inside a `pure` function
can be observed after the function returns, except through the return value.**

This is a consequence of the no-side-effects rule. A `pure` function cannot:
- Store a pointer to an intermediate in a global
- Pass a pointer out through a mutable reference parameter
- Communicate an intermediate's address via any I/O channel

Therefore, when a `pure` function returns, every allocation it made — except the return
value itself — is unreachable. They can all be freed in a single operation.

**No other mainstream language can derive this automatically.** Go, Rust, and Python have
no purity annotation the compiler can trust. Rust's borrow checker provides related
guarantees but requires explicit lifetime annotations; it cannot automatically scope all
intermediates to a single arena without programmer-managed arena types. Candor can do this
as an invisible compiler optimization, requiring zero source changes.

---

## Mechanism: Caller-Scoped Arena

When the compiler determines a function is `pure`, it allocates a small arena (stack-resident
or thread-local) before the call. All allocations made within the `pure` function's call
tree — including calls to other `pure` functions — draw from this arena instead of the
global heap.

```
caller ──► calls pure fn ──► arena created
                             │
                             ├─ let a = parse_name(line)?       ← arena alloc
                             ├─ let b = validate_score(a)?      ← arena alloc
                             ├─ let c = build_row(a, b)         ← arena alloc
                             │
                             └─ returns ok(row)                 ← return value copied out
                                                                   to caller's allocator
                                arena freed in one operation
```

The return value is the only allocation that escapes the arena. It is copied to the
caller's allocator (or stack, if the return type fits). Then the entire arena is
reclaimed — pointer decrement, not per-object traversal.

---

## Formal Specification

### PURE-ARENA-001 — Intermediate allocations in `pure` functions use a transient arena

**Condition:** A function `f` annotated `pure` makes one or more heap allocations
(strings, vecs, `box<T>`, `result<T,E>` with heap-allocated arms) during its execution.

**Guarantee:** All such allocations, except the value returned by `f`, are unreachable
after `f` returns. The compiler may place them in a call-scoped arena and reclaim the
entire arena on `f`'s return path — both the `ok(...)` and `err(...)` paths.

**Constraint:** The return value of `f` must be copied (or moved) to the caller's allocator
before the arena is freed. The compiler is responsible for this copy on all return paths.

**Visibility:** This optimization is not observable from Candor source code. A correct
program produces identical values with or without the optimization. The only observable
difference is memory usage and allocation throughput.

### PURE-ARENA-002 — Transitivity: called pure functions share the arena

When a `pure` function `f` calls another `pure` function `g`, `g`'s allocations draw from
the same arena as `f`. The arena is not created per-call — it is created once at the
outermost `pure` call in the call stack and freed when that outermost call returns.

This is the compounding benefit: a chain of five `pure` functions, each building on the
last, produces a single arena allocation + a single arena free, regardless of the number
of intermediate values created along the chain.

### PURE-ARENA-003 — `effectful` boundary flushes the arena

When a `pure` call chain is invoked from an `effects(...)` context, the arena is created
at the call site of the first `pure` function in the chain and freed when control returns
to the `effects(...)` context. The return value is promoted to the general heap allocator
at that boundary.

This ensures that values returned from `pure` chains into effectful contexts have normal
lifetime semantics — they are owned by the effectful caller and live until that caller
explicitly drops them.

### PURE-ARENA-004 — Error paths are not special-cased

An `err(e)` return from a `pure` function is subject to the same rules as `ok(v)`:
the error value is copied to the caller's allocator before the arena is freed. This
matters because error strings (`str`) are heap-allocated. The compiler must ensure the
error string is promoted before arena reclamation.

Consequence: `?` operator propagation (M14.1) through a `pure` call chain is always
safe — each `?` either continues within the arena or promotes an error value out.

---

## Arena Sizing Strategy

The arena must be large enough to hold all intermediates for the deepest `pure` call
chain in the program. Two strategies:

**Fixed-size stack arena (default):** A fixed-size buffer (e.g., 64 KB) allocated on the
call stack before the outermost `pure` call. If the `pure` chain exhausts this buffer,
fall back to the heap allocator for the overflow. The fallback is correct and safe; it
only means the full optimization is not applied for that call. The compiler can warn
(`--warn=arena-overflow`) when static analysis predicts a `pure` chain may exceed the
fixed size.

**Heap-backed arena (for large chains):** When the compiler can statically bound the
allocation size of a `pure` chain (e.g., for functions operating on `str` inputs of
known maximum length), it can size the arena precisely. This is an advanced optimization
requiring inter-procedural analysis and is deferred.

The default fixed-size stack arena covers the common case: a `pure` function doing parsing
or transformation on a single input, producing a single output. The stack allocation is
itself free (stack pointer adjustment) and has zero GC pressure.

---

## Interaction with Existing Features

| Feature | Interaction |
|---------|-------------|
| `box<T>` | `box_new` inside `pure` allocates from the arena; box is freed with the arena unless returned |
| `arc<T>` | `arc_new` inside `pure` allocates from the arena; the refcount is 1 throughout (no concurrent access possible in a `pure` fn); if returned, promoted to heap before arena free |
| `vec<T>` | `vec_push` (realloc growth) inside `pure` uses arena; the backing buffer is freed with the arena unless returned |
| `result<T,E>` | Both arms use the arena; the survivor is promoted on return; see PURE-ARENA-004 |
| `str` operations | `str_concat`, `str_substr` and similar allocate from the arena inside `pure`; final result promoted on return |
| `?` propagation | Safe across all `pure` call sites; error values promoted before arena free |
| `|>` pipeline | A chain `a |> b |> c` where all are `pure` shares a single arena for the whole chain |
| `must{}` blocks | Both arms execute within the same arena; no special handling needed |

---

## Example: Token Count Before and After

Consider `parse_row` from `examples/pipeline.cnd`:

```candor
fn parse_row(line: str) -> result<Row, str> pure {
    let parts = str_split(line, ',')           ## arena alloc: vec<str>
    if vec_len(parts) != 2 {
        return err("expected 2 fields")        ## arena alloc: str; promoted on err return
    }
    let name  = parts[0]                       ## no alloc (slice of parts)
    let score = str_to_int(parts[1])?          ## arena alloc: err str if parse fails
    if score < 0 or score > 100 {
        return err(str_concat("score out of range: ", int_to_str(score)))
    }
    ok(Row { name: name, score: score })       ## Row copied out; parts vec freed with arena
}
```

**Without pure arena:** `str_split` → `malloc` for vec + 2 str slices. `int_to_str` → 
`malloc` for digit string. `Row` constructed → `malloc` (if heap-allocated). Return → 
`free` for `parts` vec + intermediates. Total: 3–5 malloc/free pairs per call.

**With pure arena:** `str_split` draws from arena. `int_to_str` draws from arena. `Row` is
either stack-allocated or arena-allocated. Return → `Row` promoted to caller's allocator;
arena reclaimed in one pointer decrement. Total: 1 arena alloc + 1 arena free per call,
regardless of the number of intermediates.

---

## Implementation Notes (Compiler)

The optimization is implemented in the C and LLVM backends, not in the type-checker.
The type-checker already enforces `pure` — no new semantic rules are needed. The backends
replace the `malloc`/`free` pair in generated code with arena-scoped equivalents.

**C backend:** Emit a `_cnd_arena_t arena = _cnd_arena_init(buf, sizeof(buf))` before the
outermost `pure` call. Thread the arena pointer through all nested `pure` calls (either
as an implicit parameter or via a thread-local). Emit `_cnd_arena_reset(&arena)` on all
return paths of the outermost call, after the return value is promoted.

**LLVM backend:** Allocate the arena on the stack using `alloca`. The arena pointer is
a function argument in the IR for all `pure` functions, inserted by the compiler. All
`malloc` calls in `pure` fn bodies are replaced with arena bump-pointer advances.

The thread-local approach (arena pointer stored in a thread-local register) avoids adding
an implicit parameter to every `pure` function's ABI — which would break the C
interoperability guarantee. The thread-local is set at the outermost `pure` call boundary
and cleared when that boundary returns.

---

## Definition of Done

- [ ] `pure` function bodies in the C backend use arena allocation for all heap types
- [ ] Return values are promoted to the caller's allocator before arena reset
- [ ] `err(...)` return paths are promoted correctly (PURE-ARENA-004)
- [ ] Nested `pure` → `pure` calls share the enclosing arena (PURE-ARENA-002)
- [ ] Arena overflow falls back to heap allocator silently; `--warn=arena-overflow` flag available
- [ ] All existing `pure` function tests still pass (correctness unchanged)
- [ ] `candorc --profile` shows reduction in malloc/free call counts for `pure`-heavy programs
- [ ] Benchmark: `parse_row` called 100,000 times — measure wall time and peak RSS vs. without optimization

---

*Specification added April 2026. Companion to the token_compression_floor.md efficiency analysis.*
