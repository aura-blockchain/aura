"""Type definitions for VCRegistry (Verifiable Credentials) module."""

from dataclasses import dataclass
from typing import Optional, List, Dict, Any
from datetime import datetime
from enum import Enum


class VCStatus(Enum):
    """Verifiable Credential status."""

    ACTIVE = "active"
    REVOKED = "revoked"
    EXPIRED = "expired"
    SUSPENDED = "suspended"


class VCType(Enum):
    """Types of verifiable credentials."""

    IDENTITY = "identity"
    EDUCATION = "education"
    EMPLOYMENT = "employment"
    LICENSE = "license"
    CERTIFICATION = "certification"
    CUSTOM = "custom"


class PresentationStatus(Enum):
    """Presentation verification status."""

    VERIFIED = "verified"
    INVALID = "invalid"
    EXPIRED = "expired"
    REVOKED = "revoked"


@dataclass
class VCParams:
    """Parameters for creating a verifiable credential."""

    issuer: str
    subject: str
    vc_type: VCType
    claims: Dict[str, Any]
    expiry_date: Optional[datetime] = None
    metadata: Optional[Dict[str, Any]] = None


@dataclass
class VerifiableCredential:
    """Verifiable Credential information."""

    vc_id: str
    issuer: str
    subject: str
    vc_type: VCType
    status: VCStatus
    claims: Dict[str, Any]
    proof: str
    issued_at: datetime
    expires_at: Optional[datetime] = None
    revoked_at: Optional[datetime] = None
    metadata: Optional[Dict[str, Any]] = None
    schema_id: Optional[str] = None


@dataclass
class VCProof:
    """Credential cryptographic proof."""

    proof_type: str
    proof_value: str
    verification_method: str
    created: datetime
    proof_purpose: str
    challenge: Optional[str] = None
    domain: Optional[str] = None


@dataclass
class VCPresentation:
    """Verifiable Presentation."""

    presentation_id: str
    holder: str
    credentials: List[str]
    proof: VCProof
    created_at: datetime
    expires_at: Optional[datetime] = None
    challenge: Optional[str] = None
    domain: Optional[str] = None
    metadata: Optional[Dict[str, Any]] = None


@dataclass
class PresentationParams:
    """Parameters for creating a presentation."""

    holder: str
    credential_ids: List[str]
    challenge: Optional[str] = None
    domain: Optional[str] = None
    purpose: str = "authentication"


@dataclass
class VCRevocationParams:
    """Parameters for revoking a credential."""

    vc_id: str
    issuer: str
    reason: str
    proof: str


@dataclass
class VCSchema:
    """Credential schema definition."""

    schema_id: str
    name: str
    version: str
    vc_type: VCType
    required_fields: List[str]
    optional_fields: List[str]
    field_types: Dict[str, str]
    created_at: datetime
    author: str


@dataclass
class IssuerInfo:
    """Credential issuer information."""

    issuer_id: str
    name: str
    did: str
    public_key: str
    verified: bool
    total_issued: int
    total_revoked: int
    reputation_score: float
    registered_at: datetime


@dataclass
class VCVerificationResult:
    """Credential verification result."""

    vc_id: str
    valid: bool
    status: VCStatus
    issuer_verified: bool
    signature_valid: bool
    not_expired: bool
    not_revoked: bool
    verified_at: datetime
    errors: List[str]
    warnings: List[str]


@dataclass
class RegistryStats:
    """Registry statistics."""

    total_credentials: int
    active_credentials: int
    revoked_credentials: int
    expired_credentials: int
    total_issuers: int
    total_holders: int
    total_presentations: int
    credentials_by_type: Dict[str, int]


@dataclass
class VCQuery:
    """Query parameters for searching credentials."""

    issuer: Optional[str] = None
    subject: Optional[str] = None
    vc_type: Optional[VCType] = None
    status: Optional[VCStatus] = None
    issued_after: Optional[datetime] = None
    issued_before: Optional[datetime] = None
    include_revoked: bool = False
