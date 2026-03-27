# LAYER 2: SYNTAX
**Status:** FROZEN
**Depends on:** L0-AXIOMS.md, L1-LEXER.md

---

## PURPOSE

This layer defines the grammatical structure of Candor programs.
The grammar is expressed as EBNF. Every production corresponds to a
function in the reference parser. These rules are FROZEN: a program
accepted by this grammar is always accepted.

---

### RULE SYN-010: Program Structure (File)

**Status:** FROZEN
**Layer:** SYN
**Depends on:** LEX-050

A Candor source file is a flat sequence of top-level declarations.
The optional `module` declaration, if present, must appear before any `use`
declarations or other declarations. `use` declarations conventionally follow
the `module` declaration but may appear anywhere in the top-level sequence.
File-scope directives (`#intent`, `#mcp_tool`, `#test`, `#c_header`, etc.)
may appear before any `fn` or `struct` declaration to decorate it, or as
standalone declarations (`#c_header`). The parser attaches decoration directives
to the immediately following declaration.

**Grammar:**

```ebnf
File         ::= Directive* ModuleDecl? ( Directive | Decl )*
ModuleDecl   ::= 'module' Ident
Decl         ::= FnDecl
               | StructDecl
               | EnumDecl
               | ConstDecl
               | ExternFnDecl
               | ImplDecl
               | ImplForDecl
               | TraitDecl
               | CapabilityDecl
               | UseDecl
               | CHeaderDecl
```

**Invariant:** A file contains zero or more declarations. An empty file is valid.

**Compliance Test:**
PASS:
```
module mymod

use std::io

fn main() -> unit { }
```
FAIL: A file containing only `42` (bare expression at file scope is not a declaration)

**On Violation:**
`expected declaration (fn, struct, enum, module, use, extern, const, impl, trait, or cap), got <token>`

---

### RULE SYN-020: Operator Precedence

**Status:** FROZEN
**Layer:** SYN
**Depends on:** LEX-010, LEX-050

Expressions are parsed using a precedence ladder. The table below lists
levels from lowest to highest binding power. Within a level, binary
operators are left-associative. Unary operators are right-associative.

| Level | Operators | Associativity | Parser function |
|-------|-----------|---------------|-----------------|
| 1 (lowest) | `or` | left | `parseOrExpr` |
| 2 | `and` | left | `parseAndExpr` |
| 3 | `==` `!=` `<` `>` `<=` `>=` | left | `parseCmpExpr` |
| 4 | `+` `-` | left | `parseAddExpr` |
| 5 | `*` `/` `%` | left | `parseMulExpr` |
| 6 | `!` `not` `-` (unary) `&` `*` (deref) | right | `parseUnaryExpr` |
| 7 (highest) | `()` `[]` `.field` `.0` `as T` `must{}` struct-literal `{}` | left | `parsePostfixExpr` |

Notes:
- The `as` cast operator is a postfix operator at level 7 (tighter than arithmetic).
- Struct literal syntax (`Name { field: val }`) is only recognized when the receiver
  is a PascalCase identifier, preventing ambiguity with block statements.
- `must { arm => body }` is a postfix operator; the scrutinee is the expression to its left.
- `match expr { arm => body }` is a primary expression, not a postfix operator.

**Grammar:**

```ebnf
Expr         ::= OrExpr
OrExpr       ::= AndExpr ( 'or' AndExpr )*
AndExpr      ::= CmpExpr ( 'and' CmpExpr )*
CmpExpr      ::= AddExpr ( CmpOp AddExpr )*
CmpOp        ::= '==' | '!=' | '<' | '>' | '<=' | '>='
AddExpr      ::= MulExpr ( ('+' | '-') MulExpr )*
MulExpr      ::= UnaryExpr ( ('*' | '/' | '%') UnaryExpr )*
UnaryExpr    ::= ('!' | 'not' | '-' | '&' | '*') UnaryExpr
               | PostfixExpr
PostfixExpr  ::= PrimaryExpr PostfixOp*
PostfixOp    ::= '(' ArgList ')'
               | '[' Expr ']'
               | '.' Ident
               | '.' IntLit
               | 'as' Type
               | 'must' '{' MustArmList '}'
               | '{' FieldInitList '}'    (* only when receiver is PascalCase Ident *)
```

**Invariant:** Every valid expression reduces to exactly one parse tree (NP-5).

**Compliance Test:**
PASS: `1 + 2 * 3` parses as `1 + (2 * 3)` (multiplication binds tighter than addition)
PASS: `a or b and c` parses as `a or (b and c)` (and binds tighter than or)
PASS: `!x.field` parses as `!(x.field)` (postfix binds tighter than unary)
FAIL: `+x` — `+` is not a unary operator (use unary `-` instead)

---

### RULE SYN-030: Top-Level Declarations

**Status:** FROZEN
**Layer:** SYN
**Depends on:** SYN-010, LEX-050

The complete set of syntactic forms allowed at the top level of a file.
Each starts with a distinct leading token, giving the parser a one-token
look-ahead to dispatch without ambiguity.

| Leading token | Declaration kind |
|---------------|-----------------|
| `fn` | Function declaration |
| `struct` | Struct type declaration |
| `enum` | Enum type declaration |
| `const` | Constant declaration |
| `extern` | External function declaration |
| `impl` | Impl block (methods or trait impl) |
| `trait` | Trait declaration |
| `cap` | Capability declaration |
| `module` | Module declaration |
| `use` | Use import declaration |
| `#c_header` | C header interop directive |
| `TokDirective` | Decoration directive (attached to next `fn`/`struct`) |

**Grammar:**

```ebnf
Decl         ::= FnDecl
               | StructDecl
               | EnumDecl
               | ConstDecl
               | ExternFnDecl
               | ImplDecl
               | ImplForDecl
               | TraitDecl
               | CapabilityDecl
               | UseDecl
               | CHeaderDecl
```

**Invariant:** No construct other than those listed may appear at the top level.

**Compliance Test:**
PASS: `const MAX: i64 = 100` at file scope
FAIL: `let x = 5` at file scope — `let` is a statement, not a declaration

**On Violation:**
`expected declaration (fn, struct, enum, module, use, extern, const, impl, trait, or cap), got <token>`

---

### RULE SYN-040: Function Declaration

**Status:** FROZEN
**Layer:** SYN
**Depends on:** SYN-030, SYN-060, SYN-065, SYN-075, SYN-080, LEX-040, LEX-050

A function declaration introduces a named, callable unit. Generic type
parameters, an effects annotation, and contract clauses are all optional.
The return type annotation is mandatory (Candor does not infer return types
from the declaration syntax).

**Grammar:**

```ebnf
FnDecl           ::= Directive* 'fn' Ident GenericParams? '(' ParamList ')' '->' Type
                     EffectsAnnotation? ContractClauses? Block

GenericParams    ::= '<' TypeParamList '>'
TypeParamList    ::= TypeParam ( ',' TypeParam )*
TypeParam        ::= Ident ( ':' TraitBound )?
TraitBound       ::= Ident ( '+' Ident )*

ParamList        ::= ( Param ( ',' Param )* ','? )?
Param            ::= Ident ':' Type

ContractClauses  ::= ContractClause*
ContractClause   ::= ( 'requires' | 'ensures' ) Expr
```

Notes:
- Trailing commas are permitted in parameter lists.
- `self` is a valid parameter name (it is not a reserved keyword at the
  syntax layer; semantic rules govern its meaning in `impl` bodies).
- Effects and contract clauses may appear in any order after the return type
  but before the opening `{` of the body block.

**Invariant:** A function declaration always has a body block. Bodyless
function signatures are only valid inside `trait` declarations (SYN-100).

**Compliance Test:**
PASS:
```
fn add(x: i64, y: i64) -> i64 { x + y }
```
PASS (generic with bounds):
```
fn identity<T: Display>(v: T) -> T { v }
```
PASS (with effects and contracts):
```
fn read_file(path: str) -> str effects(io) requires path != "" { }
```
FAIL: `fn foo(x: i64) { }` — missing `->` return type annotation
FAIL: `fn foo() -> i64` — missing body block

**On Violation:**
`expected ->, got <token>`
`expected {, got <token>`

---

### RULE SYN-045: Struct Declaration

**Status:** FROZEN
**Layer:** SYN
**Depends on:** SYN-030, SYN-075, LEX-040, LEX-050

A struct declaration introduces a named product type with named fields.
Fields are separated by commas; a trailing comma is permitted.
Struct declarations may be preceded by decoration directives
(e.g., `#export_json`).

**Grammar:**

```ebnf
StructDecl   ::= Directive* 'struct' Ident '{' FieldList '}'
FieldList    ::= ( Field ( ',' Field )* ','? )?
Field        ::= Ident ':' Type
```

Notes:
- Structs do not currently support generic type parameters at the declaration
  site (generic types are expressed through `impl` and trait bounds).
- An empty struct `struct Unit {}` is valid and has no fields.

**Invariant:** All field names within a single struct declaration must be distinct.
This is enforced at the type-checking layer (TYP), not here.

**Compliance Test:**
PASS:
```
struct Point {
    x: f64,
    y: f64,
}
```
PASS (empty struct):
```
struct Token {}
```
FAIL: `struct Foo { x: i64 y: i64 }` — missing comma between fields

**On Violation:**
`expected , or }, got <token>`

---

### RULE SYN-050: Enum Declaration

**Status:** FROZEN
**Layer:** SYN
**Depends on:** SYN-030, SYN-075, LEX-040, LEX-050

An enum declaration introduces a named sum type. Each variant is either
a unit variant (no payload) or a tuple variant (one or more typed fields
in parentheses). Variants are separated by commas; a trailing comma
is permitted.

**Grammar:**

```ebnf
EnumDecl       ::= 'enum' Ident '{' VariantList '}'
VariantList    ::= ( EnumVariant ( ',' EnumVariant )* ','? )?
EnumVariant    ::= Ident ( '(' TypeList ')' )?
TypeList       ::= Type ( ',' Type )* ','?
```

Notes:
- Variants with a single type `Foo(T)` are single-field tuple variants.
- Variants with multiple types `Foo(T, U)` are multi-field tuple variants.
- The standard library `option<T>` and `result<T, E>` types are expressed
  as enum types; their constructor names `some`, `none`, `ok`, `err` are
  reserved keywords (LEX-050) that can appear as expressions.

**Invariant:** All variant names within a single enum must be distinct.
Enforced at the type-checking layer.

**Compliance Test:**
PASS:
```
enum Shape {
    Circle(f64),
    Rectangle(f64, f64),
    Point,
}
```
FAIL: `enum Foo { Bar Baz }` — missing comma between variants

**On Violation:**
`expected , or }, got <token>`

---

### RULE SYN-055: `impl` Block

**Status:** STABLE
**Layer:** SYN
**Depends on:** SYN-030, SYN-040, SYN-075, LEX-040, LEX-050

An `impl` block attaches methods to a named type. Two forms are recognized:

1. **Type impl:** `impl TypeName { methods }` — adds methods directly to a type.
2. **Trait impl:** `impl TraitName for TypeName { methods }` — implements a named
   trait for a type.

Methods inside an `impl` block are parsed by the same grammar as function
declarations (SYN-040) except that the `fn` keyword is consumed by
`parseImplMethod` rather than `parseFnDecl`, and effects/contract clauses
are not parsed for impl methods (the body block follows immediately after
the return type).

**Grammar:**

```ebnf
ImplDecl       ::= 'impl' Ident '{' ImplMethod* '}'
ImplForDecl    ::= 'impl' Ident 'for' Ident '{' ImplMethod* '}'
ImplMethod     ::= 'fn' Ident GenericParams? '(' ParamList ')' '->' Type Block
```

Notes:
- The first `Ident` after `impl` is always the trait name in `impl Trait for Type`
  or the type name in `impl Type { }`.
- The parser distinguishes the two forms by one-token look-ahead after the
  first identifier: `for` triggers the trait-impl path.
- Impl methods do not currently support effects annotations or contract
  clauses in the `impl` body (those are permitted on top-level `fn` declarations).

**Invariant:** Every method in an `impl` block must have a body block.

**Compliance Test:**
PASS:
```
impl Point {
    fn distance(self: Point, other: Point) -> f64 { 0.0 }
}
```
PASS (trait impl):
```
impl Display for Point {
    fn fmt(self: Point) -> str { "" }
}
```
FAIL:
```
impl Point {
    fn area() -> f64  ## missing body
}
```

**On Violation:**
`expected {, got <token>`

---

### RULE SYN-060: Effects Annotation Syntax

**Status:** STABLE
**Layer:** SYN
**Depends on:** SYN-040, LEX-050

An effects annotation immediately follows the return type in a function
declaration (before any contract clauses). Three syntactic forms exist:

| Form | Meaning |
|------|---------|
| `pure` | Function has no effects |
| `effects []` | Syntactic sugar for `pure` |
| `effects(name, ...)` | Function declares the named effects |
| `cap(name)` | Function requires the named capability |

The annotation is optional. If absent, the function's effects are unchecked
(neither asserted pure nor given an effects set).

**Grammar:**

```ebnf
EffectsAnnotation ::= 'pure'
                    | 'effects' '[' ']'
                    | 'effects' '(' EffectNameList ')'
                    | 'cap' '(' Ident ')'
EffectNameList    ::= Ident ( ',' Ident )* ','?
```

Notes:
- `effects()` with an empty list is a parse error; use `effects []` or `pure`
  for the no-effects case.
- Effect names are plain identifiers (e.g., `io`, `net`, `rand`).
- `cap(Name)` references a named capability declared with `cap Name`.

**Invariant:** `effects(...)` requires at least one effect name.

**Compliance Test:**
PASS: `fn f() -> unit pure { }`
PASS: `fn f() -> unit effects(io) { }`
PASS: `fn f() -> unit effects [] { }`
PASS: `fn f() -> unit cap(Admin) { }`
FAIL: `fn f() -> unit effects() { }` — empty effects list is a parse error

**On Violation:**
`effects() requires at least one effect name; use effects [] for pure`

---

### RULE SYN-065: Contract Clauses

**Status:** STABLE
**Layer:** SYN
**Depends on:** SYN-040, SYN-060, LEX-050

Contract clauses follow the optional effects annotation in a function
declaration. Any number of `requires` and `ensures` clauses may appear,
in any order, before the opening brace of the function body.

**Grammar:**

```ebnf
ContractClauses  ::= ContractClause*
ContractClause   ::= 'requires' Expr
                   | 'ensures' Expr
```

Notes:
- The expression in a `requires` clause is the precondition on entry.
- The expression in an `ensures` clause is the postcondition on exit.
- Inside an `ensures` expression, `old(expr)` refers to the value of `expr`
  at function entry (SYN-095).
- Contract expressions are boolean-typed (enforced at TYP layer).

**Invariant:** Contract clauses may only appear in function declarations,
not inside `impl` methods or lambda expressions at this layer.

**Compliance Test:**
PASS:
```
fn divide(x: i64, y: i64) -> i64
    requires y != 0
    ensures result * y == x
{ x / y }
```
PASS (multiple clauses):
```
fn push(v: vec<i64>, x: i64) -> vec<i64>
    requires x >= 0
    ensures result.len() == v.len() + 1
{ }
```
FAIL: `requires` appearing inside a block statement (not in a declaration position)

**On Violation:**
`expected {, got requires` (if contract keyword appears where a block is expected)

---

### RULE SYN-070: Pattern Syntax

**Status:** FROZEN
**Layer:** SYN
**Depends on:** SYN-020, LEX-010, LEX-040, LEX-050

Patterns appear in `match` and `must` arms. The parser does not maintain a
separate `parsePattern` function; instead, arm patterns are parsed as
expressions (`parseExpr`) with the addition of `_` (wildcard) as a valid
expression token. The semantic layer distinguishes pattern-position
identifiers from expression-position identifiers.

Recognized pattern forms (as parsed):

| Pattern form | Parsed as |
|---|---|
| `_` | `IdentExpr` with lexeme `_` (wildcard) |
| Literal `42`, `"s"`, `true` | Literal expression |
| `Ident` | `IdentExpr` — binds a variable or matches an enum unit variant |
| `Enum::Variant` | `PathExpr` |
| `some(P)` | `CallExpr` with callee `some` |
| `none` | `IdentExpr` with keyword `none` |
| `ok(P)` | `CallExpr` with callee `ok` |
| `err(P)` | `CallExpr` with callee `err` |

**Grammar:**

```ebnf
MatchExpr    ::= 'match' Expr '{' ArmList '}'
MustExpr     ::= PostfixExpr 'must' '{' ArmList '}'
ArmList      ::= ( Arm ( ',' Arm )* ','? )?
Arm          ::= Pattern '=>' ArmBody
Pattern      ::= '_'
               | Literal
               | Ident
               | Ident '::' Ident
               | Ident '(' ArgList ')'
ArmBody      ::= Expr
               | 'return' Expr
               | 'break'
               | '{' Stmt* '}'
```

Notes:
- The `match` keyword introduces a match expression; the scrutinee is a
  full expression.
- `must` is a postfix operator: `expr must { arm => body }`.
- Arms are separated by optional commas. A trailing comma is allowed.
- `ArmBody` may be a block (`{ stmts }`) for multi-statement arms.
- The `=>` token (`TokFatArrow`) separates pattern from body.

**Compliance Test:**
PASS:
```
match x {
    0 => "zero",
    n => "other",
}
```
PASS (must):
```
result must { ok(v) => v, err(e) => return e }
```
FAIL: `match x { }` with no arms — syntactically valid (zero arms), but
type-checking may reject non-exhaustive matches

**On Violation:**
`expected =>, got <token>`

---

### RULE SYN-075: Type Annotation Syntax

**Status:** FROZEN
**Layer:** SYN
**Depends on:** LEX-040, LEX-050

A type expression denotes a type. Type expressions appear in function
parameter lists, return type positions, `let` type annotations, `const`
declarations, struct fields, and enum variant payloads.

**Grammar:**

```ebnf
Type           ::= NamedType
               | GenericType
               | QualifiedType
               | FnType
               | TupleType

NamedType      ::= Ident
QualifiedType  ::= Ident '.' Ident              (* module-qualified: mod.TypeName *)
GenericType    ::= Ident '<' TypeArgList '>'
TypeArgList    ::= Type ( ',' Type )* ','?
FnType         ::= 'fn' '(' TypeList ')' '->' Type
TypeList       ::= ( Type ( ',' Type )* ','? )?
TupleType      ::= '(' Type ',' Type ( ',' Type )* ','? ')'
```

Notes:
- A tuple type requires at least two elements; `(T)` is a parenthesized type,
  not a one-tuple. There is no unit tuple syntax at this layer; use the `unit`
  keyword as a named type.
- Generic type arguments use `<` and `>` as delimiters (same tokens as
  comparison operators); the parser disambiguates by context (type position vs.
  expression position).
- `secret` and `cap` are keywords that are also valid as generic type constructors:
  `secret<T>`, `cap<Admin>`.
- Built-in generic type names include `vec`, `option`, `result`, `map`, `ring`,
  `tensor`, `mmap`; these are parsed identically to user-defined generic types.

**Invariant:** Every position that syntactically requires a type must produce
a well-formed `TypeExpr` node; if not, parsing halts with an error.

**Compliance Test:**
PASS: `i64` — simple named type
PASS: `vec<i64>` — generic type
PASS: `option<str>` — generic type
PASS: `fn(i64, i64) -> bool` — function type
PASS: `(i64, str)` — tuple type
PASS: `std.Point` — module-qualified type
FAIL: `(i64)` — single-element tuple is not a valid type (parsed as parenthesized `i64`)
FAIL: `fn -> i64` — missing parameter list

**On Violation:**
`expected type, got <token>`

---

### RULE SYN-080: Block

**Status:** FROZEN
**Layer:** SYN
**Depends on:** SYN-020, LEX-010

A block is a brace-delimited sequence of statements. Blocks appear as
function bodies, loop bodies, `if`/`else` bodies, and multi-statement
arm bodies. A block may also appear as a standalone expression (`BlockExpr`)
inside match/must arms.

The value of a block is the value of the final expression statement, if the
last statement is an expression statement. If the last statement is not an
expression statement (e.g., it is a `let` or `return`), the block's value
is `unit`.

**Grammar:**

```ebnf
Block        ::= '{' Stmt* '}'
BlockExpr    ::= '{' Stmt* '}'

Stmt         ::= LetStmt
               | TupleDestructureStmt
               | ReturnStmt
               | IfStmt
               | LoopStmt
               | WhileStmt
               | ForStmt
               | BreakStmt
               | ContinueStmt
               | AssertStmt
               | ExprOrAssignStmt

LetStmt      ::= 'let' 'mut'? Ident (':' Type)? '=' Expr
TupleDestructureStmt ::= 'let' 'mut'? '(' Ident (',' Ident)* ','? ')' '=' Expr
ReturnStmt   ::= 'return' Expr?
IfStmt       ::= 'if' Expr Block ( 'else' ( IfStmt | Block ) )?
LoopStmt     ::= 'loop' Block
WhileStmt    ::= 'while' Expr Block
ForStmt      ::= 'for' Ident (',' Ident)? 'in' Expr Block
BreakStmt    ::= 'break'
ContinueStmt ::= 'continue'
AssertStmt   ::= 'assert' Expr
ExprOrAssignStmt ::= Expr ( '=' Expr | CompoundAssignOp Expr )?

CompoundAssignOp ::= '+=' | '-=' | '*=' | '/=' | '%='
```

Notes:
- There are no semicolons in Candor. Statement boundaries are determined
  by the grammar structure (each statement form starts with a distinct keyword
  or is unambiguously parsed as an expression).
- Assignment targets are restricted to `IdentExpr`, `FieldExpr`, and
  `IndexExpr`. Any other lvalue is a parse error.
- `return` with no expression is valid and returns `unit`.
- `for v, i in collection` binds both the element and the index.

**Invariant:** A block always has matching braces. An unclosed block is a parse error.

**Compliance Test:**
PASS:
```
{
    let x = 1
    let y = 2
    x + y
}
```
PASS (tuple destructure):
```
let (a, b) = pair
```
PASS (compound assignment):
```
x += 1
```
FAIL: `{ let x = 5 = 6 }` — `let` result is not an assignment target

**On Violation:**
`expected }, got <token>`
`invalid assignment target`

---

### RULE SYN-085: `module` and `use` Declarations

**Status:** FROZEN
**Layer:** SYN
**Depends on:** SYN-010, LEX-040, LEX-050

**`module` declaration:**
Declares the module this file belongs to. Parsed as `module Ident`.
At most one `module` declaration is permitted per file; enforcement is
at the semantic layer (MOD).

**`use` declaration:**
Imports a name or path from another module into the current scope.
The path consists of one or more `Ident` segments separated by `::`.
No `as` alias is supported at the syntax layer in the current parser
(the `UseDecl` AST node carries only a `Path` field).

**Grammar:**

```ebnf
ModuleDecl   ::= 'module' Ident
UseDecl      ::= 'use' Ident ( '::' Ident )*
```

Notes:
- `use foo` imports the top-level module `foo`.
- `use foo::Bar` imports the name `Bar` from module `foo`.
- `use foo::bar::Baz` imports `Baz` from the nested path `foo::bar`.
- The `as` alias syntax (e.g., `use foo as f`) is not parsed by the
  current parser; it is reserved for a future amendment.

**Invariant:** A `use` path must contain at least one identifier segment.

**Compliance Test:**
PASS: `module payments`
PASS: `use std::io`
PASS: `use collections::vec`
FAIL: `module foo::bar` — module names may not contain `::` (only a single identifier)
FAIL: `use` with no path — bare `use` is a syntax error

**On Violation:**
`expected identifier, got <token>`

---

### RULE SYN-090: Constant Declaration

**Status:** FROZEN
**Layer:** SYN
**Depends on:** SYN-030, SYN-075, SYN-020, LEX-040, LEX-050

A `const` declaration introduces a named compile-time constant. The type
annotation is mandatory; the value expression must be a compile-time
constant (enforced at the semantic layer, not here).

**Grammar:**

```ebnf
ConstDecl    ::= 'const' Ident ':' Type '=' Expr
```

**Invariant:** A `const` declaration always has both a type annotation and
a value expression.

**Compliance Test:**
PASS: `const MAX: i64 = 1000`
PASS: `const PI: f64 = 3.14159`
FAIL: `const MAX = 100` — missing type annotation
FAIL: `const MAX: i64` — missing `=` and value

**On Violation:**
`expected :, got <token>`
`expected =, got <token>`

---

### RULE SYN-091: External Function Declaration

**Status:** FROZEN
**Layer:** SYN
**Depends on:** SYN-030, SYN-060, SYN-075, LEX-040, LEX-050

An `extern fn` declaration declares the signature of a function implemented
outside Candor (typically a C function accessed via the C backend).
No body block is present. An optional effects annotation may follow the
return type.

**Grammar:**

```ebnf
ExternFnDecl ::= 'extern' 'fn' Ident '(' ParamList ')' '->' Type EffectsAnnotation?
```

**Invariant:** An `extern fn` declaration never has a body block.

**Compliance Test:**
PASS: `extern fn malloc(size: i64) -> i64`
PASS: `extern fn rand() -> i64 effects(rand)`
FAIL: `extern fn foo() -> unit { }` — extern declarations must not have a body

**On Violation:**
`expected declaration ..., got {` (body found where none expected)

---

### RULE SYN-092: Trait Declaration

**Status:** STABLE
**Layer:** SYN
**Depends on:** SYN-030, SYN-075, LEX-040, LEX-050

A `trait` declaration defines a named interface consisting of method
signatures. Methods inside a `trait` block have no body; they consist
of a signature only (`fn Name(params) -> Type`).

**Grammar:**

```ebnf
TraitDecl    ::= 'trait' Ident '{' TraitMethod* '}'
TraitMethod  ::= 'fn' Ident '(' ParamList ')' '->' Type
```

Notes:
- Trait methods have no body and no effects/contract annotations.
- Implementations of trait methods are provided in `impl Trait for Type`
  blocks (SYN-055).

**Invariant:** Trait method signatures never contain a body block.

**Compliance Test:**
PASS:
```
trait Display {
    fn fmt(self: Self) -> str
}
```
FAIL:
```
trait Foo {
    fn bar() -> unit { }  ## body not allowed in trait
}
```

**On Violation:**
`expected fn, got <token>`

---

### RULE SYN-093: Capability Declaration

**Status:** STABLE
**Layer:** SYN
**Depends on:** SYN-030, LEX-040, LEX-050

A `cap` declaration introduces a named capability token into scope.
Capabilities are used with the `cap(Name)` effects annotation (SYN-060)
to gate access to effectful functions.

**Grammar:**

```ebnf
CapabilityDecl ::= 'cap' Ident
```

**Compliance Test:**
PASS: `cap Admin`
PASS: `cap NetworkAccess`
FAIL: `cap` with no following identifier

**On Violation:**
`expected identifier, got <token>`

---

### RULE SYN-094: Primary Expressions

**Status:** FROZEN
**Layer:** SYN
**Depends on:** SYN-020, SYN-070, SYN-075, LEX-010, LEX-040, LEX-050

Primary expressions are the leaves and structured literals of the expression
grammar. They are parsed by `parsePrimaryExpr`.

**Grammar:**

```ebnf
PrimaryExpr  ::= IntLit
               | FloatLit
               | StrLit
               | BoolLit
               | 'unit'
               | Ident
               | Ident '::' Ident            (* enum path: Foo::Bar *)
               | '(' Expr ')'                (* grouped expression *)
               | '(' Expr ',' ExprList ')'   (* tuple literal *)
               | '[' ExprList ']'            (* vec literal *)
               | '&' UnaryExpr              (* reference *)
               | '*' UnaryExpr              (* dereference *)
               | 'fn' '(' ParamList ')' '->' Type Block    (* lambda *)
               | 'spawn' Block              (* concurrency spawn *)
               | 'match' Expr '{' ArmList '}'
               | 'old' '(' Expr ')'         (* contract: value at entry *)
               | 'forall' Ident 'in' Expr ':' Expr
               | 'exists' Ident 'in' Expr ':' Expr
               | 'some' | 'none' | 'ok' | 'err'
                  (* constructor keywords; applied via postfix call syntax *)

ExprList     ::= ( Expr ( ',' Expr )* ','? )?
BoolLit      ::= 'true' | 'false'
```

Notes:
- `some`, `none`, `ok`, `err` are parsed as `IdentExpr` in primary position;
  they become constructor calls when postfix `(arg)` follows.
- `move`, `secret`, `reveal` are also parsed as `IdentExpr` in primary position.
- A tuple literal `(a, b, c)` requires at least two elements; `(a)` is a
  grouped expression, not a single-element tuple.
- `forall` and `exists` are quantifier expressions used in contract predicates.

**Compliance Test:**
PASS: `42` — integer literal
PASS: `(1, 2, 3)` — tuple literal
PASS: `[1, 2, 3]` — vec literal
PASS: `fn(x: i64) -> i64 { x }` — lambda
PASS: `Foo::Bar` — enum path
FAIL: `()` — empty parentheses are not a valid primary expression (not a unit literal; use the `unit` keyword or identifier)

**On Violation:**
`expected expression, got <token>`

---

### RULE SYN-095: `old` Expression (Contract Support)

**Status:** STABLE
**Layer:** SYN
**Depends on:** SYN-094, LEX-050

The `old(expr)` form is a primary expression that may only appear inside
`ensures` contract clauses. It evaluates `expr` as of function entry.

**Grammar:**

```ebnf
OldExpr ::= 'old' '(' Expr ')'
```

**Compliance Test:**
PASS: `ensures old(x) + 1 == x`
FAIL: `old x` — missing parentheses

**On Violation:**
`expected (, got <token>`

---

### RULE SYN-096: Struct Literal Expression

**Status:** FROZEN
**Layer:** SYN
**Depends on:** SYN-020, SYN-094, LEX-040

A struct literal constructs a value of a named struct type. It is recognized
as a postfix operation on a PascalCase identifier (SYN-020). The parser
restricts struct literal syntax to identifiers whose first character is an
uppercase ASCII letter (`A`–`Z`), preventing ambiguity with block statements
following a lowercase identifier.

**Grammar:**

```ebnf
StructLitExpr ::= PascalIdent '{' FieldInitList '}'
FieldInitList ::= ( FieldInit ( ',' FieldInit )* ','? ( '..' Expr )? )?
FieldInit     ::= Ident ':' Expr
PascalIdent   ::= Ident   (* where first character is [A-Z] *)
```

Notes:
- The spread syntax `..base` may appear at the end of the field list to
  copy remaining fields from another struct value of the same type.
- A trailing comma before `}` is permitted.
- Lowercase-initial identifiers followed by `{` are parsed as an expression
  statement followed by a block, not as a struct literal.

**Compliance Test:**
PASS:
```
Point { x: 1.0, y: 2.0 }
```
PASS (spread):
```
Point { x: 3.0, ..origin }
```
FAIL: `point { x: 1.0 }` — `point` starts with lowercase; `{` begins a block

**On Violation:**
(None — the `{` is simply not consumed as a struct literal opener)

---

## COMPLETE GRAMMAR SUMMARY

The following collects all EBNF productions from this layer in one place
for easy reference. Terminal tokens use their lexeme or kind name from
L1-LEXER.md (e.g., `Ident` = `TokIdent`, `IntLit` = `TokInt`).

```ebnf
(* ── File ─────────────────────────────────────────────────── *)

File         ::= ( Directive | Decl )*

Directive    ::= TokDirective StrLit?     (* #word "optional-arg" *)

Decl         ::= FnDecl
               | StructDecl
               | EnumDecl
               | ConstDecl
               | ExternFnDecl
               | ImplDecl
               | ImplForDecl
               | TraitDecl
               | CapabilityDecl
               | UseDecl
               | ModuleDecl
               | CHeaderDecl

CHeaderDecl  ::= '#c_header' StrLit?

(* ── Declarations ─────────────────────────────────────────── *)

ModuleDecl   ::= 'module' Ident

UseDecl      ::= 'use' Ident ( '::' Ident )*

ConstDecl    ::= 'const' Ident ':' Type '=' Expr

CapabilityDecl ::= 'cap' Ident

FnDecl       ::= 'fn' Ident GenericParams? '(' ParamList ')' '->' Type
                 EffectsAnnotation? ContractClauses? Block

ExternFnDecl ::= 'extern' 'fn' Ident '(' ParamList ')' '->' Type EffectsAnnotation?

StructDecl   ::= 'struct' Ident '{' FieldList '}'
FieldList    ::= ( Field ( ',' Field )* ','? )?
Field        ::= Ident ':' Type

EnumDecl     ::= 'enum' Ident '{' VariantList '}'
VariantList  ::= ( EnumVariant ( ',' EnumVariant )* ','? )?
EnumVariant  ::= Ident ( '(' TypeList ')' )?

ImplDecl     ::= 'impl' Ident '{' ImplMethod* '}'
ImplForDecl  ::= 'impl' Ident 'for' Ident '{' ImplMethod* '}'
ImplMethod   ::= 'fn' Ident GenericParams? '(' ParamList ')' '->' Type Block

TraitDecl    ::= 'trait' Ident '{' TraitMethod* '}'
TraitMethod  ::= 'fn' Ident '(' ParamList ')' '->' Type

(* ── Generics ──────────────────────────────────────────────── *)

GenericParams ::= '<' TypeParamList '>'
TypeParamList ::= TypeParam ( ',' TypeParam )* ','?
TypeParam     ::= Ident ( ':' TraitBound )?
TraitBound    ::= Ident ( '+' Ident )*

(* ── Parameters ────────────────────────────────────────────── *)

ParamList    ::= ( Param ( ',' Param )* ','? )?
Param        ::= Ident ':' Type

(* ── Effects & Contracts ───────────────────────────────────── *)

EffectsAnnotation ::= 'pure'
                    | 'effects' '[' ']'
                    | 'effects' '(' EffectNameList ')'
                    | 'cap' '(' Ident ')'
EffectNameList    ::= Ident ( ',' Ident )* ','?

ContractClauses  ::= ContractClause*
ContractClause   ::= 'requires' Expr
                   | 'ensures' Expr

(* ── Types ─────────────────────────────────────────────────── *)

Type         ::= Ident                          (* NamedType *)
               | Ident '.' Ident               (* QualifiedType: mod.Name *)
               | Ident '<' TypeArgList '>'     (* GenericType *)
               | 'fn' '(' TypeList ')' '->' Type   (* FnType *)
               | '(' Type ',' Type ( ',' Type )* ','? ')'  (* TupleType, 2+ elems *)

TypeArgList  ::= Type ( ',' Type )* ','?
TypeList     ::= ( Type ( ',' Type )* ','? )?

(* ── Blocks & Statements ───────────────────────────────────── *)

Block        ::= '{' Stmt* '}'

Stmt         ::= 'let' 'mut'? Ident (':' Type)? '=' Expr
               | 'let' 'mut'? '(' Ident ( ',' Ident )* ','? ')' '=' Expr
               | 'return' Expr?
               | 'if' Expr Block ( 'else' ( 'if' Expr Block ( 'else' ... )? | Block ) )?
               | 'loop' Block
               | 'while' Expr Block
               | 'for' Ident ( ',' Ident )? 'in' Expr Block
               | 'break'
               | 'continue'
               | 'assert' Expr
               | Expr ( '=' Expr | '+=' Expr | '-=' Expr | '*=' Expr | '/=' Expr | '%=' Expr )?

(* ── Expressions ───────────────────────────────────────────── *)

Expr         ::= OrExpr
OrExpr       ::= AndExpr ( 'or' AndExpr )*
AndExpr      ::= CmpExpr ( 'and' CmpExpr )*
CmpExpr      ::= AddExpr ( ( '==' | '!=' | '<' | '>' | '<=' | '>=' ) AddExpr )*
AddExpr      ::= MulExpr ( ( '+' | '-' ) MulExpr )*
MulExpr      ::= UnaryExpr ( ( '*' | '/' | '%' ) UnaryExpr )*
UnaryExpr    ::= ( '!' | 'not' | '-' ) UnaryExpr
               | PostfixExpr

PostfixExpr  ::= PrimaryExpr PostfixOp*
PostfixOp    ::= '(' ArgList ')'
               | '[' Expr ']'
               | '.' Ident
               | '.' IntLit
               | 'as' Type
               | 'must' '{' ArmList '}'
               | '{' FieldInitList '}'    (* PascalCase Ident receiver only *)

ArgList      ::= ( Expr ( ',' Expr )* ','? )?

PrimaryExpr  ::= IntLit
               | FloatLit
               | StrLit
               | 'true' | 'false'
               | Ident
               | Ident '::' Ident
               | '(' Expr ')'
               | '(' Expr ',' Expr ( ',' Expr )* ','? ')'
               | '[' ExprList ']'
               | '&' UnaryExpr
               | '*' UnaryExpr
               | 'fn' '(' ParamList ')' '->' Type Block
               | 'spawn' Block
               | 'match' Expr '{' ArmList '}'
               | 'old' '(' Expr ')'
               | 'forall' Ident 'in' Expr ':' Expr
               | 'exists' Ident 'in' Expr ':' Expr

ExprList     ::= ( Expr ( ',' Expr )* ','? )?

(* ── Match / Must Arms ─────────────────────────────────────── *)

MatchExpr    ::= 'match' Expr '{' ArmList '}'
MustExpr     ::= PostfixExpr 'must' '{' ArmList '}'
ArmList      ::= ( Arm ( ',' Arm )* ','? )?
Arm          ::= Pattern '=>' ArmBody
Pattern      ::= '_' | Literal | Ident | Ident '::' Ident | Ident '(' ArgList ')'
ArmBody      ::= Expr | 'return' Expr | 'break' | '{' Stmt* '}'

(* ── Struct Literal ────────────────────────────────────────── *)

StructLitExpr ::= PascalIdent '{' FieldInitList '}'
FieldInitList ::= ( FieldInit ( ',' FieldInit )* ','? ( '..' Expr )? )?
FieldInit     ::= Ident ':' Expr
PascalIdent   ::= Ident   (* first character is [A-Z] *)
```

---

*End of L2-SYNTAX.md*
