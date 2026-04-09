# Next Session Handoff
**Written: 2026-04-08 22:00 MDT**

Read this first. It tells you exactly where we are and exactly what to do next.

---

## Current State — M9.18 Achieved

Stage 2 bootstrap is working.

- `stage2.c` compiles to `stage3.exe` with **0 GCC errors** ✅
- `stage3.exe` runs and produces `stage4.c` with **exit 0, 11,616 lines** ✅
- `stage4.c` has **1 GCC error** and ~150-line diff vs stage2.c ❌ (TASK-10)

The one remaining gap is not a crash — it's a correctness issue in the void-suffix detection that causes `emit_count` to be initialized from a void expression in stage4.c.

---

## The One Bug (TASK-10)

**Root cause:** `emit_fn_body` in `src/compiler/emit_c.cnd` uses a brittle string-suffix check to decide whether to emit `return expr` or `(void)(expr)` for a function's tail expression. It checks if the emitted string ends with `"((void)0);\n}))"`. For the `emit_count` initializer in `emit_block_expr`, this suffix check fires incorrectly — the expression is NOT void, but stage3's emission produces a string that happens to match the pattern.

**GCC error:**
```
D:/tmp/stage4_compile.c:9059:28: error: void value not ignored as it ought to be
int64_t emit_count = (__extension__ ({
    ...
    (void)((__extension__ ({   ← stage3 wrapped inner match in void
```

**Where the bug is:** `src/compiler/emit_c.cnd` — `emit_fn_body` around line 1153:
```candor
let void_sfx = "((void)0);\n}))"
let e_len = str_len(e_str)
let vs_len = str_len(void_sfx)
let mut is_void: bool = false
if e_len >= vs_len {
    is_void = str_eq(str_substr(e_str, e_len - vs_len, vs_len), void_sfx)
}
```

**The fix:** Replace this string-suffix check with an AST-level check. Before emitting `e_str`, call `arm_is_terminal(e_node)` on the original `Expr` node. If the expression is terminal, emit `(void)`. Otherwise emit `return`. This is clean and correct.

```candor
## Replace the string-suffix block with:
let is_void = arm_is_terminal(e_node)
if is_void {
    emb_line(buf, str_concat("    ", str_concat("(void)(", str_concat(e_str, ");"))))
} else {
    emb_line(buf, str_concat("    ", str_concat("return ", str_concat(e_str, ";"))))
}
```

**How to verify the fix:** After applying, rebuild and regenerate:
```bash
diff /d/tmp/stage2.c /d/tmp/stage4.c
```
Should produce 0 lines of diff (full idempotency).

---

## How to Rebuild and Test

```bash
# 1. Edit src/compiler/emit_c.cnd to fix emit_fn_body void-suffix check

# 2. Rebuild Candor binary
./candorc-stage1-rebuilt.exe src/compiler/lexer.cnd src/compiler/parser.cnd \
  src/compiler/typeck.cnd src/compiler/emit_c.cnd \
  src/compiler/manifest.cnd src/compiler/main.cnd

# 3. Generate stage2.c
./src/compiler/lexer.exe src/compiler/lexer.cnd src/compiler/parser.cnd \
  src/compiler/typeck.cnd src/compiler/emit_c.cnd \
  src/compiler/manifest.cnd src/compiler/main.cnd > /d/tmp/stage2.c 2>/dev/null

# 4. Re-append runtime macros (VSCode strips them — see docs/known_compiler_bugs.md Bug 1)
#    Full block in docs/AI_GUIDE.md Step 4

# 5. Compile
PATH="/c/msys64/mingw64/bin:$PATH" /c/msys64/mingw64/bin/gcc.exe \
  -std=gnu23 -O0 -o /d/tmp/stage3.exe /d/tmp/stage2.c -I src/compiler -lm

# 6. Test stage3
/d/tmp/stage3.exe src/compiler/lexer.cnd src/compiler/parser.cnd \
  src/compiler/typeck.cnd src/compiler/emit_c.cnd \
  src/compiler/manifest.cnd src/compiler/main.cnd > /d/tmp/stage4.c
echo "EXIT: $?"; wc -l /d/tmp/stage4.c

# 7. Compile stage4 (must be 0 errors)
cat >> src/compiler/_cnd_runtime.h << 'MACROS'
[see AI_GUIDE.md Step 4]
MACROS
PATH="/c/msys64/mingw64/bin:$PATH" gcc.exe -std=gnu23 -O0 \
  -o /d/tmp/stage4.exe /d/tmp/stage4.c -I src/compiler -lm

# 8. Verify idempotency
diff /d/tmp/stage2.c /d/tmp/stage4.c
```

Success = 0 diff lines.

---

## Key Files

| File | Role |
|------|------|
| `src/compiler/emit_c.cnd` | **Primary fix target** — `emit_fn_body` void-suffix check |
| `src/compiler/lexer.exe` | Current Candor-compiled binary |
| `src/compiler/_cnd_runtime.h` | C runtime — map macros get stripped by VSCode, see Bug 1 |
| `docs/AI_GUIDE.md` | Exact commands, GCC facts, two-compiler rule |
| `docs/known_compiler_bugs.md` | All known bugs with timestamps and root causes |
| `Agents_Collab.md` | Active task list — TASK-10 is the only open bootstrap task |

---

## Do Not Forget

- GCC requires `-std=gnu23` (auto type deduction)
- GCC requires `PATH="/c/msys64/mingw64/bin:$PATH"` (assembler/linker lookup)
- `_cnd_runtime.h` map macros must be re-appended before every gcc compile
- Source file order matters: `lexer parser typeck emit_c manifest main`
- The Candor binary output is named `lexer.exe` (candorc picks the first source file's name)

---

## After Idempotency Is Confirmed

1. Run `stage4.exe` on the same inputs — verify it produces identical output to `stage3.exe`
2. Update `Agents_Collab.md` TASK-10 as Done
3. Update `docs/roadmap.md` with M9.19 — Full bootstrap idempotency verified
4. Commit with message `feat: M9.19 full bootstrap idempotency — stage4.c == stage2.c`
5. Post to Reddit
