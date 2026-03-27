# LAYER 6: FUNCTIONS
**Status:** FROZEN (declarations, calls), STABLE (generics, closures, extern)
**Depends on:** L5-STATEMENTS.md, L3-TYPES.md

---

## PURPOSE

Functions are the primary unit of computation in Candor.
This layer defines how functions are declared, called, and typed.
Effects annotations are defined here syntactically; enforcement is in L7-EFFECTS.md.

---

### RULE FN-010: Function Declaration
**Status:** FROZEN
**Layer:** FN
**Depends on:** SYN-045, TYP-060, STMT-001

A function declaration introduces a named callable with an explicit signature and a body.

**Grammar:**
```ebnf
FnDecl        ::= 'fn' Ident TypeParamList? '(' ParamList ')' '->' Type
                  EffectsAnnotation? ContractClauses? Block
TypeParamList ::= '<' TypeParamSpec (',' TypeParamSpec)* '>'
TypeParamSpec ::= Ident (':' Ident ('+' Ident)*)?
ParamList     ::= (Param (',' Param)* ','?)?
Param         ::= Ident ':' Type
```

Every function declaration requires:
- A name (identifier)
- A parenthesized parameter list (may be empty)
- An explicit return type following `->`
- A block body `{ ... }`

The return type is always explicit. There is no return-type inference.

A function with no `return` statement in any code path implicitly returns `unit`.
If the declared return type is not `unit` and any code path reaches the end of
the block without a `return`, this is a type error (see FN-040).

**Type Rule:**
```
Γ, x1:T1, ..., xn:Tn ⊢ Body : unit
RetType = unit
──────────────────────────────────────────────────────────────────
Γ ⊢ fn Name(x1:T1, ..., xn:Tn) -> RetType Body : fn(T1,...,Tn)->RetType
```

**Invariant:** `len(TypeParams) == 0` for non-generic functions (see FN-060 for generic).

**Compliance Test:**
```candor
// PASS: minimal function
fn greet(name: str) -> str {
    return "hello"
}
```
```candor
// PASS: zero-parameter, unit return
fn init() -> unit {
}
```
```candor
// PASS: explicit unit return value
fn noop() -> unit {
    return unit
}
```
```candor
// FAIL: missing '->' and return type
fn bad(x: i64) {
    return x
}
```

**On Violation:** `expected '->', got ...`

---

### RULE FN-020: Parameter List
**Status:** FROZEN
**Layer:** FN
**Depends on:** FN-010, TYP-001–TYP-060

A parameter list is a sequence of zero or more `name: Type` pairs, separated by commas.
A trailing comma after the last parameter is permitted.

Each parameter introduces a name binding in the function body scope.
Parameter names must be distinct within the same parameter list.

There are no default parameter values at this layer.
There are no variadic parameters at this layer.

The parameter list determines the input types of the function's `fn(T...) -> U` type (TYP-060).

**Grammar:**
```ebnf
ParamList ::= (Param (',' Param)* ','?)?
Param     ::= Ident ':' Type
```

**Type Rule:**
```
For each Param(xi, Ti):  Γ ⊢ Ti : Type
─────────────────────────────────────────
Γ ⊢ ParamList : [(x1,T1), ..., (xn,Tn)]
```

**Invariant:** All parameter names are pairwise distinct within the list.

**Compliance Test:**
```candor
// PASS: two parameters
fn add(a: i64, b: i64) -> i64 {
    return a + b
}
```
```candor
// PASS: trailing comma
fn f(x: i64, y: str,) -> unit {
}
```
```candor
// PASS: empty parameter list
fn nothing() -> unit {
}
```
```candor
// FAIL: parameter without type annotation
fn bad(x) -> unit {
}
```

**On Violation:** `expected ':', got ...`

---

### RULE FN-030: Calling Convention and Call Expressions
**Status:** FROZEN
**Layer:** FN
**Depends on:** FN-010, FN-020, TYP-060, TYP-100

All Candor-defined functions use the standard calling convention for the target platform.
Parameters are passed by value (copy for scalar types; by pointer for owned struct types
as determined by the code generator).

A function call expression applies a callable to an argument list.

**Grammar:**
```ebnf
CallExpr ::= Expr '(' ArgList ')'
ArgList  ::= (Expr (',' Expr)* ','?)?
```

The callee expression must have function type `fn(T1, ..., Tn) -> U`.
The argument count must match the parameter count exactly.
Each argument type must match the corresponding parameter type exactly
(after the literal-adoption rule in TYP-080; no other implicit coercion applies).

The type of a call expression is the return type `U` of the callee's function type.

**Type Rule:**
```
Γ ⊢ f : fn(T1,...,Tn)->U
Γ ⊢ a1 : T1   ...   Γ ⊢ an : Tn
──────────────────────────────────
Γ ⊢ f(a1,...,an) : U
```

**Invariant:** `len(args) == len(params)` and `type(argi) == parami` for all i.

**Compliance Test:**
```candor
// PASS
fn double(x: i64) -> i64 {
    return x * 2
}
fn main() -> unit {
    let y: i64 = double(21)
}
```
```candor
// FAIL: wrong argument count
fn f(x: i64) -> unit {}
fn main() -> unit {
    f(1, 2)
}
```
```candor
// FAIL: type mismatch on argument
fn f(x: i64) -> unit {}
fn main() -> unit {
    let s: str = "hi"
    f(s)
}
```
```candor
// FAIL: calling a non-function
fn main() -> unit {
    let x: i64 = 5
    x(1)
}
```

**On Violation:**
- Argument count: `argument count mismatch: expected N, got M`
- Argument type: `argument N: cannot use X as Y`
- Non-function callee: `cannot call non-function type X`

---

### RULE FN-040: Return Type Checking
**Status:** FROZEN
**Layer:** FN
**Depends on:** FN-010, TYP-025, TYP-100

Every `return` statement in a function body must return a value whose type matches
the function's declared return type.

A bare `return` (no value) is permitted only when the declared return type is `unit`.

If the declared return type is `unit` and the function body has no `return` statement,
the function implicitly returns `unit` at the end of the block.

If the declared return type is not `unit` and a code path reaches the end of the
function body without executing a `return`, the behavior is implicitly as if
`return unit` were appended — which is a type error because `unit != RetType`.

**Type Rule:**
```
Γ ⊢ v : T    T == RetType
──────────────────────────────────────────────────
Γ ⊢ return v : unit  (within function of RetType)

RetType == unit
──────────────────────────────────────────────────
Γ ⊢ return : unit

RetType != unit   /\ bare return used
──────────────────────────────────────────────────
ERROR: bare return in function returning RetType
```

**Invariant:** For every `return e` statement, `type(e) == declared RetType`.

**Compliance Test:**
```candor
// PASS: return type matches
fn square(x: i64) -> i64 {
    return x * x
}
```
```candor
// PASS: bare return in unit function
fn side_effect() -> unit {
    return
}
```
```candor
// PASS: implicit unit return
fn implicit() -> unit {
}
```
```candor
// FAIL: return type mismatch
fn bad() -> i64 {
    return "hello"
}
```
```candor
// FAIL: bare return in non-unit function
fn bad() -> i64 {
    return
}
```

**On Violation:**
- Type mismatch: `return type mismatch: got X, expected Y`
- Bare return in non-unit function: `bare return in function returning X`

---

### RULE FN-050: Closures (Anonymous Functions)
**Status:** STABLE
**Layer:** FN
**Depends on:** FN-010, FN-020, TYP-060

A closure is an anonymous function expression.
It has the same syntactic structure as a named function declaration
but without a name, and it appears in expression position.

**Grammar:**
```ebnf
LambdaExpr ::= 'fn' '(' ParamList ')' '->' Type Block
```

A closure has type `fn(T1,...,Tn) -> U` as defined by TYP-060.

A closure may be assigned to a `let` binding or passed as a function argument.

The closure's body is type-checked immediately at the point of definition.
The body is checked in a new scope whose parent is the enclosing scope,
giving the closure access to all names visible at the point of definition (captures).

**Type Rule:**
```
Γ, x1:T1,...,xn:Tn ⊢ Body : unit
──────────────────────────────────────────────────
Γ ⊢ fn(x1:T1,...,xn:Tn)->U Body : fn(T1,...,Tn)->U
```

**Invariant:** A closure's return type annotation is always explicit (same as FN-010).

**Compliance Test:**
```candor
// PASS: closure assigned to binding
fn main() -> unit {
    let double: fn(i64) -> i64 = fn(x: i64) -> i64 { return x * 2 }
    let y: i64 = double(5)
}
```
```candor
// PASS: closure passed as argument
fn apply(f: fn(i64) -> i64, x: i64) -> i64 {
    return f(x)
}
fn main() -> unit {
    let result: i64 = apply(fn(n: i64) -> i64 { return n + 1 }, 10)
}
```
```candor
// FAIL: missing return type on closure
fn main() -> unit {
    let f = fn(x: i64) { return x }
}
```

**On Violation:** `expected '->', got ...`

---

### RULE FN-051: Lambda Capture Rules
**Status:** STABLE
**Layer:** FN
**Depends on:** FN-050

A closure may reference any name from an enclosing scope that is visible at
the point of the closure's definition. Such names are said to be *captured*.

Immutable bindings from the enclosing scope are captured by copy.
Mutable bindings (declared `mut`) from the enclosing scope are captured by
reference (pointer to the outer binding), so that mutations inside the closure
are visible in the outer scope.

A captured name is immutable inside the closure unless the original outer
binding was declared `mut`.

Top-level function names are not captured; they are resolved globally.

**Invariant:** A mutable outer variable mutated inside a closure reflects the
mutation in the outer scope after the closure executes.

**Compliance Test:**
```candor
// PASS: capturing an immutable binding
fn main() -> unit {
    let base: i64 = 10
    let add_base: fn(i64) -> i64 = fn(x: i64) -> i64 { return x + base }
    let r: i64 = add_base(5)
}
```
```candor
// PASS: capturing a mutable binding
fn main() -> unit {
    let mut counter: i64 = 0
    let inc: fn() -> unit = fn() -> unit { counter = counter + 1 }
    inc()
}
```
```candor
// FAIL: assigning to an immutable capture
fn main() -> unit {
    let x: i64 = 5
    let f: fn() -> unit = fn() -> unit { x = 10 }
}
```

**On Violation:** `cannot assign to immutable variable "x"`

---

### RULE FN-060: Generic Functions
**Status:** STABLE
**Layer:** FN
**Depends on:** FN-010, FN-020, TYP-070

A function may be parameterized over one or more type variables by placing a
`TypeParamList` between the function name and its parameter list.

**Grammar:**
```ebnf
GenericFnDecl ::= 'fn' Ident '<' TypeParamSpec (',' TypeParamSpec)* '>'
                  '(' ParamList ')' '->' Type EffectsAnnotation? ContractClauses? Block
TypeParamSpec ::= Ident (':' Ident ('+' Ident)*)?
```

Generic functions are not compiled directly. They are instantiated at each call
site by *monomorphization*: the type checker infers concrete types for each type
parameter by unifying the declared parameter types against the actual argument
types at the call site.

All type parameters must be inferrable from the argument types at the call site.
A type parameter that cannot be inferred from the arguments is a compile-time error.

Trait bounds on type parameters (e.g., `<T: Display>`) are parsed here but their
enforcement is specified in L10-TRAITS.md.

At this layer (no bounds), type parameters are unconstrained type variables.

**Type Rule:**
```
gDecl.TypeParams = [T1,...,Tk]
subst = unify(gDecl.Params, argTypes)
dom(subst) == {T1,...,Tk}
retType = apply(subst, gDecl.RetType)
──────────────────────────────────────
Γ ⊢ g(a1,...,an) : retType
```

**Invariant:** Every type parameter appears at least once in `ParamList` so that
monomorphization can infer all type arguments from call-site argument types.

**Compliance Test:**
```candor
// PASS: generic identity function
fn identity<T>(x: T) -> T {
    return x
}
fn main() -> unit {
    let n: i64 = identity(42)
    let s: str = identity("hello")
}
```
```candor
// PASS: two type parameters
fn pair_first<A, B>(a: A, b: B) -> A {
    return a
}
```
```candor
// FAIL: type parameter cannot be inferred (does not appear in params)
fn make<T>() -> T {
    return unit as T
}
fn main() -> unit {
    let x = make()
}
```

**On Violation:**
- Inference failure: `cannot infer type parameter "T" for FnName`
- Argument count: `FnName: argument count mismatch: expected N, got M`

---

### RULE FN-070: `extern fn` Declaration
**Status:** STABLE
**Layer:** FN
**Depends on:** FN-010, FN-020, TYP-060

An `extern fn` declaration introduces a function symbol provided by an external
C-ABI library. The symbol name is used directly as the C linkage name.

**Grammar:**
```ebnf
ExternFnDecl ::= 'extern' 'fn' Ident '(' ParamList ')' '->' Type EffectsAnnotation?
```

An `extern fn` has no body. The type checker registers it as a callable with the
declared signature. At call sites it is treated identically to a Candor function
of the same type (FN-030 applies).

An `extern fn` may carry an optional effects annotation. If none is supplied,
the function is treated as having no declared effects (trusted by default —
this is the AXIOM-002 exception for FFI).

**Type Rule:**
```
──────────────────────────────────────────────────────────────────
Γ ⊢ extern fn Name(x1:T1,...,xn:Tn)->U : fn(T1,...,Tn)->U
```

**Invariant:** An `extern fn` declaration has no body block `{ ... }`.

**Compliance Test:**
```candor
// PASS: extern C function with no effects annotation
extern fn abs(x: i32) -> i32
```
```candor
// PASS: extern with effects annotation
extern fn malloc(size: i64) -> i64 effects(alloc)
```
```candor
// PASS: calling an extern function
extern fn abs(x: i32) -> i32
fn main() -> unit {
    let y: i32 = abs(-5 as i32)
}
```
```candor
// FAIL: extern fn with a body is not valid syntax
extern fn bad(x: i64) -> i64 {
    return x
}
```

**On Violation:** `expected declaration (fn, struct, enum, module, use, extern, const, impl, trait, or cap), got ...`

---

### RULE FN-080: Method Declarations (`impl` Block)
**Status:** STABLE
**Layer:** FN
**Depends on:** FN-010, FN-020, FN-030, TYP-040

Methods are `fn` declarations inside an `impl TypeName { ... }` block.
They are associated with the named struct type `TypeName`.

**Grammar:**
```ebnf
ImplDecl   ::= 'impl' Ident '{' ImplMethod* '}'
ImplMethod ::= 'fn' Ident TypeParamList? '(' ParamList ')' '->' Type Block
```

The first parameter of a method is conventionally named `self` and acts as the
receiver. Its declared type must be the struct type the `impl` block is for
(or a reference variant such as `ref<TypeName>`).

Method calls use dot syntax: `receiver.method(args)`.
Dispatch is static: the method is selected by the compile-time type of the
receiver. There is no vtable or dynamic dispatch at this layer.

Internally, the type checker registers method `m` on type `T` under the mangled
name `T_m` and includes the receiver as the first parameter in the signature.
The call `v.method(args)` requires `len(args) + 1 == len(sig.Params)` (the `+1`
accounts for the receiver/self parameter).

**Type Rule:**
```
TypeName is a declared struct
sig = fnSigs["TypeName_method"]   sig.Params = [TypeName, T1,...,Tn]
Γ ⊢ recv : TypeName
Γ ⊢ a1:T1 ... Γ ⊢ an:Tn
──────────────────────────────────────────────
Γ ⊢ recv.method(a1,...,an) : sig.Ret
```

**Invariant:** For each method `m` in `impl TypeName`, the mangled name `TypeName_m`
is registered in the type checker's function signature table.

**Compliance Test:**
```candor
// PASS: struct with impl methods
struct Counter { value: i64 }

impl Counter {
    fn increment(self: Counter) -> Counter {
        return Counter{ value: self.value + 1 }
    }
    fn get(self: Counter) -> i64 {
        return self.value
    }
}

fn main() -> unit {
    let c = Counter{ value: 0 }
    let c2 = c.increment()
    let n: i64 = c2.get()
}
```
```candor
// FAIL: wrong argument count for method call
struct Point { x: i64, y: i64 }
impl Point {
    fn translate(self: Point, dx: i64, dy: i64) -> Point {
        return Point{ x: self.x + dx, y: self.y + dy }
    }
}
fn main() -> unit {
    let p = Point{ x: 0, y: 0 }
    let q = p.translate(1)
}
```

**On Violation:** `method TypeName.method expects N argument(s), got M`

---

### RULE FN-090: Recursive Functions
**Status:** STABLE
**Layer:** FN
**Depends on:** FN-010, FN-030

A function may call itself directly (direct recursion) or call functions that
eventually call it back (mutual recursion). No special annotation is required.

Function signatures are collected in a pre-pass before bodies are checked, so
each function's own name is visible within its body with the correct type.

The type checker does not verify that recursive functions terminate.
Termination checking is undecidable in general (AXIOM-005).

**Invariant:** A recursive call satisfies FN-030: argument count and argument
types must match the function's own declared signature.

**Compliance Test:**
```candor
// PASS: recursive factorial
fn fact(n: i64) -> i64 {
    if n <= 1 {
        return 1
    }
    return n * fact(n - 1)
}
```
```candor
// PASS: mutually recursive functions
fn is_even(n: i64) -> bool {
    if n == 0 { return true }
    return is_odd(n - 1)
}
fn is_odd(n: i64) -> bool {
    if n == 0 { return false }
    return is_even(n - 1)
}
```
```candor
// FAIL: recursive call with wrong argument type
fn bad(n: i64) -> i64 {
    return bad("hello")
}
```

**On Violation:** Recursive calls that violate FN-030 produce the same errors as
any other call: `argument N: cannot use X as Y` or
`argument count mismatch: expected N, got M`.

---

*End of L6-FUNCTIONS.md*
