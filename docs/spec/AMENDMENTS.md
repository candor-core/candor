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

## (Future amendments go here, appended below this line)

---

*End of AMENDMENTS.md*
