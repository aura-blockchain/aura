"""Type definitions for Cryptography module."""

from dataclasses import dataclass
from typing import Optional, List, Dict
from datetime import datetime
from enum import Enum


class KeyType(Enum):
    """Key types."""

    ED25519 = "ed25519"
    SECP256K1 = "secp256k1"
    RSA = "rsa"
    QUANTUM_RESISTANT = "quantum_resistant"


class EncryptionAlgorithm(Enum):
    """Encryption algorithms."""

    AES256 = "aes256"
    CHACHA20 = "chacha20"
    QUANTUM = "quantum"


@dataclass
class KeyRotationParams:
    """Parameters for key rotation."""

    address: str
    new_public_key: str
    key_type: KeyType
    proof: str


@dataclass
class EncryptionParams:
    """Parameters for encryption."""

    data: str
    algorithm: EncryptionAlgorithm
    public_key: Optional[str] = None
    metadata: Optional[Dict[str, str]] = None


@dataclass
class DecryptionParams:
    """Parameters for decryption."""

    encrypted_data: str
    algorithm: EncryptionAlgorithm
    private_key: str


@dataclass
class KeyPair:
    """Cryptographic key pair."""

    public_key: str
    private_key: str
    key_type: KeyType
    created_at: datetime
    expires_at: Optional[datetime] = None


@dataclass
class EncryptedData:
    """Encrypted data structure."""

    ciphertext: str
    algorithm: EncryptionAlgorithm
    nonce: Optional[str] = None
    tag: Optional[str] = None
    metadata: Optional[Dict[str, str]] = None


@dataclass
class SignatureParams:
    """Digital signature parameters."""

    message: str
    private_key: str
    algorithm: str


@dataclass
class VerificationParams:
    """Signature verification parameters."""

    message: str
    signature: str
    public_key: str
    algorithm: str


@dataclass
class RandomParams:
    """Random number generation parameters."""

    length: int
    algorithm: str = "secure"
    seed: Optional[str] = None


@dataclass
class SecureEnclaveParams:
    """Secure enclave parameters."""

    operation: str
    data: str
    enclave_id: str
    attestation: Optional[str] = None
