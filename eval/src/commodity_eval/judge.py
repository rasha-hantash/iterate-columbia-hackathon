"""LLM judge implementation for evaluating commodity alert suggestions."""

import json
import os
from pathlib import Path
from typing import TypedDict, cast

from anthropic import Anthropic
from anthropic.types import TextBlock

from .models import JudgmentResult, Outcome

JUDGE_MODEL = "claude-sonnet-4-5-20250929"
JUDGE_PROMPTS_DIR = Path(__file__).parent.parent.parent / "judge_prompts"


class JudgeResultDict(TypedDict):
    critique: str
    outcome: str


def load_judge_prompt(version: str = "v1") -> str:
    """Load the judge prompt from the judge_prompts directory."""
    prompt_path = JUDGE_PROMPTS_DIR / f"{version}.txt"
    if not prompt_path.exists():
        raise FileNotFoundError(f"Judge prompt not found at {prompt_path}")
    return prompt_path.read_text()


def judge_response(
    scenario_description: str,
    eval_type: str,
    ground_truth: str,
    model_response: str,
    judge_prompt_version: str = "v1",
) -> JudgmentResult:
    """
    Judge whether alert suggestions are appropriate for the given scenario.

    Args:
        scenario_description: Human-readable description of the test scenario
        eval_type: "strict" or "criteria"
        ground_truth: JSON (strict) or pipe-delimited rules (criteria)
        model_response: JSON string of AnalysisResult
        judge_prompt_version: Version of the judge prompt to use

    Returns:
        JudgmentResult with critique and binary pass/fail outcome
    """
    api_key = os.environ.get("ANTHROPIC_API_KEY")
    if not api_key:
        raise ValueError("ANTHROPIC_API_KEY environment variable is required")

    client = Anthropic(api_key=api_key)
    system_prompt = load_judge_prompt(judge_prompt_version)

    user_message = f"""## Scenario
{scenario_description}

## Evaluation Type
{eval_type}

## Ground Truth
{ground_truth}

## Model Response
{model_response}

Please evaluate this response and provide your judgment."""

    response = client.messages.create(
        model=JUDGE_MODEL,
        max_tokens=1024,
        system=system_prompt,
        messages=[{"role": "user", "content": user_message}],
    )

    text_block = next(
        (block for block in response.content if isinstance(block, TextBlock)), None
    )
    if text_block is None:
        raise ValueError("No text block found in judge response")
    response_text = text_block.text

    try:
        json_start = response_text.find("{")
        json_end = response_text.rfind("}") + 1
        if json_start != -1 and json_end > json_start:
            json_str = response_text[json_start:json_end]
            result = cast(JudgeResultDict, json.loads(json_str))
            return JudgmentResult(
                critique=result["critique"],
                outcome=Outcome(result["outcome"].lower()),
            )
        else:
            raise ValueError("No JSON found in judge response")
    except (json.JSONDecodeError, KeyError, ValueError) as e:
        response_lower = response_text.lower()
        if "pass" in response_lower and "fail" not in response_lower:
            outcome = Outcome.PASS
        elif "fail" in response_lower:
            outcome = Outcome.FAIL
        else:
            outcome = Outcome.FAIL

        return JudgmentResult(
            critique=f"[Parse error: {e}] {response_text}",
            outcome=outcome,
        )
