"""Calls Claude API to analyze commodity positions and suggest alerts."""

import json
import os
from pathlib import Path

from anthropic import Anthropic
from anthropic.types import TextBlock, ToolUseBlock

from .models import AlertSuggestion, AnalysisResult

ANALYZER_MODEL = "claude-sonnet-4-5-20250929"
PROMPTS_DIR = Path(__file__).parent.parent.parent / "analyzer_prompts"

CREATE_ALERT_TOOL = {
    "name": "create_alert",
    "description": "Create a price alert for a commodity position",
    "input_schema": {
        "type": "object",
        "properties": {
            "commodity_code": {
                "type": "string",
                "description": "The commodity code (e.g., CORN)",
            },
            "condition": {
                "type": "string",
                "enum": ["above", "below"],
                "description": "Whether the alert triggers when price goes above or below the threshold",
            },
            "threshold_price": {
                "type": "number",
                "description": "The price threshold that triggers the alert",
            },
            "notes": {
                "type": "string",
                "description": "Explanation of why this alert is recommended",
            },
        },
        "required": ["commodity_code", "condition", "threshold_price", "notes"],
    },
}


def load_analyzer_prompt(version: str = "v1") -> str:
    """Load the analyzer system prompt."""
    prompt_path = PROMPTS_DIR / f"{version}.txt"
    if not prompt_path.exists():
        raise FileNotFoundError(f"Analyzer prompt not found at {prompt_path}")
    return prompt_path.read_text()


def analyze_positions(
    positions: list[dict],
    prices: list[dict],
    prompt_version: str = "v1",
) -> AnalysisResult:
    """
    Call Claude to analyze positions and suggest alerts.

    Args:
        positions: List of position dicts with commodity_code, direction, volume, entry_price
        prices: List of price dicts with commodity_code, price
        prompt_version: Version of the analyzer prompt to use

    Returns:
        AnalysisResult with reasoning and list of alert suggestions
    """
    api_key = os.environ.get("ANTHROPIC_API_KEY")
    if not api_key:
        raise ValueError("ANTHROPIC_API_KEY environment variable is required")

    client = Anthropic(api_key=api_key)
    system_prompt = load_analyzer_prompt(prompt_version)

    user_message = f"""Here are the user's current commodity positions:
{json.dumps(positions, indent=2)}

Here are the current market prices:
{json.dumps(prices, indent=2)}

Please analyze these positions and suggest appropriate price alerts. For each suggestion, use the create_alert tool."""

    response = client.messages.create(
        model=ANALYZER_MODEL,
        max_tokens=2048,
        system=system_prompt,
        messages=[{"role": "user", "content": user_message}],
        tools=[CREATE_ALERT_TOOL],
    )

    reasoning_parts: list[str] = []
    suggestions: list[AlertSuggestion] = []

    for block in response.content:
        if isinstance(block, TextBlock):
            reasoning_parts.append(block.text)
        elif isinstance(block, ToolUseBlock) and block.name == "create_alert":
            suggestions.append(AlertSuggestion(**block.input))

    return AnalysisResult(
        reasoning="\n".join(reasoning_parts),
        suggestions=suggestions,
    )
