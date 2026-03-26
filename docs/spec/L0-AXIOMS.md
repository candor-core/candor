# LAYER 0: AXIOMS
**Status:** FROZEN
**Depends on:** FRAMEWORK.md

---

## PURPOSE

These are the non-negotiable truths about Candor. They are not rules about syntax
or types. They are the properties that every rule in every layer must preserve.
If a proposed rule violates an axiom, the rule is invalid — not the axiom.

Axioms are amended only to increase precision. An amendment that changes the
meaning of an existing axiom requires a new axiom; the original is SUPERSEDED,
not modified.

---

### RULE AXIOM-001: One Meaning Per Expression

**Status:** FROZEN
**Layer:** AXIOM
**Depends on:** none

Every valid Candor expression, in any context, has exactly one type
and exactly one evaluation result.

**Invariant:**
For any expression `e` in context `Γ`:
- `Γ ⊢ e : T` holds for at most one type `T`
- The evaluation of `e` in state `σ` produces at most one value `v`

Overloading, implicit coercion, or context-dependent interpretation
that could produce multiple valid types is forbidden.

**Compliance Test:**
PASS: `let x: i64 = 5` — the literal `5` has type `i64` in this context (annotation-guided)
FAIL: Any language feature that makes `5` have two valid types simultaneously

**On Violation:**
Any rule that introduces ambiguous typing violates AXIOM-001 and is invalid.

---

### RULE AXIOM-002: Every Effect Is Declared

**Status:** FROZEN
**Layer:** AXIOM
**Depends on:** none

A function's observable interactions with external state are completely
described by its effects annotation. No function may have an observable
side effect not declared in its annotation.

**Invariant:**
`effects(declared) ⊇ effects(actual)` for all functions.

A function with no effects annotation is trusted (unchecked).
A function with `pure` produces no side effects.
A function with `effects(X, Y)` may only produce effects X and Y.

**Compliance Test:**
PASS: A function declared `effects(io)` that reads a file
FAIL: A compiler that allows a `pure` function to call `print()`

**On Violation:** AXIOM-002 is the foundation of EFF-001–EFF-099.
Any EFF rule that allows undeclared effects violates AXIOM-002.

---

### RULE AXIOM-003: Every Error Is Handled

**Status:** FROZEN
**Layer:** AXIOM
**Depends on:** none

A function returning `result<T,E>` or `option<T>` cannot have its
return value silently discarded. The caller must explicitly handle
both the success and failure paths.

**Invariant:**
For any call `f(...)` where `f` returns `result<T,E>` or `option<T>`:
- The return value must be bound (`let x = f(...)`)
- The bound value must be matched or propagated via `must{}`

A `result<T,E>` or `option<T>` cannot be used as `unit` without
explicit extraction.

**Compliance Test:**
PASS:
```candor
let v = divide(10, 2) must { ok(x) => x  err(e) => return unit }
```
FAIL:
```candor
divide(10, 2)   ## discarding a result<i64,str> silently
```

**On Violation:** AXIOM-003 is the foundation of STMT-070–STMT-079.
A compiler that silently discards `result<T,E>` values is non-conformant.

---

### RULE AXIOM-004: Explicit Ownership Transfer

**Status:** FROZEN
**Layer:** AXIOM
**Depends on:** none

Movement of heap-allocated values between scopes is always syntactically
visible. There are no implicit copies of owned values. If a value moves,
the source is no longer accessible.

**Invariant:**
After an owned value `x` is moved, any access to `x` is a compile-time error.

**Compliance Test:**
PASS: `let y = x` where `x` is `box<T>` — y owns the value, x is gone
FAIL: A compiler that allows reading `x` after `let y = x` for owned types

**On Violation:** AXIOM-004 is the foundation of OWN-001–OWN-099.

---

### RULE AXIOM-005: Compile-Time Decidability

**Status:** FROZEN
**Layer:** AXIOM
**Depends on:** none

Every type rule and every effect rule is decidable at compile time
without executing the program. No rule may require runtime information
to determine correctness.

Contracts (`requires`/`ensures`) are checked at compile time through
symbolic evaluation or flagged as `runtime-checked` when symbolic
evaluation is impossible. They are never silently skipped.

**Invariant:**
The compiler halts with a definitive ACCEPT or REJECT for any program
in finite time, given finite resources.

**Compliance Test:**
PASS: Any program the compiler accepts or rejects with an error
FAIL: A compiler that runs forever or crashes on a syntactically valid program

**On Violation:** Any rule that requires unbounded analysis violates AXIOM-005.

---

### RULE AXIOM-006: Source Is the Authority

**Status:** FROZEN
**Layer:** AXIOM
**Depends on:** none

The `.cnd` source file is the complete specification of program behavior.
No behavior may depend on compiler flags, environment variables, platform,
or build configuration except where explicitly declared in the source.

**Invariant:**
Two conformant compilers compiling the same `.cnd` source produce
programs with identical observable behavior (modulo declared
platform-specific effects).

**Compliance Test:**
PASS: A program that behaves identically on Linux and Windows given the same inputs
FAIL: A program whose behavior depends on undeclared compiler flags

**On Violation:** Any feature that introduces undeclared platform-dependent behavior
violates AXIOM-006. Such features must use `effects()` to declare their dependency.

---

### RULE AXIOM-007: Structural Layering

**Status:** FROZEN
**Layer:** AXIOM
**Depends on:** none

Every language feature belongs to exactly one layer.
A feature at layer N may only use features from layers 0 through N.
No circular dependencies between layers are permitted.

**Invariant:**
The dependency graph of layer features is a DAG rooted at AXIOM.

**Compliance Test:**
PASS: `effects(io)` annotation (Layer 7) on a function that uses `str` (Layer 3, TYP-020)
FAIL: A type system rule (Layer 3) that depends on the effects system (Layer 7)

**On Violation:** Any rule in layer N that depends on a rule in layer M > N violates AXIOM-007.

---

*End of L0-AXIOMS.md*
