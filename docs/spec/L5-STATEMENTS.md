# LAYER 5: STATEMENTS
**Status:** FROZEN (core), STABLE (must{}, for, assert)
**Depends on:** L4-EVAL.md, L3-TYPES.md, L2-SYNTAX.md

---

## PURPOSE

Statements are the building blocks of function bodies.
A statement produces effects (AXIOM-002) or binds names (AXIOM-001).
Statements do not return values except where they are also expressions.

---

## STATEMENT GRAMMAR (overview)

```ebnf
Stmt       ::= LetStmt
             | ReturnStmt
             | IfStmt
             | LoopStmt
             | WhileStmt
             | ForStmt
             | BreakStmt
             | ContinueStmt
             | AssertStmt
             | AssignStmt
             | FieldAssignStmt
             | IndexAssignStmt
             | ExprStmt

Block      ::= '{' Stmt* '}'
```

Each statement form is specified individually below.

---

### RULE STMT-001: Expression Statement
**Status:** FROZEN
**Layer:** STMT
**Depends on:** AXIOM-002, TYP-050, TYP-055

An expression used as a statement; its value is discarded after evaluation.
The expression is evaluated for its side effects only.

Discarding a value of type `result<T,E>` or `option<T>` violates AXIOM-003
(errors must be handled) and is a compile-time error.

**Grammar:**
```ebnf
ExprStmt ::= Expr
```

**Type Rule:**
```
Γ ⊢ e : T     T ∉ {result<_,_>, option<_>}
────────────────────────────────────────────
Γ ⊢ (ExprStmt e) : unit
```

**Invariant:** The expression type is not `result<T,E>` or `option<T>`.

**Compliance Test:**
PASS:
```candor
fn f() -> unit {
    print("hello")
}
```
FAIL:
```candor
fn f() -> unit {
    read_file("x.txt")   // result<str,str> silently discarded
}
```

**On Violation:** `expression result of type X must be used (result/option cannot be discarded)`

---

### RULE STMT-005: `let` Binding
**Status:** FROZEN
**Layer:** STMT
**Depends on:** AXIOM-001, TYP-080, TYP-101

`let` binds an immutable name in the current block scope.
The name is visible from the point of declaration to the end of the enclosing block.
The name may not be reassigned after initialization (see STMT-007).

A type annotation is optional when the type can be inferred from the initializer expression.
When a type annotation is present and the initializer type differs, the compiler
attempts a literal coercion (TYP-080); if that fails it emits a type mismatch error.

A `let` binding always requires an initializer expression. There is no uninitialized binding.

**Grammar:**
```ebnf
LetStmt ::= 'let' Ident (':' Type)? '=' Expr
```

**Type Rule:**
```
Γ ⊢ e : T     (annotation A absent  ∨  Coerce(T, A) = T')
──────────────────────────────────────────────────────────
Γ, x : T' (immutable) ⊢ (let x = e) : unit
```

**Invariant:** Every `let` binding has an initializer; the inferred type is fully resolved.

**Compliance Test:**
PASS: `let x = 5`
PASS: `let x: i32 = 5`
FAIL: `let x: i32 = "hello"` → type mismatch: cannot use str as i32
FAIL:
```candor
let x = 5
x = 10   // re-assignment to immutable binding
```

**On Violation:** `type mismatch: cannot use X as Y`

---

### RULE STMT-006: `let mut` Binding
**Status:** FROZEN
**Layer:** STMT
**Depends on:** STMT-005

`let mut` binds a mutable name in the current block scope.
The name may be reassigned after initialization via an assignment statement (STMT-010).
In all other respects the binding behaves identically to `let` (STMT-005).

**Grammar:**
```ebnf
LetMutStmt ::= 'let' 'mut' Ident (':' Type)? '=' Expr
```

**Type Rule:**
```
Γ ⊢ e : T     (annotation A absent  ∨  Coerce(T, A) = T')
──────────────────────────────────────────────────────────
Γ, x : T' (mutable) ⊢ (let mut x = e) : unit
```

**Invariant:** The binding is marked mutable in the scope; subsequent assignments are permitted.

**Compliance Test:**
PASS:
```candor
let mut count: i64 = 0
count = count + 1
```
FAIL: `let mut x: i32 = "hello"` → type mismatch (same rules as STMT-005)

---

### RULE STMT-007: Immutable Binding Restriction
**Status:** FROZEN
**Layer:** STMT
**Depends on:** STMT-005

A binding declared with `let` (without `mut`) may not appear on the left-hand side
of an assignment statement after its initialization.
This restriction applies to simple assignment, field assignment, index assignment,
and all augmented assignment forms.

**Invariant:** `mutable(x)` is false for any `let x` binding; assignments to such bindings are rejected.

**Compliance Test:**
PASS:
```candor
let x = 5
let y = x + 1
```
FAIL:
```candor
let x = 5
x = 10
```
FAIL:
```candor
let x = 5
x += 1
```

**On Violation:** `cannot assign to immutable variable "X"`

---

### RULE STMT-010: Assignment Statement
**Status:** FROZEN
**Layer:** STMT
**Depends on:** STMT-006, STMT-007, TYP-080

An assignment statement stores a new value into an existing mutable binding,
a mutable struct field, or a mutable indexed position.

Assignment targets:
- `name = Expr` — simple variable assignment
- `name.field = Expr` — field assignment on a mutable struct variable
- `name[idx] = Expr` — index assignment into a mutable `vec<T>` or `ring<T>`

The type of the right-hand side must exactly match the declared type of the target
(via `Coerce`; see TYP-080). No implicit widening occurs.

The binding (or the variable holding the struct/collection) must be declared `mut`.

**Grammar:**
```ebnf
AssignStmt      ::= Ident '=' Expr
FieldAssignStmt ::= Expr '.' Ident '=' Expr
IndexAssignStmt ::= Expr '[' Expr ']' '=' Expr
```

**Type Rule:**
```
mutable(x)    Γ ⊢ e : T    Coerce(T, type(x)) = T'
───────────────────────────────────────────────────
Γ ⊢ (x = e) : unit
```

**Invariant:** The target binding is mutable; the value type coerces to the target type.

**Compliance Test:**
PASS:
```candor
let mut x: i64 = 0
x = 42
```
PASS:
```candor
struct Point { x: i64, y: i64 }
let mut p = Point{ x: 0, y: 0 }
p.x = 10
```
FAIL:
```candor
let x: i64 = 0
x = 42   // immutable
```
FAIL:
```candor
let mut x: i64 = 0
x = "hello"   // type mismatch
```

**On Violation:**
- `cannot assign to immutable variable "X"`
- `type mismatch: cannot assign X to Y`

---

### RULE STMT-011: Augmented Assignment
**Status:** FROZEN
**Layer:** STMT
**Depends on:** STMT-010

Augmented assignment operators are syntactic sugar that combine a binary operation
with an assignment. The five operators are: `+=`, `-=`, `*=`, `/=`, `%=`.

Each desugars to a simple assignment:

| Written        | Desugars to         |
|----------------|---------------------|
| `x += e`       | `x = x + e`         |
| `x -= e`       | `x = x - e`         |
| `x *= e`       | `x = x * e`         |
| `x /= e`       | `x = x / e`         |
| `x %= e`       | `x = x % e`         |

All rules of STMT-010 apply to the desugared form: the target must be mutable and
the result type must match the target type exactly.

Targets may be a simple identifier, a field expression, or an index expression.
Any other left-hand side is a parse error.

**Grammar:**
```ebnf
AugAssignStmt ::= AssignTarget ('+=' | '-=' | '*=' | '/=' | '%=') Expr
AssignTarget  ::= Ident | Expr '.' Ident | Expr '[' Expr ']'
```

**Invariant:** Desugaring produces a valid STMT-010 assignment; the target is mutable.

**Compliance Test:**
PASS:
```candor
let mut x: i64 = 10
x += 5
```
PASS:
```candor
let mut v: vec<i64> = vec_new()
v[0] += 1
```
FAIL:
```candor
let x: i64 = 10
x += 5   // immutable
```
FAIL: `5 += 1` → invalid compound-assignment target

**On Violation:**
- `cannot assign to immutable variable "X"`
- `invalid compound-assignment target`

---

### RULE STMT-020: `if` Statement
**Status:** FROZEN
**Layer:** STMT
**Depends on:** TYP-020

An `if` statement conditionally executes a block.
The condition expression must have type `bool`.
An optional chain of `else if` branches and a final `else` branch may follow.

All branches are independent statements; the `if` statement as a whole has type `unit`.
(For `if` as an expression yielding a value, see L4-EVAL.)

**Grammar:**
```ebnf
IfStmt ::= 'if' Expr Block ('else' 'if' Expr Block)* ('else' Block)?
```

**Type Rule:**
```
Γ ⊢ cond : bool    Γ ⊢ then : unit    (Γ ⊢ else : unit)?
─────────────────────────────────────────────────────────
Γ ⊢ (if cond then else?) : unit
```

**Invariant:** The condition type is `bool`.

**Compliance Test:**
PASS:
```candor
let x: i64 = 5
if x > 3 {
    print("big")
} else {
    print("small")
}
```
FAIL:
```candor
if 1 {   // i64 is not bool
    print("x")
}
```

**On Violation:** `if condition must be bool, got X`

---

### RULE STMT-031: `loop` Statement
**Status:** FROZEN
**Layer:** STMT
**Depends on:** AXIOM-002

`loop` begins an unconditional infinite loop that executes its body repeatedly.
The only way to exit a `loop` is via `break` (STMT-040) or `return` (STMT-050).

**Grammar:**
```ebnf
LoopStmt ::= 'loop' Block
```

**Invariant:** A `loop` statement always contains at least one reachable `break` or `return`
for the program to terminate; this is not enforced at the type level but is a semantic
requirement for non-diverging programs.

**Compliance Test:**
PASS:
```candor
let mut i: i64 = 0
loop {
    i += 1
    if i >= 10 { break }
}
```

---

### RULE STMT-032: `while` Statement
**Status:** FROZEN
**Layer:** STMT
**Depends on:** TYP-020

`while` repeats its body as long as the condition is `true`.
The condition is re-evaluated before each iteration.
The condition expression must have type `bool`.

**Grammar:**
```ebnf
WhileStmt ::= 'while' Expr Block
```

**Type Rule:**
```
Γ ⊢ cond : bool    Γ ⊢ body : unit
────────────────────────────────────
Γ ⊢ (while cond body) : unit
```

**Invariant:** The condition type is `bool`.

**Compliance Test:**
PASS:
```candor
let mut n: i64 = 10
while n > 0 {
    n -= 1
}
```
FAIL:
```candor
while 1 {   // i64, not bool
    break
}
```

**On Violation:** `while condition must be bool, got X`

---

### RULE STMT-033: `for` Statement
**Status:** STABLE
**Layer:** STMT
**Depends on:** TYP-070

`for` iterates over the elements of a collection, binding each element to a loop
variable for the duration of the body block.

Two forms exist:

1. **Element iteration** — `for name in Expr Block`
   - The `in` expression must have type `vec<T>`, `ring<T>`, or `set<T>`.
   - `name` is bound as an immutable variable of type `T` for each iteration.

2. **Key-value iteration** — `for k, v in Expr Block`
   - The `in` expression must have type `map<K, V>`.
   - `k` is bound as type `K`; `v` is bound as type `V`.

The loop variable(s) are scoped to the body block and are not visible after it.

**Grammar:**
```ebnf
ForStmt ::= 'for' Ident (',' Ident)? 'in' Expr Block
```

**Type Rule:**
```
Γ ⊢ coll : vec<T>   (or ring<T> or set<T>)
────────────────────────────────────────────────
Γ, name : T ⊢ (for name in coll body) : unit

Γ ⊢ coll : map<K, V>
─────────────────────────────────────────────
Γ, k : K, v : V ⊢ (for k, v in coll body) : unit
```

**Invariant:** The collection expression has type `vec<T>`, `ring<T>`, `set<T>`, or `map<K,V>`.

**Compliance Test:**
PASS:
```candor
let items: vec<i64> = vec_new()
for x in items {
    print_int(x)
}
```
PASS:
```candor
let m: map<str, i64> = map_new()
for k, v in m {
    print(k)
}
```
FAIL:
```candor
for x in 42 {   // i64 is not iterable
    print_int(x)
}
```

**On Violation:**
- `for...in requires vec<T>, ring<T>, or set<T>, got X`
- `for k, v in ... requires map<K,V>, got X`

---

### RULE STMT-040: `break` Statement
**Status:** FROZEN
**Layer:** STMT
**Depends on:** STMT-031, STMT-032, STMT-033

`break` exits the innermost enclosing `loop`, `while`, or `for` statement immediately.
Control transfers to the statement following the loop.

`break` is only valid inside a `loop`, `while`, or `for` body.
Use outside any loop context is a compile-time error.

`break` inside a `must{}` arm desugars to a `BreakExpr` with type `never`,
which satisfies arm type unification without requiring a value.

**Grammar:**
```ebnf
BreakStmt ::= 'break'
```

**Invariant:** `break` appears lexically inside a `loop`, `while`, or `for` body.

**Compliance Test:**
PASS:
```candor
loop {
    break
}
```
PASS:
```candor
let items: vec<i64> = vec_new()
for x in items {
    if x > 10 { break }
}
```
FAIL:
```candor
fn f() -> unit {
    break   // not inside a loop
}
```

**On Violation:** `break outside loop`

---

### RULE STMT-041: `continue` Statement
**Status:** FROZEN
**Layer:** STMT
**Depends on:** STMT-031, STMT-032, STMT-033

`continue` skips the remainder of the current iteration of the innermost enclosing
`loop`, `while`, or `for` statement, and proceeds to the next iteration.

For `loop`: execution returns to the top of the loop body.
For `while`: execution re-evaluates the condition.
For `for`: execution advances to the next element.

`continue` is only valid inside a `loop`, `while`, or `for` body.
Use outside any loop context is a compile-time error.

**Grammar:**
```ebnf
ContinueStmt ::= 'continue'
```

**Invariant:** `continue` appears lexically inside a `loop`, `while`, or `for` body.

**Compliance Test:**
PASS:
```candor
let mut i: i64 = 0
while i < 10 {
    i += 1
    if i == 5 { continue }
    print_int(i)
}
```
FAIL:
```candor
fn f() -> unit {
    continue   // not inside a loop
}
```

**On Violation:** `continue outside loop`

---

### RULE STMT-050: `return` Statement
**Status:** FROZEN
**Layer:** STMT
**Depends on:** TYP-025, TYP-080

`return` exits the current function, optionally providing a value to the caller.

- **Bare return** (`return` with no expression): valid only when the function's declared
  return type is `unit`. Returns the single `unit` value.
- **Value return** (`return Expr`): the expression type must match (via `Coerce`)
  the function's declared return type.

The return type of the enclosing function is established at function declaration
(L6-FUNCTIONS) and is constant throughout the body.

**Grammar:**
```ebnf
ReturnStmt ::= 'return' Expr?
```

**Type Rule:**
```
retType = unit
───────────────────────────────
Γ ⊢ (return) : never

Γ ⊢ e : T    Coerce(T, retType) = T'
─────────────────────────────────────
Γ ⊢ (return e) : never
```

`return` has type `never` because control does not proceed past it.

**Invariant:**
- Bare `return` requires `retType = unit`.
- `return Expr` requires `Coerce(exprType, retType)` to succeed.

**Compliance Test:**
PASS:
```candor
fn add(a: i64, b: i64) -> i64 {
    return a + b
}
```
PASS:
```candor
fn nothing() -> unit {
    return
}
```
FAIL:
```candor
fn f() -> i64 {
    return "hello"   // str does not match i64
}
```
FAIL:
```candor
fn f() -> i64 {
    return   // bare return in non-unit function
}
```

**On Violation:**
- `return type mismatch: got X, expected Y`
- `bare return in function returning X`

---

### RULE STMT-060: `const` Declaration
**Status:** STABLE
**Layer:** STMT
**Depends on:** AXIOM-001, TYP-080

`const` declares a named compile-time constant.
The type annotation is mandatory.
The initializer expression must be a compile-time constant: a literal,
an arithmetic expression over literals, or a reference to another `const`.

`const` names are **module-scope** (top-level visibility), not block-scoped like `let`.
A `const` declared inside a function body is not supported; `const` is a top-level declaration.
`const` names are accessible throughout the module without forward-declaration restrictions.

The initializer is type-checked in an empty scope (no local variables visible),
enforcing that the value depends only on literals and other constants.

`const` names follow the same identifier rules as other bindings (LEX layer),
but by convention are written in UPPER_SNAKE_CASE.

**Grammar:**
```ebnf
ConstDecl ::= 'const' Ident ':' Type '=' Expr
```

**Type Rule:**
```
∅ ⊢ e : T    Coerce(T, A) = T'
────────────────────────────────
⊢ (const Name : A = e) registers Name : A
```

**Invariant:**
- The type annotation is present and resolves to a concrete type.
- The initializer type coerces to the annotated type.
- The initializer is evaluated in an empty scope (no local variables).

**Compliance Test:**
PASS:
```candor
const MAX_SIZE: i64 = 1024
```
PASS:
```candor
const PI: f64 = 3.14159
```
FAIL:
```candor
const X: i64 = "hello"   // str not coercible to i64
```
FAIL:
```candor
fn f() -> unit {
    let y: i64 = 5
    const Z: i64 = y   // y is not a constant expression
}
```

**On Violation:** `const X: cannot use Y as Z`

---

### RULE STMT-070: `must{}` Expression
**Status:** STABLE
**Layer:** STMT
**Depends on:** TYP-050, TYP-055, AXIOM-003

`must{}` is a postfix expression that pattern-matches a `result<T,E>` or `option<T>` value
and forces the programmer to handle every variant.

The subject expression (before `must`) must have type `result<T,E>` or `option<T>`.
Any other subject type is a compile-time error.

Each arm specifies a pattern and a body expression:
- For `result<T,E>`: arms `ok(x) => body` and `err(e) => body`
- For `option<T>`: arms `some(x) => body` and `none => body`
- A wildcard arm `_ => body` matches any remaining variant

The bound variable in `ok(x)`, `err(e)`, or `some(x)` is scoped to that arm's body.

Arms must be exhaustive: every variant of the subject type must be covered.
(A wildcard arm `_ => ...` satisfies exhaustiveness for any remaining variants.)

All arm bodies that do not diverge (type `never`) must have the same type.
The overall type of the `must{}` expression is the common arm body type.
If all arms diverge, the type is `unit`.

Arms that end with `return`, `break`, or a `{ ... }` block ending in a control-flow
statement have type `never` and are excluded from unification.

**Grammar:**
```ebnf
MustExpr  ::= Expr 'must' '{' MustArm+ '}'
MustArm   ::= MustPattern '=>' MustBody ','?
MustPattern ::= 'ok' '(' Ident ')'
              | 'err' '(' Ident ')'
              | 'some' '(' Ident ')'
              | 'none'
              | '_'
MustBody  ::= Expr | Block | 'return' Expr | 'break'
```

**Type Rule:**
```
Γ ⊢ subject : result<T, E>
Γ, x : T ⊢ ok_body : U     Γ, e : E ⊢ err_body : U  (or never)
──────────────────────────────────────────────────────────────────
Γ ⊢ subject must { ok(x) => ok_body, err(e) => err_body } : U

Γ ⊢ subject : option<T>
Γ, x : T ⊢ some_body : U     Γ ⊢ none_body : U  (or never)
─────────────────────────────────────────────────────────────
Γ ⊢ subject must { some(x) => some_body, none => none_body } : U
```

**Invariant:**
- Subject type is `result<T,E>` or `option<T>`.
- Arms are exhaustive over the subject type's variants.
- All non-diverging arm body types unify to a single type `U`.

**Compliance Test:**
PASS:
```candor
let r: result<i64, str> = ok(42)
let value = r must {
    ok(x) => x,
    err(e) => 0,
}
```
PASS:
```candor
let r: result<i64, str> = ok(42)
r must {
    ok(x)  => print_int(x),
    err(e) => return,
}
```
PASS:
```candor
let o: option<i64> = some(7)
let n = o must {
    some(x) => x,
    none    => 0,
}
```
FAIL:
```candor
let x: i64 = 5
x must {   // i64 is not result or option
    _ => 0,
}
```
FAIL:
```candor
let r: result<i64, str> = ok(1)
r must {
    ok(x) => x,
    // err arm missing — not exhaustive
}
```

**On Violation:** `must{} requires result<T,E> or option<T>, got X`

---

### RULE STMT-071: `must{}` Type Errors
**Status:** STABLE
**Layer:** STMT
**Depends on:** STMT-070

This rule specifies the exact error messages emitted for `must{}` type violations.

**Case 1: Invalid subject type.**
The subject expression is not `result<T,E>` or `option<T>`.

**On Violation:** `must{} requires result<T,E> or option<T>, got X`

**Case 2: Arm type mismatch.**
Two or more non-diverging arm bodies produce types that cannot be unified.

**On Violation:** `must{} arm type X does not match expected Y`

**Compliance Test:**
FAIL: `42 must { _ => 0 }` → `must{} requires result<T,E> or option<T>, got i64`
FAIL:
```candor
let r: result<i64, str> = ok(1)
r must {
    ok(x)  => x,       // i64
    err(e) => "bad",   // str — does not match i64
}
```
→ `must{} arm type str does not match expected i64`

---

### RULE STMT-072: `assert` Statement
**Status:** STABLE
**Layer:** STMT
**Depends on:** TYP-020, AXIOM-002

`assert` evaluates a boolean expression and panics at runtime if it is `false`.
The expression must have type `bool`.

At compile time, when the expression can be fully evaluated (constant folding / comptime pass),
a failing assertion is a compile-time error.

`assert` is distinct from `requires`/`ensures` (L11-CONTRACTS): `assert` appears inside
function bodies and always generates a runtime check; contract clauses appear in function
signatures and may generate proof obligations for formal verification.

**Grammar:**
```ebnf
AssertStmt ::= 'assert' Expr
```

**Type Rule:**
```
Γ ⊢ e : bool
──────────────────────────────
Γ ⊢ (assert e) : unit
```

**Invariant:** The expression type is `bool`.

**Runtime Semantics:** If the expression evaluates to `false`, the program terminates
immediately with a panic message identifying the assertion that failed.

**Compliance Test:**
PASS:
```candor
fn divide(a: i64, b: i64) -> i64 {
    assert b != 0
    return a / b
}
```
FAIL:
```candor
assert 42   // i64 is not bool
```
FAIL:
```candor
assert "ok"   // str is not bool
```

**On Violation:** `assert requires bool, got X`

---

*End of L5-STATEMENTS.md*
