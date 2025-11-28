"""Module for identity change operations."""

from typing import List, Optional, Dict, Any
from datetime import datetime
from ..types import (
    IdentityChangeParams,
    IdentityChangeRequest,
    IdentityChangeStatus,
    IdentityChangeType,
    IdentityVerificationParams,
    IdentityProfile,
    IdentityAttestation,
    RecoveryParams,
    TxResult,
    GasOptions
)


class IdentityChangeModule:
    """Identity change module for managing identity updates and recovery."""

    def __init__(self, client):
        """Initialize identity change module."""
        self.client = client

    async def create_identity(
        self,
        address: str,
        metadata: Optional[Dict[str, Any]] = None,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Create a new identity profile.

        Args:
            address: Address for identity
            metadata: Optional metadata
            options: Transaction options

        Returns:
            Transaction result
        """
        if not address:
            raise ValueError("Address is required")

        message = {
            "@type": "/aura.identitychange.v1beta1.MsgCreateIdentity",
            "address": address,
            "metadata": metadata or {}
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def update_identity(
        self,
        address: str,
        metadata: Dict[str, Any],
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Update identity metadata.

        Args:
            address: Address to update
            metadata: New metadata
            options: Transaction options

        Returns:
            Transaction result
        """
        if not address:
            raise ValueError("Address is required")
        if not metadata:
            raise ValueError("Metadata is required")

        message = {
            "@type": "/aura.identitychange.v1beta1.MsgUpdateIdentity",
            "address": address,
            "metadata": metadata
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def get_identity(self, address: str) -> Optional[IdentityProfile]:
        """Get identity profile for an address.

        Args:
            address: Address to query

        Returns:
            Identity profile or None
        """
        if not address:
            raise ValueError("Address is required")

        try:
            data = await self.client.get(f"/aura/identitychange/v1beta1/identity/{address}")
            identity_data = data.get("identity")

            if not identity_data:
                return None

            return IdentityProfile(
                address=identity_data.get("address", address),
                verified=identity_data.get("verified", False),
                verification_level=identity_data.get("verification_level", 0),
                created_at=datetime.fromisoformat(identity_data.get("created_at")) if identity_data.get("created_at") else datetime.now(),
                updated_at=datetime.fromisoformat(identity_data.get("updated_at")) if identity_data.get("updated_at") else datetime.now(),
                attestations=identity_data.get("attestations", []),
                metadata=identity_data.get("metadata")
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get identity: {e}")

    async def verify_identity(
        self,
        params: IdentityVerificationParams,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Verify an identity.

        Args:
            params: Verification parameters
            options: Transaction options

        Returns:
            Transaction result
        """
        if not params.address:
            raise ValueError("Address is required")
        if not params.verification_type:
            raise ValueError("Verification type is required")
        if not params.proof_data:
            raise ValueError("Proof data is required")

        message = {
            "@type": "/aura.identitychange.v1beta1.MsgVerifyIdentity",
            "address": params.address,
            "verification_type": params.verification_type,
            "proof_data": params.proof_data,
            "attestations": params.attestations or []
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def revoke_identity(
        self,
        address: str,
        reason: str,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Revoke an identity.

        Args:
            address: Address to revoke
            reason: Revocation reason
            options: Transaction options

        Returns:
            Transaction result
        """
        if not address:
            raise ValueError("Address is required")
        if not reason:
            raise ValueError("Reason is required")

        message = {
            "@type": "/aura.identitychange.v1beta1.MsgRevokeIdentity",
            "address": address,
            "reason": reason
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def get_identity_history(
        self,
        address: str,
        limit: int = 100
    ) -> List[Dict[str, Any]]:
        """Get identity change history.

        Args:
            address: Address to query
            limit: Maximum number of results

        Returns:
            List of identity changes
        """
        if not address:
            raise ValueError("Address is required")

        try:
            data = await self.client.get(f"/aura/identitychange/v1beta1/history/{address}?limit={limit}")
            return data.get("history", [])
        except Exception as e:
            raise RuntimeError(f"Failed to get identity history: {e}")

    async def request_identity_change(
        self,
        params: IdentityChangeParams,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Request an identity change.

        Args:
            params: Change request parameters
            options: Transaction options

        Returns:
            Transaction result
        """
        if not params.old_address:
            raise ValueError("Old address is required")
        if not params.new_address:
            raise ValueError("New address is required")
        if not params.proof:
            raise ValueError("Proof is required")
        if not params.reason:
            raise ValueError("Reason is required")

        message = {
            "@type": "/aura.identitychange.v1beta1.MsgRequestChange",
            "old_address": params.old_address,
            "new_address": params.new_address,
            "change_type": params.change_type.value if isinstance(params.change_type, IdentityChangeType) else params.change_type,
            "proof": params.proof,
            "reason": params.reason,
            "metadata": params.metadata or {}
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def get_change_request(self, request_id: str) -> Optional[IdentityChangeRequest]:
        """Get an identity change request.

        Args:
            request_id: Request ID

        Returns:
            Change request or None
        """
        if not request_id:
            raise ValueError("Request ID is required")

        try:
            data = await self.client.get(f"/aura/identitychange/v1beta1/requests/{request_id}")
            request_data = data.get("request")

            if not request_data:
                return None

            return IdentityChangeRequest(
                request_id=request_data.get("request_id", request_id),
                old_address=request_data.get("old_address", ""),
                new_address=request_data.get("new_address", ""),
                change_type=IdentityChangeType(request_data.get("change_type", "update")),
                status=IdentityChangeStatus(request_data.get("status", "pending")),
                proof=request_data.get("proof", ""),
                reason=request_data.get("reason", ""),
                submitted_at=datetime.fromisoformat(request_data.get("submitted_at")) if request_data.get("submitted_at") else datetime.now(),
                processed_at=datetime.fromisoformat(request_data.get("processed_at")) if request_data.get("processed_at") else None,
                approved_by=request_data.get("approved_by"),
                rejection_reason=request_data.get("rejection_reason")
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get change request: {e}")

    async def approve_change_request(
        self,
        request_id: str,
        approver: str,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Approve an identity change request.

        Args:
            request_id: Request ID
            approver: Approver address
            options: Transaction options

        Returns:
            Transaction result
        """
        if not request_id:
            raise ValueError("Request ID is required")
        if not approver:
            raise ValueError("Approver address is required")

        message = {
            "@type": "/aura.identitychange.v1beta1.MsgApproveChange",
            "request_id": request_id,
            "approver": approver
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def add_attestation(
        self,
        subject: str,
        claim_type: str,
        claim_data: Dict[str, Any],
        issuer: str,
        expires_at: Optional[datetime] = None,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Add an identity attestation.

        Args:
            subject: Subject address
            claim_type: Type of claim
            claim_data: Claim data
            issuer: Issuer address
            expires_at: Optional expiration
            options: Transaction options

        Returns:
            Transaction result
        """
        if not subject:
            raise ValueError("Subject address is required")
        if not claim_type:
            raise ValueError("Claim type is required")
        if not claim_data:
            raise ValueError("Claim data is required")
        if not issuer:
            raise ValueError("Issuer address is required")

        message = {
            "@type": "/aura.identitychange.v1beta1.MsgAddAttestation",
            "subject": subject,
            "issuer": issuer,
            "claim_type": claim_type,
            "claim_data": claim_data,
            "expires_at": expires_at.isoformat() if expires_at else None
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def initiate_recovery(
        self,
        params: RecoveryParams,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Initiate account recovery.

        Args:
            params: Recovery parameters
            options: Transaction options

        Returns:
            Transaction result
        """
        if not params.lost_address:
            raise ValueError("Lost address is required")
        if not params.recovery_address:
            raise ValueError("Recovery address is required")
        if not params.recovery_key:
            raise ValueError("Recovery key is required")
        if not params.guardians or len(params.guardians) == 0:
            raise ValueError("At least one guardian is required")

        message = {
            "@type": "/aura.identitychange.v1beta1.MsgInitiateRecovery",
            "lost_address": params.lost_address,
            "recovery_address": params.recovery_address,
            "recovery_key": params.recovery_key,
            "guardians": params.guardians,
            "signatures": params.signatures
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)
