package keeper

import (
	"encoding/json"

	"github.com/aequitas/aura/chain/x/governance/types"
)

// JSON marshaling helpers for types that don't have proto.Message interface

func (k *Keeper) marshalVote(vote *types.Vote) ([]byte, error) {
	return json.Marshal(vote)
}

func (k *Keeper) unmarshalVote(bz []byte) (*types.Vote, error) {
	var vote types.Vote
	err := json.Unmarshal(bz, &vote)
	return &vote, err
}

func (k *Keeper) marshalDeposit(deposit *types.Deposit) ([]byte, error) {
	return json.Marshal(deposit)
}

func (k *Keeper) unmarshalDeposit(bz []byte) (*types.Deposit, error) {
	var deposit types.Deposit
	err := json.Unmarshal(bz, &deposit)
	return &deposit, err
}

func (k *Keeper) marshalVoteDelegation(delegation *types.VoteDelegation) ([]byte, error) {
	return json.Marshal(delegation)
}

func (k *Keeper) unmarshalVoteDelegation(bz []byte) (*types.VoteDelegation, error) {
	var delegation types.VoteDelegation
	err := json.Unmarshal(bz, &delegation)
	return &delegation, err
}

func (k *Keeper) marshalTokenLock(lock *types.TokenLock) ([]byte, error) {
	return json.Marshal(lock)
}

func (k *Keeper) unmarshalTokenLock(bz []byte) (*types.TokenLock, error) {
	var lock types.TokenLock
	err := json.Unmarshal(bz, &lock)
	return &lock, err
}
