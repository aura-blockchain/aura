package types

import (
	"time"

	governancepb "github.com/aequitas/aura/proto/aura/governance/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ToProtoVote converts a Vote to its proto representation
func ToProtoVote(v *Vote) *governancepb.Vote {
	if v == nil {
		return nil
	}
	return &governancepb.Vote{
		ProposalId:     v.ProposalID,
		Voter:          v.Voter,
		Option:         governancepb.VoteOption(v.Option),
		Timestamp:      timestamppb.New(v.Timestamp),
		IsSecret:       v.IsSecret,
		EncryptedVote:  v.EncryptedVote,
		VoteCommitment: v.Commitment,
		VotingPower:    v.VotingPower,
	}
}

// FromProtoVote converts a proto Vote to the Go type
func FromProtoVote(pv *governancepb.Vote) *Vote {
	if pv == nil {
		return nil
	}
	var timestamp time.Time
	if pv.Timestamp != nil {
		timestamp = pv.Timestamp.AsTime()
	}
	return &Vote{
		ProposalID:    pv.ProposalId,
		Voter:         pv.Voter,
		Option:        VoteOption(pv.Option),
		Timestamp:     timestamp,
		IsSecret:      pv.IsSecret,
		EncryptedVote: pv.EncryptedVote,
		Commitment:    pv.VoteCommitment,
		VotingPower:   pv.VotingPower,
	}
}

// ToProtoDeposit converts a Deposit to its proto representation
func ToProtoDeposit(d *Deposit) *governancepb.Deposit {
	if d == nil {
		return nil
	}
	return &governancepb.Deposit{
		ProposalId: d.ProposalID,
		Depositor:  d.Depositor,
		Amount:     d.Amount,
		Timestamp:  timestamppb.New(d.Timestamp),
	}
}

// FromProtoDeposit converts a proto Deposit to the Go type
func FromProtoDeposit(pd *governancepb.Deposit) *Deposit {
	if pd == nil {
		return nil
	}
	var timestamp time.Time
	if pd.Timestamp != nil {
		timestamp = pd.Timestamp.AsTime()
	}
	return &Deposit{
		ProposalID: pd.ProposalId,
		Depositor:  pd.Depositor,
		Amount:     pd.Amount,
		Timestamp:  timestamp,
	}
}

// ToProtoVoteDelegation converts a VoteDelegation to its proto representation
func ToProtoVoteDelegation(vd *VoteDelegation) *governancepb.VoteDelegation {
	if vd == nil {
		return nil
	}
	categories := make([]governancepb.ProposalCategory, len(vd.Categories))
	for i, cat := range vd.Categories {
		categories[i] = governancepb.ProposalCategory(cat)
	}
	return &governancepb.VoteDelegation{
		Delegator:      vd.Delegator,
		Delegate:       vd.Delegate,
		DelegationTime: timestamppb.New(vd.DelegationTime),
		DelegatedPower: vd.DelegatedPower,
		Categories:     categories,
	}
}

// FromProtoVoteDelegation converts a proto VoteDelegation to the Go type
func FromProtoVoteDelegation(pvd *governancepb.VoteDelegation) *VoteDelegation {
	if pvd == nil {
		return nil
	}
	var delegationTime time.Time
	if pvd.DelegationTime != nil {
		delegationTime = pvd.DelegationTime.AsTime()
	}
	categories := make([]ProposalCategory, len(pvd.Categories))
	for i, cat := range pvd.Categories {
		categories[i] = ProposalCategory(cat)
	}
	return &VoteDelegation{
		Delegator:      pvd.Delegator,
		Delegate:       pvd.Delegate,
		DelegationTime: delegationTime,
		DelegatedPower: pvd.DelegatedPower,
		Categories:     categories,
	}
}

// ToProtoTokenLock converts a TokenLock to its proto representation
func ToProtoTokenLock(tl *TokenLock) *governancepb.TokenLock {
	if tl == nil {
		return nil
	}
	return &governancepb.TokenLock{
		Owner:        tl.Owner,
		ProposalId:   tl.ProposalID,
		LockedAmount: tl.LockedAmount,
		LockTime:     timestamppb.New(tl.LockTime),
		UnlockTime:   timestamppb.New(tl.UnlockTime),
	}
}

// FromProtoTokenLock converts a proto TokenLock to the Go type
func FromProtoTokenLock(ptl *governancepb.TokenLock) *TokenLock {
	if ptl == nil {
		return nil
	}
	var lockTime, unlockTime time.Time
	if ptl.LockTime != nil {
		lockTime = ptl.LockTime.AsTime()
	}
	if ptl.UnlockTime != nil {
		unlockTime = ptl.UnlockTime.AsTime()
	}
	return &TokenLock{
		Owner:        ptl.Owner,
		ProposalID:   ptl.ProposalId,
		LockedAmount: ptl.LockedAmount,
		LockTime:     lockTime,
		UnlockTime:   unlockTime,
	}
}
