# CANDOR SPECIFICATION INDEX
**Version:** 1.0
**Status:** STABLE
**Depends on:** FRAMEWORK.md

---

## PURPOSE

This is the single lookup table for every Candor concept.
Given a term, find the rule(s) that define it.
Given a rule ID, find the document.

An AI agent or human reading this index can answer:
"Where is [X] defined?" in constant time.

---

## QUICK LOOKUP: Concepts → Rules

### A
| Concept | Rule IDs | Document |
|---------|----------|----------|
| `arc<T>` reference-counted pointer | OWN-020 | L8-OWNERSHIP.md |
| Arithmetic operators | EVAL-010–EVAL-019 | L4-EVAL.md |
| `assert` statement | CONTR-030 | L11-CONTRACTS.md |
| Assignment statement | STMT-010 | L5-STATEMENTS.md |
| Augmented assignment (`+=`, `-=`, etc.) | STMT-011 | L5-STATEMENTS.md |

### B
| Concept | Rule IDs | Document |
|---------|----------|----------|
| `bf16` type | TYP-015 | L3-TYPES.md |
| Binary expressions | EVAL-010 | L4-EVAL.md |
| `bool` type | TYP-004 | L3-TYPES.md |
| `bool` literals (`true`, `false`) | LEX-030 | L1-LEXER.md |
| `box<T>` heap pointer | OWN-010 | L8-OWNERSHIP.md |
| `break` statement | STMT-040 | L5-STATEMENTS.md |

### C
| Concept | Rule IDs | Document |
|---------|----------|----------|
| `cap<T>` capability token | AI-010 | L17-AI.md |
| `continue` statement | STMT-041 | L5-STATEMENTS.md |
| Calling convention | FN-030 | L6-FUNCTIONS.md |
| Closures / lambdas | FN-050 | L6-FUNCTIONS.md |
| Comments | LEX-080 | L1-LEXER.md |
| Comparison operators | EVAL-020 | L4-EVAL.md |
| `const` declarations | STMT-060 | L5-STATEMENTS.md |
| Contract clauses | CONTR-010–CONTR-050 | L11-CONTRACTS.md |

### D
| Concept | Rule IDs | Document |
|---------|----------|----------|
| Declarations (top-level) | SYN-030 | L2-SYNTAX.md |

### E
| Concept | Rule IDs | Document |
|---------|----------|----------|
| Effects annotation syntax | SYN-060 | L2-SYNTAX.md |
| Effects checking rules | EFF-010–EFF-050 | L7-EFFECTS.md |
| Effects system overview | EFF-001 | L7-EFFECTS.md |
| `enum` declaration | SYN-050 | L2-SYNTAX.md |
| Enum variant patterns | EVAL-060 | L4-EVAL.md |
| Enum variant construction | EVAL-061 | L4-EVAL.md |
| `err(E)` result variant | TYP-052 | L3-TYPES.md |
| Error handling (`must{}`) | STMT-070 | L5-STATEMENTS.md |
| Expression statement | STMT-001 | L5-STATEMENTS.md |
| `extern fn` declaration | FN-070 | L6-FUNCTIONS.md |

### F
| Concept | Rule IDs | Document |
|---------|----------|----------|
| `f16` type | TYP-014 | L3-TYPES.md |
| `f32` type | TYP-012 | L3-TYPES.md |
| `f64` type | TYP-013 | L3-TYPES.md |
| Field access (`.field`) | EVAL-040 | L4-EVAL.md |
| Float literals | LEX-022 | L1-LEXER.md |
| `fn` declaration | SYN-040 | L2-SYNTAX.md |
| `for` loop | STMT-033 | L5-STATEMENTS.md |
| Function call expression | EVAL-030 | L4-EVAL.md |
| Function type | TYP-060 | L3-TYPES.md |

### G
| Concept | Rule IDs | Document |
|---------|----------|----------|
| Generic functions | FN-060 | L6-FUNCTIONS.md |
| Generic types | TYP-070 | L3-TYPES.md |

### I
| Concept | Rule IDs | Document |
|---------|----------|----------|
| `i8`, `i16`, `i32`, `i64`, `i128` | TYP-001–TYP-005 | L3-TYPES.md |
| Identifiers | LEX-040 | L1-LEXER.md |
| `if` expression | EVAL-050 | L4-EVAL.md |
| `if` statement | STMT-020 | L5-STATEMENTS.md |
| Index operator (`[]`) | EVAL-045 | L4-EVAL.md |
| Integer literals | LEX-021 | L1-LEXER.md |
| Integer overflow | EVAL-015 | L4-EVAL.md |

### K
| Concept | Rule IDs | Document |
|---------|----------|----------|
| Keywords (complete list) | LEX-050 | L1-LEXER.md |

### L
| Concept | Rule IDs | Document |
|---------|----------|----------|
| `let` statement | STMT-005 | L5-STATEMENTS.md |
| `let mut` statement | STMT-006 | L5-STATEMENTS.md |
| Lambda capture rules | FN-051 | L6-FUNCTIONS.md |
| Logical operators | EVAL-025 | L4-EVAL.md |
| `loop` statement | STMT-031 | L5-STATEMENTS.md |

### M
| Concept | Rule IDs | Document |
|---------|----------|----------|
| `map<K,V>` | COL-020 | L12-COLLECTIONS.md |
| `match` expression | EVAL-060 | L4-EVAL.md |
| `match` exhaustiveness | EVAL-061 | L4-EVAL.md |
| `mmap<T>` | STORE-010 | L15-STORAGE.md |
| `module` declaration | MOD-010 | L9-MODULES.md |
| `must{}` expression | STMT-070 | L5-STATEMENTS.md |

### N
| Concept | Rule IDs | Document |
|---------|----------|----------|
| `none` literal | TYP-051 | L3-TYPES.md |
| Numeric coercion | TYP-080 | L3-TYPES.md |
| Numeric type hierarchy | TYP-090 | L3-TYPES.md |

### O
| Concept | Rule IDs | Document |
|---------|----------|----------|
| `ok(T)` result variant | TYP-053 | L3-TYPES.md |
| `option<T>` type | TYP-050 | L3-TYPES.md |
| Operator precedence | SYN-020 | L2-SYNTAX.md |

### P
| Concept | Rule IDs | Document |
|---------|----------|----------|
| Path expressions (`Mod::Item`) | EVAL-070 | L4-EVAL.md |
| Pattern syntax | SYN-070 | L2-SYNTAX.md |
| `pure` annotation | EFF-005 | L7-EFFECTS.md |

### R
| Concept | Rule IDs | Document |
|---------|----------|----------|
| `ref<T>` immutable borrow | OWN-030 | L8-OWNERSHIP.md |
| `refmut<T>` mutable borrow | OWN-031 | L8-OWNERSHIP.md |
| `requires` clause | CONTR-010 | L11-CONTRACTS.md |
| `ensures` clause | CONTR-020 | L11-CONTRACTS.md |
| `result<T,E>` type | TYP-055 | L3-TYPES.md |
| `return` statement | STMT-050 | L5-STATEMENTS.md |
| `ring<T>` circular buffer | COL-030 | L12-COLLECTIONS.md |

### S
| Concept | Rule IDs | Document |
|---------|----------|----------|
| `secret<T>` wrapper | AI-020 | L17-AI.md |
| Substrate boundary (language/hardware interface) | SUB-001 | SUBSTRATE.md |
| Substrate profile (named target implementation) | SUB-050, SUB-051 | SUBSTRATE.md |
| Substrate extensions (qbit, quantum effects, etc.) | SUB-051 | SUBSTRATE.md |
| S0 — Primitive type algebra | SUB-010, SUB-011 | SUBSTRATE.md |
| S1 — Memory model | SUB-020, SUB-021 | SUBSTRATE.md |
| S2 — Execution model / calling convention | SUB-030, SUB-031 | SUBSTRATE.md |
| S3 — Platform capabilities / effect tag mapping | SUB-040 | SUBSTRATE.md |
| `set<T>` | COL-040 | L12-COLLECTIONS.md |
| `some(x)` option variant | TYP-051 | L3-TYPES.md |
| `spawn` expression | CONC-020 | L14-CONCURRENCY.md |
| `str` type | TYP-020 | L3-TYPES.md |
| String literals | LEX-023 | L1-LEXER.md |
| String escape sequences | LEX-024 | L1-LEXER.md |
| `struct` declaration | SYN-045 | L2-SYNTAX.md |
| Struct literal | EVAL-080 | L4-EVAL.md |

### T
| Concept | Rule IDs | Document |
|---------|----------|----------|
| `task<T>` type | CONC-010 | L14-CONCURRENCY.md |
| `tensor<T>` type | ML-010 | L13-ML.md |
| `trait` declaration | TRAIT-010 | L10-TRAITS.md |
| Tuple types | TYP-030 | L3-TYPES.md |
| Type annotations | SYN-080 | L2-SYNTAX.md |
| Type casting (`as`) | EVAL-090 | L4-EVAL.md |
| Token kinds (complete list) | LEX-010–LEX-090 | L1-LEXER.md |

### U
| Concept | Rule IDs | Document |
|---------|----------|----------|
| `u8`, `u16`, `u32`, `u64`, `u128` | TYP-006–TYP-010 | L3-TYPES.md |
| Unary operators | EVAL-005 | L4-EVAL.md |
| `unit` type | TYP-025 | L3-TYPES.md |
| `use` declaration | MOD-020 | L9-MODULES.md |

### V
| Concept | Rule IDs | Document |
|---------|----------|----------|
| `vec<T>` | COL-010 | L12-COLLECTIONS.md |

### W
| Concept | Rule IDs | Document |
|---------|----------|----------|
| `while` loop | STMT-032 | L5-STATEMENTS.md |
| Whitespace handling | LEX-005 | L1-LEXER.md |

---

## QUICK LOOKUP: Error Messages → Rules

When a compiler produces an error, find the governing rule here.

| Error Message Pattern | Rule ID |
|----------------------|---------|
| `undefined identifier "X"` | MOD-030 |
| `return type mismatch: got X, expected Y` | FN-040 |
| `function with effects(X) cannot call "Y" which requires effect "Z"` | EFF-030 |
| `must{} requires result<T,E> or option<T>, got X` | STMT-071 |
| `must{} arm type X does not match expected Y` | STMT-072 |
| `non-exhaustive match: missing arm for X` | EVAL-062 |
| `cannot use X as Y` | TYP-100 |
| `cannot assign to immutable binding "X"` | STMT-007 |
| `index out of bounds: type X is not indexable` | EVAL-046 |
| `pure function cannot call "X" which has effects` | EFF-031 |
| `"X" is from module "Y"; add 'use Y::X'` | MOD-031 |
| `type annotation required: cannot infer type of "X"` | TYP-101 |
| `unknown effect "X"` | EFF-002 |
| `requires clause violated` | CONTR-011 |

---

## QUICK LOOKUP: Rule IDs → Documents

| Range | Document |
|-------|----------|
| SUB-001 – SUB-099 | `SUBSTRATE.md` |
| AXIOM-001 – AXIOM-099 | `L0-AXIOMS.md` |
| LEX-001 – LEX-099 | `L1-LEXER.md` |
| SYN-001 – SYN-099 | `L2-SYNTAX.md` |
| TYP-001 – TYP-099 | `L3-TYPES.md` |
| EVAL-001 – EVAL-099 | `L4-EVAL.md` |
| STMT-001 – STMT-099 | `L5-STATEMENTS.md` |
| FN-001 – FN-099 | `L6-FUNCTIONS.md` |
| EFF-001 – EFF-099 | `L7-EFFECTS.md` |
| OWN-001 – OWN-099 | `L8-OWNERSHIP.md` |
| MOD-001 – MOD-099 | `L9-MODULES.md` |
| TRAIT-001 – TRAIT-099 | `L10-TRAITS.md` |
| CONTR-001 – CONTR-099 | `L11-CONTRACTS.md` |
| COL-001 – COL-099 | `L12-COLLECTIONS.md` |
| ML-001 – ML-099 | `L13-ML.md` |
| CONC-001 – CONC-099 | `L14-CONCURRENCY.md` |
| STORE-001 – STORE-099 | `L15-STORAGE.md` |
| VERIFY-001 – VERIFY-099 | `L16-VERIFICATION.md` |
| AI-001 – AI-099 | `L17-AI.md` |

---

*End of INDEX.md*
