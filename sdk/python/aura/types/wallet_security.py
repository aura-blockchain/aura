"""Type definitions for WalletSecurity module."""

from dataclasses import dataclass
from typing import Optional, List, Dict
from datetime import datetime
from enum import Enum


class SessionStatus(Enum):
    """Wallet session status."""

    ACTIVE = "active"
    EXPIRED = "expired"
    REVOKED = "revoked"
    LOCKED = "locked"


class BiometricType(Enum):
    """Biometric authentication types."""

    FINGERPRINT = "fingerprint"
    FACE = "face"
    VOICE = "voice"
    IRIS = "iris"


class SecurityFeature(Enum):
    """Wallet security features."""

    MULTISIG = "multisig"
    BIOMETRIC = "biometric"
    TWO_FACTOR = "two_factor"
    HARDWARE_KEY = "hardware_key"
    SOCIAL_RECOVERY = "social_recovery"


@dataclass
class WalletSecurityParams:
    """Wallet security configuration."""

    address: str
    security_level: int
    features: List[SecurityFeature]
    session_timeout: int
    require_confirmation: bool


@dataclass
class SessionParams:
    """Session creation parameters."""

    address: str
    device_id: str
    ip_address: str
    user_agent: str
    duration_seconds: int = 3600


@dataclass
class WalletSession:
    """Wallet session information."""

    session_id: str
    address: str
    device_id: str
    status: SessionStatus
    created_at: datetime
    expires_at: datetime
    last_activity: datetime
    ip_address: str
    location: Optional[str] = None
    user_agent: Optional[str] = None


@dataclass
class BiometricData:
    """Biometric authentication data."""

    address: str
    biometric_type: BiometricType
    template_hash: str
    registered_at: datetime
    enabled: bool
    device_id: str
    last_used: Optional[datetime] = None


@dataclass
class MultisigConfig:
    """Multi-signature configuration."""

    address: str
    threshold: int
    signers: List[str]
    weights: Dict[str, int]
    created_at: datetime
    updated_at: datetime
    total_transactions: int


@dataclass
class TransactionConfirmation:
    """Transaction confirmation requirement."""

    tx_hash: str
    address: str
    required_confirmations: int
    current_confirmations: int
    confirmed: bool
    expires_at: datetime
    confirmers: List[str]


@dataclass
class TwoFactorAuth:
    """Two-factor authentication configuration."""

    address: str
    method: str
    enabled: bool
    secret: str
    backup_codes: List[str]
    registered_at: datetime
    last_verified: Optional[datetime] = None


@dataclass
class HardwareKeyInfo:
    """Hardware key information."""

    key_id: str
    address: str
    device_type: str
    public_key: str
    registered_at: datetime
    enabled: bool
    last_used: Optional[datetime] = None
    firmware_version: Optional[str] = None


@dataclass
class SocialRecoveryConfig:
    """Social recovery configuration."""

    address: str
    guardians: List[str]
    threshold: int
    timeout_period: int
    active: bool
    configured_at: datetime


@dataclass
class RecoveryRequest:
    """Account recovery request."""

    request_id: str
    address: str
    new_owner: str
    guardians_approved: List[str]
    threshold: int
    status: str
    created_at: datetime
    expires_at: datetime
    executed_at: Optional[datetime] = None


@dataclass
class SecurityAlert:
    """Wallet security alert."""

    alert_id: str
    address: str
    alert_type: str
    severity: str
    message: str
    detected_at: datetime
    acknowledged: bool
    details: Optional[Dict[str, str]] = None


@dataclass
class AccessLog:
    """Wallet access log entry."""

    log_id: str
    address: str
    action: str
    device_id: str
    ip_address: str
    success: bool
    timestamp: datetime
    location: Optional[str] = None
    details: Optional[str] = None


@dataclass
class SecuritySettings:
    """Wallet security settings."""

    address: str
    security_level: int
    session_timeout: int
    require_biometric: bool
    require_2fa: bool
    require_confirmation: bool
    whitelist_enabled: bool
    whitelisted_addresses: List[str]
    max_daily_transfer: Optional[str] = None
