# Wallet Security Module

## Overview

The Wallet Security module provides comprehensive wallet protection and security features for the Aura blockchain, including hardware wallet integration, multi-signature wallets with weighted signing, social recovery mechanisms, transaction simulation, domain verification, spending limits, session management, biometric authentication, secure enclave storage, encrypted backups, dust attack filtering, and address checksum validation.

## Features

- **Hardware Wallet Support**: Register and verify hardware wallets (Ledger, Trezor, etc.) with firmware attestation
- **Multi-Signature Wallets**: Configurable multi-sig with signer weights, weighted thresholds, and time-locks
- **Social Recovery**: Guardian-based account recovery with threshold approvals and delay periods
- **Transaction Simulation**: Pre-execution simulation to detect malicious transactions and estimate gas
- **Domain Verification**: SSL certificate verification for dApp domains with phishing protection
- **Spending Limits**: Daily/weekly/monthly spending caps per denomination with automatic enforcement
- **Session Management**: Session-based authentication with auto-lock, timeout, and inactivity tracking
- **Biometric Authentication**: Fingerprint/face ID enrollment and verification with replay protection
- **Secure Enclave**: Hardware-backed key storage with attestation certificates
- **Encrypted Backups**: BIP39 seed encryption with configurable KDF and cloud/local storage
- **Dust Attack Protection**: Automatic filtering of dust transactions with pattern detection
- **Address Validation**: Checksum validation for multiple address formats (Bech32, EIP-55, etc.)

## State

### Hardware Wallets
- **HardwareWalletConfig**: Hardware wallet registration with device type, ID, firmware version, derivation path, and attestation signature

### Multi-Signature
- **MultiSigWallet**: Multi-sig wallet with signers, threshold, signer weights, weighted threshold, and optional time-lock
- **PendingMultiSigTransaction**: Pending transactions requiring threshold signatures with expiration

### Social Recovery
- **SocialRecoveryConfig**: Guardian configuration with recovery threshold and delay period
- **RecoveryRequest**: Active recovery request with guardian approvals and execution timestamp

### Transaction Security
- **TransactionSimulation**: Simulation results with state changes, gas estimates, and risk scores
- **DomainVerification**: Domain SSL verification status with certificate hashes and expiration

### Spending Controls
- **SpendingLimit**: Per-wallet per-denom limits (daily, weekly, monthly) with current usage tracking

### Session Management
- **SessionConfig**: Session timeout, auto-lock settings, and inactivity thresholds
- **Session**: Active session with lock status and last activity timestamp

### Biometric Authentication
- **BiometricAuth**: Biometric enrollment data with type (fingerprint, face ID) and verification templates
- **BiometricProof**: Replay protection tracking for used biometric proofs

### Secure Storage
- **SecureEnclaveConfig**: Secure enclave storage with enclave type, encrypted key material, and attestation
- **EncryptedBackup**: Encrypted seed backup with algorithm, KDF parameters, salt, iterations, and storage location

### Attack Protection
- **DustAttackFilter**: Dust filtering configuration with minimum amount thresholds and pattern detection
- **WalletSecurityMetrics**: Per-wallet security metrics including threat scores and anomaly counts

## Messages

### MsgRegisterHardwareWallet
Register hardware wallet with attestation.

**Fields**: `address`, `type`, `device_id`, `firmware_version`, `derivation_path`, `signature`

### MsgCreateMultiSigWallet
Create multi-sig wallet with weighted signing.

**Fields**: `creator`, `signers`, `threshold`, `signer_weights`, `weight_threshold`, `time_lock`

### MsgSignMultiSigTransaction
Sign pending multi-sig transaction.

**Fields**: `tx_id`, `signer`, `signature`

### MsgConfigureSocialRecovery
Configure social recovery guardians.

**Fields**: `wallet_id`, `guardians`, `recovery_threshold`, `recovery_delay`

### MsgInitiateRecovery
Initiate account recovery process.

**Fields**: `wallet_id`, `new_address`, `initiator`

### MsgApproveRecovery
Guardian approves recovery request.

**Fields**: `request_id`, `guardian`, `signature`

### MsgExecuteRecovery
Execute approved recovery request.

**Fields**: `request_id`

### MsgSimulateTransaction
Simulate transaction before execution.

**Fields**: `tx_data`, `sender`

### MsgVerifyDomain
Verify dApp domain certificate.

**Fields**: `domain`, `certificate_hash`, `verifier`

### MsgSetSpendingLimit
Configure spending limits.

**Fields**: `wallet_id`, `denom`, `daily_limit`, `weekly_limit`, `monthly_limit`

### MsgConfigureSession
Configure session parameters.

**Fields**: `wallet_id`, `timeout_duration`, `auto_lock_enabled`, `inactivity_threshold_seconds`

### MsgLockSession
Lock active session.

**Fields**: `session_id`

### MsgUnlockSession
Unlock session with authentication.

**Fields**: `session_id`, `authentication_proof`

### MsgEnrollBiometric
Enroll biometric authentication.

**Fields**: `wallet_id`, `type`, `enrollment_data`

### MsgAuthenticateBiometric
Authenticate using biometric proof.

**Fields**: `wallet_id`, `biometric_proof`

### MsgStoreInSecureEnclave
Store keys in secure enclave.

**Fields**: `wallet_id`, `enclave_type`, `encrypted_key_material`, `attestation_certificate`

### MsgCreateEncryptedBackup
Create encrypted seed backup.

**Fields**: `wallet_id`, `encrypted_seed`, `encryption_algorithm`, `key_derivation_function`, `salt`, `iterations`, `location`

### MsgConfigureDustFilter
Configure dust attack filtering.

**Fields**: `wallet_id`, `enabled`, `minimum_amount`, `max_dust_transactions_per_block`, `suspicious_pattern_threshold`

### MsgValidateAddressChecksum
Validate address checksum.

**Fields**: `address`, `algorithm`
