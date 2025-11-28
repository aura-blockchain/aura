"""Module for privacy operations."""

from typing import List, Optional, Dict, Any
from datetime import datetime
from ..types import (
    PrivacyParams,
    ConfidentialTransaction,
    StealthAddress,
    MixingPool,
    RingSignatureParams,
    RingSignature,
    PrivacyProof,
    ConfidentialBalance,
    MixingTransaction,
    PrivacySettings,
    PrivacyLevel,
    MixingStrategy,
    TxResult,
    GasOptions
)


class PrivacyModule:
    """Privacy module for confidential transactions and mixing."""

    def __init__(self, client):
        """Initialize privacy module."""
        self.client = client

    async def create_private_transaction(
        self,
        sender: str,
        recipient_commitment: str,
        amount_commitment: str,
        range_proof: str,
        params: PrivacyParams,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Create a confidential transaction.

        Args:
            sender: Sender address
            recipient_commitment: Recipient commitment
            amount_commitment: Amount commitment
            range_proof: Range proof
            params: Privacy parameters
            options: Transaction options

        Returns:
            Transaction result
        """
        if not sender:
            raise ValueError("Sender address is required")
        if not recipient_commitment:
            raise ValueError("Recipient commitment is required")
        if not amount_commitment:
            raise ValueError("Amount commitment is required")
        if not range_proof:
            raise ValueError("Range proof is required")

        message = {
            "@type": "/aura.privacy.v1beta1.MsgCreatePrivateTransaction",
            "sender": sender,
            "recipient_commitment": recipient_commitment,
            "amount_commitment": amount_commitment,
            "range_proof": range_proof,
            "privacy_level": params.level.value if isinstance(params.level, PrivacyLevel) else params.level,
            "use_mixing": params.use_mixing,
            "use_stealth_address": params.use_stealth_address,
            "mixing_rounds": params.mixing_rounds
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def mix_tokens(
        self,
        address: str,
        token: str,
        amount: str,
        strategy: MixingStrategy,
        rounds: int = 3,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Mix tokens for privacy.

        Args:
            address: Address
            token: Token denom
            amount: Amount to mix
            strategy: Mixing strategy
            rounds: Number of mixing rounds
            options: Transaction options

        Returns:
            Transaction result
        """
        if not address:
            raise ValueError("Address is required")
        if not token:
            raise ValueError("Token is required")
        if not amount or int(amount) <= 0:
            raise ValueError("Valid amount is required")

        message = {
            "@type": "/aura.privacy.v1beta1.MsgMixTokens",
            "address": address,
            "token": token,
            "amount": amount,
            "strategy": strategy.value if isinstance(strategy, MixingStrategy) else strategy,
            "rounds": rounds
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def generate_ring_signature(
        self,
        params: RingSignatureParams
    ) -> RingSignature:
        """Generate a ring signature.

        Args:
            params: Ring signature parameters

        Returns:
            Ring signature
        """
        if not params.message:
            raise ValueError("Message is required")
        if not params.ring_members or len(params.ring_members) < 2:
            raise ValueError("At least 2 ring members are required")
        if not params.key_image:
            raise ValueError("Key image is required")

        try:
            request_data = {
                "message": params.message,
                "ring_members": params.ring_members,
                "key_image": params.key_image,
                "ring_size": params.ring_size
            }

            data = await self.client.post("/aura/privacy/v1beta1/ring-signature/generate", request_data)
            sig_data = data.get("signature", {})

            return RingSignature(
                signature=sig_data.get("signature", ""),
                ring_members=sig_data.get("ring_members", params.ring_members),
                key_image=sig_data.get("key_image", params.key_image),
                ring_size=sig_data.get("ring_size", params.ring_size),
                created_at=datetime.fromisoformat(sig_data.get("created_at")) if sig_data.get("created_at") else datetime.now(),
                verified=sig_data.get("verified", False)
            )
        except Exception as e:
            raise RuntimeError(f"Failed to generate ring signature: {e}")

    async def verify_ring_signature(
        self,
        signature: str,
        message: str,
        ring_members: List[str],
        key_image: str
    ) -> bool:
        """Verify a ring signature.

        Args:
            signature: Signature to verify
            message: Original message
            ring_members: Ring members
            key_image: Key image

        Returns:
            True if signature is valid
        """
        if not signature:
            raise ValueError("Signature is required")
        if not message:
            raise ValueError("Message is required")
        if not ring_members:
            raise ValueError("Ring members are required")
        if not key_image:
            raise ValueError("Key image is required")

        try:
            request_data = {
                "signature": signature,
                "message": message,
                "ring_members": ring_members,
                "key_image": key_image
            }

            data = await self.client.post("/aura/privacy/v1beta1/ring-signature/verify", request_data)
            return data.get("valid", False)
        except Exception as e:
            raise RuntimeError(f"Failed to verify ring signature: {e}")

    async def create_stealth_address(
        self,
        owner: str,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Create a stealth address.

        Args:
            owner: Owner address
            options: Transaction options

        Returns:
            Transaction result
        """
        if not owner:
            raise ValueError("Owner address is required")

        message = {
            "@type": "/aura.privacy.v1beta1.MsgCreateStealthAddress",
            "owner": owner
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def get_stealth_address(self, address: str) -> Optional[StealthAddress]:
        """Get stealth address information.

        Args:
            address: Stealth address

        Returns:
            Stealth address info or None
        """
        if not address:
            raise ValueError("Address is required")

        try:
            data = await self.client.get(f"/aura/privacy/v1beta1/stealth/{address}")
            stealth_data = data.get("stealth")

            if not stealth_data:
                return None

            return StealthAddress(
                address=stealth_data.get("address", address),
                public_view_key=stealth_data.get("public_view_key", ""),
                public_spend_key=stealth_data.get("public_spend_key", ""),
                created_at=datetime.fromisoformat(stealth_data.get("created_at")) if stealth_data.get("created_at") else datetime.now(),
                usage_count=stealth_data.get("usage_count", 0)
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get stealth address: {e}")

    async def get_mixing_pools(self, token: Optional[str] = None) -> List[MixingPool]:
        """Get available mixing pools.

        Args:
            token: Optional token filter

        Returns:
            List of mixing pools
        """
        try:
            path = "/aura/privacy/v1beta1/mixing-pools"
            if token:
                path += f"?token={token}"

            data = await self.client.get(path)

            pools = []
            for pool_data in data.get("pools", []):
                pools.append(MixingPool(
                    pool_id=pool_data.get("pool_id", ""),
                    token=pool_data.get("token", ""),
                    denomination=pool_data.get("denomination", ""),
                    total_deposits=pool_data.get("total_deposits", 0),
                    available_liquidity=pool_data.get("available_liquidity", "0"),
                    mixing_fee=pool_data.get("mixing_fee", "0"),
                    anonymity_set_size=pool_data.get("anonymity_set_size", 0)
                ))

            return pools
        except Exception as e:
            raise RuntimeError(f"Failed to get mixing pools: {e}")

    async def get_confidential_balance(self, address: str) -> Optional[ConfidentialBalance]:
        """Get confidential balance for an address.

        Args:
            address: Address to query

        Returns:
            Confidential balance or None
        """
        if not address:
            raise ValueError("Address is required")

        try:
            data = await self.client.get(f"/aura/privacy/v1beta1/balance/{address}")
            balance_data = data.get("balance")

            if not balance_data:
                return None

            return ConfidentialBalance(
                address=balance_data.get("address", address),
                balance_commitment=balance_data.get("balance_commitment", ""),
                blinding_factor=balance_data.get("blinding_factor", ""),
                encrypted_amount=balance_data.get("encrypted_amount", ""),
                last_updated=datetime.fromisoformat(balance_data.get("last_updated")) if balance_data.get("last_updated") else datetime.now()
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get confidential balance: {e}")

    async def get_mixing_history(
        self,
        address: str,
        limit: int = 100
    ) -> List[MixingTransaction]:
        """Get mixing transaction history.

        Args:
            address: Address to query
            limit: Maximum number of results

        Returns:
            List of mixing transactions
        """
        if not address:
            raise ValueError("Address is required")

        try:
            data = await self.client.get(f"/aura/privacy/v1beta1/mixing-history/{address}?limit={limit}")

            history = []
            for mix_data in data.get("history", []):
                history.append(MixingTransaction(
                    mix_id=mix_data.get("mix_id", ""),
                    strategy=MixingStrategy(mix_data.get("strategy", "tornado")),
                    input_amount=mix_data.get("input_amount", "0"),
                    output_amount=mix_data.get("output_amount", "0"),
                    fee=mix_data.get("fee", "0"),
                    rounds=mix_data.get("rounds", 0),
                    anonymity_set=mix_data.get("anonymity_set", 0),
                    started_at=datetime.fromisoformat(mix_data.get("started_at")) if mix_data.get("started_at") else datetime.now(),
                    completed_at=datetime.fromisoformat(mix_data.get("completed_at")) if mix_data.get("completed_at") else None,
                    status=mix_data.get("status", "pending")
                ))

            return history
        except Exception as e:
            raise RuntimeError(f"Failed to get mixing history: {e}")

    async def update_privacy_settings(
        self,
        address: str,
        settings: PrivacySettings,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Update privacy settings.

        Args:
            address: Address
            settings: Privacy settings
            options: Transaction options

        Returns:
            Transaction result
        """
        if not address:
            raise ValueError("Address is required")

        message = {
            "@type": "/aura.privacy.v1beta1.MsgUpdatePrivacySettings",
            "address": address,
            "default_privacy_level": settings.default_privacy_level.value if isinstance(settings.default_privacy_level, PrivacyLevel) else settings.default_privacy_level,
            "auto_mix": settings.auto_mix,
            "min_anonymity_set": settings.min_anonymity_set,
            "max_mixing_rounds": settings.max_mixing_rounds,
            "stealth_addresses_enabled": settings.stealth_addresses_enabled,
            "ring_signature_size": settings.ring_signature_size
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def get_privacy_settings(self, address: str) -> Optional[PrivacySettings]:
        """Get privacy settings for an address.

        Args:
            address: Address to query

        Returns:
            Privacy settings or None
        """
        if not address:
            raise ValueError("Address is required")

        try:
            data = await self.client.get(f"/aura/privacy/v1beta1/settings/{address}")
            settings_data = data.get("settings")

            if not settings_data:
                return None

            return PrivacySettings(
                default_privacy_level=PrivacyLevel(settings_data.get("default_privacy_level", "basic")),
                auto_mix=settings_data.get("auto_mix", False),
                min_anonymity_set=settings_data.get("min_anonymity_set", 10),
                max_mixing_rounds=settings_data.get("max_mixing_rounds", 3),
                stealth_addresses_enabled=settings_data.get("stealth_addresses_enabled", False),
                ring_signature_size=settings_data.get("ring_signature_size", 11)
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get privacy settings: {e}")
