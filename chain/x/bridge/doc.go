// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

// Package bridge implements secure cross-chain asset transfer functionality for the AURA blockchain.
//
// The bridge module enables trustless asset transfers between AURA and other blockchain networks
// through validator attestation, fraud proofs, time-locked withdrawals, and comprehensive security measures.
//
// # Overview
//
// The bridge module manages:
//   - Cross-chain asset transfers with multi-validator attestation
//   - Wrapped token creation and management
//   - Cross-chain atomic swaps
//   - Shared identity verification across chains
//   - Relayer performance tracking and incentives
//   - Fraud proof submission and resolution
//   - Chain configuration and support
//
// # Architecture
//
// The module implements a multi-layered security approach:
//   - Validator Attestation: Multiple validators must confirm transfers
//   - Fraud Proof Window: Time period for challenging fraudulent transfers
//   - Circuit Breakers: Automatic pause on suspicious activity
//   - Time Locks: Delayed withdrawals for security
//   - Merkle Proofs: Cryptographic verification of cross-chain state
//
// # Key Components
//
// ## Cross-Chain Transfers
//
// Transfers go through multiple stages with security checks:
//
//	// Initiate a transfer
//	transfer := &types.CrossChainTransfer{
//	    SourceChain: "ethereum",
//	    TargetChain: "aura",
//	    Sender: ethAddress,
//	    Recipient: auraAddress,
//	    Amount: "1000000",
//	    Denom: "uaura",
//	    Status: types.TransferStatus_PENDING,
//	    RequiredConfirmations: 5,
//	}
//
//	// Validators attest to the transfer
//	keeper.SubmitAttestation(ctx, transferID, validatorAddr, true)
//
//	// Check if threshold reached
//	if keeper.CheckAttestationThreshold(ctx, transferID) {
//	    // Process the transfer
//	    keeper.ProcessWithdrawal(ctx, recipient, amount, transferID)
//	}
//
// ## Fraud Proofs
//
// Challenge suspicious transfers with cryptographic evidence:
//
//	// Submit fraud proof
//	err := keeper.SubmitFraudProof(ctx, transferID, challenger, proofData)
//
//	// Resolve fraud proof (by validators)
//	proof, err := keeper.ResolveFraudProof(ctx, transferID, valid)
//	if valid {
//	    // Challenger receives reward
//	    // Transfer is marked fraudulent
//	}
//
// ## Wrapped Tokens
//
// Create wrapped representations of foreign chain assets:
//
//	wrappedToken := &types.WrappedToken{
//	    WrappedDenom: "wrapped-eth",
//	    OriginalDenom: "ETH",
//	    OriginChain: "ethereum",
//	    TotalSupply: "1000000",
//	    BurnEnabled: true,
//	}
//	keeper.setWrappedToken(ctx, wrappedToken)
//
// ## Cross-Chain Swaps
//
// Atomic swaps between chains using hash time-locked contracts:
//
//	swap := &types.CrossChainSwap{
//	    SwapId: "swap-123",
//	    SourceChain: "aura",
//	    TargetChain: "ethereum",
//	    HashLock: hashLock,
//	    TimeLock: timestamppb.New(time.Now().Add(24 * time.Hour)),
//	    Status: types.SwapStatus_SWAP_PENDING,
//	}
//	keeper.setSwap(ctx, swap)
//
// ## Shared Identity
//
// Link identities across multiple chains:
//
//	identity := &types.SharedIdentity{
//	    Address: auraAddress,
//	    ChainAddresses: map[string]string{
//	        "ethereum": ethAddress,
//	        "bitcoin": btcAddress,
//	    },
//	    Verified: true,
//	}
//	keeper.setSharedIdentity(ctx, identity)
//
// ## Relayer Management
//
// Track and incentivize bridge relayers:
//
//	keeper.recordRelayerStats(ctx, relayerAddr, true, amount)
//	stats, _ := keeper.getRelayerStats(ctx, relayerAddr)
//	// stats.TotalTransfersRelayed, stats.SuccessfulTransfers, etc.
//
// ## Chain Configuration
//
// Add support for new chains:
//
//	chainConfig := types.ChainConfig{
//	    ChainId: "ethereum-mainnet",
//	    Enabled: true,
//	    MinConfirmations: 12,
//	    BridgeAddress: "0x...",
//	}
//	keeper.AddSupportedChain(ctx, chainConfig)
//
// # Security Features
//
// The bridge module implements comprehensive security:
//
// ## Reentrancy Protection
//
//	keeper.reentrancyGuard.Enter(ctx)
//	defer keeper.reentrancyGuard.Exit(ctx)
//
// ## Pause Mechanism
//
//	if err := keeper.pauseGuard.RequireNotPaused(ctx); err != nil {
//	    return err
//	}
//
// ## Input Validation
//
//	if err := keeper.inputValidator.ValidateAddress(address); err != nil {
//	    return err
//	}
//
// ## Safe Math Operations
//
//	result, err := keeper.safeMath.Add(amount1, amount2)
//
// ## Gas Limit Protection
//
//	if err := keeper.gasLimitGuard.CheckGasLimit(ctx); err != nil {
//	    return err
//	}
//
// ## Access Control
//
//	if err := keeper.accessControl.RequireRole(ctx, address, "relayer"); err != nil {
//	    return err
//	}
//
// # State Structure
//
// The module stores data with the following key prefixes:
//   - TransferPrefix: Cross-chain transfers
//   - ChainConfigPrefix: Supported chain configurations
//   - SharedIdentityPrefix: Cross-chain identity mappings
//   - SwapPrefix: Atomic swap data
//   - WrappedTokenPrefix: Wrapped token metadata
//   - RelayerPrefix: Relayer statistics
//   - ValidatorPrefix: Bridge validator set
//   - AttestationPrefix: Validator attestations
//   - FraudProofPrefix: Fraud proof submissions
//
// # Performance Optimizations
//
// Performance features include:
//   - Indexed lookups by transfer ID and hash
//   - Efficient prefix iteration for batch queries
//   - Cached chain configurations
//   - Optimized Merkle proof verification
//   - Batched attestation processing
//
// # Events
//
// The module emits events for:
//   - transfer_initiated
//   - transfer_completed
//   - attestation_submitted
//   - fraud_proof_submitted
//   - fraud_proof_resolved
//   - chain_added
//   - wrapped_token_created
//   - swap_initiated
//   - swap_completed
//
// # Queries
//
// Available queries:
//   - Transfer: Get transfer by ID
//   - TransferByHash: Lookup transfer by hash
//   - Attestations: Get validator attestations
//   - FraudProof: Query fraud proof status
//   - SupportedChains: List all supported chains
//   - WrappedTokens: Query wrapped token info
//   - RelayerStats: Get relayer performance data
//   - SharedIdentity: Lookup cross-chain identity
//
// # Transactions
//
// Available transactions:
//   - InitiateTransfer: Start cross-chain transfer
//   - SubmitAttestation: Validator confirms transfer
//   - InitiateWithdrawal: Request withdrawal
//   - ExecuteWithdrawal: Complete time-locked withdrawal
//   - SubmitFraudProof: Challenge fraudulent transfer
//   - ResolveFraudProof: Finalize fraud investigation
//   - AddSupportedChain: Add new chain support
//   - DisableChain: Temporarily disable a chain
//   - CreateWrappedToken: Register wrapped asset
//
// # Module Parameters
//
// Configurable parameters:
//   - BridgeEnabled: Global bridge on/off switch
//   - MinConfirmations: Required validator confirmations
//   - MaxTransferAmount: Circuit breaker threshold
//   - BridgeFeeBasisPoints: Fee percentage (in basis points)
//   - FraudProofWindow: Time to submit fraud proofs
//   - FraudProofReward: Reward for valid fraud proof
//   - WithdrawalTimelock: Delay before withdrawal execution
//
// # Integration Example
//
//	import (
//	    bridgekeeper "github.com/aequitas/aura/chain/x/bridge/keeper"
//	    bridgetypes "github.com/aequitas/aura/chain/x/bridge/types"
//	)
//
//	// In app.go
//	app.BridgeKeeper = bridgekeeper.NewKeeper(
//	    appCodec,
//	    keys[bridgetypes.StoreKey],
//	    &app.ParamsKeeper.Subspace,
//	    app.BankKeeper,
//	    app.AccountKeeper,
//	    app.VCKeeper,
//	)
//
// # Security Considerations
//
// When using the bridge module:
//  1. Set appropriate MinConfirmations based on source chain security
//  2. Monitor fraud proof submissions for unusual activity
//  3. Regularly review relayer performance metrics
//  4. Implement circuit breakers for maximum transfer amounts
//  5. Use time-locks for large withdrawals
//  6. Verify shared identities through multiple sources
//  7. Keep fraud proof window long enough for detection
//  8. Ensure validator set is properly incentivized
//
// # Compliance Features
//
// The module supports compliance requirements:
//   - Complete audit trail of all transfers
//   - Fraud detection and resolution process
//   - Configurable transfer limits
//   - Identity verification integration
//   - Validator accountability through attestations
//
// For detailed documentation, see:
// https://github.com/aequitas/aura/tree/main/docs/modules/bridge
package bridge
