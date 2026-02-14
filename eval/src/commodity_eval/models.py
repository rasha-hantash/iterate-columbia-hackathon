"""Data models for the commodity alert evaluation system."""

from enum import Enum

from pydantic import BaseModel


class Outcome(str, Enum):
    """Binary pass/fail judgment outcome."""

    PASS = "pass"
    FAIL = "fail"


class AlertSuggestion(BaseModel):
    """A single alert suggestion from the analyzer."""

    commodity_code: str
    condition: str  # "above" or "below"
    threshold_price: float
    notes: str


class AnalysisResult(BaseModel):
    """Output from the Claude position analyzer."""

    reasoning: str
    suggestions: list[AlertSuggestion]


class JudgmentResult(BaseModel):
    """Output from the LLM judge."""

    critique: str
    outcome: Outcome


class EvalRow(BaseModel):
    """One row from the golden CSV evaluation dataset."""

    scenario_id: str
    description: str
    user_name: str
    positions_json: str
    prices_json: str
    eval_type: str  # "strict" or "criteria"
    ground_truth: str
    model_response: str | None = None
    model_critique: str | None = None
    model_outcome: Outcome | None = None
    human_critique: str | None = None
    human_outcome: Outcome | None = None

    @classmethod
    def from_csv_row(cls, row: dict[str, str]) -> "EvalRow":
        """Create an EvalRow from a CSV row dictionary."""
        return cls(
            scenario_id=row.get("Scenario ID", ""),
            description=row.get("Description", ""),
            user_name=row.get("User Name", ""),
            positions_json=row.get("Positions JSON", ""),
            prices_json=row.get("Prices JSON", ""),
            eval_type=row.get("Eval Type", ""),
            ground_truth=row.get("Ground Truth", ""),
            model_response=row.get("Model Response") or None,
            model_critique=row.get("Model Critique") or None,
            model_outcome=(
                Outcome(row["Model Outcome"].lower())
                if row.get("Model Outcome")
                else None
            ),
            human_critique=row.get("Human Critique") or None,
            human_outcome=(
                Outcome(row["Human Outcome"].lower())
                if row.get("Human Outcome")
                else None
            ),
        )

    def to_csv_row(self) -> dict[str, str]:
        """Convert to a CSV row dictionary."""
        return {
            "Scenario ID": self.scenario_id,
            "Description": self.description,
            "User Name": self.user_name,
            "Positions JSON": self.positions_json,
            "Prices JSON": self.prices_json,
            "Eval Type": self.eval_type,
            "Ground Truth": self.ground_truth,
            "Model Response": self.model_response or "",
            "Model Critique": self.model_critique or "",
            "Model Outcome": self.model_outcome.value if self.model_outcome else "",
            "Human Critique": self.human_critique or "",
            "Human Outcome": self.human_outcome.value if self.human_outcome else "",
        }
