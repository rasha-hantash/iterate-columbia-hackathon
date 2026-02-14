# Commodity Alert Eval Framework

Offline evaluation framework that tests the quality of Claude's commodity alert suggestions using an LLM-as-judge approach.

## What This Tests

The commodity alert system uses Claude to analyze a user's CORN positions and suggest price alerts (stop-loss, take-profit). This eval framework measures whether those suggestions are:

1. **Directionally correct** — long positions get stop-loss below entry, short positions get stop-loss above entry
2. **Reasonably priced** — thresholds fall within sensible ranges relative to entry/current prices
3. **Complete** — all positions in a portfolio are addressed
4. **Well-reasoned** — the LLM's reasoning mentions relevant risk factors (position size, P&L status)

## Prerequisites

- Python 3.11+
- [uv](https://docs.astral.sh/uv/) package manager
- `ANTHROPIC_API_KEY` environment variable set

## Quick Start

```bash
cd eval
uv sync
ANTHROPIC_API_KEY=your-key uv run python scripts/run_eval.py --verbose
```

## Architecture

Two-phase pipeline:

```
Golden CSV  ──►  Phase 1: Analyzer  ──►  Phase 2: Judge  ──►  Results CSV
(scenarios)      (Claude API +           (Claude API          (pass/fail +
                  create_alert tool)      evaluates output)    critiques)
```

**Phase 1 (Analyzer):** For each scenario, sends positions + prices to Claude with a `create_alert` tool. Claude returns reasoning + alert suggestions. This replicates what the Go `/analyze-positions` endpoint does.

**Phase 2 (Judge):** A separate Claude call evaluates whether the suggestions are appropriate, using the ground truth from the golden dataset. Returns pass/fail with a detailed critique.

## Golden Dataset

10 scenarios in `golden/scenarios.csv`, all using CORN positions with different configurations.

### Strict Scenarios (S01-S05)

Test basic directional correctness. Ground truth specifies an exact commodity, condition, and acceptable price range.

| ID | Scenario | What Must Happen |
|----|----------|-----------------|
| S01 | Long 50k CORN @4.50, price dropped to 4.25 | Stop-loss below entry (3.80-4.25) |
| S02 | Short 20k CORN @4.40, price rose to 4.60 | Stop-loss above entry (4.60-4.85) |
| S03 | Long 30k CORN @4.55, price at 5.10 (profit) | Take-profit above (5.10-5.80) |
| S04 | Long 100k CORN @4.50, price at 4.52 (large) | Protective stop-loss below (4.20-4.50) |
| S05 | Short 15k CORN @4.80, price at 4.35 (profit) | Take-profit below (4.10-4.40) |

### Criteria Scenarios (S06-S10)

Test reasoning quality with qualitative rules (pipe-delimited). ALL rules must be satisfied to pass.

| ID | Scenario | Rules |
|----|----------|-------|
| S06 | Long + short CORN positions | Must address both with alerts in both directions |
| S07 | 100k bushel long CORN (very large) | Must mention size/concentration risk |
| S08 | Long CORN, 12% profit | Must suggest take-profit, mention gains |
| S09 | Long CORN, 10% loss | Must suggest stop-loss, mention risk |
| S10 | Long CORN, at breakeven | Must suggest both stop-loss and take-profit |

## Eval Types

### Strict
Ground truth is JSON: `{"commodity_code": "CORN", "expected_condition": "below", "price_min": 3.80, "price_max": 4.25}`

The judge checks that at least one suggestion matches all three constraints (commodity, condition, price range).

### Criteria
Ground truth is pipe-delimited rules: `must suggest stop-loss|reasoning must mention risk|threshold below 4.50`

The judge checks each rule individually. ALL must pass.

## Prompt Versioning

Both the analyzer and judge prompts are versioned:

```
analyzer_prompts/v1.txt    # System prompt for position analysis
judge_prompts/v1.txt       # System prompt for the judge
```

To A/B test a prompt change:
1. Create `analyzer_prompts/v2.txt` with your new prompt
2. Run: `uv run python scripts/run_eval.py --analyzer-prompt v2 -v`
3. Compare pass rates between v1 and v2 results

## Adding New Scenarios

1. Open `golden/scenarios.csv`
2. Add a new row with:
   - `Scenario ID`: Next ID (e.g., S11)
   - `Description`: What this scenario tests
   - `User Name`: Alice, Bob, or Carol
   - `Positions JSON`: JSON array of positions
   - `Prices JSON`: JSON array of current prices
   - `Eval Type`: `strict` or `criteria`
   - `Ground Truth`: JSON (strict) or pipe-delimited rules (criteria)
3. Leave Model Response, Model Critique, Model Outcome, Human Critique, Human Outcome empty

## Interpreting Results

Output CSVs are saved to `eval_results/` with timestamps. Key columns:

- **Model Response**: Full JSON of Claude's analysis (reasoning + suggestions)
- **Model Critique**: The judge's explanation of what passed/failed
- **Model Outcome**: `pass` or `fail`
- **Human Critique / Human Outcome**: Fill these in manually to measure judge-human alignment

## CLI Options

```
--csv PATH              Path to golden CSV (default: golden/scenarios.csv)
--analyzer-prompt VER   Analyzer prompt version (default: v1)
--judge-prompt VER      Judge prompt version (default: v1)
--output-dir PATH       Output directory (default: eval_results/)
--verbose, -v           Print progress
--skip-analysis         Only run judge on existing responses
--skip-judge            Only run analysis, skip judging
```
