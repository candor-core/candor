# LAYER 3: TYPE SYSTEM
**Status:** FROZEN (primitives), STABLE (generics, coercion)
**Depends on:** L0-AXIOMS.md, L1-LEXER.md, L2-SYNTAX.md

---

## PURPOSE

This layer defines every type in Candor: what it is, what values it contains,
how it is written in source, and what operations are valid on it.
The type of every expression is determined by these rules.

Types are checked at compile time. No type operation requires runtime inspection.

---

## PRIMITIVE TYPES

### RULE TYP-001: Type `i8`
**Status:** FROZEN | **Layer:** TYP | **Depends on:** AXIOM-001

`i8` is the type of 8-bit signed two's-complement integers.
Value range: -128 to 127 inclusive.

**Grammar:** `i8`

**Invariant:** All values of type `i8` are in [-128, 127].

**Compliance Test:**
PASS: `let x: i8 = 127`
FAIL: `let x: i8 = 128` → overflow error at compile time if literal, runtime-wrapped otherwise

---

### RULE TYP-002: Type `i16`
**Status:** FROZEN | **Layer:** TYP | **Depends on:** AXIOM-001

`i16` is the type of 16-bit signed two's-complement integers.
Value range: -32768 to 32767 inclusive.

---

### RULE TYP-003: Type `i32`
**Status:** FROZEN | **Layer:** TYP | **Depends on:** AXIOM-001

`i32` is the type of 32-bit signed two's-complement integers.
Value range: -2,147,483,648 to 2,147,483,647 inclusive.

---

### RULE TYP-004: Type `i64`
**Status:** FROZEN | **Layer:** TYP | **Depends on:** AXIOM-001

`i64` is the type of 64-bit signed two's-complement integers.
Value range: -9,223,372,036,854,775,808 to 9,223,372,036,854,775,807 inclusive.

`i64` is the **default integer type**. An integer literal without a type
annotation defaults to `i64` when the context does not specify another type.

---

### RULE TYP-005: Type `i128`
**Status:** FROZEN | **Layer:** TYP | **Depends on:** AXIOM-001

`i128` is the type of 128-bit signed two's-complement integers.

---

### RULE TYP-006: Type `u8`
**Status:** FROZEN | **Layer:** TYP | **Depends on:** AXIOM-001

`u8` is the type of 8-bit unsigned integers. Range: 0 to 255.

---

### RULE TYP-007: Type `u16`
**Status:** FROZEN | **Layer:** TYP | **Depends on:** AXIOM-001

`u16` is the type of 16-bit unsigned integers. Range: 0 to 65535.

---

### RULE TYP-008: Type `u32`
**Status:** FROZEN | **Layer:** TYP | **Depends on:** AXIOM-001

`u32` is the type of 32-bit unsigned integers. Range: 0 to 4,294,967,295.

---

### RULE TYP-009: Type `u64`
**Status:** FROZEN | **Layer:** TYP | **Depends on:** AXIOM-001

`u64` is the type of 64-bit unsigned integers. Range: 0 to 2^64 - 1.

---

### RULE TYP-010: Type `u128`
**Status:** FROZEN | **Layer:** TYP | **Depends on:** AXIOM-001

`u128` is the type of 128-bit unsigned integers. Range: 0 to 2^128 - 1.

---

### RULE TYP-011: Integer Type Hierarchy

**Status:** FROZEN
**Layer:** TYP
**Depends on:** TYP-001–TYP-010

Signed and unsigned integers are distinct types.
`i64` and `u64` are not interchangeable without an explicit `as` cast (EVAL-090).

The complete integer types are: `i8 i16 i32 i64 i128 u8 u16 u32 u64 u128`.

No implicit integer promotion exists. Every widening or narrowing conversion
requires an explicit `as` cast.

**Compliance Test:**
PASS: `let x: i32 = 5 as i32`
FAIL: `let x: i32 = 5i64_value` where `5i64_value: i64` — type mismatch without cast

---

### RULE TYP-012: Type `f32`
**Status:** FROZEN | **Layer:** TYP | **Depends on:** AXIOM-001

`f32` is the type of 32-bit IEEE 754 binary floating-point numbers.

---

### RULE TYP-013: Type `f64`
**Status:** FROZEN | **Layer:** TYP | **Depends on:** AXIOM-001

`f64` is the type of 64-bit IEEE 754 binary floating-point numbers.

`f64` is the **default float type**. A float literal without a type annotation
defaults to `f64` when the context does not specify another type.

---

### RULE TYP-014: Type `f16`
**Status:** STABLE | **Layer:** TYP | **Depends on:** AXIOM-001

`f16` is the type of 16-bit IEEE 754 binary floating-point numbers (half precision).
Primary use: ML weight storage, reduced-precision compute.

---

### RULE TYP-015: Type `bf16`
**Status:** STABLE | **Layer:** TYP | **Depends on:** AXIOM-001

`bf16` is the type of 16-bit bfloat (Brain Float) numbers:
8 exponent bits, 7 mantissa bits, same exponent range as `f32`.
Primary use: ML training with hardware accelerators (NVIDIA, Google TPU).

---

### RULE TYP-020: Type `str`
**Status:** FROZEN | **Layer:** TYP | **Depends on:** AXIOM-001

`str` is the type of immutable UTF-8 strings.
A `str` value is a sequence of Unicode code points.
String length is measured in bytes, not code points.

`str` values are immutable. There is no in-place mutation of `str`.
String concatenation produces a new `str`.

**Compliance Test:**
PASS: `let s: str = "hello"`
FAIL: `s[0] = 'H'` → `str` is not mutable by index

---

### RULE TYP-025: Type `unit`
**Status:** FROZEN | **Layer:** TYP | **Depends on:** AXIOM-001

`unit` is the type with exactly one value, also written `unit`.
It represents the absence of meaningful return value.
Functions that produce no value return `unit`.

`unit` is not `void`. `unit` is a first-class value.

**Compliance Test:**
PASS: `fn f() -> unit { return unit }`
PASS: `let x: unit = unit`

---

### RULE TYP-030: Tuple Types
**Status:** STABLE | **Layer:** TYP | **Depends on:** AXIOM-001

A tuple type is an ordered product of types: `(T1, T2, ..., TN)` for N ≥ 2.

A tuple value is a parenthesized, comma-separated list of expressions.
Tuple elements are accessed by index: `t.0`, `t.1`, etc.

**Grammar:**
```ebnf
TupleType  ::= '(' Type ',' Type (',' Type)* ')'
TupleLit   ::= '(' Expr ',' Expr (',' Expr)* ')'
```

**Compliance Test:**
PASS: `let p: (i64, str) = (42, "hello")`
PASS: `let x = p.0` → `x: i64 = 42`

---

## COMPOSITE TYPES

### RULE TYP-040: Struct Types
**Status:** FROZEN | **Layer:** TYP | **Depends on:** SYN-045

A struct type is a named product type declared with `struct`.
Struct fields are named and typed.
Struct values are created with struct literal syntax (EVAL-080).

A struct type's name is its identity. Two structs with identical fields
but different names are different types.

**Compliance Test:**
PASS:
```candor
struct Point { x: i64, y: i64 }
let p = Point{ x: 1, y: 2 }
```
FAIL: `Point{ x: 1 }` when `y` is required and has no default → missing field error

---

### RULE TYP-041: Enum Types
**Status:** FROZEN | **Layer:** TYP | **Depends on:** SYN-050

An enum type is a named sum type declared with `enum`.
Each variant is named and carries zero or one payload.

**Grammar:**
```ebnf
EnumVariant ::= Ident ( '(' Type ')' )?
```

An enum value is constructed by naming the variant with its payload:
`Variant::Name(payload)` or `Variant::Name` for payload-free variants.

When the enum type is unambiguous from context, the type name may be omitted:
`Name(payload)` instead of `Variant::Name(payload)`.

**Compliance Test:**
PASS:
```candor
enum Color { Red, Green, Blue }
let c = Color::Red
```
PASS:
```candor
enum Shape { Circle(f64), Rectangle(f64, f64) }
let s = Shape::Circle(3.14)
```

---

## ALGEBRAIC WRAPPER TYPES

### RULE TYP-050: Type `option<T>`
**Status:** FROZEN | **Layer:** TYP | **Depends on:** TYP-041, AXIOM-003

`option<T>` is the standard nullable type. Values:
- `some(v)` where `v: T` — the value is present
- `none` — the value is absent

There is no null pointer in Candor. `option<T>` is the only way to
express optionality. An `option<T>` value must be matched before use.

**Grammar:**
```ebnf
SomeExpr ::= 'some' '(' Expr ')'
NoneExpr ::= 'none'
```

**Type Rule:**
```
Γ ⊢ v : T
─────────────────────
Γ ⊢ some(v) : option<T>

──────────────────────
Γ ⊢ none : option<T>
```

**Compliance Test:**
PASS: `let x: option<i64> = some(42)`
PASS: `let x: option<i64> = none`
FAIL: `let x: i64 = none` → type mismatch

---

### RULE TYP-051: `none` and `some` Are Not Constructors

**Status:** FROZEN | **Layer:** TYP | **Depends on:** TYP-050

`none` is a keyword, not a function or identifier.
`some` is a keyword, not a function. `some(x)` is special syntax.
They cannot be used as values, passed as arguments, or shadowed.

**Compliance Test:**
FAIL: `let some = 5` → `some` is a keyword
FAIL: `let f = some` → `some` is not a value

---

### RULE TYP-055: Type `result<T, E>`
**Status:** FROZEN | **Layer:** TYP | **Depends on:** TYP-041, AXIOM-003

`result<T, E>` is the standard error type. Values:
- `ok(v)` where `v: T` — success
- `err(e)` where `e: E` — failure

A `result<T,E>` value must be matched or propagated before use (AXIOM-003).

**Type Rule:**
```
Γ ⊢ v : T
──────────────────────────
Γ ⊢ ok(v) : result<T, E>

Γ ⊢ e : E
──────────────────────────
Γ ⊢ err(e) : result<T, E>
```

**Compliance Test:**
PASS: `let r: result<i64, str> = ok(42)`
PASS: `let r: result<i64, str> = err("oops")`
FAIL: `let x: i64 = ok(42)` → cannot use result<i64,str> as i64 without extracting

---

## GENERIC TYPES

### RULE TYP-060: Function Types
**Status:** STABLE | **Layer:** TYP | **Depends on:** TYP-001–TYP-055

A function type describes a callable value with a specific signature.

```ebnf
FnType ::= 'fn' '(' (Type (',' Type)*)? ')' '->' Type
```

A function type includes parameter types and return type.
Effects are not part of the function type at this layer.
(Effects on function types are addressed in EFF-020.)

**Compliance Test:**
PASS: `let f: fn(i64) -> str = int_to_str`

---

### RULE TYP-070: Generic Type Parameters
**Status:** STABLE | **Layer:** TYP | **Depends on:** TYP-040–TYP-060

A type may be parameterized with one or more type variables.
Type variables are resolved at use sites (monomorphization).

```ebnf
TypeParams   ::= '<' Ident (',' Ident)* '>'
GenericType  ::= Ident TypeParams
```

**Compliance Test:**
PASS:
```candor
struct Pair<T> { first: T, second: T }
let p = Pair<i64>{ first: 1, second: 2 }
```

---

## TYPE COMPATIBILITY

### RULE TYP-080: Explicit Coercion Only
**Status:** FROZEN | **Layer:** TYP | **Depends on:** AXIOM-001, TYP-011

There is no implicit type coercion between distinct types.
Every type conversion requires an explicit `as` cast (EVAL-090)
or an explicit conversion function.

The only implicit behavior: integer literals and float literals
without type annotations adopt the type required by context.
If context provides no type, defaults apply (TYP-004 for integers, TYP-013 for floats).

**Compliance Test:**
PASS: `let x: i32 = 5` → literal 5 adopts type i32
FAIL: `fn f(x: i32) -> unit {} f(5i64_val)` where `5i64_val: i64` → mismatch

---

### RULE TYP-090: Numeric Type Ordering
**Status:** STABLE | **Layer:** TYP | **Depends on:** TYP-001–TYP-015

For cast validation and overflow checking, types are ordered by their
value range (not bit width alone). Signed and unsigned are separate orderings.

Signed: `i8 < i16 < i32 < i64 < i128`
Unsigned: `u8 < u16 < u32 < u64 < u128`
Float: `f16 < bf16 < f32 < f64`

A narrowing cast (from wider to narrower type) is always permitted syntactically
but may produce a warning if the source value cannot be statically proven to fit.

---

### RULE TYP-100: Type Mismatch Error
**Status:** FROZEN | **Layer:** TYP | **Depends on:** AXIOM-001

When an expression of type `T` is used where type `U` is required
and `T ≠ U` and no implicit adoption applies (TYP-080),
the compiler emits:

```
cannot use TYPE_T as TYPE_U
```

**On Violation:** `cannot use X as Y`

---

### RULE TYP-101: Type Inference Failure
**Status:** FROZEN | **Layer:** TYP | **Depends on:** AXIOM-001

When the type of an expression cannot be determined from context alone,
and no annotation is provided, the compiler emits:

```
type annotation required: cannot infer type of "X"
```

**On Violation:** `type annotation required: cannot infer type of "X"`

---

*End of L3-TYPES.md*
