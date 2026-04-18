# CANDOR SPECIFICATION FRAMEWORK
**Version:** 1.0
**Status:** FROZEN
**Authority:** Scott W. Corley

---

## PURPOSE

This document defines the meta-rules that govern all Candor specification documents.
It is the foundation of foundations. It does not describe the Candor language.
It describes how Candor is described.

Any AI agent, human, or tool consuming Candor specification documents
must read this document first. Everything else depends on it.

---

## 1. THE CORE CONTRACT

> **If you follow every rule in the applicable spec layers and your program
> is rejected by a conformant Candor compiler, the spec is wrong, not your program.
> File an amendment.**

> **If you follow every rule and your program compiles but produces incorrect behavior,
> the spec is incomplete. File an amendment.**

> **If you violate a rule and your program is accepted, the compiler is non-conformant.
> File a compiler bug.**

The spec is not aspirational. It is a contract with a test suite.

---

## 2. RULE ANATOMY

Every rule in every spec document has this structure:

```
### RULE [LAYER-NNN]: Rule Name
**Status:** FROZEN | STABLE | DRAFT
**Layer:** [layer name]
**Depends on:** [comma-separated rule IDs, or "none"]

[Definition: exact, unambiguous statement of what is true]

**Grammar** (if syntactic):
[EBNF production]

**Type Rule** (if semantic):
[formal notation: Γ ⊢ expr : Type]

**Invariant:**
[what must hold, as a boolean predicate]

**Compliance Test:**
PASS: [minimal Candor program or expression that must be accepted]
FAIL: [minimal Candor program or expression that must be rejected]

**On Violation:**
[exact compiler error message template]
```

Fields are omitted only when genuinely not applicable (e.g., axioms have no grammar).

---

## 3. LAYER SYSTEM

Layers are numbered. A higher layer may depend on lower layers.
A higher layer may NEVER contradict a lower layer.
A lower layer may NEVER be amended to break a higher layer without
simultaneously amending the higher layer with a migration path.

| Layer | ID      | Status   | Covers                                    |
|-------|---------|----------|-------------------------------------------|
| 0     | AXIOM   | FROZEN   | Philosophical invariants; non-negotiable  |
| 1     | LEX     | FROZEN   | Tokenization; character-level rules       |
| 2     | SYN     | FROZEN   | Grammar; syntactic structure              |
| 3     | TYP     | FROZEN   | Type system; primitives and type rules    |
| 4     | EVAL    | STABLE   | Expression evaluation semantics           |
| 5     | STMT    | STABLE   | Statement semantics                       |
| 6     | FN      | STABLE   | Functions, closures, calling convention   |
| 7     | EFF     | STABLE   | Effects system                            |
| 8     | OWN     | STABLE   | Ownership, borrows, lifetimes             |
| 9     | MOD     | STABLE   | Module system, namespacing                |
| 10    | TRAIT   | STABLE   | Traits and impl blocks                    |
| 11    | CONTR   | STABLE   | Contracts (requires/ensures/assert)       |
| 12    | COL     | STABLE   | Standard collections (vec, map, ring)     |
| 13    | ML      | DRAFT    | Tensor, SIMD, ML primitives               |
| 14    | CONC    | DRAFT    | Concurrency (task, spawn, async)          |
| 15    | STORE   | DRAFT    | Storage (mmap, colstore, NIXL)            |
| 16    | VERIFY  | DRAFT    | Formal verification (SMT, refinement)     |
| 17    | AI      | DRAFT    | AI-layer (MCP, intent, cap tokens)        |

### The Substrate Boundary

Below Layer 0 sits the **Substrate Boundary** — the formal interface between
the Candor language and the physical hardware on which it executes.

The substrate is NOT a language layer. It is a target interface that compiler
implementations satisfy. Language layers (0–17) are permanent and axiom-governed.
The substrate is target-specific and replaceable without changing the language.

| Substrate Layer | ID  | Status | Covers |
|-----------------|-----|--------|--------|
| S0 | SUB/S0 | DRAFT | Primitive type algebra; concrete representations |
| S1 | SUB/S1 | DRAFT | Memory model; layout, alignment, atomics |
| S2 | SUB/S2 | DRAFT | Execution model; calling convention, execution units |
| S3 | SUB/S3 | DRAFT | Platform capabilities; effect tag → platform mapping |

A **substrate profile** is a named, conformant implementation of S0–S3
for a specific hardware target (e.g., `x86_64-win64`, `arm64-linux`,
`quantum-hybrid-v1`). New hardware paradigms (quantum, ternary, GPU) are
absorbed as new substrate profiles — the language spec does not change.

See `SUBSTRATE.md` for the full substrate specification and profile roadmap.

### Layer Status Meanings

**FROZEN:** Rules may only be amended to add precision or fix logical contradictions.
No amendment may change the meaning of valid existing programs.
Additions must be backward compatible.
Requires SPEC-AMEND with justification signed by authority.

**STABLE:** Rules may be amended to fix bugs or add features.
Amendments must include a migration path for programs that relied on prior behavior.
New rules may be added freely.
Requires SPEC-AMEND.

**DRAFT:** Rules are under active development.
May change without migration path.
Implementation is not required for conformance.
No SPEC-AMEND required; file an issue.

---

## 4. RULE IDs

Rule IDs are permanent. Once assigned, a rule ID is never reused, even if the rule is deleted.

Format: `[LAYER]-[NNN]`
Example: `LEX-001`, `TYP-042`, `EFF-007`

Deleted rules are marked `**Status: DELETED**` and their body replaced with:
> "This rule was deleted in amendment [AMEND-NNN]. See AMENDMENTS.md."

Superseded rules point to their replacement:
> "This rule was superseded by [RULE-ID] in amendment [AMEND-NNN]."

---

## 5. AMENDMENT PROTOCOL

### Filing an Amendment

An amendment is required when:
- A FROZEN or STABLE rule is incorrect
- A FROZEN or STABLE rule is ambiguous
- A new rule is needed in a FROZEN or STABLE layer
- A rule must be deleted from a FROZEN or STABLE layer

An amendment is NOT required for:
- New rules in DRAFT layers
- Fixing typos that don't change meaning
- Adding examples

### Amendment Format

Amendments live in `docs/spec/AMENDMENTS.md`.

```
## AMEND-[NNN]: Title
**Date:** YYYY-MM-DD
**Author:** [name]
**Affects:** [list of rule IDs]
**Type:** CORRECTION | ADDITION | DELETION | CLARIFICATION

### Problem
[what was wrong or missing]

### Change
[exact change to rule text]

### Migration
[for FROZEN/STABLE: how existing programs are affected]
[NONE if no existing programs affected]

### Test
[new compliance test if applicable]
```

---

## 6. COMPLIANCE DEFINITION

A **conformant Candor compiler** is one that:

1. Accepts every program whose construction follows all FROZEN and STABLE rules.
2. Rejects every program that violates any FROZEN or STABLE rule with the specified error.
3. Produces output semantically equivalent to the evaluation rules for every accepted program.

A **conformant Candor program** is one that:

1. Satisfies all FROZEN and STABLE rules for its declared feature set.
2. Compiles and runs without error on any conformant compiler.

A program is **partially conformant** if it satisfies FROZEN and STABLE rules but not DRAFT rules.
Partial conformance is acceptable. DRAFT features are not required.

---

## 7. READING ORDER FOR AI AGENTS

When an AI agent needs to understand a Candor construct, it should:

1. Look up the construct in `INDEX.md` → get rule IDs
2. Read those rules in layer order (lowest layer first)
3. Read the rules each rule depends on if not already read
4. Stop when all dependencies are satisfied

An AI agent generating Candor code should:

1. Determine which layers are needed for the construct
2. Enumerate all rules in those layers
3. Generate code satisfying all rules simultaneously
4. If a conflict arises between rules, the lower-layer rule wins

An AI agent encountering a compiler rejection should:

1. Match the error message to the `On Violation` field of the relevant rule
2. Identify the rule ID
3. Fix the violation per the rule's definition
4. If no rule covers the error, the compiler may be non-conformant — file a bug

---

## 8. THE NON-NEGOTIABLE PROPERTIES

These properties hold at every layer, for every version of Candor, forever.
They are not rules — they are the conditions under which rules are valid.

**NP-1: Decidability.** Every syntactic and type rule must be decidable at compile time
without executing the program. A rule that requires execution to check is invalid.

**NP-2: Locality.** Every rule is checkable from local context.
A rule that requires whole-program analysis is DRAFT until an efficient algorithm is specified.

**NP-3: Explainability.** Every compiler rejection must map to exactly one rule.
A compiler that rejects a program without citing a rule is non-conformant.

**NP-4: Monotonicity.** Adding a correct program to a project cannot make another
correct program incorrect. The correctness of a program depends only on its own rules.

**NP-5: One meaning.** A valid Candor expression has exactly one type and
exactly one evaluation result in any given context. Overloading that produces
ambiguity is forbidden.

---

## 9. DOCUMENT MAP

| Document | Layers | Status |
|----------|--------|--------|
| `FRAMEWORK.md` | meta | FROZEN |
| `INDEX.md` | all | STABLE |
| `SUBSTRATE.md` | SUB / S0–S3 | DRAFT |
| `L0-AXIOMS.md` | AXIOM | FROZEN |
| `L1-LEXER.md` | LEX | FROZEN |
| `L2-SYNTAX.md` | SYN | FROZEN |
| `L3-TYPES.md` | TYP | FROZEN |
| `L4-EVAL.md` | EVAL | STABLE |
| `L5-STATEMENTS.md` | STMT | STABLE |
| `L6-FUNCTIONS.md` | FN | STABLE |
| `L7-EFFECTS.md` | EFF | STABLE |
| `L8-OWNERSHIP.md` | OWN | STABLE |
| `L9-MODULES.md` | MOD | STABLE |
| `L10-TRAITS.md` | TRAIT | STABLE |
| `L11-CONTRACTS.md` | CONTR | STABLE |
| `L12-COLLECTIONS.md` | COL | STABLE |
| `L13-ML.md` | ML | DRAFT |
| `L14-CONCURRENCY.md` | CONC | DRAFT |
| `L15-STORAGE.md` | STORE | DRAFT |
| `L16-VERIFICATION.md` | VERIFY | DRAFT |
| `L17-AI.md` | AI | DRAFT |
| `AMENDMENTS.md` | meta | STABLE |

---

*End of FRAMEWORK.md*
