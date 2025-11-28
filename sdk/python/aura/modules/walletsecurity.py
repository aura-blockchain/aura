"""Module for wallet security operations."""

from typing import List, Optional, Dict, Any
from datetime import datetime
from ..types import (
    WalletSecurityParams,
    SessionParams,
    WalletSession,
    BiometricData,
    MultisigConfig,
    TransactionConfirmation,
    TwoFactorAuth,
    HardwareKeyInfo,
    SocialRecoveryConfig,
    RecoveryRequest,
    SecurityAlert,
    AccessLog,
    SecuritySettings,
    SessionStatus,
    BiometricType,
    SecurityFeature,
    TxResult,
    GasOptions
)


class WalletSecurityModule:
    """Wallet security module for protecting user wallets."""

    def __init__(self, client):
        """Initialize wallet security module."""
        self.client = client

    async def enable_multisig(
        self,
        address: str,
        threshold: int,
        signers: List[str],
        weights: Optional[Dict[str, int]] = None,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Enable multi-signature protection.

        Args:
            address: Wallet address
            threshold: Signature threshold
            signers: List of signer addresses
            weights: Optional signer weights
            options: Transaction options

        Returns:
            Transaction result
        """
        if not address:
            raise ValueError("Address is required")
        if threshold <= 0:
            raise ValueError("Valid threshold is required")
        if not signers or len(signers) == 0:
            raise ValueError("At least one signer is required")

        message = {
            "@type": "/aura.walletsecurity.v1beta1.MsgEnableMultisig",
            "address": address,
            "threshold": threshold,
            "signers": signers,
            "weights": weights or {}
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def create_session(
        self,
        params: SessionParams,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Create a wallet session.

        Args:
            params: Session parameters
            options: Transaction options

        Returns:
            Transaction result
        """
        if not params.address:
            raise ValueError("Address is required")
        if not params.device_id:
            raise ValueError("Device ID is required")
        if not params.ip_address:
            raise ValueError("IP address is required")

        message = {
            "@type": "/aura.walletsecurity.v1beta1.MsgCreateSession",
            "address": params.address,
            "device_id": params.device_id,
            "ip_address": params.ip_address,
            "user_agent": params.user_agent,
            "duration_seconds": params.duration_seconds
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def validate_biometric(
        self,
        address: str,
        biometric_type: BiometricType,
        template_hash: str,
        device_id: str,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Validate biometric authentication.

        Args:
            address: Wallet address
            biometric_type: Type of biometric
            template_hash: Biometric template hash
            device_id: Device ID
            options: Transaction options

        Returns:
            Transaction result
        """
        if not address:
            raise ValueError("Address is required")
        if not template_hash:
            raise ValueError("Template hash is required")
        if not device_id:
            raise ValueError("Device ID is required")

        message = {
            "@type": "/aura.walletsecurity.v1beta1.MsgValidateBiometric",
            "address": address,
            "biometric_type": biometric_type.value if isinstance(biometric_type, BiometricType) else biometric_type,
            "template_hash": template_hash,
            "device_id": device_id
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def get_security_settings(self, address: str) -> Optional[SecuritySettings]:
        """Get wallet security settings.

        Args:
            address: Wallet address

        Returns:
            Security settings or None
        """
        if not address:
            raise ValueError("Address is required")

        try:
            data = await self.client.get(f"/aura/walletsecurity/v1beta1/settings/{address}")
            settings_data = data.get("settings")

            if not settings_data:
                return None

            return SecuritySettings(
                address=settings_data.get("address", address),
                security_level=settings_data.get("security_level", 1),
                session_timeout=settings_data.get("session_timeout", 3600),
                require_biometric=settings_data.get("require_biometric", False),
                require_2fa=settings_data.get("require_2fa", False),
                require_confirmation=settings_data.get("require_confirmation", False),
                max_daily_transfer=settings_data.get("max_daily_transfer"),
                whitelist_enabled=settings_data.get("whitelist_enabled", False),
                whitelisted_addresses=settings_data.get("whitelisted_addresses", [])
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get security settings: {e}")

    async def update_security_params(
        self,
        address: str,
        params: WalletSecurityParams,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Update wallet security parameters.

        Args:
            address: Wallet address
            params: Security parameters
            options: Transaction options

        Returns:
            Transaction result
        """
        if not address:
            raise ValueError("Address is required")

        message = {
            "@type": "/aura.walletsecurity.v1beta1.MsgUpdateSecurityParams",
            "address": address,
            "security_level": params.security_level,
            "features": [f.value if isinstance(f, SecurityFeature) else f for f in params.features],
            "session_timeout": params.session_timeout,
            "require_confirmation": params.require_confirmation
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def get_active_sessions(self, address: str) -> List[WalletSession]:
        """Get active sessions for a wallet.

        Args:
            address: Wallet address

        Returns:
            List of active sessions
        """
        if not address:
            raise ValueError("Address is required")

        try:
            data = await self.client.get(f"/aura/walletsecurity/v1beta1/sessions/{address}")

            sessions = []
            for session_data in data.get("sessions", []):
                sessions.append(WalletSession(
                    session_id=session_data.get("session_id", ""),
                    address=session_data.get("address", address),
                    device_id=session_data.get("device_id", ""),
                    status=SessionStatus(session_data.get("status", "active")),
                    created_at=datetime.fromisoformat(session_data.get("created_at")) if session_data.get("created_at") else datetime.now(),
                    expires_at=datetime.fromisoformat(session_data.get("expires_at")) if session_data.get("expires_at") else datetime.now(),
                    last_activity=datetime.fromisoformat(session_data.get("last_activity")) if session_data.get("last_activity") else datetime.now(),
                    ip_address=session_data.get("ip_address", ""),
                    location=session_data.get("location"),
                    user_agent=session_data.get("user_agent")
                ))

            return sessions
        except Exception as e:
            raise RuntimeError(f"Failed to get active sessions: {e}")

    async def revoke_session(
        self,
        session_id: str,
        address: str,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Revoke a wallet session.

        Args:
            session_id: Session ID
            address: Wallet address
            options: Transaction options

        Returns:
            Transaction result
        """
        if not session_id:
            raise ValueError("Session ID is required")
        if not address:
            raise ValueError("Address is required")

        message = {
            "@type": "/aura.walletsecurity.v1beta1.MsgRevokeSession",
            "session_id": session_id,
            "address": address
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def register_biometric(
        self,
        biometric_data: BiometricData,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Register biometric authentication.

        Args:
            biometric_data: Biometric data
            options: Transaction options

        Returns:
            Transaction result
        """
        if not biometric_data.address:
            raise ValueError("Address is required")
        if not biometric_data.template_hash:
            raise ValueError("Template hash is required")
        if not biometric_data.device_id:
            raise ValueError("Device ID is required")

        message = {
            "@type": "/aura.walletsecurity.v1beta1.MsgRegisterBiometric",
            "address": biometric_data.address,
            "biometric_type": biometric_data.biometric_type.value if isinstance(biometric_data.biometric_type, BiometricType) else biometric_data.biometric_type,
            "template_hash": biometric_data.template_hash,
            "device_id": biometric_data.device_id,
            "enabled": biometric_data.enabled
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def configure_social_recovery(
        self,
        config: SocialRecoveryConfig,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Configure social recovery.

        Args:
            config: Social recovery configuration
            options: Transaction options

        Returns:
            Transaction result
        """
        if not config.address:
            raise ValueError("Address is required")
        if not config.guardians or len(config.guardians) == 0:
            raise ValueError("At least one guardian is required")
        if config.threshold <= 0 or config.threshold > len(config.guardians):
            raise ValueError("Invalid threshold")

        message = {
            "@type": "/aura.walletsecurity.v1beta1.MsgConfigureSocialRecovery",
            "address": config.address,
            "guardians": config.guardians,
            "threshold": config.threshold,
            "timeout_period": config.timeout_period,
            "active": config.active
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def initiate_recovery(
        self,
        address: str,
        new_owner: str,
        guardian_signatures: List[str],
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Initiate wallet recovery.

        Args:
            address: Wallet address to recover
            new_owner: New owner address
            guardian_signatures: Guardian signatures
            options: Transaction options

        Returns:
            Transaction result
        """
        if not address:
            raise ValueError("Address is required")
        if not new_owner:
            raise ValueError("New owner address is required")
        if not guardian_signatures or len(guardian_signatures) == 0:
            raise ValueError("At least one guardian signature is required")

        message = {
            "@type": "/aura.walletsecurity.v1beta1.MsgInitiateRecovery",
            "address": address,
            "new_owner": new_owner,
            "guardian_signatures": guardian_signatures
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def get_multisig_config(self, address: str) -> Optional[MultisigConfig]:
        """Get multisig configuration.

        Args:
            address: Wallet address

        Returns:
            Multisig configuration or None
        """
        if not address:
            raise ValueError("Address is required")

        try:
            data = await self.client.get(f"/aura/walletsecurity/v1beta1/multisig/{address}")
            config_data = data.get("config")

            if not config_data:
                return None

            return MultisigConfig(
                address=config_data.get("address", address),
                threshold=config_data.get("threshold", 0),
                signers=config_data.get("signers", []),
                weights=config_data.get("weights", {}),
                created_at=datetime.fromisoformat(config_data.get("created_at")) if config_data.get("created_at") else datetime.now(),
                updated_at=datetime.fromisoformat(config_data.get("updated_at")) if config_data.get("updated_at") else datetime.now(),
                total_transactions=config_data.get("total_transactions", 0)
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get multisig config: {e}")

    async def get_security_alerts(
        self,
        address: str,
        limit: int = 100
    ) -> List[SecurityAlert]:
        """Get security alerts for a wallet.

        Args:
            address: Wallet address
            limit: Maximum number of results

        Returns:
            List of security alerts
        """
        if not address:
            raise ValueError("Address is required")

        try:
            data = await self.client.get(f"/aura/walletsecurity/v1beta1/alerts/{address}?limit={limit}")

            alerts = []
            for alert_data in data.get("alerts", []):
                alerts.append(SecurityAlert(
                    alert_id=alert_data.get("alert_id", ""),
                    address=alert_data.get("address", address),
                    alert_type=alert_data.get("alert_type", ""),
                    severity=alert_data.get("severity", "info"),
                    message=alert_data.get("message", ""),
                    detected_at=datetime.fromisoformat(alert_data.get("detected_at")) if alert_data.get("detected_at") else datetime.now(),
                    acknowledged=alert_data.get("acknowledged", False),
                    details=alert_data.get("details")
                ))

            return alerts
        except Exception as e:
            raise RuntimeError(f"Failed to get security alerts: {e}")

    async def get_access_logs(
        self,
        address: str,
        limit: int = 100
    ) -> List[AccessLog]:
        """Get access logs for a wallet.

        Args:
            address: Wallet address
            limit: Maximum number of results

        Returns:
            List of access logs
        """
        if not address:
            raise ValueError("Address is required")

        try:
            data = await self.client.get(f"/aura/walletsecurity/v1beta1/access-logs/{address}?limit={limit}")

            logs = []
            for log_data in data.get("logs", []):
                logs.append(AccessLog(
                    log_id=log_data.get("log_id", ""),
                    address=log_data.get("address", address),
                    action=log_data.get("action", ""),
                    device_id=log_data.get("device_id", ""),
                    ip_address=log_data.get("ip_address", ""),
                    location=log_data.get("location"),
                    success=log_data.get("success", False),
                    timestamp=datetime.fromisoformat(log_data.get("timestamp")) if log_data.get("timestamp") else datetime.now(),
                    details=log_data.get("details")
                ))

            return logs
        except Exception as e:
            raise RuntimeError(f"Failed to get access logs: {e}")
