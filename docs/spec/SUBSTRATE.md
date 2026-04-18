# SUBSTRATE SPECIFICATION
**Version:** 0.1
**Status:** DRAFT
**Authority:** Scott W. Corley
**Depends on:** FRAMEWORK.md, L0-AXIOMS.md

---

## PURPOSE

This document defines the **Substrate Boundary** — the formal interface between
the Candor language and the physical hardware on which it executes.

Everything above the boundary (Layers 0–17) is governed by the 7 frozen axioms
and is permanent. Everything below the boundary is target-specific and replaceable.

A new hardware paradigm (quantum computing, ternary logic, new CPU architectures)
is absorbed by defining a new **substrate profile** that satisfies this interface.
The language spec does not change. The axioms do not change.

This document defines what every conformant substrate profile must provide.
It does not define any specific substrate profile — those live in separate
`SUBSTRATE-<name>.md` documents.

---

## THE BOUNDARY DEFINED

```
  ╔══════════════════════════════════════════════╗
  ║  Layer 17 — AI                               ║
  ║  Layer 16 — Formal Verification              ║
  ║  Layer 15 — Storage                          ║
  ║  Layer 14 — Concurrency                      ║
  ║  Layer 13 — ML / SIMD                        ║
  ║  Layer 12 — Collections                      ║
  ║  Layer 11 — Contracts                        ║
  ║  Layer 10 — Traits                           ║
  ║  Layer  9 — Modules                          ║
  ║  Layer  8 — Ownership                        ║
  ║  Layer  7 — Effects                          ║
  ║  Layer  6 — Functions                        ║
  ║  Layer  5 — Statements                       ║
  ║  Layer  4 — Evaluation                       ║
  ║  Layer  3 — Types                            ║
  ║  Layer  2 — Syntax                           ║
  ║  Layer  1 — Lexer                            ║
  ║  Layer  0 — AXIOMS (frozen, eternal)         ║
  ╠══════════ SUBSTRATE BOUNDARY ════════════════╣
  ║  S0 — Primitive Type Algebra                 ║
  ║  S1 — Memory Model                           ║
  ║  S2 — Execution Model                        ║
  ║  S3 — Platform Capabilities                  ║
  ╠══════════════════════════════════════════════╣
  ║  Hardware: x86_64 / ARM / RISC-V /           ║
  ║            Quantum Hybrid / Ternary / GPU    ║
  ╚══════════════════════════════════════════════╝
```

The language never references hardware directly. It references abstract
types, effects, and operations. The substrate maps these to hardware reality.

---

## S0 — PRIMITIVE TYPE ALGEBRA

### RULE SUB-010: Substrate Must Define Primitive Types
**Status:** DRAFT
**Layer:** SUBSTRATE / S0
**Depends on:** TYP-001–TYP-025

A conformant substrate profile must provide concrete representations for
every primitive type declared in Layer 3 (TYP):

| Candor Type | Minimum width | Representation requirement |
|-------------|--------------|---------------------------|
| `bool`      | 1 bit        | 0 = false, 1 = true; no other values valid |
| `i8`        | 8 bits       | Two's complement signed integer |
| `i16`       | 16 bits      | Two's complement signed integer |
| `i32`       | 32 bits      | Two's complement signed integer |
| `i64`       | 64 bits      | Two's complement signed integer |
| `i128`      | 128 bits     | Two's complement signed integer |
| `u8`        | 8 bits       | Unsigned integer |
| `u16`       | 16 bits      | Unsigned integer |
| `u32`       | 32 bits      | Unsigned integer |
| `u64`       | 64 bits      | Unsigned integer |
| `u128`      | 128 bits     | Unsigned integer |
| `f32`       | 32 bits      | IEEE 754 binary32 |
| `f64`       | 64 bits      | IEEE 754 binary64 |
| `f16`       | 16 bits      | IEEE 754 binary16 |
| `unit`      | 0 bits       | No runtime representation required |
| `str`       | pointer+len  | UTF-8 encoded; immutable; null-termination optional |

A substrate may provide ADDITIONAL types (e.g., `qbit` on a quantum substrate,
trit types on a ternary substrate) as substrate extensions. These extensions
are declared in the substrate profile and may be used only in code annotated
with the corresponding substrate capability effect.

**Invariant:**
Every TYP-layer primitive type is representable on any conformant substrate.

---

### RULE SUB-011: Arithmetic Must Be Defined
**Status:** DRAFT
**Layer:** SUBSTRATE / S0
**Depends on:** SUB-010, EVAL-010–EVAL-019

A conformant substrate must define the behavior of all arithmetic operations
on each primitive type, including:
- Overflow semantics for integer types (wrapping, trap, or saturating —
  the profile must declare which, and Candor's Layer 4 EVAL-015 governs
  which is required)
- NaN and infinity semantics for floating-point types (IEEE 754 compliant)
- Bit shift semantics (shift amount ≥ bit width is defined per profile)

---

## S1 — MEMORY MODEL

### RULE SUB-020: Substrate Must Define an Address Space
**Status:** DRAFT
**Layer:** SUBSTRATE / S1
**Depends on:** OWN-001–OWN-099

A conformant substrate profile must declare:

1. **Pointer width** — the size in bits of a memory address (e.g., 64 on x86_64)
2. **Alignment rules** — required alignment for each primitive type
3. **Endianness** — byte order for multi-byte types (little-endian, big-endian,
   or bi-endian with declared default)
4. **Atomic operations** — which types support atomic load/store/compare-exchange
   and at what granularity (used by Layer 14 CONC)
5. **Address space regions** — stack, heap, static, and any substrate-specific
   regions (e.g., quantum register space on a quantum substrate)

**Note on quantum substrates:**
A quantum substrate declares a separate **quantum register space** distinct from
classical memory. `qbit` values reside in quantum register space and are not
addressable by classical pointers. AXIOM-004 (explicit ownership) enforces
the no-clone theorem: a `qbit` cannot be copied because owned values are
never implicitly copied.

---

### RULE SUB-021: Stack and Heap Semantics
**Status:** DRAFT
**Layer:** SUBSTRATE / S1
**Depends on:** SUB-020, OWN-010

A conformant substrate must provide:
- A **call stack** for function frames (LIFO; frame lifetime = function lifetime)
- A **heap allocator** supporting at minimum: allocate, reallocate, free
- Heap allocation failure must be surfaceable as a `result<T, AllocError>`;
  a substrate that traps on OOM is conformant only if declared as such in its profile

---

## S2 — EXECUTION MODEL

### RULE SUB-030: Substrate Must Define a Calling Convention
**Status:** DRAFT
**Layer:** SUBSTRATE / S2
**Depends on:** FN-030, SUB-010, SUB-020

A conformant substrate profile must declare:

1. **Parameter passing** — how arguments are passed (registers, stack, or both)
2. **Return value passing** — how return values are passed back to the caller
3. **Caller/callee-saved registers** — which registers the callee must preserve
4. **Stack alignment** — required stack alignment at call sites

The Candor compiler emits code conforming to the declared calling convention.
`extern fn` declarations (FN-070) may declare an alternate calling convention
for interoperability.

---

### RULE SUB-031: Execution Units
**Status:** DRAFT
**Layer:** SUBSTRATE / S2
**Depends on:** SUB-030, EFF-001

A conformant substrate profile must declare what **execution units** it provides:

| Unit type | Effect tag | Description |
|-----------|-----------|-------------|
| Classical core | (none / default) | Standard sequential execution |
| SIMD unit | `effects(simd)` | Vector/SIMD instruction execution |
| GPU | `effects(gpu_exec)` | Massively parallel GPU execution |
| QPU (quantum) | `effects(quantum_exec)` | Quantum circuit execution |
| TPU / neural | `effects(neural_exec)` | Neural accelerator execution |

Functions targeted at a non-default execution unit must declare the corresponding
effect. A substrate that does not provide a unit must reject programs that declare
effects for it at compile time (SUB conformance check, not type check).

**Design intent:**
This is how Candor absorbs new hardware paradigms. A QPU is not a language
feature — it is a declared execution unit in a substrate profile, accessed
via an effect tag. The language layer never changes.

---

### RULE SUB-032: Quantum Execution Semantics
**Status:** DRAFT
**Layer:** SUBSTRATE / S2
**Depends on:** SUB-031, AXIOM-002, AXIOM-004, AXIOM-005

Quantum substrates must declare the following additional semantics:

1. **`qbit` type**: A quantum bit residing in quantum register space.
   Non-copyable by AXIOM-004. Non-addressable by classical pointers (SUB-020).

2. **Measurement**: Reading a `qbit` collapses its quantum state to a classical
   `bool`. Measurement is non-deterministic from a classical perspective.
   Any function that measures a `qbit` must declare `effects(quantum_measure)`.
   AXIOM-001 is preserved: measurement has exactly one *type* (`bool`), even
   though the *value* is probabilistic. Probabilistic outcome ≠ ambiguous type.

3. **Circuit decidability**: The structure of a quantum computation (which gates
   apply to which qbits in what order) is fully decidable at compile time.
   AXIOM-005 is preserved: the compiler can fully analyze and validate a quantum
   program at compile time, even though measurement outcomes are not known.

4. **No-clone enforcement**: AXIOM-004 makes `qbit` non-copyable by construction.
   A language-level copy of a `qbit` is a compile error — the quantum no-clone
   theorem is enforced without a separate rule.

---

## S3 — PLATFORM CAPABILITIES

### RULE SUB-040: Substrate Must Declare Platform Capabilities
**Status:** DRAFT
**Layer:** SUBSTRATE / S3
**Depends on:** EFF-001, SUB-031

A conformant substrate profile must declare which platform capabilities it
provides, and map each to the corresponding Candor effect tag:

| Candor Effect Tag     | Classical mapping | Notes |
|-----------------------|-------------------|-------|
| `effects(fs_read)`    | POSIX read / Win32 ReadFile | Read from filesystem |
| `effects(fs_write)`   | POSIX write / Win32 WriteFile | Write to filesystem |
| `effects(net_in)`     | POSIX recv / Win32 recv | Receive from network |
| `effects(net_out)`    | POSIX send / Win32 send | Send to network |
| `effects(proc_spawn)` | fork/exec / CreateProcess | Spawn child process |
| `effects(registry_write)` | Win32 RegSetValueEx | Write to Windows registry |
| `effects(registry_read)`  | Win32 RegQueryValueEx | Read from Windows registry |
| `effects(time)`       | clock_gettime / GetSystemTime | Read system time |
| `effects(env_read)`   | getenv / GetEnvironmentVariable | Read env vars |
| `effects(random)`     | /dev/urandom / BCryptGenRandom | Cryptographic randomness |
| `effects(quantum_exec)` | QPU dispatch | Quantum substrate only |
| `effects(quantum_measure)` | Measurement collapse | Quantum substrate only |
| `effects(simd)`       | SSE/AVX / NEON | Vector execution |
| `effects(gpu_exec)`   | CUDA / Vulkan Compute | GPU execution |

A substrate that does not provide a capability must cause the compiler to
reject programs that declare the corresponding effect (compile-time substrate
conformance check).

**Design intent:**
The effect tag is the stable, language-level name. The substrate mapping is
the hardware-specific implementation. This is the mechanism by which
`effects(net_in)` means the same thing in a Candor program regardless of
whether the substrate is Linux sockets, Windows Winsock, or a future
quantum networking protocol.

---

## SUBSTRATE PROFILES

### RULE SUB-050: A Substrate Profile Is a Named Conformant Implementation
**Status:** DRAFT
**Layer:** SUBSTRATE

A **substrate profile** is a document `SUBSTRATE-<name>.md` that:

1. Names the target (e.g., `x86_64-win64`, `arm64-linux`, `quantum-hybrid-v1`)
2. Declares all S0 type representations for this target
3. Declares all S1 memory model parameters for this target
4. Declares all S2 execution units and calling conventions for this target
5. Declares all S3 capabilities this target provides, with their platform mappings
6. Declares any **substrate extensions** — types or effects that exist on this
   target but are not in the base language spec

A Candor compiler targets exactly one substrate profile at a time.
The profile is selected at compile time (not at runtime).
AXIOM-006 (source is the authority) is preserved because the substrate
profile is a fixed, declared parameter of the compilation — not an
undeclared environment dependency.

---

### RULE SUB-051: Substrate Extensions Are Opt-In
**Status:** DRAFT
**Layer:** SUBSTRATE
**Depends on:** SUB-050, EFF-001

Substrate-specific types (e.g., `qbit`) and effects (e.g., `quantum_exec`)
are only accessible in source files that declare:

```candor
use substrate::quantum   ## or whatever the profile names its extension
```

Programs that do not import a substrate extension are portable across all
conformant substrates that satisfy the base S0–S3 interface.
Programs that import a substrate extension are portable only across substrates
that declare that extension.

The compiler enforces this at compile time (AXIOM-005).

---

## ROADMAP

This section is non-normative. It describes the intended trajectory of
substrate profile development.

### Phase 1 — Classical (current)

| Profile | Status | Notes |
|---------|--------|-------|
| `x86_64-win64` | Active | Primary development target; Win11 |
| `x86_64-linux` | Active | Linux development and CI |
| `arm64-darwin` | Planned | Apple Silicon |
| `arm64-linux`  | Planned | Raspberry Pi, embedded ARM |

These profiles share identical S0 type algebras (classical binary).
Differences are confined to S1 (calling convention, alignment) and
S3 (platform capability mappings).

### Phase 2 — Accelerated Classical (near-term)

| Profile | Status | Notes |
|---------|--------|-------|
| `x86_64-win64+simd` | Planned | Adds `effects(simd)` / AVX-512 |
| `gpu-cuda`     | Draft  | Adds `effects(gpu_exec)` via CUDA |
| `gpu-vulkan`   | Draft  | Adds `effects(gpu_exec)` via Vulkan Compute |

### Phase 3 — Quantum Hybrid (future)

| Profile | Status | Notes |
|---------|--------|-------|
| `quantum-hybrid-v1` | Research | Classical + QPU coprocessor |

A quantum hybrid substrate extends the classical substrate with:
- `qbit` type in quantum register space (S0 extension)
- Quantum register space declaration (S1 extension)
- QPU as an execution unit (S2)
- `effects(quantum_exec)`, `effects(quantum_measure)` (S3)

Classical Candor code is unaffected. Quantum code is isolated in
modules that `use substrate::quantum`.

Key insight: AXIOM-004 (explicit ownership) enforces the quantum no-clone
theorem without any new rule. The axiom that was designed for memory safety
in 2026 also enforces quantum physics in 2035. This is not coincidence —
both derive from the same first principle: values with unique identity
cannot be silently duplicated.

### Phase 4 — Alternative Arithmetic Substrates (long-term research)

| Profile | Status | Notes |
|---------|--------|-------|
| `ternary-v1`     | Research | Base-3 arithmetic substrate |
| `quaternary-v1`  | Research | Base-4 arithmetic substrate |

Ternary and quaternary substrates redefine S0 — what primitive types exist
and how arithmetic works at the physical level. The Candor language above
the boundary never says "binary." It says `i64` and `+`. The substrate
maps those to the physical representation.

For these substrates, S0 would declare:
- A `trit` / `quad` fundamental unit replacing `bit`
- Integer types with equivalent numeric ranges but different physical encoding
- Arithmetic operations with ternary/quaternary overflow semantics

The language, type system, effects, ownership, and all 7 axioms remain
unchanged. The substrate handles the physical encoding.

---

## CONFORMANCE

A compiler targeting a substrate profile is **substrate-conformant** if:

1. It maps every Candor primitive type to the S0 representation declared in the profile.
2. It generates memory accesses according to the S1 alignment and endianness rules.
3. It generates function calls and returns according to the S2 calling convention.
4. It rejects programs that declare S3 effects the profile does not provide.
5. It rejects programs that use substrate extensions the profile does not declare.
6. It accepts programs that use only base types and base effects on any conformant profile.

A program is **substrate-portable** if it uses only base types and base effects
(no substrate extensions). A substrate-portable program must compile and run
correctly on any conformant classical substrate.

---

*End of SUBSTRATE.md*
