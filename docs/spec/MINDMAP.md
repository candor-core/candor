# CANDOR SPECIFICATION MIND MAP
**Status:** STABLE
**Depends on:** FRAMEWORK.md, INDEX.md

---

## PURPOSE

This document is a navigation guide over the Candor specification.
It shows how layers relate, which rules an AI or human needs for common tasks,
and how to find anything quickly.

---

## THE LAYER STACK

```
┌─────────────────────────────────────────────────────────────┐
│  AI/AGENT LAYER (L17-AI)         cap<T>, #intent, #mcp_tool │
│  VERIFICATION (L16-VERIFY)        SMT, refinement types      │
│  STORAGE (L15-STORE)              mmap<T>, colstore, NIXL    │
│  CONCURRENCY (L14-CONC)           task<T>, spawn, async      │
│  ML PRIMITIVES (L13-ML)           tensor<T>, SIMD, f16/bf16  │
│  COLLECTIONS (L12-COL)            vec<T>, map<K,V>, ring<T>  │
│  CONTRACTS (L11-CONTR)            requires, ensures, assert  │
│  TRAITS (L10-TRAIT)               trait decl, impl T for S   │
│  MODULES (L9-MOD)                 module, use, namespacing    │
│  OWNERSHIP (L8-OWN)               box<T>, arc<T>, ref<T>     │
├─────────────────────────────────────────────────────────────┤
│  EFFECTS (L7-EFF)                 pure, effects(io,sys,...)  │ ← enforcement
│  FUNCTIONS (L6-FN)                fn, closures, extern fn    │ ← callables
│  STATEMENTS (L5-STMT)             let, if, loop, must{}      │ ← control flow
│  EVALUATION (L4-EVAL)             expressions, operators     │ ← values
├─────────────────────────────────────────────────────────────┤
│  TYPE SYSTEM (L3-TYP)             i64, str, option<T>...     │ ← what things are
│  SYNTAX (L2-SYN)                  EBNF grammar               │ ← structure
│  LEXER (L1-LEX)                   tokens, keywords           │ ← characters → tokens
│  AXIOMS (L0-AXIOM)                non-negotiable truths      │ ← the foundation
└─────────────────────────────────────────────────────────────┘
```

**Reading rule:** Any layer may only use features from layers BELOW it.
When a rule at layer N is ambiguous, the rule at layer M < N always wins.

---

## COMMON TASK GUIDES

### "I want to write a function"

```
1. SYN-040    fn declaration syntax
2. FN-010     function declaration semantics
3. FN-020     parameter list rules
4. FN-040     return type checking
5. TYP-060    function type
6. EFF-010    add effects annotation if needed
```

### "I want to handle an error"

```
1. TYP-055    result<T,E> type
2. TYP-050    option<T> type
3. STMT-070   must{} expression
4. AXIOM-003  why you must handle it (the law)
```

### "I want to define a type"

```
   Struct:  TYP-040, SYN-045
   Enum:    TYP-041, SYN-050
   Alias:   (not yet in spec — use newtype struct)
   Generic: TYP-070, SYN-040
```

### "I want to use ownership"

```
1. AXIOM-004  why ownership exists
2. OWN-010    box<T> — heap-owned pointer
3. OWN-020    arc<T> — reference-counted pointer
4. OWN-030    ref<T> — immutable borrow
5. OWN-031    refmut<T> — mutable borrow
```

### "I want to call C code"

```
1. SYN-091    extern fn declaration
2. FN-070     extern fn semantics
3. AXIOM-002  effects: extern fns are trusted (unchecked)
```

### "I want to do I/O"

```
1. EFF-002    io effect name
2. EFF-010    effects(io) annotation syntax
3. EFF-040    which builtins require io
```

### "I want to understand an error message"

```
INDEX.md → "QUICK LOOKUP: Error Messages → Rules"
→ read the rule's "On Violation" field
→ read the rule's "Compliance Test" FAIL case
```

---

## CONCEPT CLUSTERS

### Numeric Types
```
i8 i16 i32 i64* i128          TYP-001–TYP-005   L3-TYPES.md
u8 u16 u32 u64  u128          TYP-006–TYP-010   L3-TYPES.md
f16 bf16 f32 f64*             TYP-012–TYP-015   L3-TYPES.md
* = default types
No implicit promotion → TYP-080 (use `as` cast)
```

### String/Text
```
str (immutable UTF-8)          TYP-020           L3-TYPES.md
String literals                LEX-023           L1-LEXER.md
Escape sequences               LEX-024           L1-LEXER.md
```

### Optionality & Errors
```
option<T>  → some(v) | none    TYP-050–TYP-051   L3-TYPES.md
result<T,E>→ ok(v) | err(e)    TYP-055           L3-TYPES.md
must{}     → forced unwrap      STMT-070          L5-STATEMENTS.md
```

### Control Flow
```
if / else                       STMT-020          L5-STATEMENTS.md
loop                            STMT-031          L5-STATEMENTS.md
while                           STMT-032          L5-STATEMENTS.md
for x in collection             STMT-033          L5-STATEMENTS.md
break / continue                STMT-040–041      L5-STATEMENTS.md
return                          STMT-050          L5-STATEMENTS.md
match expr { pattern => expr }  EVAL-060          L4-EVAL.md
```

### Binding & Mutation
```
let x = ...    (immutable)      STMT-005          L5-STATEMENTS.md
let mut x = .. (mutable)        STMT-006          L5-STATEMENTS.md
x = expr       (assignment)     STMT-010          L5-STATEMENTS.md
const X: T = . (compile-time)   STMT-060          L5-STATEMENTS.md
```

### Effects
```
pure            → no effects    EFF-005           L7-EFFECTS.md
effects(io)     → I/O allowed   EFF-010           L7-EFFECTS.md
effects(sys)    → OS calls      EFF-010           L7-EFFECTS.md
effects(net)    → network       EFF-010           L7-EFFECTS.md
Propagation check               EFF-030           L7-EFFECTS.md
```

### Collections
```
vec<T>          dynamic array   COL-010           L12-COLLECTIONS.md
map<K,V>        hash map        COL-020           L12-COLLECTIONS.md
ring<T>         circular buffer COL-030           L12-COLLECTIONS.md
set<T>          hash set        COL-040           L12-COLLECTIONS.md
```

### Concurrency
```
task<T>         async task      CONC-010          L14-CONCURRENCY.md
spawn expr      launch task     CONC-020          L14-CONCURRENCY.md
```

### ML/Compute
```
tensor<T>       n-dim array     ML-010            L13-ML.md
f16/bf16        reduced float   TYP-014–TYP-015   L3-TYPES.md
```

---

## THE 7 AXIOMS (MEMORIZE THESE)

| Axiom | Name | One-liner |
|-------|------|-----------|
| AXIOM-001 | One Meaning | Every expression has exactly one type |
| AXIOM-002 | Declare Effects | Every side effect must be in the annotation |
| AXIOM-003 | Handle Errors | `result`/`option` cannot be silently discarded |
| AXIOM-004 | Own Explicitly | Ownership transfer is always visible in source |
| AXIOM-005 | Compile-Time | All type + effect rules are decidable at compile time |
| AXIOM-006 | Source Authority | `.cnd` file is the complete truth; no hidden behavior |
| AXIOM-007 | Layered | Layer N may only use layers 0..N-1 |

**These are the laws. Every other rule derives from them.**

---

## SPEC DOCUMENT MAP

```
FRAMEWORK.md    ← Read this first. Meta-rules for all spec documents.
INDEX.md        ← O(1) lookup: concept → rule → document
AMENDMENTS.md   ← Append-only change log

L0-AXIOMS.md    FROZEN  7 axioms
L1-LEXER.md     FROZEN  LEX-001–LEX-090  (tokenization)
L2-SYNTAX.md    FROZEN  SYN-010–SYN-096  (grammar)
L3-TYPES.md     FROZEN  TYP-001–TYP-101  (type system)
L4-EVAL.md      STABLE  EVAL-001–EVAL-099 (expressions)
L5-STATEMENTS.md STABLE STMT-001–STMT-072 (statements)
L6-FUNCTIONS.md STABLE  FN-010–FN-090    (functions)
L7-EFFECTS.md   STABLE  EFF-001–EFF-070  (effects)
L8-OWNERSHIP.md STABLE  OWN-010–OWN-099  (ownership)
L9-MODULES.md   STABLE  MOD-010–MOD-099  (modules)
L10-TRAITS.md   STABLE  TRAIT-010–TRAIT-099 (traits)
L11-CONTRACTS.md STABLE CONTR-010–CONTR-099 (contracts)
L12-COLLECTIONS.md STABLE COL-010–COL-099 (collections)
L13-ML.md       DRAFT   ML-010–ML-099    (ML/tensor)
L14-CONCURRENCY.md DRAFT CONC-010–CONC-099 (tasks/spawn)
L15-STORAGE.md  DRAFT   STORE-010–STORE-099 (mmap/colstore)
L16-VERIFICATION.md DRAFT VERIFY-010–VERIFY-099 (SMT/refinement)
L17-AI.md       DRAFT   AI-010–AI-099    (cap<T>/intent)
```

---

## THE AMENDMENT PROTOCOL (QUICK REFERENCE)

1. Find the rule that's wrong or missing
2. Write the fix using the template in FRAMEWORK.md §5
3. Append to AMENDMENTS.md as AMEND-NNN (NNN = next sequential number)
4. Update the rule in its layer document
5. Update INDEX.md if the rule's error message or concept changed

**FROZEN layer?** Amendment requires justification + migration path.
**STABLE layer?** Amendment requires migration path.
**DRAFT layer?** No amendment needed — edit directly.

---

## FOR AI AGENTS: QUICK START

```
Generate valid Candor:
  1. Read FRAMEWORK.md (meta)
  2. Read INDEX.md (lookup)
  3. Read L0 + L1 + L2 + L3 (foundation — always needed)
  4. Read the layer(s) for your construct
  5. Generate code satisfying all rules

Fix a compiler error:
  1. Look up the error message in INDEX.md "Error Messages" table
  2. Read the rule's "On Violation" + "Compliance Test FAIL"
  3. Fix per the rule's definition

Discover a spec gap:
  1. Write a minimal program that should work but doesn't
  2. Identify which rule is missing or wrong
  3. File AMEND-NNN per FRAMEWORK.md §5
```

---

*End of MINDMAP.md*
