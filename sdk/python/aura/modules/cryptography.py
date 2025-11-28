"""Module for cryptography operations."""

from typing import Optional, Dict, Any
from datetime import datetime
from ..types import (
    KeyRotationParams,
    EncryptionParams,
    DecryptionParams,
    KeyPair,
    EncryptedData,
    SignatureParams,
    VerificationParams,
    RandomParams,
    SecureEnclaveParams,
    KeyType,
    EncryptionAlgorithm,
    TxResult,
    GasOptions
)


class CryptographyModule:
    """Cryptography module for encryption and key management."""

    def __init__(self, client):
        """Initialize cryptography module."""
        self.client = client

    async def generate_keypair(
        self,
        key_type: KeyType = KeyType.ED25519,
        expires_in_days: Optional[int] = None
    ) -> KeyPair:
        """Generate a new cryptographic key pair.

        Args:
            key_type: Type of key to generate
            expires_in_days: Optional expiration in days

        Returns:
            Generated key pair
        """
        try:
            params = {
                "key_type": key_type.value if isinstance(key_type, KeyType) else key_type
            }
            if expires_in_days:
                params["expires_in_days"] = expires_in_days

            query_str = "&".join([f"{k}={v}" for k, v in params.items()])
            data = await self.client.get(f"/aura/cryptography/v1beta1/generate/keypair?{query_str}")

            keypair_data = data.get("keypair", {})

            return KeyPair(
                public_key=keypair_data.get("public_key", ""),
                private_key=keypair_data.get("private_key", ""),
                key_type=KeyType(keypair_data.get("key_type", "ed25519")),
                created_at=datetime.fromisoformat(keypair_data.get("created_at")) if keypair_data.get("created_at") else datetime.now(),
                expires_at=datetime.fromisoformat(keypair_data.get("expires_at")) if keypair_data.get("expires_at") else None
            )
        except Exception as e:
            raise RuntimeError(f"Failed to generate keypair: {e}")

    async def encrypt_data(self, params: EncryptionParams) -> EncryptedData:
        """Encrypt data.

        Args:
            params: Encryption parameters

        Returns:
            Encrypted data
        """
        if not params.data:
            raise ValueError("Data is required")

        try:
            request_data = {
                "data": params.data,
                "algorithm": params.algorithm.value if isinstance(params.algorithm, EncryptionAlgorithm) else params.algorithm,
                "public_key": params.public_key,
                "metadata": params.metadata or {}
            }

            data = await self.client.post("/aura/cryptography/v1beta1/encrypt", request_data)

            encrypted = data.get("encrypted", {})

            return EncryptedData(
                ciphertext=encrypted.get("ciphertext", ""),
                algorithm=EncryptionAlgorithm(encrypted.get("algorithm", "aes256")),
                nonce=encrypted.get("nonce"),
                tag=encrypted.get("tag"),
                metadata=encrypted.get("metadata")
            )
        except Exception as e:
            raise RuntimeError(f"Failed to encrypt data: {e}")

    async def decrypt_data(self, params: DecryptionParams) -> str:
        """Decrypt data.

        Args:
            params: Decryption parameters

        Returns:
            Decrypted data
        """
        if not params.encrypted_data:
            raise ValueError("Encrypted data is required")
        if not params.private_key:
            raise ValueError("Private key is required")

        try:
            request_data = {
                "encrypted_data": params.encrypted_data,
                "algorithm": params.algorithm.value if isinstance(params.algorithm, EncryptionAlgorithm) else params.algorithm,
                "private_key": params.private_key
            }

            data = await self.client.post("/aura/cryptography/v1beta1/decrypt", request_data)
            return data.get("decrypted_data", "")
        except Exception as e:
            raise RuntimeError(f"Failed to decrypt data: {e}")

    async def rotate_keys(
        self,
        params: KeyRotationParams,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Rotate cryptographic keys.

        Args:
            params: Key rotation parameters
            options: Transaction options

        Returns:
            Transaction result
        """
        if not params.address:
            raise ValueError("Address is required")
        if not params.new_public_key:
            raise ValueError("New public key is required")
        if not params.proof:
            raise ValueError("Proof is required")

        message = {
            "@type": "/aura.cryptography.v1beta1.MsgRotateKeys",
            "address": params.address,
            "new_public_key": params.new_public_key,
            "key_type": params.key_type.value if isinstance(params.key_type, KeyType) else params.key_type,
            "proof": params.proof
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def get_public_key(self, address: str) -> Optional[str]:
        """Get public key for an address.

        Args:
            address: Address to query

        Returns:
            Public key or None
        """
        if not address:
            raise ValueError("Address is required")

        try:
            data = await self.client.get(f"/aura/cryptography/v1beta1/keys/{address}")
            return data.get("public_key")
        except Exception:
            return None

    async def sign_data(self, params: SignatureParams) -> str:
        """Sign data with private key.

        Args:
            params: Signature parameters

        Returns:
            Signature
        """
        if not params.message:
            raise ValueError("Message is required")
        if not params.private_key:
            raise ValueError("Private key is required")

        try:
            request_data = {
                "message": params.message,
                "private_key": params.private_key,
                "algorithm": params.algorithm
            }

            data = await self.client.post("/aura/cryptography/v1beta1/sign", request_data)
            return data.get("signature", "")
        except Exception as e:
            raise RuntimeError(f"Failed to sign data: {e}")

    async def verify_signature(self, params: VerificationParams) -> bool:
        """Verify a digital signature.

        Args:
            params: Verification parameters

        Returns:
            True if signature is valid
        """
        if not params.message:
            raise ValueError("Message is required")
        if not params.signature:
            raise ValueError("Signature is required")
        if not params.public_key:
            raise ValueError("Public key is required")

        try:
            request_data = {
                "message": params.message,
                "signature": params.signature,
                "public_key": params.public_key,
                "algorithm": params.algorithm
            }

            data = await self.client.post("/aura/cryptography/v1beta1/verify", request_data)
            return data.get("valid", False)
        except Exception as e:
            raise RuntimeError(f"Failed to verify signature: {e}")

    async def generate_random(self, params: RandomParams) -> str:
        """Generate cryptographically secure random data.

        Args:
            params: Random generation parameters

        Returns:
            Random data (hex encoded)
        """
        if not params.length or params.length <= 0:
            raise ValueError("Valid length is required")

        try:
            request_data = {
                "length": params.length,
                "algorithm": params.algorithm,
                "seed": params.seed
            }

            data = await self.client.post("/aura/cryptography/v1beta1/random", request_data)
            return data.get("random_data", "")
        except Exception as e:
            raise RuntimeError(f"Failed to generate random data: {e}")

    async def secure_enclave_operation(
        self,
        params: SecureEnclaveParams,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Perform secure enclave operation.

        Args:
            params: Secure enclave parameters
            options: Transaction options

        Returns:
            Transaction result
        """
        if not params.operation:
            raise ValueError("Operation is required")
        if not params.data:
            raise ValueError("Data is required")
        if not params.enclave_id:
            raise ValueError("Enclave ID is required")

        message = {
            "@type": "/aura.cryptography.v1beta1.MsgSecureEnclaveOp",
            "operation": params.operation,
            "data": params.data,
            "enclave_id": params.enclave_id,
            "attestation": params.attestation or ""
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def get_enclave_status(self, enclave_id: str) -> Dict[str, Any]:
        """Get secure enclave status.

        Args:
            enclave_id: Enclave ID

        Returns:
            Enclave status
        """
        if not enclave_id:
            raise ValueError("Enclave ID is required")

        try:
            data = await self.client.get(f"/aura/cryptography/v1beta1/enclave/{enclave_id}")
            return data.get("enclave", {})
        except Exception as e:
            raise RuntimeError(f"Failed to get enclave status: {e}")
