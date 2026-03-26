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

## (Future amendments go here, appended below this line)

---

*End of AMENDMENTS.md*
