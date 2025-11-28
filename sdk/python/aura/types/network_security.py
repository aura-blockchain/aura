"""Type definitions for NetworkSecurity module."""

from dataclasses import dataclass
from typing import Optional, List, Dict
from datetime import datetime
from enum import Enum


class ThreatLevel(Enum):
    """Security threat levels."""

    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    CRITICAL = "critical"


class NetworkStatus(Enum):
    """Network operational status."""

    NORMAL = "normal"
    UNDER_ATTACK = "under_attack"
    RECOVERING = "recovering"
    MAINTENANCE = "maintenance"


class ThreatType(Enum):
    """Types of security threats."""

    SYBIL = "sybil"
    ECLIPSE = "eclipse"
    DOS = "dos"
    DDOS = "ddos"
    SPAM = "spam"
    MALICIOUS_PEER = "malicious_peer"


@dataclass
class NetworkSecurityParams:
    """Network security configuration."""

    rate_limit_enabled: bool
    max_peers: int
    max_connections_per_ip: int
    blacklist: List[str]
    whitelist: List[str]


@dataclass
class SecurityThreat:
    """Security threat information."""

    threat_id: str
    threat_type: ThreatType
    threat_level: ThreatLevel
    source: str
    target: Optional[str]
    detected_at: datetime
    description: str
    resolved_at: Optional[datetime] = None
    mitigated: bool = False
    mitigation_action: Optional[str] = None


@dataclass
class PeerReputation:
    """Peer reputation score."""

    peer_id: str
    ip_address: str
    reputation_score: float
    successful_interactions: int
    failed_interactions: int
    last_seen: datetime
    is_trusted: bool
    is_blacklisted: bool


@dataclass
class RateLimitConfig:
    """Rate limiting configuration."""

    max_requests_per_second: int
    max_requests_per_minute: int
    burst_size: int
    penalty_duration: int


@dataclass
class RateLimitStatus:
    """Rate limit status for an entity."""

    entity_id: str
    current_rate: float
    limit_exceeded: bool
    reset_at: datetime
    blocked_until: Optional[datetime] = None


@dataclass
class MempoolStatus:
    """Mempool security status."""

    size: int
    max_size: int
    pending_txs: int
    spam_detected: int
    rejected_txs: int
    average_fee: str


@dataclass
class GossipMetrics:
    """Gossip protocol metrics."""

    messages_sent: int
    messages_received: int
    messages_dropped: int
    peer_count: int
    bandwidth_used: int
    latency_avg: float


@dataclass
class ForkDetection:
    """Fork detection information."""

    fork_detected: bool
    fork_height: Optional[int] = None
    fork_hash: Optional[str] = None
    detected_at: Optional[datetime] = None
    resolution_status: str = "none"
