"""Type definitions for ValidatorSecurity module."""

from dataclasses import dataclass
from typing import Optional, List, Dict
from datetime import datetime
from enum import Enum


class ValidatorStatus(Enum):
    """Validator operational status."""

    ACTIVE = "active"
    INACTIVE = "inactive"
    JAILED = "jailed"
    TOMBSTONED = "tombstoned"
    UNBONDING = "unbonding"


class SlashingReason(Enum):
    """Reasons for validator slashing."""

    DOUBLE_SIGN = "double_sign"
    DOWNTIME = "downtime"
    BYZANTINE_BEHAVIOR = "byzantine_behavior"
    INVALID_SIGNATURE = "invalid_signature"
    CENSORSHIP = "censorship"


class SecurityLevel(Enum):
    """Validator security levels."""

    BASIC = "basic"
    STANDARD = "standard"
    ENHANCED = "enhanced"
    MAXIMUM = "maximum"


@dataclass
class ValidatorSecurityParams:
    """Validator security configuration."""

    validator_address: str
    security_level: SecurityLevel
    enable_sentry_nodes: bool
    enable_hsm: bool
    enable_monitoring: bool


@dataclass
class ValidatorInfo:
    """Validator information."""

    operator_address: str
    consensus_address: str
    status: ValidatorStatus
    jailed: bool
    tombstoned: bool
    commission_rate: str
    voting_power: int
    uptime: float
    last_active: datetime
    jail_time: Optional[datetime] = None


@dataclass
class SlashingEvent:
    """Validator slashing event."""

    event_id: str
    validator_address: str
    reason: SlashingReason
    slash_amount: str
    slash_percentage: float
    occurred_at: datetime
    height: int
    jailed: bool
    tombstoned: bool
    evidence: Optional[str] = None


@dataclass
class JailingParams:
    """Jailing operation parameters."""

    validator_address: str
    reason: str
    duration_blocks: int
    evidence: Optional[str] = None


@dataclass
class UnjailParams:
    """Unjailing operation parameters."""

    validator_address: str
    proof_of_correction: str


@dataclass
class ValidatorMonitoring:
    """Validator monitoring data."""

    validator_address: str
    uptime_percentage: float
    missed_blocks: int
    total_blocks: int
    last_signed_block: int
    consecutive_misses: int
    health_score: float
    alerts: List[str]
    last_checked: datetime


@dataclass
class SentryNodeConfig:
    """Sentry node configuration."""

    sentry_id: str
    validator_address: str
    ip_address: str
    port: int
    active: bool
    last_heartbeat: datetime
    bandwidth_usage: int
    connection_count: int


@dataclass
class DoubleSignEvidence:
    """Double signing evidence."""

    validator_address: str
    block_height: int
    vote_a: str
    vote_b: str
    timestamp_a: datetime
    timestamp_b: datetime
    proof: str


@dataclass
class ValidatorPerformance:
    """Validator performance metrics."""

    validator_address: str
    blocks_signed: int
    blocks_missed: int
    uptime_percentage: float
    average_response_time: float
    rewards_earned: str
    slashing_events: int
    period_start: datetime
    period_end: datetime


@dataclass
class HSMConfig:
    """Hardware Security Module configuration."""

    hsm_id: str
    validator_address: str
    hsm_type: str
    enabled: bool
    key_slot: int
    firmware_version: str
    last_verified: datetime


@dataclass
class ValidatorAlert:
    """Validator security alert."""

    alert_id: str
    validator_address: str
    alert_type: str
    severity: str
    message: str
    created_at: datetime
    acknowledged: bool
    resolved: bool
