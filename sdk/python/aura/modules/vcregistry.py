"""Module for Verifiable Credentials Registry operations."""

from typing import List, Optional, Dict, Any
from datetime import datetime
from ..types import (
    VCParams,
    VerifiableCredential,
    VCPresentation,
    PresentationParams,
    VCRevocationParams,
    VCStatus,
    VCType,
    VCQuery,
    VCVerificationResult,
    RegistryStats,
    IssuerInfo,
    TxResult,
    GasOptions,
    PresentationStatus
)


class VCRegistryModule:
    """Verifiable Credentials Registry module."""

    def __init__(self, client):
        """Initialize VC Registry module."""
        self.client = client

    async def mint_vc(
        self,
        params: VCParams,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Mint a new verifiable credential.

        Args:
            params: Credential parameters
            options: Transaction options

        Returns:
            Transaction result
        """
        if not params.issuer:
            raise ValueError("Issuer address is required")
        if not params.subject:
            raise ValueError("Subject address is required")
        if not params.claims:
            raise ValueError("Claims are required")

        message = {
            "@type": "/aura.vcregistry.v1beta1.MsgMintVC",
            "issuer": params.issuer,
            "subject": params.subject,
            "vc_type": params.vc_type.value if isinstance(params.vc_type, VCType) else params.vc_type,
            "claims": params.claims,
            "expiry_date": params.expiry_date.isoformat() if params.expiry_date else None,
            "metadata": params.metadata or {}
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def revoke_vc(
        self,
        params: VCRevocationParams,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Revoke a verifiable credential.

        Args:
            params: Revocation parameters
            options: Transaction options

        Returns:
            Transaction result
        """
        if not params.vc_id:
            raise ValueError("VC ID is required")
        if not params.issuer:
            raise ValueError("Issuer address is required")
        if not params.reason:
            raise ValueError("Revocation reason is required")

        message = {
            "@type": "/aura.vcregistry.v1beta1.MsgRevokeVC",
            "vc_id": params.vc_id,
            "issuer": params.issuer,
            "reason": params.reason,
            "proof": params.proof
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def verify_vc(self, vc_id: str) -> VCVerificationResult:
        """Verify a verifiable credential.

        Args:
            vc_id: Credential ID

        Returns:
            Verification result
        """
        if not vc_id:
            raise ValueError("VC ID is required")

        try:
            data = await self.client.get(f"/aura/vcregistry/v1beta1/verify/{vc_id}")

            return VCVerificationResult(
                vc_id=data.get("vc_id", vc_id),
                valid=data.get("valid", False),
                status=VCStatus(data.get("status", "active")),
                issuer_verified=data.get("issuer_verified", False),
                signature_valid=data.get("signature_valid", False),
                not_expired=data.get("not_expired", True),
                not_revoked=data.get("not_revoked", True),
                verified_at=datetime.fromisoformat(data.get("verified_at")) if data.get("verified_at") else datetime.now(),
                errors=data.get("errors", []),
                warnings=data.get("warnings", [])
            )
        except Exception as e:
            raise RuntimeError(f"Failed to verify VC: {e}")

    async def get_vc(self, vc_id: str) -> Optional[VerifiableCredential]:
        """Get a verifiable credential by ID.

        Args:
            vc_id: Credential ID

        Returns:
            Verifiable credential or None
        """
        if not vc_id:
            raise ValueError("VC ID is required")

        try:
            data = await self.client.get(f"/aura/vcregistry/v1beta1/credentials/{vc_id}")
            vc_data = data.get("credential")

            if not vc_data:
                return None

            return VerifiableCredential(
                vc_id=vc_data.get("vc_id", vc_id),
                issuer=vc_data.get("issuer", ""),
                subject=vc_data.get("subject", ""),
                vc_type=VCType(vc_data.get("vc_type", "custom")),
                status=VCStatus(vc_data.get("status", "active")),
                claims=vc_data.get("claims", {}),
                proof=vc_data.get("proof", ""),
                issued_at=datetime.fromisoformat(vc_data.get("issued_at")) if vc_data.get("issued_at") else datetime.now(),
                expires_at=datetime.fromisoformat(vc_data.get("expires_at")) if vc_data.get("expires_at") else None,
                revoked_at=datetime.fromisoformat(vc_data.get("revoked_at")) if vc_data.get("revoked_at") else None,
                metadata=vc_data.get("metadata"),
                schema_id=vc_data.get("schema_id")
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get VC: {e}")

    async def list_vcs(
        self,
        query: Optional[VCQuery] = None,
        limit: int = 100
    ) -> List[VerifiableCredential]:
        """List verifiable credentials with optional filtering.

        Args:
            query: Query parameters
            limit: Maximum number of results

        Returns:
            List of verifiable credentials
        """
        try:
            params = {"limit": limit}

            if query:
                if query.issuer:
                    params["issuer"] = query.issuer
                if query.subject:
                    params["subject"] = query.subject
                if query.vc_type:
                    params["vc_type"] = query.vc_type.value if isinstance(query.vc_type, VCType) else query.vc_type
                if query.status:
                    params["status"] = query.status.value if isinstance(query.status, VCStatus) else query.status
                if query.issued_after:
                    params["issued_after"] = query.issued_after.isoformat()
                if query.issued_before:
                    params["issued_before"] = query.issued_before.isoformat()
                params["include_revoked"] = query.include_revoked

            # Build query string
            query_str = "&".join([f"{k}={v}" for k, v in params.items()])
            data = await self.client.get(f"/aura/vcregistry/v1beta1/credentials?{query_str}")

            credentials = []
            for vc_data in data.get("credentials", []):
                credentials.append(VerifiableCredential(
                    vc_id=vc_data.get("vc_id", ""),
                    issuer=vc_data.get("issuer", ""),
                    subject=vc_data.get("subject", ""),
                    vc_type=VCType(vc_data.get("vc_type", "custom")),
                    status=VCStatus(vc_data.get("status", "active")),
                    claims=vc_data.get("claims", {}),
                    proof=vc_data.get("proof", ""),
                    issued_at=datetime.fromisoformat(vc_data.get("issued_at")) if vc_data.get("issued_at") else datetime.now(),
                    expires_at=datetime.fromisoformat(vc_data.get("expires_at")) if vc_data.get("expires_at") else None,
                    revoked_at=datetime.fromisoformat(vc_data.get("revoked_at")) if vc_data.get("revoked_at") else None,
                    metadata=vc_data.get("metadata"),
                    schema_id=vc_data.get("schema_id")
                ))

            return credentials
        except Exception as e:
            raise RuntimeError(f"Failed to list VCs: {e}")

    async def submit_presentation(
        self,
        params: PresentationParams,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Submit a verifiable presentation.

        Args:
            params: Presentation parameters
            options: Transaction options

        Returns:
            Transaction result
        """
        if not params.holder:
            raise ValueError("Holder address is required")
        if not params.credential_ids or len(params.credential_ids) == 0:
            raise ValueError("At least one credential ID is required")

        message = {
            "@type": "/aura.vcregistry.v1beta1.MsgSubmitPresentation",
            "holder": params.holder,
            "credential_ids": params.credential_ids,
            "challenge": params.challenge,
            "domain": params.domain,
            "purpose": params.purpose
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def verify_presentation(self, presentation_id: str) -> Dict[str, Any]:
        """Verify a verifiable presentation.

        Args:
            presentation_id: Presentation ID

        Returns:
            Verification result
        """
        if not presentation_id:
            raise ValueError("Presentation ID is required")

        try:
            data = await self.client.get(f"/aura/vcregistry/v1beta1/presentations/{presentation_id}/verify")
            return {
                "presentation_id": presentation_id,
                "valid": data.get("valid", False),
                "status": data.get("status", "invalid"),
                "holder_verified": data.get("holder_verified", False),
                "credentials_valid": data.get("credentials_valid", []),
                "credentials_invalid": data.get("credentials_invalid", []),
                "verified_at": data.get("verified_at", datetime.now().isoformat()),
                "errors": data.get("errors", [])
            }
        except Exception as e:
            raise RuntimeError(f"Failed to verify presentation: {e}")

    async def get_issuer_info(self, issuer_address: str) -> Optional[IssuerInfo]:
        """Get issuer information.

        Args:
            issuer_address: Issuer address

        Returns:
            Issuer information or None
        """
        if not issuer_address:
            raise ValueError("Issuer address is required")

        try:
            data = await self.client.get(f"/aura/vcregistry/v1beta1/issuers/{issuer_address}")
            issuer_data = data.get("issuer")

            if not issuer_data:
                return None

            return IssuerInfo(
                issuer_id=issuer_data.get("issuer_id", issuer_address),
                name=issuer_data.get("name", ""),
                did=issuer_data.get("did", ""),
                public_key=issuer_data.get("public_key", ""),
                verified=issuer_data.get("verified", False),
                total_issued=issuer_data.get("total_issued", 0),
                total_revoked=issuer_data.get("total_revoked", 0),
                reputation_score=issuer_data.get("reputation_score", 0.0),
                registered_at=datetime.fromisoformat(issuer_data.get("registered_at")) if issuer_data.get("registered_at") else datetime.now()
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get issuer info: {e}")

    async def get_registry_stats(self) -> RegistryStats:
        """Get registry statistics.

        Returns:
            Registry statistics
        """
        try:
            data = await self.client.get("/aura/vcregistry/v1beta1/stats")

            return RegistryStats(
                total_credentials=data.get("total_credentials", 0),
                active_credentials=data.get("active_credentials", 0),
                revoked_credentials=data.get("revoked_credentials", 0),
                expired_credentials=data.get("expired_credentials", 0),
                total_issuers=data.get("total_issuers", 0),
                total_holders=data.get("total_holders", 0),
                total_presentations=data.get("total_presentations", 0),
                credentials_by_type=data.get("credentials_by_type", {})
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get registry stats: {e}")

    async def get_credentials_by_subject(self, subject: str) -> List[VerifiableCredential]:
        """Get all credentials for a subject.

        Args:
            subject: Subject address

        Returns:
            List of credentials
        """
        query = VCQuery(subject=subject, include_revoked=False)
        return await self.list_vcs(query=query)

    async def get_credentials_by_issuer(self, issuer: str) -> List[VerifiableCredential]:
        """Get all credentials issued by an issuer.

        Args:
            issuer: Issuer address

        Returns:
            List of credentials
        """
        query = VCQuery(issuer=issuer, include_revoked=True)
        return await self.list_vcs(query=query)
