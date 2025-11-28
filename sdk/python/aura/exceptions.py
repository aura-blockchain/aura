"""Custom exceptions for AURA SDK."""


class AuraSDKError(Exception):
    """Base exception for AURA SDK."""

    pass


class BridgeError(AuraSDKError):
    """Exception for bridge operations."""

    pass


class ComplianceError(AuraSDKError):
    """Exception for compliance operations."""

    pass


class ConfidenceScoreError(AuraSDKError):
    """Exception for confidence score operations."""

    pass


class CryptographyError(AuraSDKError):
    """Exception for cryptography operations."""

    pass


class DataRegistryError(AuraSDKError):
    """Exception for data registry operations."""

    pass


class EconomicSecurityError(AuraSDKError):
    """Exception for economic security operations."""

    pass


class IdentityChangeError(AuraSDKError):
    """Exception for identity change operations."""

    pass


class InclusionRoutinesError(AuraSDKError):
    """Exception for inclusion routines operations."""

    pass


class MonitoringError(AuraSDKError):
    """Exception for monitoring operations."""

    pass


class NetworkSecurityError(AuraSDKError):
    """Exception for network security operations."""

    pass


class PrevalidationError(AuraSDKError):
    """Exception for prevalidation operations."""

    pass


class PrivacyError(AuraSDKError):
    """Exception for privacy operations."""

    pass


class ValidatorSecurityError(AuraSDKError):
    """Exception for validator security operations."""

    pass


class VCRegistryError(AuraSDKError):
    """Exception for VC registry operations."""

    pass


class WalletSecurityError(AuraSDKError):
    """Exception for wallet security operations."""

    pass
