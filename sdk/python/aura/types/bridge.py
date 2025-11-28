"""Type definitions for Bridge module."""

from dataclasses import dataclass
from typing import Optional, List
from enum import Enum
from datetime import datetime


class BridgeStatus(Enum):
    """Bridge transfer status."""

    PENDING = "pending"
    CONFIRMED = "confirmed"
    COMPLETED = "completed"
    FAILED = "failed"
    REFUNDED = "refunded"


class ChainType(Enum):
    """Supported chain types."""

    ETHEREUM = "ethereum"
    BINANCE = "binance"
    POLYGON = "polygon"
    COSMOS = "cosmos"


@dataclass
class BridgeTransferParams:
    """Parameters for initiating bridge transfer."""

    source_chain: str
    destination_chain: str
    token: str
    amount: str
    recipient: str
    memo: Optional[str] = None
    timeout_height: Optional[int] = None
    timeout_timestamp: Optional[int] = None


@dataclass
class BridgeTransfer:
    """Bridge transfer information."""

    transfer_id: str
    source_chain: str
    destination_chain: str
    sender: str
    recipient: str
    token: str
    amount: str
    status: BridgeStatus
    created_at: datetime
    updated_at: datetime
    transaction_hash: Optional[str] = None
    destination_hash: Optional[str] = None
    proof: Optional[str] = None
    error_message: Optional[str] = None


@dataclass
class BridgeParams:
    """Bridge module parameters."""

    enabled: bool
    supported_chains: List[str]
    min_transfer_amount: str
    max_transfer_amount: str
    transfer_fee: str
    timeout_duration: int
    confirmation_blocks: int
    relayer_addresses: List[str]


@dataclass
class BridgeBalance:
    """Bridge balance information."""

    chain: str
    token: str
    locked_amount: str
    available_amount: str
    total_supply: str


@dataclass
class BridgeProof:
    """Bridge transfer proof."""

    transfer_id: str
    merkle_root: str
    merkle_proof: List[str]
    signatures: List[str]
    validators: List[str]


@dataclass
class RelayerInfo:
    """Bridge relayer information."""

    address: str
    chains: List[str]
    active: bool
    total_relayed: int
    success_rate: float
    last_relay_time: Optional[datetime] = None
