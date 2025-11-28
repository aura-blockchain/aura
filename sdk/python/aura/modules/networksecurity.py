"""Module for network security operations."""

from typing import List, Optional, Dict, Any
from datetime import datetime
from ..types import (
    NetworkSecurityParams,
    SecurityThreat,
    NetworkStatus,
    PeerReputation,
    RateLimitConfig,
    RateLimitStatus,
    MempoolStatus,
    GossipMetrics,
    ForkDetection,
    ThreatLevel,
    ThreatType,
    TxResult,
    GasOptions
)


class NetworkSecurityModule:
    """Network security module for protecting against attacks."""

    def __init__(self, client):
        """Initialize network security module."""
        self.client = client

    async def get_reputation_score(self, peer_id: str) -> Optional[PeerReputation]:
        """Get reputation score for a peer.

        Args:
            peer_id: Peer ID

        Returns:
            Peer reputation or None
        """
        if not peer_id:
            raise ValueError("Peer ID is required")

        try:
            data = await self.client.get(f"/aura/networksecurity/v1beta1/reputation/{peer_id}")
            rep_data = data.get("reputation")

            if not rep_data:
                return None

            return PeerReputation(
                peer_id=rep_data.get("peer_id", peer_id),
                ip_address=rep_data.get("ip_address", ""),
                reputation_score=rep_data.get("reputation_score", 0.0),
                successful_interactions=rep_data.get("successful_interactions", 0),
                failed_interactions=rep_data.get("failed_interactions", 0),
                last_seen=datetime.fromisoformat(rep_data.get("last_seen")) if rep_data.get("last_seen") else datetime.now(),
                is_trusted=rep_data.get("is_trusted", False),
                is_blacklisted=rep_data.get("is_blacklisted", False)
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get reputation score: {e}")

    async def report_malicious_node(
        self,
        peer_id: str,
        threat_type: ThreatType,
        evidence: str,
        reporter: str,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Report a malicious node.

        Args:
            peer_id: Malicious peer ID
            threat_type: Type of threat
            evidence: Evidence of malicious behavior
            reporter: Reporter address
            options: Transaction options

        Returns:
            Transaction result
        """
        if not peer_id:
            raise ValueError("Peer ID is required")
        if not evidence:
            raise ValueError("Evidence is required")
        if not reporter:
            raise ValueError("Reporter address is required")

        message = {
            "@type": "/aura.networksecurity.v1beta1.MsgReportMaliciousNode",
            "peer_id": peer_id,
            "threat_type": threat_type.value if isinstance(threat_type, ThreatType) else threat_type,
            "evidence": evidence,
            "reporter": reporter
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def get_rate_limits(self, entity_id: str) -> Optional[RateLimitStatus]:
        """Get rate limit status for an entity.

        Args:
            entity_id: Entity ID

        Returns:
            Rate limit status or None
        """
        if not entity_id:
            raise ValueError("Entity ID is required")

        try:
            data = await self.client.get(f"/aura/networksecurity/v1beta1/ratelimits/{entity_id}")
            limit_data = data.get("rate_limit")

            if not limit_data:
                return None

            return RateLimitStatus(
                entity_id=limit_data.get("entity_id", entity_id),
                current_rate=limit_data.get("current_rate", 0.0),
                limit_exceeded=limit_data.get("limit_exceeded", False),
                reset_at=datetime.fromisoformat(limit_data.get("reset_at")) if limit_data.get("reset_at") else datetime.now(),
                blocked_until=datetime.fromisoformat(limit_data.get("blocked_until")) if limit_data.get("blocked_until") else None
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get rate limits: {e}")

    async def update_security_params(
        self,
        params: NetworkSecurityParams,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Update network security parameters (admin only).

        Args:
            params: Security parameters
            options: Transaction options

        Returns:
            Transaction result
        """
        message = {
            "@type": "/aura.networksecurity.v1beta1.MsgUpdateSecurityParams",
            "rate_limit_enabled": params.rate_limit_enabled,
            "max_peers": params.max_peers,
            "max_connections_per_ip": params.max_connections_per_ip,
            "blacklist": params.blacklist,
            "whitelist": params.whitelist
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def get_security_threats(
        self,
        threat_level: Optional[ThreatLevel] = None,
        limit: int = 100
    ) -> List[SecurityThreat]:
        """Get security threats.

        Args:
            threat_level: Optional threat level filter
            limit: Maximum number of results

        Returns:
            List of security threats
        """
        try:
            params = {"limit": limit}
            if threat_level:
                params["threat_level"] = threat_level.value if isinstance(threat_level, ThreatLevel) else threat_level

            query_str = "&".join([f"{k}={v}" for k, v in params.items()])
            data = await self.client.get(f"/aura/networksecurity/v1beta1/threats?{query_str}")

            threats = []
            for threat_data in data.get("threats", []):
                threats.append(SecurityThreat(
                    threat_id=threat_data.get("threat_id", ""),
                    threat_type=ThreatType(threat_data.get("threat_type", "spam")),
                    threat_level=ThreatLevel(threat_data.get("threat_level", "low")),
                    source=threat_data.get("source", ""),
                    target=threat_data.get("target"),
                    detected_at=datetime.fromisoformat(threat_data.get("detected_at")) if threat_data.get("detected_at") else datetime.now(),
                    resolved_at=datetime.fromisoformat(threat_data.get("resolved_at")) if threat_data.get("resolved_at") else None,
                    description=threat_data.get("description", ""),
                    mitigated=threat_data.get("mitigated", False),
                    mitigation_action=threat_data.get("mitigation_action")
                ))

            return threats
        except Exception as e:
            raise RuntimeError(f"Failed to get security threats: {e}")

    async def get_network_status(self) -> NetworkStatus:
        """Get current network security status.

        Returns:
            Network status
        """
        try:
            data = await self.client.get("/aura/networksecurity/v1beta1/status")
            status_str = data.get("status", "normal")
            return NetworkStatus(status_str)
        except Exception as e:
            raise RuntimeError(f"Failed to get network status: {e}")

    async def get_mempool_status(self) -> MempoolStatus:
        """Get mempool security status.

        Returns:
            Mempool status
        """
        try:
            data = await self.client.get("/aura/networksecurity/v1beta1/mempool")
            mempool_data = data.get("mempool", {})

            return MempoolStatus(
                size=mempool_data.get("size", 0),
                max_size=mempool_data.get("max_size", 0),
                pending_txs=mempool_data.get("pending_txs", 0),
                spam_detected=mempool_data.get("spam_detected", 0),
                rejected_txs=mempool_data.get("rejected_txs", 0),
                average_fee=mempool_data.get("average_fee", "0")
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get mempool status: {e}")

    async def get_gossip_metrics(self) -> GossipMetrics:
        """Get gossip protocol metrics.

        Returns:
            Gossip metrics
        """
        try:
            data = await self.client.get("/aura/networksecurity/v1beta1/gossip")
            gossip_data = data.get("metrics", {})

            return GossipMetrics(
                messages_sent=gossip_data.get("messages_sent", 0),
                messages_received=gossip_data.get("messages_received", 0),
                messages_dropped=gossip_data.get("messages_dropped", 0),
                peer_count=gossip_data.get("peer_count", 0),
                bandwidth_used=gossip_data.get("bandwidth_used", 0),
                latency_avg=gossip_data.get("latency_avg", 0.0)
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get gossip metrics: {e}")

    async def check_fork_detection(self) -> ForkDetection:
        """Check for network forks.

        Returns:
            Fork detection information
        """
        try:
            data = await self.client.get("/aura/networksecurity/v1beta1/fork-detection")
            fork_data = data.get("fork", {})

            return ForkDetection(
                fork_detected=fork_data.get("fork_detected", False),
                fork_height=fork_data.get("fork_height"),
                fork_hash=fork_data.get("fork_hash"),
                detected_at=datetime.fromisoformat(fork_data.get("detected_at")) if fork_data.get("detected_at") else None,
                resolution_status=fork_data.get("resolution_status", "none")
            )
        except Exception as e:
            raise RuntimeError(f"Failed to check fork detection: {e}")

    async def blacklist_peer(
        self,
        peer_id: str,
        reason: str,
        duration_blocks: int,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Blacklist a peer.

        Args:
            peer_id: Peer ID
            reason: Blacklist reason
            duration_blocks: Duration in blocks
            options: Transaction options

        Returns:
            Transaction result
        """
        if not peer_id:
            raise ValueError("Peer ID is required")
        if not reason:
            raise ValueError("Reason is required")

        message = {
            "@type": "/aura.networksecurity.v1beta1.MsgBlacklistPeer",
            "peer_id": peer_id,
            "reason": reason,
            "duration_blocks": duration_blocks
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def whitelist_peer(
        self,
        peer_id: str,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Whitelist a trusted peer.

        Args:
            peer_id: Peer ID
            options: Transaction options

        Returns:
            Transaction result
        """
        if not peer_id:
            raise ValueError("Peer ID is required")

        message = {
            "@type": "/aura.networksecurity.v1beta1.MsgWhitelistPeer",
            "peer_id": peer_id
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def configure_rate_limit(
        self,
        config: RateLimitConfig,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Configure rate limiting.

        Args:
            config: Rate limit configuration
            options: Transaction options

        Returns:
            Transaction result
        """
        message = {
            "@type": "/aura.networksecurity.v1beta1.MsgConfigureRateLimit",
            "max_requests_per_second": config.max_requests_per_second,
            "max_requests_per_minute": config.max_requests_per_minute,
            "burst_size": config.burst_size,
            "penalty_duration": config.penalty_duration
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)
