"""Module for cross-chain bridge operations."""

from typing import List, Optional, Dict, Any
from datetime import datetime
from ..types import (
    BridgeTransferParams,
    BridgeTransfer,
    BridgeParams,
    BridgeStatus,
    BridgeBalance,
    BridgeProof,
    RelayerInfo,
    TxResult,
    GasOptions
)


class BridgeModule:
    """Cross-chain bridge module."""

    def __init__(self, client):
        """Initialize bridge module."""
        self.client = client

    async def lock_tokens(
        self,
        params: BridgeTransferParams,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Lock tokens for cross-chain transfer.

        Args:
            params: Transfer parameters
            options: Transaction options

        Returns:
            Transaction result
        """
        if not params.source_chain:
            raise ValueError("Source chain is required")
        if not params.destination_chain:
            raise ValueError("Destination chain is required")
        if not params.token:
            raise ValueError("Token is required")
        if not params.amount or int(params.amount) <= 0:
            raise ValueError("Valid amount is required")
        if not params.recipient:
            raise ValueError("Recipient address is required")

        message = {
            "@type": "/aura.bridge.v1beta1.MsgLockTokens",
            "source_chain": params.source_chain,
            "destination_chain": params.destination_chain,
            "token": params.token,
            "amount": params.amount,
            "recipient": params.recipient,
            "memo": params.memo or "",
            "timeout_height": params.timeout_height,
            "timeout_timestamp": params.timeout_timestamp
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def mint_tokens(
        self,
        transfer_id: str,
        proof: str,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Mint tokens on destination chain.

        Args:
            transfer_id: Transfer ID
            proof: Merkle proof
            options: Transaction options

        Returns:
            Transaction result
        """
        if not transfer_id:
            raise ValueError("Transfer ID is required")
        if not proof:
            raise ValueError("Proof is required")

        message = {
            "@type": "/aura.bridge.v1beta1.MsgMintTokens",
            "transfer_id": transfer_id,
            "proof": proof
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def unlock_tokens(
        self,
        transfer_id: str,
        proof: str,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Unlock tokens after successful transfer.

        Args:
            transfer_id: Transfer ID
            proof: Burn proof from destination chain
            options: Transaction options

        Returns:
            Transaction result
        """
        if not transfer_id:
            raise ValueError("Transfer ID is required")
        if not proof:
            raise ValueError("Proof is required")

        message = {
            "@type": "/aura.bridge.v1beta1.MsgUnlockTokens",
            "transfer_id": transfer_id,
            "proof": proof
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def burn_tokens(
        self,
        transfer_id: str,
        amount: str,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Burn tokens to return to source chain.

        Args:
            transfer_id: Transfer ID
            amount: Amount to burn
            options: Transaction options

        Returns:
            Transaction result
        """
        if not transfer_id:
            raise ValueError("Transfer ID is required")
        if not amount or int(amount) <= 0:
            raise ValueError("Valid amount is required")

        message = {
            "@type": "/aura.bridge.v1beta1.MsgBurnTokens",
            "transfer_id": transfer_id,
            "amount": amount
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def link_address(
        self,
        local_address: str,
        remote_chain: str,
        remote_address: str,
        proof: str,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Link addresses across chains.

        Args:
            local_address: Local chain address
            remote_chain: Remote chain ID
            remote_address: Remote chain address
            proof: Ownership proof
            options: Transaction options

        Returns:
            Transaction result
        """
        if not local_address:
            raise ValueError("Local address is required")
        if not remote_chain:
            raise ValueError("Remote chain is required")
        if not remote_address:
            raise ValueError("Remote address is required")
        if not proof:
            raise ValueError("Proof is required")

        message = {
            "@type": "/aura.bridge.v1beta1.MsgLinkAddress",
            "local_address": local_address,
            "remote_chain": remote_chain,
            "remote_address": remote_address,
            "proof": proof
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def cross_chain_swap(
        self,
        source_token: str,
        source_amount: str,
        destination_chain: str,
        destination_token: str,
        recipient: str,
        min_amount_out: str,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Perform cross-chain token swap.

        Args:
            source_token: Source token denom
            source_amount: Amount to swap
            destination_chain: Destination chain ID
            destination_token: Destination token denom
            recipient: Recipient address
            min_amount_out: Minimum amount to receive
            options: Transaction options

        Returns:
            Transaction result
        """
        if not source_token:
            raise ValueError("Source token is required")
        if not source_amount or int(source_amount) <= 0:
            raise ValueError("Valid source amount is required")
        if not destination_chain:
            raise ValueError("Destination chain is required")
        if not destination_token:
            raise ValueError("Destination token is required")
        if not recipient:
            raise ValueError("Recipient is required")

        message = {
            "@type": "/aura.bridge.v1beta1.MsgCrossChainSwap",
            "source_token": source_token,
            "source_amount": source_amount,
            "destination_chain": destination_chain,
            "destination_token": destination_token,
            "recipient": recipient,
            "min_amount_out": min_amount_out
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def relay_transfer(
        self,
        transfer_id: str,
        relayer: str,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Relay a cross-chain transfer.

        Args:
            transfer_id: Transfer ID
            relayer: Relayer address
            options: Transaction options

        Returns:
            Transaction result
        """
        if not transfer_id:
            raise ValueError("Transfer ID is required")
        if not relayer:
            raise ValueError("Relayer address is required")

        message = {
            "@type": "/aura.bridge.v1beta1.MsgRelayTransfer",
            "transfer_id": transfer_id,
            "relayer": relayer
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def get_bridge_transfer(self, transfer_id: str) -> Optional[BridgeTransfer]:
        """Get bridge transfer by ID.

        Args:
            transfer_id: Transfer ID

        Returns:
            Bridge transfer information or None
        """
        if not transfer_id:
            raise ValueError("Transfer ID is required")

        try:
            data = await self.client.get(f"/aura/bridge/v1beta1/transfers/{transfer_id}")
            transfer_data = data.get("transfer")

            if not transfer_data:
                return None

            return BridgeTransfer(
                transfer_id=transfer_data.get("transfer_id", transfer_id),
                source_chain=transfer_data.get("source_chain", ""),
                destination_chain=transfer_data.get("destination_chain", ""),
                sender=transfer_data.get("sender", ""),
                recipient=transfer_data.get("recipient", ""),
                token=transfer_data.get("token", ""),
                amount=transfer_data.get("amount", "0"),
                status=BridgeStatus(transfer_data.get("status", "pending")),
                created_at=datetime.fromisoformat(transfer_data.get("created_at")) if transfer_data.get("created_at") else datetime.now(),
                updated_at=datetime.fromisoformat(transfer_data.get("updated_at")) if transfer_data.get("updated_at") else datetime.now(),
                transaction_hash=transfer_data.get("transaction_hash"),
                destination_hash=transfer_data.get("destination_hash"),
                proof=transfer_data.get("proof"),
                error_message=transfer_data.get("error_message")
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get bridge transfer: {e}")

    async def get_bridge_params(self) -> BridgeParams:
        """Get bridge module parameters.

        Returns:
            Bridge parameters
        """
        try:
            data = await self.client.get("/aura/bridge/v1beta1/params")
            params = data.get("params", {})

            return BridgeParams(
                enabled=params.get("enabled", True),
                supported_chains=params.get("supported_chains", []),
                min_transfer_amount=params.get("min_transfer_amount", "0"),
                max_transfer_amount=params.get("max_transfer_amount", "0"),
                transfer_fee=params.get("transfer_fee", "0"),
                timeout_duration=params.get("timeout_duration", 3600),
                confirmation_blocks=params.get("confirmation_blocks", 12),
                relayer_addresses=params.get("relayer_addresses", [])
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get bridge params: {e}")

    async def get_bridge_stats(self) -> Dict[str, Any]:
        """Get bridge statistics.

        Returns:
            Bridge statistics
        """
        try:
            data = await self.client.get("/aura/bridge/v1beta1/stats")
            return {
                "total_transfers": data.get("total_transfers", 0),
                "pending_transfers": data.get("pending_transfers", 0),
                "completed_transfers": data.get("completed_transfers", 0),
                "failed_transfers": data.get("failed_transfers", 0),
                "total_volume": data.get("total_volume", "0"),
                "total_fees": data.get("total_fees", "0"),
                "transfers_by_chain": data.get("transfers_by_chain", {}),
                "active_relayers": data.get("active_relayers", 0)
            }
        except Exception as e:
            raise RuntimeError(f"Failed to get bridge stats: {e}")

    async def get_bridge_balance(self, chain: str, token: str) -> Optional[BridgeBalance]:
        """Get bridge balance for a specific chain and token.

        Args:
            chain: Chain ID
            token: Token denom

        Returns:
            Bridge balance or None
        """
        if not chain:
            raise ValueError("Chain is required")
        if not token:
            raise ValueError("Token is required")

        try:
            data = await self.client.get(f"/aura/bridge/v1beta1/balances/{chain}/{token}")
            balance_data = data.get("balance")

            if not balance_data:
                return None

            return BridgeBalance(
                chain=balance_data.get("chain", chain),
                token=balance_data.get("token", token),
                locked_amount=balance_data.get("locked_amount", "0"),
                available_amount=balance_data.get("available_amount", "0"),
                total_supply=balance_data.get("total_supply", "0")
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get bridge balance: {e}")

    async def get_linked_address(self, address: str, chain: str) -> Optional[str]:
        """Get linked address for a given address on another chain.

        Args:
            address: Address to query
            chain: Target chain

        Returns:
            Linked address or None
        """
        if not address:
            raise ValueError("Address is required")
        if not chain:
            raise ValueError("Chain is required")

        try:
            data = await self.client.get(f"/aura/bridge/v1beta1/linked/{address}/{chain}")
            return data.get("linked_address")
        except Exception:
            return None

    async def get_relayer_info(self, relayer_address: str) -> Optional[RelayerInfo]:
        """Get relayer information.

        Args:
            relayer_address: Relayer address

        Returns:
            Relayer information or None
        """
        if not relayer_address:
            raise ValueError("Relayer address is required")

        try:
            data = await self.client.get(f"/aura/bridge/v1beta1/relayers/{relayer_address}")
            relayer_data = data.get("relayer")

            if not relayer_data:
                return None

            return RelayerInfo(
                address=relayer_data.get("address", relayer_address),
                chains=relayer_data.get("chains", []),
                active=relayer_data.get("active", False),
                total_relayed=relayer_data.get("total_relayed", 0),
                success_rate=relayer_data.get("success_rate", 0.0),
                last_relay_time=datetime.fromisoformat(relayer_data.get("last_relay_time")) if relayer_data.get("last_relay_time") else None
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get relayer info: {e}")
