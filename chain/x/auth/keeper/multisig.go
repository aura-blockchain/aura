package keeper

import (
	"context"
	"fmt"
	"time"

	"github.com/aequitas/aura/chain/x/auth/types"
	authproto "github.com/aequitas/aura/proto/aura/auth/v1beta1"
)

// CreateMultisigWallet creates a new multisig wallet
func (k *Keeper) CreateMultisigWallet(ctx context.Context, creator string, signers []string, threshold uint32, walletType authproto.WalletType) (*authproto.MultisigWallet, error) {
	// Validate creator has permission
	if err := k.RequirePermission(ctx, creator, types.PermissionManageMultisig); err != nil {
		return nil, err
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	now := time.Now()
	walletID := types.GenerateID("wallet", creator, now.String())

	wallet := &authproto.MultisigWallet{
		Id:         walletID,
		Signers:    signers,
		Threshold:  threshold,
		CreatedAt:  &now,
		CreatedBy:  creator,
		WalletType: walletType,
	}

	// Validate wallet configuration
	if err := types.ValidateMultisigWallet(wallet); err != nil {
		k.LogAudit(ctx, creator, "create_multisig_wallet", walletID, "failed", nil, err.Error())
		return nil, fmt.Errorf("%w: %v", types.ErrInvalidMultisigWallet, err)
	}

	k.multisigWallets[walletID] = wallet
	k.LogAudit(ctx, creator, "create_multisig_wallet", walletID, "success", map[string]string{
		"signers":   fmt.Sprintf("%v", signers),
		"threshold": fmt.Sprintf("%d", threshold),
		"type":      walletType.String(),
	}, "")

	return wallet, nil
}

// GetMultisigWallet retrieves a multisig wallet by ID

// ListMultisigWallets returns all multisig wallets
func (k *Keeper) ListMultisigWallets() []*authproto.MultisigWallet {
	k.mu.RLock()
	defer k.mu.RUnlock()

	wallets := make([]*authproto.MultisigWallet, 0, len(k.multisigWallets))
	for _, wallet := range k.multisigWallets {
		wallets = append(wallets, wallet)
	}

	return wallets
}

// CreateMultisigProposal creates a new proposal for a multisig wallet
func (k *Keeper) CreateMultisigProposal(ctx context.Context, proposer, walletID, title, description string, payload []byte, expiresInSeconds int64) (*authproto.MultisigProposal, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	// Verify wallet exists
	wallet, ok := k.multisigWallets[walletID]
	if !ok {
		k.LogAudit(ctx, proposer, "create_multisig_proposal", walletID, "failed", nil, "wallet not found")
		return nil, types.ErrMultisigWalletNotFound
	}

	// Verify proposer is a signer
	isSigner := false
	for _, signer := range wallet.Signers {
		if signer == proposer {
			isSigner = true
			break
		}
	}
	if !isSigner {
		k.LogAudit(ctx, proposer, "create_multisig_proposal", walletID, "failed", nil, "not a signer")
		return nil, types.ErrNotWalletSigner
	}

	now := time.Now()
	proposalID := types.GenerateID("proposal", walletID, proposer, now.String())

	// Calculate expiry time
	var expiresAt *time.Time
	if expiresInSeconds > 0 {
		expiry := now.Add(time.Duration(expiresInSeconds) * time.Second)
		expiresAt = &expiry
	} else {
		// Use default from params
		expiry := now.Add(time.Duration(k.params.MultisigProposalExpirySeconds) * time.Second)
		expiresAt = &expiry
	}

	proposal := &authproto.MultisigProposal{
		Id:          proposalID,
		WalletId:    walletID,
		Title:       title,
		Description: description,
		Payload:     payload,
		Signatures:  []string{proposer}, // Proposer auto-signs
		Status:      authproto.ProposalStatus_PROPOSAL_STATUS_PENDING,
		CreatedAt:   &now,
		ExpiresAt:   expiresAt,
	}

	// Validate proposal
	if err := types.ValidateMultisigProposal(proposal); err != nil {
		k.LogAudit(ctx, proposer, "create_multisig_proposal", proposalID, "failed", nil, err.Error())
		return nil, fmt.Errorf("%w: %v", types.ErrInvalidProposal, err)
	}

	// Check if auto-approved (threshold = 1)
	if types.IsProposalApproved(proposal, wallet) {
		proposal.Status = authproto.ProposalStatus_PROPOSAL_STATUS_APPROVED
	}

	k.multisigProposals[proposalID] = proposal
	k.LogAudit(ctx, proposer, "create_multisig_proposal", proposalID, "success", map[string]string{
		"wallet_id": walletID,
		"title":     title,
	}, "")

	return proposal, nil
}

// SignMultisigProposal adds a signature to a proposal
func (k *Keeper) SignMultisigProposal(ctx context.Context, signer, proposalID string) (*authproto.MultisigProposal, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	// Get proposal
	proposal, ok := k.multisigProposals[proposalID]
	if !ok {
		k.LogAudit(ctx, signer, "sign_multisig_proposal", proposalID, "failed", nil, "proposal not found")
		return nil, types.ErrProposalNotFound
	}

	// Check if proposal is expired
	if types.IsProposalExpired(proposal) {
		proposal.Status = authproto.ProposalStatus_PROPOSAL_STATUS_EXPIRED
		k.LogAudit(ctx, signer, "sign_multisig_proposal", proposalID, "failed", nil, "proposal expired")
		return nil, types.ErrProposalExpired
	}

	// Check if already executed
	if proposal.Status == authproto.ProposalStatus_PROPOSAL_STATUS_EXECUTED {
		k.LogAudit(ctx, signer, "sign_multisig_proposal", proposalID, "failed", nil, "already executed")
		return nil, types.ErrProposalAlreadyExecuted
	}

	// Get wallet
	wallet, ok := k.multisigWallets[proposal.WalletId]
	if !ok {
		return nil, types.ErrMultisigWalletNotFound
	}

	// Verify signer is authorized
	isSigner := false
	for _, s := range wallet.Signers {
		if s == signer {
			isSigner = true
			break
		}
	}
	if !isSigner {
		k.LogAudit(ctx, signer, "sign_multisig_proposal", proposalID, "failed", nil, "not a signer")
		return nil, types.ErrNotWalletSigner
	}

	// Check if already signed
	for _, sig := range proposal.Signatures {
		if sig == signer {
			k.LogAudit(ctx, signer, "sign_multisig_proposal", proposalID, "failed", nil, "already signed")
			return nil, types.ErrAlreadySigned
		}
	}

	// Add signature
	proposal.Signatures = append(proposal.Signatures, signer)

	// Check if threshold reached
	if types.IsProposalApproved(proposal, wallet) {
		proposal.Status = authproto.ProposalStatus_PROPOSAL_STATUS_APPROVED
	}

	k.LogAudit(ctx, signer, "sign_multisig_proposal", proposalID, "success", map[string]string{
		"signature_count": fmt.Sprintf("%d", len(proposal.Signatures)),
		"threshold":       fmt.Sprintf("%d", wallet.Threshold),
		"approved":        fmt.Sprintf("%v", proposal.Status == authproto.ProposalStatus_PROPOSAL_STATUS_APPROVED),
	}, "")

	return proposal, nil
}

// ExecuteMultisigProposal executes an approved proposal
func (k *Keeper) ExecuteMultisigProposal(ctx context.Context, executor, proposalID string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	// Get proposal
	proposal, ok := k.multisigProposals[proposalID]
	if !ok {
		k.LogAudit(ctx, executor, "execute_multisig_proposal", proposalID, "failed", nil, "proposal not found")
		return types.ErrProposalNotFound
	}

	// Check if proposal is expired
	if types.IsProposalExpired(proposal) {
		proposal.Status = authproto.ProposalStatus_PROPOSAL_STATUS_EXPIRED
		k.LogAudit(ctx, executor, "execute_multisig_proposal", proposalID, "failed", nil, "proposal expired")
		return types.ErrProposalExpired
	}

	// Check if already executed
	if proposal.Status == authproto.ProposalStatus_PROPOSAL_STATUS_EXECUTED {
		k.LogAudit(ctx, executor, "execute_multisig_proposal", proposalID, "failed", nil, "already executed")
		return types.ErrProposalAlreadyExecuted
	}

	// Get wallet
	wallet, ok := k.multisigWallets[proposal.WalletId]
	if !ok {
		return types.ErrMultisigWalletNotFound
	}

	// Verify proposal is approved
	if !types.IsProposalApproved(proposal, wallet) {
		k.LogAudit(ctx, executor, "execute_multisig_proposal", proposalID, "failed", nil, "not approved")
		return types.ErrProposalNotApproved
	}

	// Mark as executed
	now := time.Now()
	proposal.Status = authproto.ProposalStatus_PROPOSAL_STATUS_EXECUTED
	proposal.ExecutedAt = &now

	k.LogAudit(ctx, executor, "execute_multisig_proposal", proposalID, "success", map[string]string{
		"wallet_id": proposal.WalletId,
		"title":     proposal.Title,
	}, "")

	// Note: In a real implementation, the payload would be decoded and executed here
	// This would involve unmarshaling the payload and executing the appropriate action

	return nil
}

// GetMultisigProposal retrieves a proposal by ID
// ListMultisigProposals returns all proposals for a wallet
func (k *Keeper) ListMultisigProposals(walletID string, status authproto.ProposalStatus) []*authproto.MultisigProposal {
	k.mu.RLock()
	defer k.mu.RUnlock()

	proposals := make([]*authproto.MultisigProposal, 0)
	for _, proposal := range k.multisigProposals {
		if walletID != "" && proposal.WalletId != walletID {
			continue
		}
		if status != authproto.ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED && proposal.Status != status {
			continue
		}
		proposals = append(proposals, proposal)
	}

	return proposals
}
