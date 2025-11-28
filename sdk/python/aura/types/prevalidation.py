"""Type definitions for Prevalidation module."""

from dataclasses import dataclass
from typing import Optional, List, Dict, Any
from datetime import datetime
from enum import Enum


class ValidationStatus(Enum):
    """Validation status."""

    VALID = "valid"
    INVALID = "invalid"
    WARNING = "warning"
    PENDING = "pending"


class ValidationLevel(Enum):
    """Validation thoroughness level."""

    BASIC = "basic"
    STANDARD = "standard"
    COMPREHENSIVE = "comprehensive"


@dataclass
class PrevalidationParams:
    """Parameters for transaction prevalidation."""

    transaction_data: str
    validation_level: ValidationLevel
    check_balance: bool = True
    check_nonce: bool = True
    check_signature: bool = True
    check_gas: bool = True


@dataclass
class PrevalidationResult:
    """Prevalidation result."""

    status: ValidationStatus
    valid: bool
    errors: List[str]
    warnings: List[str]
    estimated_gas: int
    estimated_fee: str
    validated_at: datetime
    checks_performed: List[str]
    metadata: Optional[Dict[str, Any]] = None


@dataclass
class ValidationRule:
    """Validation rule definition."""

    rule_id: str
    name: str
    description: str
    rule_type: str
    condition: str
    severity: str
    enabled: bool
    error_message: str


@dataclass
class TransactionCheck:
    """Individual transaction check result."""

    check_name: str
    passed: bool
    message: str
    severity: str
    timestamp: datetime


@dataclass
class ValidationPolicy:
    """Validation policy configuration."""

    policy_id: str
    name: str
    rules: List[str]
    strict_mode: bool
    auto_reject: bool
    notify_on_fail: bool


@dataclass
class GasEstimation:
    """Gas estimation details."""

    base_gas: int
    computation_gas: int
    storage_gas: int
    total_gas: int
    gas_price: str
    total_fee: str
    confidence: float


@dataclass
class NonceValidation:
    """Nonce validation result."""

    address: str
    current_nonce: int
    expected_nonce: int
    valid: bool
    gap: int


@dataclass
class BalanceValidation:
    """Balance validation result."""

    address: str
    required_balance: str
    actual_balance: str
    sufficient: bool
    shortfall: Optional[str] = None
