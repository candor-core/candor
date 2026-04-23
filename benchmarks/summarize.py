#!/usr/bin/env python3
"""
Summarize benchmark results from benchmarks/results/*.json

Usage:
    python benchmarks/summarize.py [--run RUN_ID]
    python benchmarks/summarize.py --list
"""

import argparse
import json
from pathlib import Path

RESULTS_DIR = Path(__file__).parent / "results"

DISPLAY_ORDER = ["Candor", "Rust", "Go", "C++", "C"]


def load_run(run_id: str) -> list[dict]:
    path = RESULTS_DIR / f"run_{run_id}.json"
    if not path.exists():
        raise FileNotFoundError(f"No run found: {path}")
    with open(path) as f:
        return json.load(f)


def latest_run_id() -> str:
    runs = sorted(RESULTS_DIR.glob("run_*.json"), key=lambda p: p.stat().st_mtime)
    if not runs:
        raise FileNotFoundError("No benchmark runs found in benchmarks/results/")
    return runs[-1].stem.removeprefix("run_")


def list_runs() -> None:
    runs = sorted(RESULTS_DIR.glob("run_*.json"), key=lambda p: p.stat().st_mtime)
    if not runs:
        print("No benchmark runs found.")
        return
    print(f"{'Run ID':<20} {'Languages':<40} {'All correct'}")
    print(f"{'-'*20} {'-'*40} {'-'*11}")
    for run in runs:
        with open(run) as f:
            data = json.load(f)
        run_id = run.stem.removeprefix("run_")
        langs = ", ".join(r["display"] for r in data)
        all_correct = all(r["final_correct"] for r in data)
        print(f"{run_id:<20} {langs:<40} {'YES' if all_correct else 'NO'}")


def print_report(results: list[dict]) -> None:
    # Candor semantic summary
    candor_results = [r for r in results if r.get("language") == "candor" and r.get("final_correct")]
    if candor_results:
        cr = candor_results[0]
        for rnd in cr.get("rounds", []):
            sem = rnd.get("semantics")
            if sem:
                print(f"\n  Candor semantic check:")
                print(f"    pure functions:    {sem['pure_fns']}")
                print(f"    effects(io) fns:   {sem['effects_fns']}")
                print(f"    must{{}} blocks:    {sem['must_blocks']}")
                print(f"    result<T,E> types: {sem['result_types']}")
                print(f"    option<T> types:   {sem['option_types']}")
                print(f"    verdict:           {'PASS — uses Candor semantics' if sem['passed'] else 'WARN — missing pure/effects/must'}")

    # Sort by display order, then alphabetically for unknowns
    order = {name: i for i, name in enumerate(DISPLAY_ORDER)}
    results = sorted(results, key=lambda r: (order.get(r["display"], 99), r["display"]))

    print("\n" + "=" * 70)
    print("  TOKEN DENSITY BENCHMARK — Log Batch Processor")
    print("=" * 70)

    # Main table
    print(f"\n  {'Language':<10} {'Rounds':>6} {'Input':>8} {'Output':>8} {'Total':>8}  {'Status'}")
    print(f"  {'-'*10} {'-'*6} {'-'*8} {'-'*8} {'-'*8}  {'-'*16}")
    for r in results:
        status = "✓ correct" if r["final_correct"] else ("compiled" if r["final_compiles"] else "✗ failed")
        print(
            f"  {r['display']:<10} {len(r['rounds']):>6} "
            f"{r['total_input_tokens']:>8} {r['total_output_tokens']:>8} "
            f"{r['total_tokens']:>8}  {status}"
        )

    # Relative cost table (output tokens — the AI's work)
    print("\n  Output token comparison (AI-generated code size):")
    correct = [r for r in results if r["final_correct"]]
    if not correct:
        print("  No correct results to compare.")
        return

    best = min(correct, key=lambda r: r["total_output_tokens"])
    print(f"\n  {'Language':<10} {'Output tok':>10} {'vs best':>10}")
    print(f"  {'-'*10} {'-'*10} {'-'*10}")
    for r in correct:
        ratio = r["total_output_tokens"] / best["total_output_tokens"]
        marker = " ← baseline" if r is best else f" ({ratio:.2f}×)"
        print(f"  {r['display']:<10} {r['total_output_tokens']:>10}{marker}")

    # Total cost (input + output × rounds — the full API bill)
    print("\n  Total token cost (all rounds — the API bill):")
    best_total = min(correct, key=lambda r: r["total_tokens"])
    print(f"\n  {'Language':<10} {'Total tok':>10} {'vs best':>10}")
    print(f"  {'-'*10} {'-'*10} {'-'*10}")
    for r in correct:
        ratio = r["total_tokens"] / best_total["total_tokens"]
        marker = " ← baseline" if r is best_total else f" ({ratio:.2f}×)"
        print(f"  {r['display']:<10} {r['total_tokens']:>10}{marker}")

    # Per-round breakdown for multi-round runs
    multi = [r for r in results if len(r["rounds"]) > 1]
    if multi:
        print("\n  Multi-round detail:")
        for r in multi:
            print(f"\n  {r['display']} ({len(r['rounds'])} rounds):")
            for rnd in r["rounds"]:
                outcome = "correct" if rnd["correct"] else ("compiled" if rnd["compiles"] else "compile error")
                print(f"    Round {rnd['round']}: {rnd['input_tokens']} in + {rnd['output_tokens']} out = {rnd['input_tokens'] + rnd['output_tokens']} total — {outcome}")

    print()


def main() -> None:
    parser = argparse.ArgumentParser(description="Summarize benchmark results")
    parser.add_argument("--run", help="Run ID (default: latest)")
    parser.add_argument("--list", action="store_true", help="List all runs")
    args = parser.parse_args()

    if args.list:
        list_runs()
        return

    run_id = args.run or latest_run_id()
    print(f"  Run: {run_id}")
    results = load_run(run_id)
    print_report(results)


if __name__ == "__main__":
    main()
