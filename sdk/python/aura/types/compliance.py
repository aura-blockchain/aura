"""Type definitions for Compliance module."""

from dataclasses import dataclass
from typing import Optional, List, Dict, Any
from enum import Enum
from datetime import datetime


class ComplianceStatus(Enum):
    """Compliance check status."""

    APPROVED = "approved"
    PENDING = "pending"
    REJECTED = "rejected"
    UNDER_REVIEW = "under_review"


class ComplianceLevel(Enum):
    """Compliance level."""

    BASIC = "basic"
    STANDARD = "standard"
    ENHANCED = "enhanced"


@dataclass
class ComplianceCheckParams:
    """Parameters for compliance check."""

    address: str
    check_type: str
    level: ComplianceLevel
    metadata: Optional[Dict[str, Any]] = None


@dataclass
class ComplianceReport:
    """Compliance report."""

    report_id: str
    address: str
    status: ComplianceStatus
    level: ComplianceLevel
    checks_passed: List[str]
    checks_failed: List[str]
    risk_score: float
    created_at: datetime
    expires_at: Optional[datetime] = None
    notes: Optional[str] = None


@dataclass
class KYCAMLStatus:
    """KYC/AML status."""

    address: str
    kyc_verified: bool
    aml_cleared: bool
    verification_date: Optional[datetime] = None
    expiry_date: Optional[datetime] = None
    provider: Optional[str] = None


@dataclass
class SanctionCheck:
    """Sanction screening result."""

    address: str
    is_sanctioned: bool
    lists_matched: List[str]
    checked_at: datetime
    details: Optional[str] = None


@dataclass
class TransactionMonitoring:
    """Transaction monitoring data."""

    tx_hash: str
    address: str
    amount: str
    risk_level: str
    flags: List[str]
    timestamp: datetime
    reviewed: bool = False
