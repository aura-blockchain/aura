"""Type definitions for InclusionRoutines module."""

from dataclasses import dataclass
from typing import Optional, List, Dict, Any
from datetime import datetime
from enum import Enum


class RoutineStatus(Enum):
    """Inclusion routine status."""

    ACTIVE = "active"
    INACTIVE = "inactive"
    PAUSED = "paused"
    COMPLETED = "completed"
    CANCELLED = "cancelled"


class RoutinePriority(Enum):
    """Routine execution priority."""

    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    CRITICAL = "critical"


@dataclass
class InclusionRoutineParams:
    """Parameters for creating inclusion routine."""

    creator: str
    name: str
    description: str
    code: str
    priority: RoutinePriority
    max_gas: int
    prerequisites: Optional[List[str]] = None
    schedule: Optional[str] = None
    metadata: Optional[Dict[str, Any]] = None


@dataclass
class InclusionRoutine:
    """Inclusion routine information."""

    routine_id: str
    creator: str
    name: str
    description: str
    status: RoutineStatus
    priority: RoutinePriority
    code_hash: str
    created_at: datetime
    updated_at: datetime
    executions: int
    success_rate: float
    average_gas: int
    last_execution: Optional[datetime] = None


@dataclass
class RoutineExecution:
    """Routine execution record."""

    execution_id: str
    routine_id: str
    executor: str
    status: str
    started_at: datetime
    gas_used: int
    completed_at: Optional[datetime] = None
    result: Optional[str] = None
    error: Optional[str] = None


@dataclass
class RoutineUpdateParams:
    """Parameters for updating routine."""

    routine_id: str
    code: Optional[str] = None
    priority: Optional[RoutinePriority] = None
    schedule: Optional[str] = None
    status: Optional[RoutineStatus] = None


@dataclass
class RoutineRateLimitParams:
    """Rate limiting parameters."""

    routine_id: str
    max_executions_per_block: int
    cooldown_blocks: int


@dataclass
class RoutinePrerequisite:
    """Routine prerequisite check."""

    prerequisite_id: str
    routine_id: str
    condition: str
    required: bool


@dataclass
class RoutineRegistry:
    """Routine registry information."""

    total_routines: int
    active_routines: int
    total_executions: int
    average_success_rate: float
    top_routines: List[str]
