"""Type definitions for EconomicSecurity module."""

from dataclasses import dataclass
from typing import Optional, List, Dict
from datetime import datetime


@dataclass
class DynamicFeeParams:
    """Dynamic fee calculation parameters."""

    transaction_size: int
    priority: int
    network_congestion: float


@dataclass
class MEVProtectionParams:
    """MEV protection parameters."""

    transaction_data: str
    max_slippage: float
    deadline: int
    use_private_mempool: bool = True


@dataclass
class FeeStructure:
    """Fee structure information."""

    base_fee: str
    priority_fee: str
    total_fee: str
    gas_price: str
    estimated_time: int


@dataclass
class WhaleProtectionParams:
    """Whale protection parameters."""

    address: str
    amount: str
    token: str


@dataclass
class EconomicMetrics:
    """Economic security metrics."""

    total_value_locked: str
    circulating_supply: str
    inflation_rate: float
    staking_ratio: float
    average_fee: str
    mev_prevented: str
    timestamp: datetime


@dataclass
class FeeConfiguration:
    """Fee configuration."""

    min_gas_price: str
    max_gas_price: str
    base_fee_multiplier: float
    congestion_threshold: float
    whale_threshold: str


@dataclass
class MEVReport:
    """MEV protection report."""

    tx_hash: str
    mev_detected: bool
    mev_amount: str
    protection_applied: bool
    timestamp: datetime
    details: Optional[str] = None


@dataclass
class CircuitBreakerParams:
    """Circuit breaker parameters."""

    threshold: str
    window_duration: int
    cooldown_period: int


@dataclass
class CircuitBreakerStatus:
    """Circuit breaker status."""

    active: bool
    triggered_at: Optional[datetime] = None
    reason: Optional[str] = None
    reset_at: Optional[datetime] = None
