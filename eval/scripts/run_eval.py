#!/usr/bin/env python3
"""CLI entry point for running commodity alert evaluations."""

import argparse
import sys
from datetime import datetime, timezone
from pathlib import Path

# Add src to path so we can import commodity_eval
sys.path.insert(0, str(Path(__file__).parent.parent / "src"))

from commodity_eval import Outcome, load_eval_rows, run_analysis, run_judge, save_eval_rows


def main() -> None:
    parser = argparse.ArgumentParser(description="Run commodity alert eval pipeline")
    parser.add_argument(
        "--csv",
        type=Path,
        default=Path(__file__).parent.parent / "golden" / "scenarios.csv",
        help="Path to golden dataset CSV (default: golden/scenarios.csv)",
    )
    parser.add_argument(
        "--analyzer-prompt",
        default="v1",
        help="Analyzer prompt version (default: v1)",
    )
    parser.add_argument(
        "--judge-prompt",
        default="v1",
        help="Judge prompt version (default: v1)",
    )
    parser.add_argument(
        "--output-dir",
        type=Path,
        default=Path(__file__).parent.parent / "eval_results",
        help="Output directory for results (default: eval_results/)",
    )
    parser.add_argument(
        "--verbose", "-v",
        action="store_true",
        help="Print progress information",
    )
    parser.add_argument(
        "--skip-analysis",
        action="store_true",
        help="Skip analysis phase, only run judge on existing responses",
    )
    parser.add_argument(
        "--skip-judge",
        action="store_true",
        help="Skip judge phase, only run analysis",
    )

    args = parser.parse_args()

    # Load golden dataset
    if args.verbose:
        print(f"Loading scenarios from {args.csv}")
    rows = load_eval_rows(args.csv)
    if args.verbose:
        print(f"Loaded {len(rows)} scenarios\n")

    # Phase 1: Analysis
    if not args.skip_analysis:
        if args.verbose:
            print("=" * 60)
            print("PHASE 1: Running analyzer")
            print("=" * 60)
        rows = list(run_analysis(rows, args.analyzer_prompt, args.verbose))
        if args.verbose:
            print()

    # Phase 2: Judge
    if not args.skip_judge:
        if args.verbose:
            print("=" * 60)
            print("PHASE 2: Running judge")
            print("=" * 60)
        rows = list(run_judge(rows, args.judge_prompt, args.verbose))
        if args.verbose:
            print()

    # Save results
    timestamp = datetime.now(timezone.utc).strftime("%Y%m%d_%H%M%S")
    output_path = args.output_dir / f"analyzer-{args.analyzer_prompt}_judge-{args.judge_prompt}_{timestamp}.csv"
    save_eval_rows(rows, output_path)
    print(f"Results saved to {output_path}")

    # Print summary
    judged = [r for r in rows if r.model_outcome is not None]
    passed = [r for r in judged if r.model_outcome == Outcome.PASS]
    failed = [r for r in judged if r.model_outcome == Outcome.FAIL]

    print("\n" + "=" * 60)
    print("SUMMARY")
    print("=" * 60)
    print(f"Total scenarios:  {len(rows)}")
    print(f"Judged:           {len(judged)}")
    print(f"Passed:           {len(passed)}")
    print(f"Failed:           {len(failed)}")
    if judged:
        print(f"Pass rate:        {len(passed) / len(judged) * 100:.1f}%")

    if failed:
        print(f"\nFailed scenarios:")
        for r in failed:
            print(f"  {r.scenario_id}: {r.description}")


if __name__ == "__main__":
    main()
