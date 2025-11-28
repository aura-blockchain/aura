"""Module for validator security operations."""

from typing import List, Optional, Dict, Any
from datetime import datetime
from ..types import (
    ValidatorSecurityParams,
    ValidatorInfo,
    SlashingEvent,
    JailingParams,
    UnjailParams,
    ValidatorMonitoring,
    SentryNodeConfig,
    DoubleSignEvidence,
    ValidatorPerformance,
    HSMConfig,
    ValidatorAlert,
    ValidatorStatus,
    SlashingReason,
    SecurityLevel,
    TxResult,
    GasOptions
)


class ValidatorSecurityModule:
    """Validator security module for validator operations and monitoring."""

    def __init__(self, client):
        """Initialize validator security module."""
        self.client = client

    async def get_validator_status(self, validator_address: str) -> Optional[ValidatorInfo]:
        """Get validator status and information.

        Args:
            validator_address: Validator address

        Returns:
            Validator information or None
        """
        if not validator_address:
            raise ValueError("Validator address is required")

        try:
            data = await self.client.get(f"/aura/validatorsecurity/v1beta1/validators/{validator_address}")
            val_data = data.get("validator")

            if not val_data:
                return None

            return ValidatorInfo(
                operator_address=val_data.get("operator_address", validator_address),
                consensus_address=val_data.get("consensus_address", ""),
                status=ValidatorStatus(val_data.get("status", "active")),
                jailed=val_data.get("jailed", False),
                tombstoned=val_data.get("tombstoned", False),
                commission_rate=val_data.get("commission_rate", "0"),
                voting_power=val_data.get("voting_power", 0),
                uptime=val_data.get("uptime", 0.0),
                last_active=datetime.fromisoformat(val_data.get("last_active")) if val_data.get("last_active") else datetime.now(),
                jail_time=datetime.fromisoformat(val_data.get("jail_time")) if val_data.get("jail_time") else None
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get validator status: {e}")

    async def jail_validator(
        self,
        params: JailingParams,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Jail a validator for misbehavior.

        Args:
            params: Jailing parameters
            options: Transaction options

        Returns:
            Transaction result
        """
        if not params.validator_address:
            raise ValueError("Validator address is required")
        if not params.reason:
            raise ValueError("Reason is required")

        message = {
            "@type": "/aura.validatorsecurity.v1beta1.MsgJailValidator",
            "validator_address": params.validator_address,
            "reason": params.reason,
            "duration_blocks": params.duration_blocks,
            "evidence": params.evidence or ""
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def unjail_validator(
        self,
        params: UnjailParams,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Unjail a validator.

        Args:
            params: Unjailing parameters
            options: Transaction options

        Returns:
            Transaction result
        """
        if not params.validator_address:
            raise ValueError("Validator address is required")
        if not params.proof_of_correction:
            raise ValueError("Proof of correction is required")

        message = {
            "@type": "/aura.validatorsecurity.v1beta1.MsgUnjailValidator",
            "validator_address": params.validator_address,
            "proof_of_correction": params.proof_of_correction
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def slash_validator(
        self,
        validator_address: str,
        reason: SlashingReason,
        slash_percentage: float,
        evidence: Optional[str] = None,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Slash a validator for protocol violations.

        Args:
            validator_address: Validator address
            reason: Slashing reason
            slash_percentage: Percentage to slash
            evidence: Optional evidence
            options: Transaction options

        Returns:
            Transaction result
        """
        if not validator_address:
            raise ValueError("Validator address is required")
        if slash_percentage <= 0 or slash_percentage > 100:
            raise ValueError("Slash percentage must be between 0 and 100")

        message = {
            "@type": "/aura.validatorsecurity.v1beta1.MsgSlashValidator",
            "validator_address": validator_address,
            "reason": reason.value if isinstance(reason, SlashingReason) else reason,
            "slash_percentage": str(slash_percentage),
            "evidence": evidence or ""
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def get_slashing_history(
        self,
        validator_address: str,
        limit: int = 100
    ) -> List[SlashingEvent]:
        """Get slashing history for a validator.

        Args:
            validator_address: Validator address
            limit: Maximum number of results

        Returns:
            List of slashing events
        """
        if not validator_address:
            raise ValueError("Validator address is required")

        try:
            data = await self.client.get(f"/aura/validatorsecurity/v1beta1/slashing/{validator_address}?limit={limit}")

            events = []
            for event_data in data.get("events", []):
                events.append(SlashingEvent(
                    event_id=event_data.get("event_id", ""),
                    validator_address=event_data.get("validator_address", validator_address),
                    reason=SlashingReason(event_data.get("reason", "downtime")),
                    slash_amount=event_data.get("slash_amount", "0"),
                    slash_percentage=event_data.get("slash_percentage", 0.0),
                    occurred_at=datetime.fromisoformat(event_data.get("occurred_at")) if event_data.get("occurred_at") else datetime.now(),
                    height=event_data.get("height", 0),
                    evidence=event_data.get("evidence"),
                    jailed=event_data.get("jailed", False),
                    tombstoned=event_data.get("tombstoned", False)
                ))

            return events
        except Exception as e:
            raise RuntimeError(f"Failed to get slashing history: {e}")

    async def get_validator_monitoring(
        self,
        validator_address: str
    ) -> Optional[ValidatorMonitoring]:
        """Get monitoring data for a validator.

        Args:
            validator_address: Validator address

        Returns:
            Monitoring data or None
        """
        if not validator_address:
            raise ValueError("Validator address is required")

        try:
            data = await self.client.get(f"/aura/validatorsecurity/v1beta1/monitoring/{validator_address}")
            mon_data = data.get("monitoring")

            if not mon_data:
                return None

            return ValidatorMonitoring(
                validator_address=mon_data.get("validator_address", validator_address),
                uptime_percentage=mon_data.get("uptime_percentage", 0.0),
                missed_blocks=mon_data.get("missed_blocks", 0),
                total_blocks=mon_data.get("total_blocks", 0),
                last_signed_block=mon_data.get("last_signed_block", 0),
                consecutive_misses=mon_data.get("consecutive_misses", 0),
                health_score=mon_data.get("health_score", 0.0),
                alerts=mon_data.get("alerts", []),
                last_checked=datetime.fromisoformat(mon_data.get("last_checked")) if mon_data.get("last_checked") else datetime.now()
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get validator monitoring: {e}")

    async def configure_sentry_node(
        self,
        config: SentryNodeConfig,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Configure a sentry node.

        Args:
            config: Sentry node configuration
            options: Transaction options

        Returns:
            Transaction result
        """
        if not config.validator_address:
            raise ValueError("Validator address is required")
        if not config.ip_address:
            raise ValueError("IP address is required")

        message = {
            "@type": "/aura.validatorsecurity.v1beta1.MsgConfigureSentry",
            "sentry_id": config.sentry_id,
            "validator_address": config.validator_address,
            "ip_address": config.ip_address,
            "port": config.port,
            "active": config.active
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def report_double_sign(
        self,
        evidence: DoubleSignEvidence,
        reporter: str,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Report double signing evidence.

        Args:
            evidence: Double signing evidence
            reporter: Reporter address
            options: Transaction options

        Returns:
            Transaction result
        """
        if not evidence.validator_address:
            raise ValueError("Validator address is required")
        if not evidence.proof:
            raise ValueError("Proof is required")

        message = {
            "@type": "/aura.validatorsecurity.v1beta1.MsgReportDoubleSign",
            "validator_address": evidence.validator_address,
            "block_height": evidence.block_height,
            "vote_a": evidence.vote_a,
            "vote_b": evidence.vote_b,
            "proof": evidence.proof,
            "reporter": reporter
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def get_validator_performance(
        self,
        validator_address: str,
        start_date: Optional[datetime] = None,
        end_date: Optional[datetime] = None
    ) -> Optional[ValidatorPerformance]:
        """Get validator performance metrics.

        Args:
            validator_address: Validator address
            start_date: Optional start date
            end_date: Optional end date

        Returns:
            Performance metrics or None
        """
        if not validator_address:
            raise ValueError("Validator address is required")

        try:
            params = {"validator": validator_address}
            if start_date:
                params["start"] = start_date.isoformat()
            if end_date:
                params["end"] = end_date.isoformat()

            query_str = "&".join([f"{k}={v}" for k, v in params.items()])
            data = await self.client.get(f"/aura/validatorsecurity/v1beta1/performance?{query_str}")

            perf_data = data.get("performance")

            if not perf_data:
                return None

            return ValidatorPerformance(
                validator_address=perf_data.get("validator_address", validator_address),
                blocks_signed=perf_data.get("blocks_signed", 0),
                blocks_missed=perf_data.get("blocks_missed", 0),
                uptime_percentage=perf_data.get("uptime_percentage", 0.0),
                average_response_time=perf_data.get("average_response_time", 0.0),
                rewards_earned=perf_data.get("rewards_earned", "0"),
                slashing_events=perf_data.get("slashing_events", 0),
                period_start=datetime.fromisoformat(perf_data.get("period_start")) if perf_data.get("period_start") else (start_date or datetime.now()),
                period_end=datetime.fromisoformat(perf_data.get("period_end")) if perf_data.get("period_end") else (end_date or datetime.now())
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get validator performance: {e}")

    async def configure_hsm(
        self,
        config: HSMConfig,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Configure Hardware Security Module.

        Args:
            config: HSM configuration
            options: Transaction options

        Returns:
            Transaction result
        """
        if not config.validator_address:
            raise ValueError("Validator address is required")
        if not config.hsm_type:
            raise ValueError("HSM type is required")

        message = {
            "@type": "/aura.validatorsecurity.v1beta1.MsgConfigureHSM",
            "hsm_id": config.hsm_id,
            "validator_address": config.validator_address,
            "hsm_type": config.hsm_type,
            "enabled": config.enabled,
            "key_slot": config.key_slot,
            "firmware_version": config.firmware_version
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def get_validator_alerts(
        self,
        validator_address: str,
        limit: int = 100
    ) -> List[ValidatorAlert]:
        """Get security alerts for a validator.

        Args:
            validator_address: Validator address
            limit: Maximum number of results

        Returns:
            List of alerts
        """
        if not validator_address:
            raise ValueError("Validator address is required")

        try:
            data = await self.client.get(f"/aura/validatorsecurity/v1beta1/alerts/{validator_address}?limit={limit}")

            alerts = []
            for alert_data in data.get("alerts", []):
                alerts.append(ValidatorAlert(
                    alert_id=alert_data.get("alert_id", ""),
                    validator_address=alert_data.get("validator_address", validator_address),
                    alert_type=alert_data.get("alert_type", ""),
                    severity=alert_data.get("severity", "info"),
                    message=alert_data.get("message", ""),
                    created_at=datetime.fromisoformat(alert_data.get("created_at")) if alert_data.get("created_at") else datetime.now(),
                    acknowledged=alert_data.get("acknowledged", False),
                    resolved=alert_data.get("resolved", False)
                ))

            return alerts
        except Exception as e:
            raise RuntimeError(f"Failed to get validator alerts: {e}")

    async def update_security_params(
        self,
        validator_address: str,
        params: ValidatorSecurityParams,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Update validator security parameters.

        Args:
            validator_address: Validator address
            params: Security parameters
            options: Transaction options

        Returns:
            Transaction result
        """
        if not validator_address:
            raise ValueError("Validator address is required")

        message = {
            "@type": "/aura.validatorsecurity.v1beta1.MsgUpdateSecurityParams",
            "validator_address": validator_address,
            "security_level": params.security_level.value if isinstance(params.security_level, SecurityLevel) else params.security_level,
            "enable_sentry_nodes": params.enable_sentry_nodes,
            "enable_hsm": params.enable_hsm,
            "enable_monitoring": params.enable_monitoring
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)
