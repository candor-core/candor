# CandorCore Trust Model

**Version 0.1 | April 2026**

---

## The Problem

Trust in software has two failure modes, and AI-assisted development has made both worse.

**Failure mode 1: Hidden side effects.** A function does something it doesn't say it does. It writes to disk. It makes a network call. It modifies global state. In most languages, you cannot know this without reading the implementation. In large codebases, nobody reads the implementation. In AI-generated code, nobody wrote the implementation with full understanding either.

**Failure mode 2: Swallowed errors.** An operation fails silently. The return value is ignored. The null pointer is dereferenced later. The wrong default is used. In most languages, ignoring a failure is valid syntax. The compiler doesn't care.

These aren't new problems. What's new is scale. A developer working with an AI coding assistant can accept a 200-line function in seconds. The speed of adoption has outrun the speed of review. The gap between "what the code does" and "what the developer thinks it does" has never been wider.

CandorCore is a direct answer to this gap. Not a linter. Not a style guide. A structural answer: make the gap impossible to close without explicit acknowledgment.

---

## The Five Structural Guarantees

### 1. Every side effect is declared and compiler-enforced

```candor
fn read_config(path: str) -> result<str, str> effects(fs) {
    return read_file(path)
}
```

A function that reads from the filesystem must declare `effects(fs)`. A function that makes network calls must declare `effects(network)`. A function declared `pure` cannot call any function that carries any effect. The compiler enforces this through the entire call graph.

What this means in practice: if you see a function with no `effects` declaration, you know it cannot touch the filesystem, the network, the environment, or spawn processes. That is a verifiable guarantee, not a comment.

### 2. Errors cannot be ignored

```candor
let data = read_config("app.toml") must {
    ok(v)  => v
    err(e) => return err(str_concat("config error: ", e))
}
```

Discarding a `result<T, E>` or `option<T>` without a `must{}` block is a compile error. There are no exceptions to this rule. You cannot shadow it with a cast. You cannot opt out of it per-call-site. Every failure path in the program is visible at every call site.

### 3. Preconditions are machine-readable, not comments

```candor
fn divide(a: i64, b: i64) -> i64
    requires b != 0
{
    return a / b
}
```

A `requires` clause is not a doc comment. It is not a convention. It is a machine-readable precondition that generates an assertion in debug builds and is visible to any tool that reads the source. An AI agent writing a caller of `divide` sees `requires b != 0` in the function signature — not buried in documentation that may or may not exist.

### 4. The ecosystem names who wrote it and whether it was audited

Module names in the CandorCore ecosystem carry structural information:

- `cc-module` — Core team, meets certification requirements by definition
- `ccPar-Nvidia-tensor` — Formal partner, full audit + partnership verified
- `ccMod-alice-fastmath` — Community author, no endorsement; may carry a cert if the author paid for an audit

The name tells you the governance tier. The cert tells you whether a human auditor checked the module against the published checklist. Neither is a quality guarantee — but both are honest claims about provenance and process.

### 5. The runtime refuses to run flagged modules without explicit consent

If a certified module is later found to have been compromised or to violate the certification checklist:

- The cert is immediately revoked
- All subscribers are notified within 24 hours
- The runtime blocks execution of the flagged module
- Execution is only possible in explicitly user-initiated diagnostic or developer mode

This is the difference between "unsigned installer" (warning) and "known malicious" (hard block). The system distinguishes them structurally.

---

## Why This Matters for AI-Generated Code

The security research community has demonstrated that frontier AI models can find and exploit real vulnerabilities — a 27-year-old OpenBSD kernel bug, a 16-year-old FFmpeg flaw — at a scale and speed that outpaces human review. The question is not whether AI will write code with security consequences. It already does. The question is whether the language gives reviewers — human or AI — the information they need to catch those consequences.

CandorCore's structural guarantees directly address the audit surface:

| Question | How Candor answers it |
|----------|-----------------------|
| Does this function touch the filesystem? | Check for `effects(fs)` in the signature. If absent: no. |
| Can this function fail silently? | If it returns `result<T,E>`, all callers must handle both arms. |
| What inputs does this function reject? | Read the `requires` clause — it is in the signature. |
| Who wrote this module? | Read the prefix: `cc-`, `ccPar-`, or `ccMod-username`. |
| Was this module audited? | Check the cert status. |
| Is this module flagged? | The runtime tells you at load time. |

These are not answers that require reading the implementation. They are answers that live in the declaration surface — the part that is always visible, always current, and always machine-parseable.

An AI agent generating a call to `read_config` sees `effects(fs)` and knows: this will affect the filesystem. It cannot pretend otherwise. A human reviewer sees the same thing in the same place. The information is not hidden anywhere.

---

## What the Trust Model Does Not Claim

Being explicit about scope is itself part of the trust model.

**Correctness** is not guaranteed. A certified module may have bugs. `requires` clauses are a declaration, not a proof. The compiler emits assertions in debug builds; it does not prove invariants.

**Completeness** is not guaranteed. A function that only declares `effects(fs)` might also have logical errors. The absence of `effects(network)` means no network calls happen — but the filesystem operations might still be wrong.

**Performance** is not guaranteed. The trust model says nothing about how fast code runs or how much memory it uses.

**Future behavior** is not guaranteed. Certification applies to a specific release. A new major release requires a full re-audit.

The trust model is a claim about **transparency**. A certified module does not hide what it does. If it later develops a vulnerability, the cert lapses or is revoked. Core stands behind that claim.

---

## The AI Collaboration Case

CandorCore was designed in part to make AI-assisted development auditable. This is not a defensive claim about AI — it is a design goal. An AI agent working in Candor cannot accidentally introduce a hidden network call in a function not declared to have network effects. The compiler catches it. An AI agent generating error handling cannot silently discard a `result` — the compiler catches it.

This means the review burden on humans working with AI-generated Candor is structurally lower than the review burden on humans working with AI-generated code in languages without these guarantees. The explicit surface of a Candor program is also the complete surface. There is no implicit layer to check.

The ecosystem naming convention extends this to third-party modules: a `ccMod-` module from an unknown author is not hidden behind a generic package name. The name itself says: community author, no Core relationship. The presence or absence of a cert says: audited, or not. That information is always in the name — not in a registry, not in a separate metadata file.

---

## The Revocation Promise

Core commits to:

1. Notifying all subscribers of a cert revocation within 24 hours
2. Providing best-effort support to help a revoked module return to certified condition
3. Treating revocation as a hard runtime block, not a soft warning
4. Publishing revocation records (the fact of revocation, not the full audit details) publicly

This is the operational complement to the structural guarantees. Structure prevents hidden behavior from being written. Revocation prevents hidden behavior from staying hidden after the fact.

---

*The cert is not a trophy. It is a claim. Core stands behind that claim.*
