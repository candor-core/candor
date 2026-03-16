# Candor Compiler Roadmap

> **Living document.** Updated as milestones complete. Each item maps to concrete compiler work.

---

## Current State (v0.1 + post-release work)

### Fully implemented
| Feature | Pipeline stage |
|---|---|
| All primitive types (i8–i128, u8–u128, f32, f64, bool, str, unit, never) | typeck, emit_c |
| vec\<T\>, map\<K,V\>, ring\<T\>, set\<T\> with full operation suites | typeck, emit_c |
| option\<T\>, result\<T,E\> with `must{}` pattern | typeck, emit_c |
| ref\<T\>, refmut\<T\> ownership / borrow model | typeck, emit_c |
| secret\<T\> / reveal information-flow enforcement | typeck, emit_c |
| Structs, enums (sum types with data), struct literals | typeck, emit_c |
| impl blocks + method call syntax (`x.method(args)`) | typeck, emit_c |
| Generic functions (monomorphization at call sites) | typeck, emit_c |
| Lambdas / closures (by-value and by-ref capture, fat-pointer calling convention) | typeck, emit_c |
| Effects system (effects(io), pure, cap, enforcement in call graph) | typeck, emit_c |
| Contracts (requires, ensures, assert — compiled to C assert()) | typeck, emit_c |
| Pattern matching: enums, option, result, bool, int/str literals, wildcards | typeck, emit_c |
| if/else, loop, while, for…in (vec, map, ring), break, continue, return | typeck, emit_c |
| Module system: `module`, `use mod::Name`, multi-file compilation | typeck, emit_c |
| extern fn (C interop, no-body declarations) | typeck, emit_c |
| const declarations (static compile-time values) | typeck, emit_c |
| as cast operator, implicit numeric widening (i8→i64, f32→f64, etc.) | typeck, emit_c |
| Vec literals `[a, b, c]` | typeck, emit_c |
| Tuple types `(T, U)`, tuple literals `(a, b)`, index access `.0` `.1`, destructuring | typeck, emit_c |
| Struct update syntax `Point { ..base, x: 1 }` | typeck, emit_c |
| Compound assignment `+=`, `-=`, `*=`, `/=`, `%=` (scalars and collections) | parser, typeck, emit_c |
| Map index assignment `m[k] = v` | typeck, emit_c |
| Compile-time evaluation of pure calls with constant args | typeck, emit_c |
| i128 / u128 via compiler `__int128` extension | emit_c |
| Stdin I/O: read_line, read_int, read_f64, try_read_* (EOF-safe) | typeck, emit_c |
| File I/O: read_file, write_file, append_file | typeck, emit_c |
| String builtins: concat, len, eq, starts_with, find, substr, from_u8, to_int | typeck, emit_c |
| M2 std lib: math, str, io, os, time, rand, path | typeck, emit_c |
| M3 Trait / Interface system: trait decl, impl Trait for Type, trait bounds on generics | lexer, parser, typeck, emit_c |
| M4.1 Diagnostic quality: source snippets + carets, did-you-mean hints, unused-var & shadow warnings, multi-error | diagnostics, typeck, main |
| M4.2 Build system: `Candor.toml`, `candorc build [--release]`, auto source discovery | manifest, main |
| M4.3 LSP server: `candorc lsp`, JSON-RPC 2.0, diagnostics, hover, go-to-def, completion | lsp |
| M4.4 Formatter: `candorc fmt`, AST pretty-printer, idempotent canonical output | emit_c |
| M4.5 Test framework: `#test` directive, `candorc test` runner, pass/fail harness | parser, emit_c, main |

### Known gaps in the current compiler
- Named-return / early-exit in closures
- Invariant clauses (token exists, not wired)
- forall / exists (tokens exist for spec only, not runtime)

---

## M1 — Language Core Completeness

> Goal: close the gap between "tokens exist" and "fully working." No new language surface; finish what was started.

### M1.1 Compound assignment operators
`+=`, `-=`, `*=`, `/=`, `%=` — tokens lexed, need parser + typeck + emit_c.
- Parser: add `CompoundAssignStmt` AST node or reuse `AssignStmt` with operator field
- Typeck: check mutability + numeric type match
- Emit C: `name op= val;`

### M1.2 Closures by reference
Current closures only capture by value. Add `ref<T>` and `refmut<T>` capture support.
- Typeck: in `inferLambdaExpr`, when captured var is a ref type, keep the pointer not the value
- Emit C: capture struct stores `T*` / `T*`, unpack with deref

### M1.3 Tuple destructuring
`let (a, b) = some_tuple` and tuple patterns in match arms.
- Parser: `DestructureTupleStmt`, tuple pattern in `parsePattern`
- Typeck: infer element types from `TupleType`
- Emit C: emit individual field accesses `._0`, `._1`, etc.

### M1.4 Struct update syntax
`let p2 = Point { ..p, x: 5.0 }` — creates a new struct copying fields not listed.
- Parser: `StructLitExpr` with optional spread field `..base`
- Typeck: verify base struct type matches, override listed fields
- Emit C: emit all fields, using base value for unlisted ones

### M1.5 Index assignment for maps and compound collections
`m[key] = value` for map (currently only vec has IndexAssignStmt support).
- Typeck: extend `checkIndexAssignStmt` to handle `map<K,V>` (call `map_insert`)
- Emit C: desugar to `map_insert(&m, key, value)`

### M1.6 Ring iteration
`for x in my_ring { }` — not yet wired in typeck's `checkForStmt`.
- Typeck: extend the for-stmt collection type check to include `ring<T>`
- Emit C: emit index-based loop over `._len` elements using modular index

### M1.7 Compound assignment for collections
`v[i] += delta` — sugar for read-modify-write without a temp binding.
- Parser: detect compound assignment on index targets → `IndexCompoundAssignStmt`
- Typeck: check element type is numeric, check mutability
- Emit C: `(v)._data[i] += delta`

---

## M2 — Standard Library (`std`)

> Goal: useful programs without reaching for extern fn. Implemented as Candor files compiled into the standard module path.

Each module is a `.cnd` file in `std/` compiled to C headers; the build tool links them.

### M2.1 `std::math`
```candor
fn abs(x: f64) -> f64 pure effects []
fn sqrt(x: f64) -> f64 pure effects []   requires x >= 0.0
fn pow(base: f64, exp: f64) -> f64 pure effects []
fn floor(x: f64) -> f64 pure effects []
fn ceil(x: f64) -> f64 pure effects []
fn sin(x: f64) -> f64 pure effects []
fn cos(x: f64) -> f64 pure effects []
fn min_i64(a: i64, b: i64) -> i64 pure effects []
fn max_i64(a: i64, b: i64) -> i64 pure effects []
fn clamp(v: f64, lo: f64, hi: f64) -> f64 pure effects []
```
Emit C: thin wrappers around `<math.h>`.

### M2.2 `std::str` (extended string ops)
```candor
fn str_repeat(s: str, n: i64) -> str pure effects []
fn str_trim(s: str) -> str pure effects []
fn str_split(s: str, sep: str) -> vec<str> effects []
fn str_replace(s: str, from: str, to: str) -> str pure effects []
fn str_to_upper(s: str) -> str pure effects []
fn str_to_lower(s: str) -> str pure effects []
fn str_contains(s: str, sub: str) -> bool pure effects []
fn str_format(fmt: str, args: vec<str>) -> str pure effects []
```

### M2.3 `std::io` (I/O utilities)
```candor
fn print_err(s: str) -> unit effects(io)
fn print_line(s: str) -> unit effects(io)   // alias for print
fn read_all_lines() -> vec<str> effects(io)
fn read_csv_line() -> vec<str> effects(io)
fn flush_stdout() -> unit effects(io)
```

### M2.4 `std::os` (process and env)
```candor
fn args() -> vec<str> effects(io)
fn getenv(name: str) -> option<str> effects(io)
fn exit(code: i32) -> never effects(io)
fn cwd() -> result<str, str> effects(io)
```
Emit C: wrappers around `stdlib.h`, `unistd.h`.

### M2.5 `std::time`
```candor
fn now_ms() -> i64 effects(io)      // wall clock, milliseconds since epoch
fn now_mono_ns() -> i64 effects(io) // monotonic nanoseconds (for benchmarks)
fn sleep_ms(ms: i64) -> unit effects(io)
```

### M2.6 `std::rand`
```candor
fn rand_u64() -> u64 effects(io)
fn rand_f64() -> f64 effects(io)   // [0, 1)
fn rand_range(lo: i64, hi: i64) -> i64 effects(io)
fn set_seed(s: u64) -> unit effects(io)
```

### M2.7 `std::path`
```candor
fn path_join(a: str, b: str) -> str pure effects []
fn path_dir(p: str) -> str pure effects []
fn path_filename(p: str) -> str pure effects []
fn path_ext(p: str) -> str pure effects []
fn path_exists(p: str) -> bool effects(io)
fn list_dir(p: str) -> result<vec<str>, str> effects(io)
fn mkdir(p: str) -> result<unit, str> effects(io)
fn remove_file(p: str) -> result<unit, str> effects(io)
```

---

## M3 — Trait / Interface System

> Goal: polymorphism without monomorphization explosion; enable `std` to define reusable contracts.

This is the largest single language addition.

### Design
```candor
trait Display {
    fn fmt(self: ref<Self>) -> str
}

trait Hash {
    fn hash(self: ref<Self>) -> u64
}

impl Display for Point {
    fn fmt(self: ref<Point>) -> str {
        return str_concat("Point(", str_concat(int_to_str(self.x), ")"))
    }
}

fn print_it<T: Display>(val: ref<T>) -> unit effects(io) {
    print(val.fmt())
}
```

### Compiler work
1. **Lexer/parser**: `trait` keyword, `for` in impl (`impl Trait for Type`), `Self` type alias, `<T: Trait>` bounds in fn and impl declarations
2. **AST**: `TraitDecl`, `ImplForDecl` (separate from `ImplDecl`)
3. **Typeck pass 1**: collect trait definitions, collect trait impls per type
4. **Typeck pass 2**: resolve trait bounds on generic functions; dispatch to the right impl
5. **Emit C**: trait calls → vtable struct or direct monomorphized dispatch (static dispatch first, vtable later)

### Built-in traits to define
- `Display` — `fmt(self) -> str`
- `Debug` — `debug(self) -> str`
- `Hash` — `hash(self) -> u64`
- `Eq` — `eq(self, other: ref<Self>) -> bool`
- `Ord` — `cmp(self, other: ref<Self>) -> i32`
- `Clone` — `clone(self) -> Self`
- `Default` — `default() -> Self`
- `Iterator<T>` — `next(self: refmut<Self>) -> option<T>`

---

## M4 — Developer Experience & Tooling

### M4.1 Diagnostic quality
- Rich error messages with source snippets and carets (like Rust's `rustc`)
- Suggestions: "did you mean X?", "add `mut` here", "missing use import"
- Warning system (unused variables, unreachable arms, shadowed bindings)
- Multiple errors collected per file (not bail-on-first)

### M4.2 Build system (`candorc build`)
Replace the bare `candorc file.cnd` with a project-aware build tool:
- `Candor.toml` project manifest (name, version, dependencies, entry point)
- Incremental compilation (only re-check changed files)
- Dependency resolution (local paths for now, registry later)
- `candorc test` — run test modules
- `candorc fmt` — canonical formatter

### M4.3 LSP server (`candorc lsp`)
- Diagnostics (type errors, effects violations, unused imports) in-editor
- Hover: show inferred type of any expression
- Go-to-definition for functions, structs, consts
- Autocomplete: struct fields, method names, import paths
- Code actions: add missing use, wrap in must{}, annotate effects

### M4.4 Formatter (`candorc fmt`)
Opinionated, deterministic, zero-config:
- 4-space indent
- Trailing commas in struct/enum/fn param lists
- `{` on same line
- Blank line between top-level declarations
- Normalize keyword spacing

### M4.5 Test framework
```candor
module tests::math

fn test_abs() -> unit {
    assert abs(-5.0) == 5.0
    assert abs(3.0) == 3.0
}
```
- `#test` directive marks test functions
- `candorc test` discovers and runs them, reports pass/fail counts

---

## M5 — Performance & Backends

### M5.1 LLVM backend ✅
Textual LLVM IR emitter (`compiler/emit_llvm/emit_llvm.go`):
- No CGo, no LLVM dev libraries — emits `.ll` text, compiled with `clang`
- `candorc build --backend=llvm` flag; `CLANG` env var override
- Implemented: all primitive types, arithmetic/comparison/bitwise ops, control flow
  (if/else/while/loop/break/continue), let bindings (alloca-based), structs
  (insertvalue/extractvalue), enums (tagged union), string globals, extern fn
  declarations, casts (trunc/sext/zext/sitofp/fptosi/etc.), function calls,
  method calls, generic instances, tuple literals, assert→llvm.trap
- `fn main()->unit` → `@_cnd_main` + `@main` wrapper returning `i32 0`
- Deferred to M5.2+: closures/lambdas, for-in, vec/map runtime, match payload binding

Priority ordering for remaining IR lowering:
1. Closures (fat pointer) — M5.2
2. Collections (vec, map via runtime) — M5.3
3. Debug info (DWARF) — M5.4
4. Cross-compilation via target triple — M5.5

### M5.2 Debug builds vs release builds ✅
Three build modes exposed via `candorc build` flags:

| Flag | CC flags | Assertions |
|---|---|---|
| *(default)* | *(none)* | on (`assert.h` active) |
| `--debug` | `-g -O0` | on (+ DWARF debug info) |
| `--release` | `-O2 -DNDEBUG` | off (stripped by `NDEBUG`) |

- `BuildConfig` struct with `ccFlags() []string` shared between C and LLVM backends
- Direct file mode (`candorc file.cnd`) also accepts `--debug` / `--release`
- Flags may be combined with `--backend=llvm`

### M5.3 Sanitizer integration ✅
`--sanitize=<kind>` flag enables runtime instrumentation via compiler sanitizers:

| Value | CC flags added |
|---|---|
| `address` | `-fsanitize=address -fno-omit-frame-pointer` |
| `undefined` (or `ub`) | `-fsanitize=undefined` |
| `memory` | `-fsanitize=memory` |
| `leak` | `-fsanitize=leak` |
| `thread` | `-fsanitize=thread` |

- Comma-separated combinations: `--sanitize=address,undefined`
- Any sanitizer implies `-g -O0` unless `--release` is given
- Parsed by `parseSanitizers(args)` into `BuildConfig.Sanitizers []string`
- Shared between C backend (`runCompile`) and LLVM backend (`runCompileLLVM`)

### M5.4 Cross-compilation
- `--target=aarch64-linux`, `--target=x86_64-windows`, etc.
- Leverage LLVM's target triple system
- Cross-compile standard library

### M5.5 WebAssembly target
- Emit WASM via LLVM wasm32-unknown-unknown
- Provide Candor↔JS interop layer via `extern fn`
- `std::wasm` module with browser bindings

---

## M6 — Formal Verification

> Goal: move contracts from runtime assertions to compile-time proofs where possible.

### M6.1 Symbolic contract evaluation
For simple arithmetic contracts (`requires x > 0`, `ensures result >= 0`):
- Evaluate at compile time when all arguments are constants
- Already partially done via `ComptimeValues`; extend to contract conditions

### M6.2 SMT integration (Z3 / CVC5)
- Translate `requires`/`ensures` clauses to SMT-LIB queries
- Query solver at compile time for pure functions
- Report: "this requires clause can never be violated" or "counterexample found"

### M6.3 Refinement types
```candor
type NonZero = i64 where self != 0
type Percentage = f64 where self >= 0.0 and self <= 1.0

fn safe_div(a: i64, b: NonZero) -> i64 { return a / b }
```
- Type alias with predicate; compiler verifies predicate at assignment site
- Propagates through type system without runtime cost when provable

### M6.4 `forall` / `exists` runtime support
The tokens already exist for spec-level quantifiers. Wire them to:
- Runtime assertion generation in debug mode
- SMT queries in verification mode

---

## M7 — AI Integration Layer

> Goal: make Candor the canonical language for agentic AI pipelines.

### M7.1 MCP-native annotations
```candor
#[mcp_tool(name="search", schema="...")]
fn search(query: str) -> result<str, str> effects(io) { ... }
```
- `#[mcp_tool]` directive generates JSON Schema and MCP tool manifest from fn signature
- `candorc mcp` emits `tools.json` alongside the binary

### M7.2 Semantic context embedding
```candor
#[intent("Computes the edit distance between two strings")]
fn levenshtein(a: str, b: str) -> i64 pure effects [] { ... }
```
- `#[intent]` attaches natural-language descriptions to functions
- `candorc doc` extracts these into a semantic context file
- Context file is machine-readable for RAG, embedding, or tool-use

### M7.3 Effects as capability tokens
Expose `effects()` annotations as first-class runtime tokens:
```candor
fn run_with_io<F: fn() -> unit effects(io)>(f: F, cap: cap<io>) -> unit { f() }
```
- A `cap<io>` value is a proof that the caller has been granted the capability
- Passed explicitly; cannot be forged; enables sandbox enforcement

### M7.4 `#[export_json]` for typed interfaces
```candor
#[export_json]
struct Config {
    name: str,
    limit: i64,
    tags: vec<str>,
}
```
- Auto-generate `config_from_json(str) -> result<Config, str>` and `config_to_json(Config) -> str`
- Useful for AI agents exchanging structured data without FFI boilerplate

---

## M8 — Ecosystem

### M8.1 Package registry
- Hosted at `candorpkg.io` (future)
- `Candor.toml` declares `[dependencies]` by name and version semver
- `candorc fetch` downloads and pins to `Candor.lock`
- Local cache at `~/.candor/pkg/`

### M8.2 C/C++ interop improvements
- `#[c_header("foo.h")]` auto-generates extern fn stubs from C headers
- Struct layout compatibility for plain-old-data structs

### M8.3 Documentation generator
- `candorc doc --html` generates HTML reference from `///` doc comments
- Extracts `#[intent]` annotations, function signatures, effects, contracts

---

## Milestone Timeline (rough, not calendar-bound)

```
Done:    v0.1  — Core language complete, full C emission, closures, effects, contracts
         M1    — Compound assignment, tuple destructuring, struct update, map index assign,
                 ring iteration, closures by reference
         M2    — Standard library (math, str, io, os, time, rand, path)
         M3    — Trait system (trait decl, impl Trait for Type, trait bounds on generics,
                 static dispatch via monomorphization)

Next:    M4.1  — Diagnostic quality (rich errors, warnings)
         M4.1  — Diagnostic quality (rich errors, warnings)
         M4.2  — Build system (Candor.toml, incremental)
         M4.3  — LSP server (hover + errors)
         M4.4  — Formatter

Later:   M5.1  — LLVM backend
         M5.4  — Cross-compilation
         M6.1  — Symbolic contract evaluation
         M6.2  — SMT integration

Future:  M5.5  — WebAssembly
         M6.3  — Refinement types
         M7.x  — AI integration layer
         M8.x  — Package registry
```

---

## Contribution priorities (for new contributors)

| Item | Difficulty | Impact | Start here |
|---|---|---|---|
| Compound assignment `+=` etc. | Low | High | M1.1 |
| Ring iteration | Low | Medium | M1.6 |
| `std::math` | Low | High | M2.1 |
| `std::str` extensions | Low | High | M2.2 |
| Rich error messages | Medium | Very high | M4.1 |
| Formatter | Medium | High | M4.4 |
| Test framework (`#test`) | Medium | High | M4.5 |
| Struct update syntax | Medium | Medium | M1.4 |
| Trait system | High | Very high | M3 |
| LLVM backend | Very high | Very high | M5.1 |

---

*Candor is open source. This roadmap reflects current priorities and will shift as the language grows. Items are not promises — they are intentions.*
