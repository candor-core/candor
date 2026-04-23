#!/usr/bin/env python3
"""
Candor Tokenizer Alignment Benchmark

Measures real Claude token counts for Candor language constructs.
Tracks which keywords are single tokens, quantifies Agent Form savings,
and flags regressions when run against a new model.

Usage:
    python benchmarks/tokenizer/token_analysis.py
    python benchmarks/tokenizer/token_analysis.py --model claude-opus-4-7
    python benchmarks/tokenizer/token_analysis.py --compare results/2026-04-22.json

Requires: pip install anthropic
          ANTHROPIC_API_KEY environment variable set
"""

import argparse
import json
import os
import sys
import time
from datetime import datetime, timezone
from pathlib import Path

import anthropic

REPO_ROOT = Path(__file__).parent.parent.parent
TOKENIZER_DIR = Path(__file__).parent
CONSTRUCTS_PATH = TOKENIZER_DIR / "constructs.json"
RESULTS_DIR = TOKENIZER_DIR / "results"


def count_tokens(client: anthropic.Anthropic, model: str, text: str, baseline: int) -> int:
    """Count tokens for a single construct, subtracting the per-request overhead."""
    resp = client.messages.count_tokens(
        model=model,
        messages=[{"role": "user", "content": "X " + text}]
    )
    return resp.input_tokens - baseline


def get_baseline(client: anthropic.Anthropic, model: str) -> int:
    """Count tokens for the minimal request ('X') to establish overhead."""
    resp = client.messages.count_tokens(
        model=model,
        messages=[{"role": "user", "content": "X"}]
    )
    return resp.input_tokens


def measure_group(client, model, baseline, group):
    """Measure all items or comparisons in a group. Returns structured results."""
    results = {}
    alert_threshold = group.get("alert_if_above", None)

    if "items" in group:
        for item in group["items"]:
            n = count_tokens(client, model, item, baseline)
            results[item] = {"tokens": n, "alert": alert_threshold and n > alert_threshold}
            time.sleep(0.05)  # avoid rate limits on rapid small calls

    if "comparisons" in group:
        for comp in group["comparisons"]:
            v_tok = count_tokens(client, model, comp["verification"], baseline)
            a_tok = count_tokens(client, model, comp["agent_form"], baseline)
            savings = round((v_tok - a_tok) / v_tok * 100, 1) if v_tok > 0 else 0.0
            results[comp["label"]] = {
                "verification": comp["verification"],
                "verification_tokens": v_tok,
                "agent_form": comp["agent_form"],
                "agent_form_tokens": a_tok,
                "tokens_saved": v_tok - a_tok,
                "savings_pct": savings,
            }
            time.sleep(0.05)

    return results


def run_benchmark(model: str) -> dict:
    client = anthropic.Anthropic()
    constructs = json.loads(CONSTRUCTS_PATH.read_text(encoding="utf-8"))

    print(f"Model: {model}")
    print(f"Establishing baseline...", flush=True)
    baseline = get_baseline(client, model)
    print(f"Baseline overhead: {baseline} tokens\n")

    all_results = {
        "timestamp": datetime.now(timezone.utc).isoformat(),
        "model": model,
        "baseline_overhead": baseline,
        "groups": {}
    }

    for group in constructs["groups"]:
        name = group["name"]
        desc = group["description"]
        print(f"  Measuring: {name}...", end="", flush=True)
        group_results = measure_group(client, model, baseline, group)
        all_results["groups"][name] = {
            "description": desc,
            "results": group_results
        }
        print(f" {len(group_results)} items")

    return all_results


def compare_results(current: dict, previous: dict) -> list[dict]:
    """Return a list of changes between two benchmark runs."""
    changes = []
    for group_name, group_data in current["groups"].items():
        prev_group = previous.get("groups", {}).get(group_name, {})
        prev_results = prev_group.get("results", {})
        for key, cur_val in group_data["results"].items():
            if key not in prev_results:
                continue
            prev_val = prev_results[key]
            # Compare items (single token counts)
            if "tokens" in cur_val and "tokens" in prev_val:
                delta = cur_val["tokens"] - prev_val["tokens"]
                if delta != 0:
                    changes.append({
                        "group": group_name,
                        "item": key,
                        "before": prev_val["tokens"],
                        "after": cur_val["tokens"],
                        "delta": delta,
                        "severity": "REGRESSION" if delta > 0 else "IMPROVEMENT",
                    })
            # Compare comparisons (savings)
            if "savings_pct" in cur_val and "savings_pct" in prev_val:
                delta = round(cur_val["savings_pct"] - prev_val["savings_pct"], 1)
                if abs(delta) >= 5.0:  # only flag meaningful shifts
                    changes.append({
                        "group": group_name,
                        "item": key,
                        "before_savings_pct": prev_val["savings_pct"],
                        "after_savings_pct": cur_val["savings_pct"],
                        "delta_pct": delta,
                        "severity": "REGRESSION" if delta < 0 else "IMPROVEMENT",
                    })
    return changes


def print_report(results: dict, changes: list | None = None):
    model = results["model"]
    ts = results["timestamp"]
    print(f"\n{'='*70}")
    print(f"Candor Tokenizer Alignment Report")
    print(f"Model: {model}   Timestamp: {ts[:19]}Z")
    print(f"{'='*70}\n")

    for group_name, group_data in results["groups"].items():
        print(f"## {group_name.replace('_', ' ').title()}")
        print(f"   {group_data['description'][:80]}")
        print()

        for key, val in group_data["results"].items():
            if "tokens" in val:
                alert = " ⚠ ALERT" if val.get("alert") else ""
                print(f"   {key:<35} {val['tokens']} token(s){alert}")
            elif "verification_tokens" in val:
                v = val["verification_tokens"]
                a = val["agent_form_tokens"]
                saved = val["tokens_saved"]
                pct = val["savings_pct"]
                print(f"   {key}")
                print(f"      Verification: {val['verification']!r:<45} {v} tok")
                print(f"      Agent Form:   {val['agent_form']!r:<45} {a} tok")
                print(f"      Savings: {saved} tokens ({pct}%)")
        print()

    if changes:
        print(f"## Changes vs Previous Run")
        regressions = [c for c in changes if c["severity"] == "REGRESSION"]
        improvements = [c for c in changes if c["severity"] == "IMPROVEMENT"]
        if regressions:
            print(f"   REGRESSIONS ({len(regressions)}):")
            for c in regressions:
                if "delta" in c:
                    print(f"   ⚠ {c['group']}.{c['item']}: {c['before']} → {c['after']} tokens (+{c['delta']})")
                else:
                    print(f"   ⚠ {c['group']}.{c['item']}: savings {c['before_savings_pct']}% → {c['after_savings_pct']}% ({c['delta_pct']:+.1f}%)")
        if improvements:
            print(f"   IMPROVEMENTS ({len(improvements)}):")
            for c in improvements:
                if "delta" in c:
                    print(f"   ✓ {c['group']}.{c['item']}: {c['before']} → {c['after']} tokens ({c['delta']})")
                else:
                    print(f"   ✓ {c['group']}.{c['item']}: savings {c['before_savings_pct']}% → {c['after_savings_pct']}% ({c['delta_pct']:+.1f}%)")
        if not regressions and not improvements:
            print("   No significant changes.")
        print()

    print(f"## Agent Form Efficiency Summary")
    af_group = results["groups"].get("full_return_signatures", {}).get("results", {})
    if af_group:
        total_v = sum(v["verification_tokens"] for v in af_group.values() if "verification_tokens" in v)
        total_a = sum(v["agent_form_tokens"] for v in af_group.values() if "agent_form_tokens" in v)
        if total_v > 0:
            overall = round((total_v - total_a) / total_v * 100, 1)
            print(f"   Signature tokens (Verification Form): {total_v}")
            print(f"   Signature tokens (Agent Form):        {total_a}")
            print(f"   Overall signature savings:            {overall}%")
    print()


def save_results(results: dict) -> Path:
    RESULTS_DIR.mkdir(parents=True, exist_ok=True)
    date_str = results["timestamp"][:10]
    model_slug = results["model"].replace("/", "-").replace(":", "-")
    path = RESULTS_DIR / f"{date_str}_{model_slug}.json"
    # If same date+model exists, append a counter
    if path.exists():
        for i in range(2, 100):
            path = RESULTS_DIR / f"{date_str}_{model_slug}_{i}.json"
            if not path.exists():
                break
    path.write_text(json.dumps(results, indent=2), encoding="utf-8")
    return path


def main():
    # Force UTF-8 output on Windows (CP1252 terminal can't print ⚠/✓)
    if sys.stdout.encoding and sys.stdout.encoding.lower() not in ("utf-8", "utf-8-sig"):
        import io
        sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding="utf-8", errors="replace")

    parser = argparse.ArgumentParser(description="Candor tokenizer alignment benchmark")
    parser.add_argument("--model", default="claude-sonnet-4-6",
                        help="Model to benchmark against")
    parser.add_argument("--compare", metavar="PATH",
                        help="Path to a previous results JSON to compare against")
    parser.add_argument("--no-save", action="store_true",
                        help="Don't save results to disk")
    args = parser.parse_args()

    if not os.environ.get("ANTHROPIC_API_KEY"):
        print("Error: ANTHROPIC_API_KEY environment variable not set", file=sys.stderr)
        sys.exit(1)

    previous = None
    if args.compare:
        compare_path = Path(args.compare)
        if not compare_path.exists():
            print(f"Error: comparison file not found: {compare_path}", file=sys.stderr)
            sys.exit(1)
        previous = json.loads(compare_path.read_text(encoding="utf-8"))
        print(f"Comparing against: {compare_path.name} (model: {previous['model']}, date: {previous['timestamp'][:10]})")

    results = run_benchmark(args.model)

    changes = None
    if previous:
        changes = compare_results(results, previous)

    print_report(results, changes)

    if not args.no_save:
        path = save_results(results)
        print(f"Results saved: {path.relative_to(REPO_ROOT)}")


if __name__ == "__main__":
    main()
