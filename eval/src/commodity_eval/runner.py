"""Orchestrates the evaluation pipeline."""

import csv
import json
from collections.abc import Iterator
from pathlib import Path

from .analyzer import analyze_positions
from .judge import judge_response
from .models import EvalRow


def load_eval_rows(csv_path: Path) -> list[EvalRow]:
    """Load evaluation rows from a CSV file."""
    rows = []
    with open(csv_path, newline="", encoding="utf-8") as f:
        reader = csv.DictReader(f)
        for row in reader:
            rows.append(EvalRow.from_csv_row(row))
    return rows


def save_eval_rows(rows: list[EvalRow], csv_path: Path) -> None:
    """Save evaluation rows to a CSV file."""
    if not rows:
        return

    fieldnames = [
        "Scenario ID",
        "Description",
        "User Name",
        "Positions JSON",
        "Prices JSON",
        "Eval Type",
        "Ground Truth",
        "Model Response",
        "Model Critique",
        "Model Outcome",
        "Human Critique",
        "Human Outcome",
    ]

    csv_path.parent.mkdir(parents=True, exist_ok=True)
    with open(csv_path, "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames)
        writer.writeheader()
        for row in rows:
            writer.writerow(row.to_csv_row())


def run_analysis(
    rows: list[EvalRow],
    prompt_version: str = "v1",
    verbose: bool = False,
) -> Iterator[EvalRow]:
    """
    Phase 1: Call Claude analyzer for each scenario, populate model_response.

    Skips rows that already have a model_response.
    """
    for i, row in enumerate(rows):
        if verbose:
            print(f"[{i + 1}/{len(rows)}] Analyzing: {row.description[:60]}...")

        if row.model_response:
            if verbose:
                print("  (skipping - already has response)")
            yield row
            continue

        positions = json.loads(row.positions_json)
        prices = json.loads(row.prices_json)

        result = analyze_positions(positions, prices, prompt_version)

        if verbose:
            print(f"  -> {len(result.suggestions)} suggestions generated")

        yield row.model_copy(update={"model_response": result.model_dump_json()})


def run_judge(
    rows: list[EvalRow],
    judge_prompt_version: str = "v1",
    verbose: bool = False,
) -> Iterator[EvalRow]:
    """
    Phase 2: Call Claude judge for each scenario, populate model_critique + model_outcome.

    Skips rows that already have a model_outcome or have no model_response.
    """
    for i, row in enumerate(rows):
        if verbose:
            print(f"[{i + 1}/{len(rows)}] Judging: {row.description[:60]}...")

        if not row.model_response:
            if verbose:
                print("  (skipping - no model response)")
            yield row
            continue

        if row.model_outcome is not None:
            if verbose:
                print("  (skipping - already judged)")
            yield row
            continue

        judgment = judge_response(
            scenario_description=row.description,
            eval_type=row.eval_type,
            ground_truth=row.ground_truth,
            model_response=row.model_response,
            judge_prompt_version=judge_prompt_version,
        )

        if verbose:
            print(f"  -> {judgment.outcome.value}")

        yield row.model_copy(
            update={
                "model_critique": judgment.critique,
                "model_outcome": judgment.outcome,
            }
        )
