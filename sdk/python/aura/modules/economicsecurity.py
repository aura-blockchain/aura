"""Module for economic security operations."""

from typing import Optional, Dict, Any
from datetime import datetime
from ..types import (
    DynamicFeeParams,
    MEVProtectionParams,
    FeeStructure,
    WhaleProtectionParams,
    EconomicMetrics,
    FeeConfiguration,
    MEVReport,
    CircuitBreakerParams,
    CircuitBreakerStatus,
    TxResult,
    GasOptions
)


class EconomicSecurityModule:
    """Economic security module for fee management and MEV protection."""

    def __init__(self, client):
        """Initialize economic security module."""
        self.client = client

    async def get_dynamic_fees(self, params: DynamicFeeParams) -> FeeStructure:
        """Calculate dynamic fees based on network conditions.

        Args:
            params: Fee calculation parameters

        Returns:
            Fee structure
        """
        if params.transaction_size <= 0:
            raise ValueError("Valid transaction size is required")

        try:
            request_data = {
                "transaction_size": params.transaction_size,
                "priority": params.priority,
                "network_congestion": params.network_congestion
            }

            data = await self.client.post("/aura/economicsecurity/v1beta1/fees/calculate", request_data)
            fee_data = data.get("fee_structure", {})

            return FeeStructure(
                base_fee=fee_data.get("base_fee", "0"),
                priority_fee=fee_data.get("priority_fee", "0"),
                total_fee=fee_data.get("total_fee", "0"),
                gas_price=fee_data.get("gas_price", "0"),
                estimated_time=fee_data.get("estimated_time", 0)
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get dynamic fees: {e}")

    async def update_fee_params(
        self,
        config: FeeConfiguration,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Update fee parameters (admin only).

        Args:
            config: Fee configuration
            options: Transaction options

        Returns:
            Transaction result
        """
        if not config.min_gas_price:
            raise ValueError("Min gas price is required")

        message = {
            "@type": "/aura.economicsecurity.v1beta1.MsgUpdateFeeParams",
            "min_gas_price": config.min_gas_price,
            "max_gas_price": config.max_gas_price,
            "base_fee_multiplier": str(config.base_fee_multiplier),
            "congestion_threshold": str(config.congestion_threshold),
            "whale_threshold": config.whale_threshold
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def get_mev_protection(self, tx_data: str) -> MEVReport:
        """Get MEV protection analysis for a transaction.

        Args:
            tx_data: Transaction data

        Returns:
            MEV report
        """
        if not tx_data:
            raise ValueError("Transaction data is required")

        try:
            request_data = {"tx_data": tx_data}
            data = await self.client.post("/aura/economicsecurity/v1beta1/mev/analyze", request_data)

            report_data = data.get("report", {})

            return MEVReport(
                tx_hash=report_data.get("tx_hash", ""),
                mev_detected=report_data.get("mev_detected", False),
                mev_amount=report_data.get("mev_amount", "0"),
                protection_applied=report_data.get("protection_applied", False),
                timestamp=datetime.fromisoformat(report_data.get("timestamp")) if report_data.get("timestamp") else datetime.now(),
                details=report_data.get("details")
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get MEV protection: {e}")

    async def enable_mev_protection(
        self,
        params: MEVProtectionParams,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Enable MEV protection for a transaction.

        Args:
            params: MEV protection parameters
            options: Transaction options

        Returns:
            Transaction result
        """
        if not params.transaction_data:
            raise ValueError("Transaction data is required")

        message = {
            "@type": "/aura.economicsecurity.v1beta1.MsgEnableMEVProtection",
            "transaction_data": params.transaction_data,
            "max_slippage": str(params.max_slippage),
            "deadline": params.deadline,
            "use_private_mempool": params.use_private_mempool
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def report_whale_activity(
        self,
        params: WhaleProtectionParams,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Report whale activity for monitoring.

        Args:
            params: Whale protection parameters
            options: Transaction options

        Returns:
            Transaction result
        """
        if not params.address:
            raise ValueError("Address is required")
        if not params.amount:
            raise ValueError("Amount is required")
        if not params.token:
            raise ValueError("Token is required")

        message = {
            "@type": "/aura.economicsecurity.v1beta1.MsgReportWhaleActivity",
            "address": params.address,
            "amount": params.amount,
            "token": params.token
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def get_economic_stats(self) -> EconomicMetrics:
        """Get economic security statistics.

        Returns:
            Economic metrics
        """
        try:
            data = await self.client.get("/aura/economicsecurity/v1beta1/stats")
            stats = data.get("metrics", {})

            return EconomicMetrics(
                total_value_locked=stats.get("total_value_locked", "0"),
                circulating_supply=stats.get("circulating_supply", "0"),
                inflation_rate=stats.get("inflation_rate", 0.0),
                staking_ratio=stats.get("staking_ratio", 0.0),
                average_fee=stats.get("average_fee", "0"),
                mev_prevented=stats.get("mev_prevented", "0"),
                timestamp=datetime.fromisoformat(stats.get("timestamp")) if stats.get("timestamp") else datetime.now()
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get economic stats: {e}")

    async def get_fee_config(self) -> FeeConfiguration:
        """Get current fee configuration.

        Returns:
            Fee configuration
        """
        try:
            data = await self.client.get("/aura/economicsecurity/v1beta1/fees/config")
            config = data.get("config", {})

            return FeeConfiguration(
                min_gas_price=config.get("min_gas_price", "0"),
                max_gas_price=config.get("max_gas_price", "0"),
                base_fee_multiplier=config.get("base_fee_multiplier", 1.0),
                congestion_threshold=config.get("congestion_threshold", 0.8),
                whale_threshold=config.get("whale_threshold", "0")
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get fee config: {e}")

    async def set_circuit_breaker(
        self,
        params: CircuitBreakerParams,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Configure circuit breaker parameters.

        Args:
            params: Circuit breaker parameters
            options: Transaction options

        Returns:
            Transaction result
        """
        if not params.threshold:
            raise ValueError("Threshold is required")

        message = {
            "@type": "/aura.economicsecurity.v1beta1.MsgSetCircuitBreaker",
            "threshold": params.threshold,
            "window_duration": params.window_duration,
            "cooldown_period": params.cooldown_period
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def get_circuit_breaker_status(self) -> CircuitBreakerStatus:
        """Get circuit breaker status.

        Returns:
            Circuit breaker status
        """
        try:
            data = await self.client.get("/aura/economicsecurity/v1beta1/circuit-breaker/status")
            status = data.get("status", {})

            return CircuitBreakerStatus(
                active=status.get("active", False),
                triggered_at=datetime.fromisoformat(status.get("triggered_at")) if status.get("triggered_at") else None,
                reason=status.get("reason"),
                reset_at=datetime.fromisoformat(status.get("reset_at")) if status.get("reset_at") else None
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get circuit breaker status: {e}")

    async def get_whale_alerts(self, limit: int = 100) -> list:
        """Get whale activity alerts.

        Args:
            limit: Maximum number of results

        Returns:
            List of whale alerts
        """
        try:
            data = await self.client.get(f"/aura/economicsecurity/v1beta1/whale-alerts?limit={limit}")
            return data.get("alerts", [])
        except Exception as e:
            raise RuntimeError(f"Failed to get whale alerts: {e}")
