"""Type definitions for IdentityChange module."""

from dataclasses import dataclass
from typing import Optional, List, Dict, Any
from datetime import datetime
from enum import Enum


class IdentityChangeStatus(Enum):
    """Identity change request status."""

    PENDING = "pending"
    APPROVED = "approved"
    REJECTED = "rejected"
    PROCESSING = "processing"
    COMPLETED = "completed"


class IdentityChangeType(Enum):
    """Type of identity change."""

    UPDATE = "update"
    RECOVERY = "recovery"
    MIGRATION = "migration"
    REVOCATION = "revocation"


@dataclass
class IdentityChangeParams:
    """Parameters for identity change request."""

    old_address: str
    new_address: str
    change_type: IdentityChangeType
    proof: str
    reason: str
    metadata: Optional[Dict[str, Any]] = None


@dataclass
class IdentityChangeRequest:
    """Identity change request."""

    request_id: str
    old_address: str
    new_address: str
    change_type: IdentityChangeType
    status: IdentityChangeStatus
    proof: str
    reason: str
    submitted_at: datetime
    processed_at: Optional[datetime] = None
    approved_by: Optional[str] = None
    rejection_reason: Optional[str] = None


@dataclass
class IdentityVerificationParams:
    """Identity verification parameters."""

    address: str
    verification_type: str
    proof_data: str
    attestations: Optional[List[str]] = None


@dataclass
class IdentityProfile:
    """Identity profile information."""

    address: str
    verified: bool
    verification_level: int
    created_at: datetime
    updated_at: datetime
    attestations: List[str]
    metadata: Optional[Dict[str, Any]] = None


@dataclass
class IdentityAttestation:
    """Identity attestation."""

    attestation_id: str
    subject: str
    issuer: str
    claim_type: str
    claim_data: Dict[str, Any]
    issued_at: datetime
    expires_at: Optional[datetime] = None
    revoked: bool = False


@dataclass
class RecoveryParams:
    """Account recovery parameters."""

    lost_address: str
    recovery_address: str
    recovery_key: str
    guardians: List[str]
    signatures: List[str]
