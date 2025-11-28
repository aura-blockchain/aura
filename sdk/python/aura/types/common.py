"""Common type definitions for AURA SDK."""

from dataclasses import dataclass
from typing import Optional, List, Dict, Any
from datetime import datetime
from enum import IntEnum


@dataclass
class ChainConfig:
    """Chain configuration."""
    rpc_endpoint: str
    rest_endpoint: Optional[str] = None
    chain_id: str = "aura-1"
    prefix: str = "aura"
    gas_price: str = "0.025uaura"
    gas_adjustment: float = 1.5


@dataclass
class Coin:
    """Coin representation."""
    denom: str
    amount: str


@dataclass
class TxResult:
    """Transaction result."""
    transaction_hash: str
    height: int
    code: int
    raw_log: Optional[str] = None
    gas_used: int = 0
    gas_wanted: int = 0


@dataclass
class GasOptions:
    """Gas configuration options."""
    gas_limit: Optional[int] = None
    gas_price: Optional[str] = None
    memo: str = ""


@dataclass
class SendParams:
    """Parameters for sending tokens."""
    recipient: str
    amount: str
    denom: str = "uaura"
    memo: str = ""


@dataclass
class TxResponse:
    """Transaction response."""

    tx_hash: str
    height: int
    code: int
    raw_log: Optional[str] = None
    gas_used: int = 0
    gas_wanted: int = 0
    data: Optional[str] = None
    timestamp: Optional[datetime] = None


@dataclass
class QueryResponse:
    """Generic query response."""

    data: Dict[str, Any]
    height: int
    timestamp: Optional[datetime] = None


@dataclass
class Pagination:
    """Pagination parameters."""

    key: Optional[str] = None
    offset: int = 0
    limit: int = 100
    count_total: bool = False
    reverse: bool = False


@dataclass
class PaginationResponse:
    """Pagination response."""

    next_key: Optional[str] = None
    total: int = 0


@dataclass
class Pool:
    """DEX pool."""
    id: str
    token_a: str
    token_b: str
    reserve_a: str
    reserve_b: str
    total_shares: str
    swap_fee: str


@dataclass
class ValidatorDescription:
    """Validator description."""
    moniker: str
    identity: str = ""
    website: str = ""
    security_contact: str = ""
    details: str = ""


@dataclass
class ValidatorCommission:
    """Validator commission."""
    rate: str
    max_rate: str
    max_change_rate: str


@dataclass
class Validator:
    """Validator information."""
    operator_address: str
    consensus_pubkey: str
    jailed: bool
    status: int
    tokens: str
    delegator_shares: str
    description: ValidatorDescription
    commission: ValidatorCommission


@dataclass
class TallyResult:
    """Proposal tally result."""
    yes: str
    abstain: str
    no: str
    no_with_veto: str


@dataclass
class Proposal:
    """Governance proposal."""
    proposal_id: str
    status: int
    final_tally_result: TallyResult
    submit_time: datetime
    deposit_end_time: datetime
    total_deposit: List[Coin]
    voting_start_time: datetime
    voting_end_time: datetime


class VoteOption(IntEnum):
    """Vote options for governance."""
    UNSPECIFIED = 0
    YES = 1
    ABSTAIN = 2
    NO = 3
    NO_WITH_VETO = 4


@dataclass
class PoolParams:
    """Parameters for creating a pool."""
    token_a: str
    token_b: str
    amount_a: str
    amount_b: str


@dataclass
class SwapParams:
    """Parameters for swapping tokens."""
    pool_id: str
    token_in: str
    amount_in: str
    min_amount_out: str
    recipient: Optional[str] = None


@dataclass
class AddLiquidityParams:
    """Parameters for adding liquidity."""
    pool_id: str
    amount_a: str
    amount_b: str
    min_shares: str


@dataclass
class RemoveLiquidityParams:
    """Parameters for removing liquidity."""
    pool_id: str
    shares: str
    min_amount_a: str
    min_amount_b: str


@dataclass
class DelegateParams:
    """Parameters for delegation."""
    validator_address: str
    amount: str
    denom: str = "uaura"


@dataclass
class UndelegateParams:
    """Parameters for undelegation."""
    validator_address: str
    amount: str
    denom: str = "uaura"


@dataclass
class RedelegateParams:
    """Parameters for redelegation."""
    src_validator_address: str
    dst_validator_address: str
    amount: str
    denom: str = "uaura"


@dataclass
class VoteParams:
    """Parameters for voting."""
    proposal_id: str
    option: VoteOption
    metadata: str = ""


@dataclass
class DepositParams:
    """Parameters for proposal deposit."""
    proposal_id: str
    amount: str
    denom: str = "uaura"
