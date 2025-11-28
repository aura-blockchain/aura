"""Type definitions for ConfidenceScore module."""

from dataclasses import dataclass
from typing import Optional, List, Dict, Any
from datetime import datetime


@dataclass
class ConfidenceScoreParams:
    """Parameters for confidence score calculation."""

    address: str
    include_history: bool = False
    include_breakdown: bool = False


@dataclass
class ConfidenceScore:
    """Confidence score information."""

    address: str
    score: float
    rank: int
    total_participants: int
    last_updated: datetime
    factors: Optional[Dict[str, float]] = None
    metadata: Optional[Dict[str, Any]] = None


@dataclass
class ScoreHistory:
    """Historical confidence score data."""

    address: str
    scores: List[Dict[str, Any]]
    start_date: datetime
    end_date: datetime
    average_score: float
    peak_score: float
    lowest_score: float


@dataclass
class ScoreBreakdown:
    """Detailed score breakdown."""

    participation_score: float
    reliability_score: float
    reputation_score: float
    activity_score: float
    penalties: float
    bonuses: float


@dataclass
class RewardParams:
    """Reward distribution parameters."""

    address: str
    routine_id: str
    completion_rate: float


@dataclass
class SlashParams:
    """Slashing parameters."""

    address: str
    reason: str
    penalty_amount: str
    evidence: Optional[str] = None


@dataclass
class IRCompletionParams:
    """Inclusion routine completion parameters."""

    routine_id: str
    participant: str
    success: bool
    execution_time: int
    gas_used: int
