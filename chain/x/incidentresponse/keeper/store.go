package keeper

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"

	storetypes "cosmossdk.io/store/types"
	"github.com/aequitas/aura/chain/x/incidentresponse/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Store provides typed KV accessors for incident response state.
type Store struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec
}

// NewStore constructs a Store wrapper.
func NewStore(storeKey storetypes.StoreKey, cdc codec.BinaryCodec) Store {
	return Store{storeKey: storeKey, cdc: cdc}
}

func (s Store) kv(ctx context.Context) storetypes.KVStore {
	return sdk.UnwrapSDKContext(ctx).KVStore(s.storeKey)
}

// ============================
// INCIDENT MANAGEMENT
// ============================

// SetIncident stores an incident
func (s Store) SetIncident(ctx context.Context, incident *types.Incident) error {
	bz, err := json.Marshal(incident)
	if err != nil {
		return err
	}
	s.kv(ctx).Set(types.IncidentKey(incident.ID), bz)
	return nil
}

// GetIncident retrieves an incident
func (s Store) GetIncident(ctx context.Context, incidentID string) (*types.Incident, bool) {
	bz := s.kv(ctx).Get(types.IncidentKey(incidentID))
	if bz == nil {
		return nil, false
	}
	var incident types.Incident
	if err := json.Unmarshal(bz, &incident); err != nil {
		return nil, false
	}
	return &incident, true
}

// DeleteIncident deletes an incident
func (s Store) DeleteIncident(ctx context.Context, incidentID string) {
	s.kv(ctx).Delete(types.IncidentKey(incidentID))
}

// IterateIncidents iterates over all incidents
func (s Store) IterateIncidents(ctx context.Context) []*types.Incident {
	it := storetypes.KVStorePrefixIterator(s.kv(ctx), types.IncidentKeyPrefix)
	defer it.Close()

	incidents := make([]*types.Incident, 0, 64)
	for ; it.Valid(); it.Next() {
		var incident types.Incident
		if err := json.Unmarshal(it.Value(), &incident); err == nil {
			incidentCopy := incident
			incidents = append(incidents, &incidentCopy)
		}
	}
	return incidents
}

// ============================
// PAUSE STATE MANAGEMENT
// ============================

// SetPauseState stores the chain pause state
func (s Store) SetPauseState(ctx context.Context, state *types.ChainPauseState) error {
	bz, err := json.Marshal(state)
	if err != nil {
		return err
	}
	s.kv(ctx).Set(types.PauseStateKey, bz)
	return nil
}

// GetPauseState retrieves the chain pause state
func (s Store) GetPauseState(ctx context.Context) (*types.ChainPauseState, bool) {
	bz := s.kv(ctx).Get(types.PauseStateKey)
	if bz == nil {
		return nil, false
	}
	var state types.ChainPauseState
	if err := json.Unmarshal(bz, &state); err != nil {
		return nil, false
	}
	return &state, true
}

// ============================
// PAUSE VOTE MANAGEMENT
// ============================

// SetPauseVote records a vote for a pause request
func (s Store) SetPauseVote(ctx context.Context, pauseRequestID string, signer string) {
	s.kv(ctx).Set(types.PauseVoteKey(pauseRequestID, signer), []byte{1})
}

// GetPauseVotes retrieves all votes for a pause request
func (s Store) GetPauseVotes(ctx context.Context, pauseRequestID string) []string {
	prefix := types.PauseVotePrefixKey(pauseRequestID)
	it := storetypes.KVStorePrefixIterator(s.kv(ctx), prefix)
	defer it.Close()

	votes := make([]string, 0, 64)
	for ; it.Valid(); it.Next() {
		suffix := bytes.TrimPrefix(it.Key(), prefix)
		suffix = bytes.TrimPrefix(suffix, []byte{':'})
		votes = append(votes, string(suffix))
	}
	return votes
}

// DeletePauseVotes deletes all votes for a pause request
func (s Store) DeletePauseVotes(ctx context.Context, pauseRequestID string) {
	prefix := types.PauseVotePrefixKey(pauseRequestID)
	it := storetypes.KVStorePrefixIterator(s.kv(ctx), prefix)
	defer it.Close()

	keys := [][]byte{}
	for ; it.Valid(); it.Next() {
		keys = append(keys, it.Key())
	}

	for _, key := range keys {
		s.kv(ctx).Delete(key)
	}
}

// ============================
// WALLET LIMITS MANAGEMENT
// ============================

// SetWalletLimit stores wallet limits
func (s Store) SetWalletLimit(ctx context.Context, limits *types.WalletLimits) error {
	bz, err := json.Marshal(limits)
	if err != nil {
		return err
	}
	s.kv(ctx).Set(types.WalletLimitKey(limits.Address), bz)
	return nil
}

// GetWalletLimit retrieves wallet limits
func (s Store) GetWalletLimit(ctx context.Context, address string) (*types.WalletLimits, bool) {
	bz := s.kv(ctx).Get(types.WalletLimitKey(address))
	if bz == nil {
		return nil, false
	}
	var limits types.WalletLimits
	if err := json.Unmarshal(bz, &limits); err != nil {
		return nil, false
	}
	return &limits, true
}

// DeleteWalletLimit deletes wallet limits
func (s Store) DeleteWalletLimit(ctx context.Context, address string) {
	s.kv(ctx).Delete(types.WalletLimitKey(address))
}

// IterateWalletLimits iterates over all wallet limits
func (s Store) IterateWalletLimits(ctx context.Context) []*types.WalletLimits {
	it := storetypes.KVStorePrefixIterator(s.kv(ctx), types.WalletLimitKeyPrefix)
	defer it.Close()

	limits := make([]*types.WalletLimits, 0, 64)
	for ; it.Valid(); it.Next() {
		var limit types.WalletLimits
		if err := json.Unmarshal(it.Value(), &limit); err == nil {
			limitCopy := limit
			limits = append(limits, &limitCopy)
		}
	}
	return limits
}

// ============================
// INCIDENT ID COUNTER
// ============================

// GetNextIncidentID retrieves and increments the incident ID counter
func (s Store) GetNextIncidentID(ctx context.Context) uint64 {
	bz := s.kv(ctx).Get(types.NextIncidentIDKey)
	if bz == nil {
		return 1
	}
	return binary.BigEndian.Uint64(bz)
}

// SetNextIncidentID stores the incident ID counter
func (s Store) SetNextIncidentID(ctx context.Context, id uint64) {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, id)
	s.kv(ctx).Set(types.NextIncidentIDKey, bz)
}

// ============================
// PARAMS MANAGEMENT
// ============================

// SetParams stores module parameters
func (s Store) SetParams(ctx context.Context, params *types.IncidentResponseParams) error {
	bz, err := json.Marshal(params)
	if err != nil {
		return err
	}
	s.kv(ctx).Set(types.ParamsKey, bz)
	return nil
}

// GetParams retrieves module parameters
func (s Store) GetParams(ctx context.Context) (*types.IncidentResponseParams, bool) {
	bz := s.kv(ctx).Get(types.ParamsKey)
	if bz == nil {
		return nil, false
	}
	var params types.IncidentResponseParams
	if err := json.Unmarshal(bz, &params); err != nil {
		return nil, false
	}
	return &params, true
}
