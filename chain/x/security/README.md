# Security Module

## Overview

The Security module consolidates six security domains into a unified layer: network security, validator security, wallet security, incident response, cryptography, and privacy. It provides peer reputation management, double-sign detection, multi-sig wallets, social recovery, cryptographic operations, mixing pools, and privacy-preserving transactions.

## Features

- **Network Security**: Trusted peers, peer banning, reputation scoring, fork/partition detection
- **Validator Security**: Sentry nodes, double-sign reporting, validator alerts, automatic failover
- **Wallet Security**: Hardware wallet registration, multi-sig wallets, social recovery, spending limits, biometric authentication
- **Incident Response**: Security incident tracking, emergency response actions, audit logging
- **Cryptography**: Key rotation scheduling, threshold signature schemes, zero-knowledge proof circuits
- **Privacy**: Mixing pools, stealth addresses, ring signatures, confidential transactions

## State

### Network Security
- **TrustedPeer**: Whitelisted peer with trust score
- **BannedPeer**: Blacklisted peer with ban reason and duration
- **PeerReputation**: Reputation score based on behavior
- **ForkAlert**: Detected blockchain fork with height and hashes
- **PartitionAlert**: Network partition detection

### Validator Security
- **ValidatorSecurity**: Validator security profile with sentry nodes
- **SentryNode**: Sentry node registration for validator protection
- **DoubleSignReport**: Evidence of double-signing with slashing
- **ValidatorAlert**: Security alerts for validators

### Wallet Security
- **HardwareWallet**: Registered hardware wallet device
- **MultiSigWallet**: Multi-signature wallet with threshold
- **MultiSigTransaction**: Pending multi-sig transaction
- **SocialRecovery**: Social recovery configuration with guardians
- **RecoveryRequest**: Active recovery attempt with approvals
- **SpendingLimit**: Daily/transaction spending limits
- **BiometricAuth**: Biometric authentication registration

### Cryptography
- **KeyRotationSchedule**: Automated key rotation schedule
- **ThresholdScheme**: Threshold signature scheme (t-of-n)
- **ZKProofCircuit**: Zero-knowledge proof circuit registration
- **QuantumResistantKey**: Post-quantum cryptographic keys

### Privacy
- **MixingPool**: Privacy mixing pool for transaction obfuscation
- **StealthAddress**: One-time stealth address for privacy
- **RingSignature**: Ring signature for sender anonymity
- **ConfidentialTransaction**: Transaction with hidden amounts

## Messages

### Network Security
- **MsgAddTrustedPeer**: Add peer to trusted list
- **MsgBanPeer**: Ban malicious peer
- **MsgUpdatePeerReputation**: Update peer reputation score
- **MsgResolveForkAlert**: Resolve fork detection alert
- **MsgResolvePartitionAlert**: Resolve network partition

### Validator Security
- **MsgRegisterValidatorSecurity**: Register validator security profile
- **MsgRegisterSentryNode**: Register sentry node
- **MsgReportDoubleSign**: Report double-signing evidence
- **MsgTriggerFailover**: Trigger validator failover

### Wallet Security
- **MsgRegisterHardwareWallet**: Register hardware wallet
- **MsgCreateMultiSigWallet**: Create multi-sig wallet
- **MsgProposeMultiSigTransaction**: Propose multi-sig transaction
- **MsgSignMultiSigTransaction**: Sign multi-sig transaction
- **MsgConfigureSocialRecovery**: Set up social recovery
- **MsgInitiateRecovery**: Start wallet recovery process
- **MsgSetSpendingLimits**: Configure spending limits

### Cryptography
- **MsgCreateKeyRotationSchedule**: Schedule key rotation
- **MsgRotateKey**: Execute key rotation
- **MsgCreateThresholdScheme**: Create threshold signature scheme
- **MsgRegisterZKProofCircuit**: Register ZK proof circuit
- **MsgGenerateQuantumResistantKey**: Generate post-quantum key

### Privacy
- **MsgCreateMixingPool**: Create privacy mixing pool
- **MsgJoinMixingPool**: Join mixing pool
- **MsgGenerateStealthAddress**: Generate stealth address
- **MsgCreateRingSignature**: Create ring signature
- **MsgCreateConfidentialTransaction**: Create confidential transaction

## BeginBlock Operations

- Process cryptographic key rotations
- Update network security metrics
- Check validator security health
- Process wallet security checks
- Update incident response state
- Refresh privacy mixing pools

## EndBlock Operations

- Clean up expired sessions
- Finalize security metrics

## Events

### EventPeerBanned
Emitted when malicious peer is banned.

**Attributes**: `peer_id`, `reason`, `duration`

### EventDoubleSignReported
Emitted when double-signing evidence is submitted.

**Attributes**: `validator`, `height`, `slash_amount`

### EventMultiSigTransactionProposed
Emitted when multi-sig transaction is proposed.

**Attributes**: `wallet_id`, `tx_id`, `proposer`

### EventMultiSigTransactionSigned
Emitted when multi-sig transaction receives signature.

**Attributes**: `wallet_id`, `tx_id`, `signer`, `signatures_count`

### EventSocialRecoveryInitiated
Emitted when wallet recovery process begins.

**Attributes**: `wallet`, `initiator`, `required_approvals`

### EventKeyRotated
Emitted when cryptographic key is rotated.

**Attributes**: `key_id`, `old_key_hash`, `new_key_hash`

### EventMixingPoolCreated
Emitted when new privacy mixing pool is created.

**Attributes**: `pool_id`, `min_participants`, `denomination`

### EventStealthAddressGenerated
Emitted when stealth address is generated.

**Attributes**: `recipient`, `stealth_address`

## Integration Notes

- Multi-sig transactions require threshold signatures before execution
- Social recovery requires guardian approvals (configurable threshold)
- Spending limits are enforced before transaction execution
- Key rotations are processed automatically on schedule
- Mixing pools provide transaction privacy through obfuscation
- Stealth addresses enable unlinkable payments
- Ring signatures provide sender anonymity
