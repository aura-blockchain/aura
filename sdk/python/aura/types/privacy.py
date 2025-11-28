"""Type definitions for Privacy module."""

from dataclasses import dataclass
from typing import Optional, List, Dict
from datetime import datetime
from enum import Enum


class PrivacyLevel(Enum):
    """Privacy protection levels."""

    NONE = "none"
    BASIC = "basic"
    STANDARD = "standard"
    MAXIMUM = "maximum"


class MixingStrategy(Enum):
    """Token mixing strategies."""

    TORNADO = "tornado"
    COINJOIN = "coinjoin"
    RING_SIGNATURE = "ring_signature"
    STEALTH = "stealth"


@dataclass
class PrivacyParams:
    """Privacy operation parameters."""

    level: PrivacyLevel
    use_mixing: bool = True
    use_stealth_address: bool = False
    mixing_rounds: int = 3


@dataclass
class ConfidentialTransaction:
    """Confidential transaction details."""

    tx_id: str
    sender_commitment: str
    receiver_commitment: str
    amount_commitment: str
    range_proof: str
    created_at: datetime
    privacy_level: PrivacyLevel
    memo: Optional[str] = None


@dataclass
class StealthAddress:
    """Stealth address information."""

    address: str
    public_view_key: str
    public_spend_key: str
    created_at: datetime
    usage_count: int


@dataclass
class MixingPool:
    """Token mixing pool information."""

    pool_id: str
    token: str
    denomination: str
    total_deposits: int
    available_liquidity: str
    mixing_fee: str
    anonymity_set_size: int


@dataclass
class RingSignatureParams:
    """Ring signature parameters."""

    message: str
    ring_members: List[str]
    key_image: str
    ring_size: int


@dataclass
class RingSignature:
    """Ring signature data."""

    signature: str
    ring_members: List[str]
    key_image: str
    ring_size: int
    created_at: datetime
    verified: bool = False


@dataclass
class PrivacyProof:
    """Zero-knowledge privacy proof."""

    proof_type: str
    proof_data: str
    public_inputs: List[str]
    verified: bool
    created_at: datetime
    verifier_address: Optional[str] = None


@dataclass
class ConfidentialBalance:
    """Confidential balance information."""

    address: str
    balance_commitment: str
    blinding_factor: str
    encrypted_amount: str
    last_updated: datetime


@dataclass
class MixingTransaction:
    """Mixing transaction record."""

    mix_id: str
    strategy: MixingStrategy
    input_amount: str
    output_amount: str
    fee: str
    rounds: int
    anonymity_set: int
    started_at: datetime
    status: str
    completed_at: Optional[datetime] = None


@dataclass
class PrivacySettings:
    """Privacy configuration settings."""

    default_privacy_level: PrivacyLevel
    auto_mix: bool
    min_anonymity_set: int
    max_mixing_rounds: int
    stealth_addresses_enabled: bool
    ring_signature_size: int
