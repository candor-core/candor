# CANDOR SPECIFICATION AMENDMENTS
**Status:** STABLE
**Depends on:** FRAMEWORK.md

---

## PURPOSE

Every change to a FROZEN or STABLE rule is recorded here.
Each amendment has a unique ID that is never reused.
The log is append-only. Old entries are never modified.

---

## AMEND-001: Initial Specification

**Date:** 2026-03-25
**Author:** Scott W. Corley
**Affects:** All (initial population)
**Type:** ADDITION

### Problem
No formal stratified specification existed. The prose specification.md
described features but did not provide rule IDs, compliance tests,
layer dependencies, or an amendment protocol.

### Change
Created the full spec framework:
- `FRAMEWORK.md` — meta-rules
- `INDEX.md` — concept lookup
- `L0-AXIOMS.md` — 7 axioms
- `L1-LEXER.md` — tokenization rules (LEX-001–LEX-090)
- `L3-TYPES.md` — type system (TYP-001–TYP-101)
- `AMENDMENTS.md` — this file

Remaining layers (L2-SYNTAX through L17-AI) are skeleton documents
pending population. The INDEX.md contains rule ID placeholders for all
planned rules.

### Migration
NONE — this is the initial population. No existing programs are affected.

### Test
Any conformant compiler that implements the FROZEN rules in L1 and L3
must pass all PASS tests and reject all FAIL tests in those documents.

---

## AMEND-002: Layer Population — L2, L5, L6, L7, MINDMAP

**Date:** 2026-03-26
**Author:** Scott W. Corley
**Affects:** L2-SYNTAX.md, L5-STATEMENTS.md, L6-FUNCTIONS.md, L7-EFFECTS.md, MINDMAP.md (new)
**Type:** ADDITION

### Problem
AMEND-001 created skeleton stubs for L2, L5, L6, L7 but did not populate them.
Without these layers, AI agents and humans had no formal grammar, statement semantics,
function rules, or effects enforcement rules to reference.

### Change
Populated four layer documents with complete formal rules:

- **L2-SYNTAX.md** — 20 rules (SYN-010–SYN-096): full EBNF grammar derived
  from `compiler/parser/parser.go`. Covers file structure, operator precedence,
  all declaration forms, all statement forms, all expression forms, pattern syntax,
  type annotation syntax, and a complete grammar summary section.

- **L5-STATEMENTS.md** — rules STMT-001–STMT-072: `let`/`let mut`, assignment,
  augmented assignment, `if`, `loop`, `while`, `for`, `break`, `continue`,
  `return`, `const`, `must{}`, and `assert`. Derived from `compiler/typeck/typeck.go`.

- **L6-FUNCTIONS.md** — rules FN-010–FN-090: function declarations, parameter lists,
  calling convention, return type checking, closures, generic functions, `extern fn`,
  method declarations, and recursive functions. Derived from parser + typeck.

- **L7-EFFECTS.md** — rules EFF-001–EFF-070: effects overview, known effect names,
  `pure` annotation, effects declaration syntax, propagation checking, builtin
  effect assignments, `cap(X)` capability annotation, `secret<T>` enforcement.
  Derived directly from `compiler/typeck/typeck.go` `KnownEffects` map and
  `checkEffectsCompat` / `checkCapCompat` functions.

- **MINDMAP.md** (new) — human/AI navigation guide: layer stack diagram,
  common task guides, concept clusters, axiom table, document map,
  amendment quick reference, and AI agent quick-start instructions.

### Migration
NONE — all additions. No existing rules modified. No existing programs affected.

### Test
Any AI agent that reads FRAMEWORK.md + INDEX.md + the populated layer documents
should be able to generate syntactically and semantically valid Candor programs
for the covered feature set (L0–L7 excluding ownership/modules/traits).

---

## AMEND-003: M9.8/M9.9 Bootstrap — typeck.cnd Bundled + Pipeline Wiring

**Date:** 2026-03-27
**Author:** Scott W. Corley
**Affects:** stage1 bootstrap build (`src/compiler/`), EFF-001 (trusted/unknown mode)
**Type:** IMPLEMENTATION

### Problem

`typeck.cnd` was written as a standalone module with its own simplified AST type
definitions duplicating those in `parser.cnd`. In a multi-source `candorc build`,
all source files share a single namespace, so duplicate struct/enum definitions
caused name collisions. Additionally:

- `module typeck` prefixed all exported names as `typeck_X`, incompatible with
  `parser.cnd`'s root-namespace AST types.
- `TK_PRIM/TK_GEN/TK_NAMED/TK_VAR` (type-kind constants) collided with
  `parser.cnd`'s lexer token constants (e.g., `TK_STR`, `TK_IDENT`).
- All `infer_*` / `check_*` functions used the old simplified AST field names,
  variant names, and structural layouts rather than parser.cnd's actual types.

### Changes

**M9.8 — typeck.cnd bundled with parser.cnd (commit 4b6ba9d):**

1. Removed `module typeck` declaration — all functions are now root-namespace.
2. Removed all duplicate AST type definitions from typeck.cnd (the old simplified
   `Stmt`, `Expr`, `Decl`, `ParsedFile`, `GenTy`, `FnTy`, `TupleTy`, `Param`,
   `StructFieldDecl`, `EnumVariant`, `FnDecl`, `StructDecl`, `EnumDecl`,
   `ExternDecl`, `ConstDecl`, `TraitDecl`, `ImplDecl`, `TypeExpr` stubs).
   All of these are now taken directly from `parser.cnd`.
3. Renamed type-kind constants `TK_PRIM/TK_GEN/TK_NAMED/TK_VAR` → `TYK_PRIM/TYK_GEN/TYK_NAMED/TYK_VAR`
   throughout typeck.cnd to eliminate token-constant collisions.
4. Rewrote all `infer_*` / `check_*` functions to match parser.cnd's actual
   field names, variant names, and struct layouts:
   - `BinExpr.left/right` are `Expr` directly (not boxed)
   - `UnExpr.operand` is `Expr` directly; `op` is `i64`
   - `IfStmt.else_: option<box<IfOrBlock>>` with new `check_if_or_block()`
   - `ForStmt.key: str`, `.val: option<str>` (was `.name`)
   - `Expr` variant names: `Int/Float/Bool/Str/Unit/None/Binary/Unary/Field/Call/
     StructLit/SomeExpr/OkExpr/ErrExpr/Cast/Index/VecLit/TupleLit/AddrOf/Lambda/
     Block/Match/Must/Return/Break/Continue/OldExpr`
   - `Stmt` variant names: `LetS/LetTupS/AssignS/IfS/WhileS/LoopS/ForS/BreakS/
     ContinueS/ReturnS/AssertS/ExprS`
5. Added `callee_name(e: Expr) -> option<str>` helper with exhaustive match
   (replacing `must{}` on a plain enum, which is not legal — `must{}` only works
   on `result<T,E>` and `option<T>`).
6. Added `typeck.cnd` to `[build].sources` in `src/compiler/Candor.toml`.

**M9.9 — typecheck() wired into stage1 pipeline (commit 2f2272a):**

1. Updated `src/compiler/main.cnd` to call `typecheck(refmut(pf))` between parse
   and emit_c. (`refmut<T>` coerces to `ref<T>` — `ref()` is not a builtin function.)
2. Added `print_str_list()` helper to iterate `tf.warnings` and `tf.errors`.
3. Pipeline aborts with non-zero exit if `tf.errors` is non-empty.

**False-positive suppression (trusted/unannotated mode per EFF-001):**

Because builtins (`print`, `lex`, `parse`, `emit_c`, etc.) and extern functions are
not registered in the TypeEnv, `infer_ident` and `infer_call_fn` now return
`ty_unknown()` instead of an error for unrecognised names. `ty_unknown()` propagates
permissively through:
- Arithmetic (`infer_arith`): if either operand is `ty_unknown()`, result is `ty_unknown()`
- Comparisons (`infer_cmp`): if either operand is `ty_unknown()`, result is `ty_bool()`
- Field access (`infer_fld`): if receiver is `ty_unknown()` or `TYK_GEN`, result is `ty_unknown()`

This produces zero false positives on all five compiler source files
(`lexer.cnd`, `parser.cnd`, `typeck.cnd`, `emit_c.cnd`, `main.cnd`).

### Migration

NONE for users — this is internal stage1 compiler infrastructure.

Any future Candor typechecker implementation must honour the trusted/unknown-mode
semantics: unknown function calls and unknown identifiers return `ty_unknown()`,
which propagates permissively. This is the formal specification of EFF-001's
"unannotated context" behaviour.

### Test

```
candorc-stage1 src/compiler/lexer.cnd    # zero type errors
candorc-stage1 src/compiler/parser.cnd   # zero type errors
candorc-stage1 src/compiler/typeck.cnd   # zero type errors
candorc-stage1 src/compiler/emit_c.cnd   # zero type errors
candorc-stage1 src/compiler/main.cnd     # zero type errors
```

The stage1 binary itself is produced by:
```
candorc build src/compiler/Candor.toml
```

---

## (Future amendments go here, appended below this line)

---

*End of AMENDMENTS.md*
