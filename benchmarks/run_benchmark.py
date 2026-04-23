#!/usr/bin/env python3
"""
Candor Token Density Benchmark
Measures tokens consumed by Claude to produce a correct, compiling Log Batch Processor
in each of: Candor, Go, Rust, C, C++.

Usage:
    python benchmarks/run_benchmark.py [--languages candor,go,rust,c,cpp] [--dry-run]

Requires:
    pip install anthropic
    ANTHROPIC_API_KEY environment variable set
"""

import argparse
import json
import os
import subprocess
import sys
import time
from datetime import datetime
from pathlib import Path

import anthropic

REPO_ROOT = Path(__file__).parent.parent
BENCHMARK_DIR = Path(__file__).parent
CONFIG_PATH = BENCHMARK_DIR / "benchmark_config.json"
RESULTS_DIR = BENCHMARK_DIR / "results"
TEST_LOGS_DIR = BENCHMARK_DIR / "test_logs"


def load_config() -> dict:
    with open(CONFIG_PATH) as f:
        return json.load(f)


def build_user_prompt(template: str, lang: dict) -> str:
    return template.replace("{LANGUAGE}", lang["display"]).replace(
        "{EXTENSION}", lang["extension"]
    )


def build_system_prompt(base: str, lang: dict) -> str:
    """Base system prompt + optional language reference (simulates trained knowledge)."""
    ctx_file = lang.get("language_context_file")
    if ctx_file:
        ref_path = REPO_ROOT / ctx_file
        if ref_path.exists():
            ref = ref_path.read_text(encoding="utf-8")
            return base + "\n\n" + ref
    return base


def call_claude(
    client: anthropic.Anthropic,
    model: str,
    max_tokens: int,
    system: str,
    messages: list,
) -> tuple[str, int, int]:
    """Returns (content, input_tokens, output_tokens)."""
    response = client.messages.create(
        model=model,
        max_tokens=max_tokens,
        system=system,
        messages=messages,
    )
    content = response.content[0].text
    return content, response.usage.input_tokens, response.usage.output_tokens


def strip_fences(code: str) -> str:
    lines = code.splitlines()
    if lines and lines[0].startswith("```"):
        lines = lines[1:]
    if lines and lines[-1].strip() == "```":
        lines = lines[:-1]
    return "\n".join(lines).strip() + "\n"


def write_source(lang: dict, code: str, run_id: str) -> Path:
    source_path = RESULTS_DIR / f"{lang['name']}_{run_id}.{lang['extension']}"
    source_path.write_text(strip_fences(code), encoding="utf-8")
    return source_path


def candor_agent_feedback(lang: dict, source_path: Path) -> str | None:
    """Run candorc agent-json and return a compact structured error string.

    Returns None if agent-json is unavailable or produces no errors.
    The returned string is ready to embed in the correction prompt.
    """
    compiler = lang["compile"][0]
    candidate = REPO_ROOT / compiler
    if candidate.exists():
        compiler = str(candidate)

    cmd = [compiler, "agent-json", str(source_path)]
    try:
        result = subprocess.run(
            cmd, capture_output=True, text=True, cwd=str(REPO_ROOT), timeout=30
        )
        data = json.loads(result.stdout)
    except Exception:
        return None

    if data.get("status") != "error":
        return None

    errors = data.get("errors", [])
    if not errors:
        return None

    lines = ["Structured compiler errors (Candor agent-json):"]
    for e in errors:
        loc = f"{e.get('file', '?')}:{e.get('line', 0)}:{e.get('col', 0)}"
        lines.append(f"  [{e.get('rule', 'ERR')}] {loc}: {e.get('message', '')}")
        if ctx := e.get("context"):
            lines.append(f"    {ctx}")
        if hint := e.get("fix_hint"):
            lines.append(f"    Fix: {hint}")
    return "\n".join(lines)


def compile_source(lang: dict, source_path: Path) -> tuple[bool, str]:
    """Returns (success, stderr)."""
    output_path = source_path.with_suffix(f".{lang['output_ext']}")

    cmd = [
        part.replace("{source}", str(source_path)).replace(
            "{output}", str(output_path)
        )
        for part in lang["compile"]
    ]

    # Resolve compiler relative to repo root if not on PATH
    candidate = REPO_ROOT / cmd[0]
    if candidate.exists():
        cmd[0] = str(candidate)

    try:
        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            cwd=str(REPO_ROOT),
        )
        stderr = (result.stdout + "\n" + result.stderr).strip()
        return result.returncode == 0, stderr
    except FileNotFoundError:
        return False, f"compiler not found: {cmd[0]}"


def check_candor_semantics(source_path: Path) -> dict:
    """Verify generated Candor code actually uses Candor's semantic features."""
    code = source_path.read_text(encoding="utf-8")
    lines = code.splitlines()

    def count_pattern(pattern: str) -> int:
        return sum(1 for l in lines if pattern in l)

    # pure appears after return type: "-> str pure {" or "-> option<T> pure {"
    pure_fns     = count_pattern(" pure {") + count_pattern(" pure{")
    effects_fns  = count_pattern("effects(")
    must_blocks  = count_pattern("must {") + count_pattern("must{")
    result_types = count_pattern("result<")
    option_types = count_pattern("option<")

    # Minimum expectations: at least one pure fn, one effects fn, one must block
    passed = pure_fns >= 1 and effects_fns >= 1 and must_blocks >= 1

    return {
        "passed": passed,
        "pure_fns": pure_fns,
        "effects_fns": effects_fns,
        "must_blocks": must_blocks,
        "result_types": result_types,
        "option_types": option_types,
    }


def run_correctness_test(lang: dict, source_path: Path, expected: str) -> tuple[bool, str]:
    """Returns (correct, actual_output)."""
    output_path = source_path.with_suffix(f".{lang['output_ext']}")

    run_cmd = [
        part.replace("{output}", str(output_path))
        for part in lang["run"]
    ]
    run_cmd.append(str(TEST_LOGS_DIR / "app.log"))
    run_cmd.append(str(TEST_LOGS_DIR / "db.log"))

    result = subprocess.run(
        run_cmd,
        capture_output=True,
        text=True,
        cwd=str(REPO_ROOT),
    )
    actual = result.stdout.strip()
    return actual == expected.strip(), actual


def run_language(
    client: anthropic.Anthropic,
    config: dict,
    lang: dict,
    run_id: str,
    dry_run: bool = False,
) -> dict:
    print(f"\n{'='*60}")
    print(f"  {lang['display']}")
    print(f"{'='*60}")

    model = config["model"]
    max_tokens = config["max_tokens"]
    base_system = config["system_prompt"] + " Output ONLY the raw source code. Do not include any explanation, prose, or markdown code fences — just the code itself starting at line 1."
    system = build_system_prompt(base_system, lang)
    user_prompt = build_user_prompt(config["user_prompt_template"], lang)
    expected = config["expected_output"]
    max_rounds = config["max_rounds"]

    ctx_file = lang.get("language_context_file")
    ref_tokens = 0
    if ctx_file:
        ref_path = REPO_ROOT / ctx_file
        if ref_path.exists():
            ref_tokens = len(ref_path.read_text(encoding="utf-8")) // 4  # rough estimate

    result = {
        "language": lang["name"],
        "display": lang["display"],
        "model": model,
        "run_id": run_id,
        "timestamp": datetime.now().isoformat(),
        "reference_tokens_est": ref_tokens,
        "rounds": [],
        "total_input_tokens": 0,
        "total_output_tokens": 0,
        "total_tokens": 0,
        "final_compiles": False,
        "final_correct": False,
    }

    if dry_run:
        print(f"  [DRY RUN] Would send {len(user_prompt)} chars to {model}")
        print(f"  User prompt preview: {user_prompt[:120]}...")
        return result

    messages = [{"role": "user", "content": user_prompt}]

    for round_num in range(1, max_rounds + 1):
        print(f"\n  Round {round_num}...")
        t0 = time.time()

        code, input_tok, output_tok = call_claude(
            client, model, max_tokens, system, messages
        )
        elapsed = time.time() - t0

        print(f"  Received {output_tok} output tokens in {elapsed:.1f}s")

        source_path = write_source(lang, code, f"{run_id}_r{round_num}")

        compiles, compile_stderr = compile_source(lang, source_path)
        print(f"  Compile: {'OK' if compiles else 'FAIL'}")
        if not compiles:
            print(f"  Compiler output: {compile_stderr[:300]}")

        correct = False
        actual_output = ""
        semantics = {}
        if compiles:
            correct, actual_output = run_correctness_test(lang, source_path, expected)
            print(f"  Correct: {'YES' if correct else 'NO'}")
            if not correct:
                print(f"  Expected:\n    {expected}")
                print(f"  Got:\n    {actual_output[:300]}")
            if lang["name"] == "candor" and correct:
                semantics = check_candor_semantics(source_path)
                sem_ok = semantics["passed"]
                print(f"  Semantics: {'OK' if sem_ok else 'WARN'} "
                      f"(pure={semantics['pure_fns']} effects={semantics['effects_fns']} "
                      f"must={semantics['must_blocks']})")

        round_data = {
            "round": round_num,
            "input_tokens": input_tok,
            "output_tokens": output_tok,
            "compiles": compiles,
            "correct": correct,
            "semantics": semantics,
            "compile_stderr": compile_stderr if not compiles else "",
            "actual_output": actual_output if compiles and not correct else "",
        }
        result["rounds"].append(round_data)
        result["total_input_tokens"] += input_tok
        result["total_output_tokens"] += output_tok
        result["total_tokens"] += input_tok + output_tok
        result["final_compiles"] = compiles
        result["final_correct"] = correct

        if correct:
            print(f"  [OK] Done in {round_num} round(s)")
            break

        if round_num < max_rounds:
            # Build correction message for next round
            if not compiles:
                agent_fb = None
                if lang["name"] == "candor":
                    agent_fb = candor_agent_feedback(lang, source_path)
                if agent_fb:
                    feedback = (
                        f"{agent_fb}\n\n"
                        f"Output the complete corrected source file. Raw code only — no prose, no markdown fences."
                    )
                else:
                    feedback = (
                        f"That code did not compile. Compiler error:\n{compile_stderr}\n\n"
                        f"Output the complete corrected source file. Raw code only — no prose, no markdown fences."
                    )
            else:
                feedback = (
                    f"The code compiled but produced incorrect output.\n"
                    f"Expected:\n{expected}\n\nGot:\n{actual_output}\n\n"
                    f"Output the complete corrected source file. Raw code only — no prose, no markdown fences."
                )
            messages.append({"role": "assistant", "content": code})
            messages.append({"role": "user", "content": feedback})

    return result


def print_summary(results: list[dict]) -> None:
    print(f"\n\n{'='*72}")
    print("  RESULTS")
    print(f"{'='*72}")
    print(
        f"  {'Language':<10} {'Rnd':>4} {'Ref*':>6} {'Task in':>8} "
        f"{'Out tok':>8} {'Total':>8}  {'Status'}"
    )
    print(f"  {'-'*10} {'-'*4} {'-'*6} {'-'*8} {'-'*8} {'-'*8}  {'-'*10}")
    for r in results:
        ref = r.get("reference_tokens_est", 0)
        task_in = r["total_input_tokens"] - ref
        status = "correct" if r["final_correct"] else ("compiled" if r["final_compiles"] else "FAILED")
        print(
            f"  {r['display']:<10} {len(r['rounds']):>4} {ref:>6} {task_in:>8} "
            f"{r['total_output_tokens']:>8} {r['total_tokens']:>8}  {status}"
        )
    print(f"\n  * Ref = language reference doc tokens (zero in a trained model)")
    print()


def main() -> None:
    parser = argparse.ArgumentParser(description="Candor token density benchmark")
    parser.add_argument(
        "--languages",
        default="candor,go,rust,c,cpp",
        help="Comma-separated list of languages to run",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Print prompts without calling the API",
    )
    args = parser.parse_args()

    requested = set(args.languages.split(","))
    config = load_config()
    languages = [l for l in config["languages"] if l["name"] in requested]

    if not languages:
        print(f"No matching languages. Available: {[l['name'] for l in config['languages']]}")
        sys.exit(1)

    RESULTS_DIR.mkdir(exist_ok=True)

    api_key = os.environ.get("ANTHROPIC_API_KEY")
    if not api_key and not args.dry_run:
        print("Error: ANTHROPIC_API_KEY not set")
        sys.exit(1)

    client = anthropic.Anthropic(api_key=api_key or "dry-run") if not args.dry_run else None
    run_id = datetime.now().strftime("%Y%m%d_%H%M%S")

    results = []
    for lang in languages:
        try:
            result = run_language(client, config, lang, run_id, dry_run=args.dry_run)
        except Exception as e:
            print(f"\n  ERROR running {lang['display']}: {e}")
            result = {"language": lang["name"], "display": lang["display"], "error": str(e),
                      "rounds": [], "total_input_tokens": 0, "total_output_tokens": 0,
                      "total_tokens": 0, "final_compiles": False, "final_correct": False}
        results.append(result)

        # Save per-language result immediately
        result_path = RESULTS_DIR / f"{lang['name']}_{run_id}.json"
        result_path.write_text(json.dumps(result, indent=2), encoding="utf-8")

    # Save combined results
    combined_path = RESULTS_DIR / f"run_{run_id}.json"
    combined_path.write_text(json.dumps(results, indent=2), encoding="utf-8")
    print(f"\nResults saved to {combined_path}")

    if not args.dry_run:
        print_summary(results)


if __name__ == "__main__":
    main()
