# Benchmark Results

**Machine:** Windows 11 Pro, 6-year-old processor  
**Date:** 2026-04-09  
**Candor build:** M9.19 (fully self-hosting)  
**GCC:** 14.2 (mingw64), `-O2`  
**Python:** 3.14  
**Runs per measurement:** 5 (avg)

---

## 1. Compiler Throughput

| Input | Lines | Avg time | Throughput |
|-------|-------|----------|------------|
| Full compiler (6 files) | 5,963 | 755ms | ~7,900 lines/sec |
| Single small file (fib.cnd) | 8 | 61ms | — (startup overhead) |

**Notes:**
- ~61ms is fixed startup cost (parse + typecheck init) regardless of file size
- Marginal throughput above startup: (5963 lines) / (755 - 61)ms ≈ **8,600 lines/sec**
- The compiler is written in Candor and compiled to C by itself — no interpreter overhead

---

## 2. Runtime Performance vs Python

| Benchmark | Candor (O2) | Python 3.14 | Speedup |
|-----------|-------------|-------------|---------|
| fib(40) recursive | 293ms | 19,129ms | **65x** |
| sieve(1,000,000) | 72ms | 242ms | **3x** |

**Notes:**
- fib(40): pure recursion, 102,334,155 calls. Candor compiles to native via GCC -O2.
  Python has per-call overhead from the interpreter frame; Candor does not.
- sieve: memory-bound (1M vec writes). Python's list is native C underneath,
  so the gap is smaller. Candor still 3x faster due to tighter loop overhead.
- Candor programs are C programs under the hood — the speedup is GCC vs CPython,
  not anything special about the Candor runtime.

---

## 3. Compiler Throughput vs Typical Compilers

For context (approximate, different machines):

| Compiler | Typical throughput |
|----------|--------------------|
| Python (parse only) | ~100k lines/sec |
| TypeScript `tsc` | ~30-50k lines/sec |
| **Candor lexer.exe** | **~7,900 lines/sec** |
| Rust `rustc` | ~5-15k lines/sec |
| GCC (C) | ~50-200k lines/sec |

Candor is slower than tsc but faster than rustc on this machine. The 61ms startup
floor means it's not suited for incremental single-file compilation today.

---

## Next Measurements Needed

- [ ] Agent eval: first-attempt pass rate, tokens/working-line (see `tests/agent_eval/`)
- [ ] `-O0` vs `-O2` runtime comparison
- [ ] Compiler throughput after warm cache (OS file cache effects)
- [ ] Memory usage during self-compile
