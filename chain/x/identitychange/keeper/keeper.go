package keeper

import (
	"errors"
	"fmt"
	"sync"

	"github.com/aequitas/aura/chain/x/identitychange/params"
	"github.com/aequitas/aura/chain/x/identitychange/types"
)

var errRequestsSuspended = errors.New("identity change requests are suspended")

type Keeper struct {
	mu          sync.RWMutex
	records     map[string]types.IdentityRecord
	requests    map[string]types.IdentityChangeRequest
	history     []types.IdentityChangeHistory
	paramsStore *params.Store
	suspended   bool
}

func NewKeeper(store *params.Store) *Keeper {
	if store == nil {
		store = params.NewStore(types.DefaultParams())
	}
	return &Keeper{
		records:     make(map[string]types.IdentityRecord),
		requests:    make(map[string]types.IdentityChangeRequest),
		paramsStore: store,
	}
}

func (k *Keeper) CreateRequest(request types.IdentityChangeRequest) (types.IdentityChangeRequest, error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.suspended {
		return types.IdentityChangeRequest{}, errRequestsSuspended
	}
	if request.RequestID == "" {
		return types.IdentityChangeRequest{}, errors.New("request id required")
	}
	count := k.countRequests(request.Requester)
	params := k.getParams()
	if int32(count) >= params.MaxRequestsPerWalletPerMonth {
		return types.IdentityChangeRequest{}, fmt.Errorf("request limit exceeded for %s", request.Requester)
	}
	request.Status = types.IdentityChangeStatusPendingVerification
	k.requests[request.RequestID] = request
	return request, nil
}

func (k *Keeper) SubmitProof(requestID string, assistant string, success bool, confidenceDelta int64, reason string) (types.IdentityChangeRequest, error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	request, ok := k.requests[requestID]
	if !ok {
		return types.IdentityChangeRequest{}, fmt.Errorf("request %s not found", requestID)
	}
	if request.Status != types.IdentityChangeStatusPendingVerification {
		return request, fmt.Errorf("request %s already processed", requestID)
	}
	request.Assistant = assistant
	request.VerdictHeight = 0
	if success {
		request.Status = types.IdentityChangeStatusReadyToApply
	} else {
		request.Status = types.IdentityChangeStatusRejected
		request.Reason = reason
	}
	k.requests[requestID] = request
	return request, nil
}
func (k *Keeper) ApplyChange(requestID string) (types.IdentityRecord, error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	request, ok := k.requests[requestID]
	if !ok {
		return types.IdentityRecord{}, fmt.Errorf("request %s missing", requestID)
	}
	if request.Status != types.IdentityChangeStatusReadyToApply {
		return types.IdentityRecord{}, fmt.Errorf("request %s not ready", requestID)
	}
	record := k.records[request.TargetDID]
	prevScore := record.ConfidenceScore
	params := k.getParams()
	if prevScore < params.MinConfidenceAfterChange {
		record.ConfidenceScore = params.MinConfidenceAfterChange
	}
	record.DID = request.TargetDID
	record.Owner = request.Requester
	record.MetadataHash = request.RequestMetaHash
	record.LatestIRVersion = request.IRID
	record.LastChangedHeight++
	record.Status = types.IdentityChangeStatusApplied
	k.records[request.TargetDID] = record
	history := types.IdentityChangeHistory{
		RequestID:           request.RequestID,
		TargetDID:           request.TargetDID,
		PrevConfidenceScore: prevScore,
		NewConfidenceScore:  record.ConfidenceScore,
		TransitionReason:    "applied",
		ChangedHeight:       record.LastChangedHeight,
	}
	k.history = append(k.history, history)
	request.Status = types.IdentityChangeStatusApplied
	k.requests[requestID] = request
	return record, nil
}

func (k *Keeper) RejectChange(requestID, reason string) (types.IdentityChangeRequest, error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	request, ok := k.requests[requestID]
	if !ok {
		return types.IdentityChangeRequest{}, fmt.Errorf("request %s missing", requestID)
	}
	if request.Status == types.IdentityChangeStatusApplied {
		return request, fmt.Errorf("cannot reject applied request %s", requestID)
	}
	request.Status = types.IdentityChangeStatusRejected
	request.Reason = reason
	k.requests[requestID] = request
	return request, nil
}

func (k *Keeper) SetSuspended(suspended bool) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.suspended = suspended
}

func (k *Keeper) countRequests(requester string) int {
	count := 0
	for _, req := range k.requests {
		if req.Requester == requester {
			count++
		}
	}
	return count
}

func (k *Keeper) GetIdentityRecord(did string) (types.IdentityRecord, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	record, ok := k.records[did]
	return record, ok
}

func (k *Keeper) GetRequest(requestID string) (types.IdentityChangeRequest, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	req, ok := k.requests[requestID]
	return req, ok
}

func (k *Keeper) ListHistory(did string) []types.IdentityChangeHistory {
	k.mu.RLock()
	defer k.mu.RUnlock()
	entries := make([]types.IdentityChangeHistory, 0, len(k.history))
	for _, item := range k.history {
		if item.TargetDID == did {
			entries = append(entries, item)
		}
	}
	return entries
}

func (k *Keeper) getParams() types.Params {
	if k.paramsStore != nil {
		return k.paramsStore.GetParams()
	}
	return types.DefaultParams()
}

func (k *Keeper) SetParams(params types.Params) error {
	if k.paramsStore == nil {
		return fmt.Errorf("params store not initialized")
	}
	return k.paramsStore.SetParams(params)
}
