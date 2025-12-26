// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"bytes"
	"context"
	"encoding/binary"

	storetypes "cosmossdk.io/store/types"
	"github.com/aequitas/aura/chain/x/common/gasmetering"
	"github.com/aequitas/aura/chain/x/vcregistry/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/cosmos/gogoproto/proto"
)

// Store provides typed KV accessors for VC registry state.
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

// utility to append and copy slices to avoid aliasing
func concat(prefix []byte, tail []byte) []byte {
	buf := make([]byte, 0, len(prefix)+len(tail))
	buf = append(buf, prefix...)
	buf = append(buf, tail...)
	return buf
}

// helper to append newline-delimited strings
func appendLine(buf []byte, s string) []byte {
	buf = append(buf, s...)
	buf = append(buf, '\n')
	return buf
}

// parseMintKey extracts address and day timestamp from a mint count key.
func parseMintKey(key []byte) (string, int64, bool) {
	if !bytes.HasPrefix(key, types.UserMintCountKeyPrefix) {
		return "", 0, false
	}
	body := key[len(types.UserMintCountKeyPrefix):]
	if len(body) < 8 {
		return "", 0, false
	}
	addrBytes := body[:len(body)-8]
	dayBytes := body[len(body)-8:]
	day := int64(binary.BigEndian.Uint64(dayBytes))
	return string(addrBytes), day, true
}

func (s Store) setBinary(ctx context.Context, key []byte, msg proto.Message) {
	bz := s.cdc.MustMarshal(msg)
	s.kv(ctx).Set(key, bz)
}

func (s Store) getBinary(ctx context.Context, key []byte, out proto.Message) bool {
	bz := s.kv(ctx).Get(key)
	if bz == nil {
		return false
	}
	if err := s.cdc.Unmarshal(bz, out); err != nil {
		// Log error but return false to indicate failure
		// This is safer than panicking as corrupted state can be handled
		return false
	}
	return true
}

// VC records
func (s Store) setVCRecord(ctx context.Context, rec types.VCRecord) {
	s.setBinary(ctx, types.VCRecordKey(rec.VcId), &rec)
}

func (s Store) getVCRecord(ctx context.Context, vcID string) (types.VCRecord, bool) {
	var rec types.VCRecord
	if !s.getBinary(ctx, types.VCRecordKey(vcID), &rec) {
		return types.VCRecord{}, false
	}
	return rec, true
}

// iterateVCRecords returns all VC records keyed by ID
// Gas metered to prevent unbounded iteration attacks
func (s Store) iterateVCRecords(ctx context.Context) []types.VCRecord {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	gasConfig := gasmetering.DefaultGasConfig()

	// Charge base iteration cost
	sdkCtx.GasMeter().ConsumeGas(gasConfig.IterationBaseCost, "vc_records_iteration_base")

	it := storetypes.KVStorePrefixIterator(s.kv(ctx), types.VCRecordKeyPrefix)
	defer it.Close()

	recs := make([]types.VCRecord, 0, 64)
	count := uint32(0)
	maxResults := gasConfig.MaxIterationResults

	for ; it.Valid(); it.Next() {
		// Enforce iteration limit to prevent DoS
		if count >= maxResults {
			break
		}

		// Charge gas per iteration
		sdkCtx.GasMeter().ConsumeGas(gasConfig.StoreIterationCost, "vc_records_iteration")

		var rec types.VCRecord
		if err := s.cdc.Unmarshal(it.Value(), &rec); err == nil {
			// Charge gas for unmarshal based on data size
			sdkCtx.GasMeter().ConsumeGas(
				uint64(len(it.Value()))*gasConfig.UnmarshalCostPerByte,
				"vc_records_unmarshal",
			)
			recs = append(recs, rec)
			count++
		}
	}
	return recs
}

// Presentations
func (s Store) setPresentation(ctx context.Context, pres types.VCPresentation) {
	s.setBinary(ctx, types.PresentationKey(pres.PresentationId), &pres)
}

func (s Store) getPresentation(ctx context.Context, presID string) (types.VCPresentation, bool) {
	var pres types.VCPresentation
	if !s.getBinary(ctx, types.PresentationKey(presID), &pres) {
		return types.VCPresentation{}, false
	}
	return pres, true
}

func (s Store) iteratePresentations(ctx context.Context) []types.VCPresentation {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	gasConfig := gasmetering.DefaultGasConfig()
	sdkCtx.GasMeter().ConsumeGas(gasConfig.IterationBaseCost, "presentations_iteration_base")

	it := storetypes.KVStorePrefixIterator(s.kv(ctx), types.PresentationKeyPrefix)
	defer it.Close()

	pres := make([]types.VCPresentation, 0, 64)
	count := uint32(0)
	for ; it.Valid(); it.Next() {
		if count >= gasConfig.MaxIterationResults {
			break
		}
		sdkCtx.GasMeter().ConsumeGas(gasConfig.StoreIterationCost, "presentations_iteration")

		var p types.VCPresentation
		if err := s.cdc.Unmarshal(it.Value(), &p); err == nil {
			sdkCtx.GasMeter().ConsumeGas(
				uint64(len(it.Value()))*gasConfig.UnmarshalCostPerByte,
				"presentations_unmarshal",
			)
			pres = append(pres, p)
			count++
		}
	}
	return pres
}

// User presentation index
func (s Store) appendUserPresentation(ctx context.Context, address, presID string) {
	key := types.UserPresentationIndexKey(address)
	buf := appendLine(s.kv(ctx).Get(key), presID)
	s.kv(ctx).Set(key, buf)
}

func (s Store) listUserPresentations(ctx context.Context, address string) []string {
	bz := s.kv(ctx).Get(types.UserPresentationIndexKey(address))
	if len(bz) == 0 {
		return nil
	}
	parts := bytes.Split(bz, []byte("\n"))
	res := make([]string, 0, len(parts))
	for _, p := range parts {
		if len(p) == 0 {
			continue
		}
		res = append(res, string(p))
	}
	return res
}

// Pending disclosure requests index
func (s Store) appendPendingDisclosure(ctx context.Context, address, requestID string) {
	key := types.PendingDisclosureRequestKey(address, requestID)
	s.kv(ctx).Set(key, []byte{1})
}

func (s Store) listPendingDisclosures(ctx context.Context, address string) []string {
	it := storetypes.KVStorePrefixIterator(s.kv(ctx), concat(types.PendingDisclosureRequestKeyPrefix, []byte(address)))
	defer it.Close()
	reqs := make([]string, 0, 64)
	for ; it.Valid(); it.Next() {
		suffix := bytes.TrimPrefix(it.Key(), concat(types.PendingDisclosureRequestKeyPrefix, []byte(address)))
		suffix = bytes.TrimPrefix(suffix, []byte{':'})
		reqs = append(reqs, string(suffix))
	}
	return reqs
}

func (s Store) deletePendingDisclosure(ctx context.Context, address, requestID string) {
	key := types.PendingDisclosureRequestKey(address, requestID)
	s.kv(ctx).Delete(key)
}

// iteratePendingDisclosures returns map[holder][]requestID for pending items.
func (s Store) iteratePendingDisclosures(ctx context.Context) map[string][]string {
	res := make(map[string][]string)
	it := storetypes.KVStorePrefixIterator(s.kv(ctx), types.PendingDisclosureRequestKeyPrefix)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		suffix := bytes.TrimPrefix(it.Key(), types.PendingDisclosureRequestKeyPrefix)
		parts := bytes.SplitN(suffix, []byte{':'}, 2)
		if len(parts) != 2 {
			continue
		}
		holder := string(parts[0])
		reqID := string(parts[1])
		res[holder] = append(res[holder], reqID)
	}
	return res
}

// Attribute VCs
func (s Store) setAttributeVC(ctx context.Context, avc types.AttributeVC) {
	s.setBinary(ctx, types.AttributeVCKey(avc.AttributeVcId), &avc)
}

func (s Store) getAttributeVC(ctx context.Context, avcID string) (types.AttributeVC, bool) {
	var avc types.AttributeVC
	if !s.getBinary(ctx, types.AttributeVCKey(avcID), &avc) {
		return types.AttributeVC{}, false
	}
	return avc, true
}

func (s Store) iterateAttributeVCs(ctx context.Context) []types.AttributeVC {
	it := storetypes.KVStorePrefixIterator(s.kv(ctx), types.AttributeVCKeyPrefix)
	defer it.Close()
	avcs := make([]types.AttributeVC, 0, 64)
	for ; it.Valid(); it.Next() {
		var avc types.AttributeVC
		if err := s.cdc.Unmarshal(it.Value(), &avc); err == nil {
			avcs = append(avcs, avc)
		}
	}
	return avcs
}

// User attribute VC index
//
// ATOMICITY NOTE: appendUserAttributeVC performs a read-modify-write on the user's attribute VC index.
// The operation is atomic within a single transaction because Cosmos SDK processes
// transactions serially and the KVStore changes are committed atomically.
// However, if this function is called after storing the attribute VC and then fails,
// the attribute VC exists but is not indexed. The CreateAttributeVC function in keeper.go
// handles both operations together to maintain consistency.
func (s Store) appendUserAttributeVC(ctx context.Context, address, avcID string) {
	key := types.UserAttributeVCIndexKey(address)
	buf := appendLine(s.kv(ctx).Get(key), avcID)
	s.kv(ctx).Set(key, buf)
}

func (s Store) listUserAttributeVCs(ctx context.Context, address string) []string {
	bz := s.kv(ctx).Get(types.UserAttributeVCIndexKey(address))
	if len(bz) == 0 {
		return nil
	}
	parts := bytes.Split(bz, []byte("\n"))
	res := make([]string, 0, len(parts))
	for _, p := range parts {
		if len(p) == 0 {
			continue
		}
		res = append(res, string(p))
	}
	return res
}

// Disclosure policy
func (s Store) setDisclosurePolicy(ctx context.Context, policy types.DisclosurePolicy) {
	s.setBinary(ctx, types.DisclosurePolicyKey(policy.HolderAddress), &policy)
}

func (s Store) getDisclosurePolicy(ctx context.Context, holder string) (types.DisclosurePolicy, bool) {
	var pol types.DisclosurePolicy
	if !s.getBinary(ctx, types.DisclosurePolicyKey(holder), &pol) {
		return types.DisclosurePolicy{}, false
	}
	return pol, true
}

func (s Store) iterateDisclosurePolicies(ctx context.Context) []types.DisclosurePolicy {
	it := storetypes.KVStorePrefixIterator(s.kv(ctx), types.DisclosurePolicyKeyPrefix)
	defer it.Close()
	res := make([]types.DisclosurePolicy, 0, 64)
	for ; it.Valid(); it.Next() {
		var pol types.DisclosurePolicy
		if err := s.cdc.Unmarshal(it.Value(), &pol); err == nil {
			res = append(res, pol)
		}
	}
	return res
}

// Disclosure requests
func (s Store) setDisclosureRequest(ctx context.Context, req types.DisclosureRequest) {
	s.setBinary(ctx, types.DisclosureRequestKey(req.RequestId), &req)
}

func (s Store) getDisclosureRequest(ctx context.Context, id string) (types.DisclosureRequest, bool) {
	var req types.DisclosureRequest
	if !s.getBinary(ctx, types.DisclosureRequestKey(id), &req) {
		return types.DisclosureRequest{}, false
	}
	return req, true
}

func (s Store) iterateDisclosureRequests(ctx context.Context) []types.DisclosureRequest {
	it := storetypes.KVStorePrefixIterator(s.kv(ctx), types.DisclosureRequestKeyPrefix)
	defer it.Close()
	res := make([]types.DisclosureRequest, 0, 64)
	for ; it.Valid(); it.Next() {
		var req types.DisclosureRequest
		if err := s.cdc.Unmarshal(it.Value(), &req); err == nil {
			res = append(res, req)
		}
	}
	return res
}

// Disclosure responses
func (s Store) setDisclosureResponse(ctx context.Context, resp types.DisclosureResponse) {
	s.setBinary(ctx, types.DisclosureResponseKey(resp.RequestId), &resp)
}

func (s Store) getDisclosureResponse(ctx context.Context, id string) (types.DisclosureResponse, bool) {
	var resp types.DisclosureResponse
	if !s.getBinary(ctx, types.DisclosureResponseKey(id), &resp) {
		return types.DisclosureResponse{}, false
	}
	return resp, true
}

func (s Store) iterateDisclosureResponses(ctx context.Context) []types.DisclosureResponse {
	it := storetypes.KVStorePrefixIterator(s.kv(ctx), types.DisclosureResponseKeyPrefix)
	defer it.Close()
	res := make([]types.DisclosureResponse, 0, 64)
	for ; it.Valid(); it.Next() {
		var resp types.DisclosureResponse
		if err := s.cdc.Unmarshal(it.Value(), &resp); err == nil {
			res = append(res, resp)
		}
	}
	return res
}

// User VC index
//
// ATOMICITY NOTE: appendUserVC performs a read-modify-write on the user's VC index.
// The operation is atomic within a single transaction because Cosmos SDK processes
// transactions serially and the KVStore changes are committed atomically.
// However, if this function is called after storing the VC record and then fails,
// the VC exists but is not indexed. Callers should use SetVCRecord which handles
// both operations to maintain consistency.
func (s Store) appendUserVC(ctx context.Context, address, vcID string) {
	key := types.UserVCIndexKey(address)
	// store as length-prefixed list of vcIDs (simple append)
	existing := s.kv(ctx).Get(key)
	buf := appendLine(existing, vcID)
	s.kv(ctx).Set(key, buf)
}

func (s Store) listUserVCs(ctx context.Context, address string) []string {
	key := types.UserVCIndexKey(address)
	bz := s.kv(ctx).Get(key)
	if len(bz) == 0 {
		return nil
	}
	parts := bytes.Split(bz, []byte("\n"))
	res := make([]string, 0, len(parts))
	for _, p := range parts {
		if len(p) == 0 {
			continue
		}
		res = append(res, string(p))
	}
	return res
}

// Revocation records
func (s Store) setRevocationRecord(ctx context.Context, rec types.RevocationRecord) {
	s.setBinary(ctx, types.RevocationRecordKey(rec.VcId), &rec)
}

func (s Store) getRevocationRecord(ctx context.Context, vcID string) (types.RevocationRecord, bool) {
	var rec types.RevocationRecord
	if !s.getBinary(ctx, types.RevocationRecordKey(vcID), &rec) {
		return types.RevocationRecord{}, false
	}
	return rec, true
}

func (s Store) iterateRevocationRecords(ctx context.Context) map[string]types.RevocationRecord {
	records := make(map[string]types.RevocationRecord)
	it := storetypes.KVStorePrefixIterator(s.kv(ctx), types.RevocationRecordKeyPrefix)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		var rec types.RevocationRecord
		if err := s.cdc.Unmarshal(it.Value(), &rec); err == nil {
			records[rec.VcId] = rec
		}
	}
	return records
}

// Revocation list
func (s Store) setRevocationList(ctx context.Context, list types.RevocationList) {
	s.setBinary(ctx, types.RevocationListKey, &list)
}

func (s Store) getRevocationList(ctx context.Context) (types.RevocationList, bool) {
	var list types.RevocationList
	if !s.getBinary(ctx, types.RevocationListKey, &list) {
		return types.RevocationList{}, false
	}
	return list, true
}

// VC policies
func (s Store) setVCPolicy(ctx context.Context, policy types.VCPolicy) {
	s.setBinary(ctx, types.VCPolicyKey(policy.VcTypeName), &policy)
}

func (s Store) getVCPolicy(ctx context.Context, vcTypeName string) (types.VCPolicy, bool) {
	var policy types.VCPolicy
	if !s.getBinary(ctx, types.VCPolicyKey(vcTypeName), &policy) {
		return types.VCPolicy{}, false
	}
	return policy, true
}

func (s Store) iterateVCPolicies(ctx context.Context) []types.VCPolicy {
	policies := make([]types.VCPolicy, 0, 64)
	it := storetypes.KVStorePrefixIterator(s.kv(ctx), types.VCPolicyKeyPrefix)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		var policy types.VCPolicy
		if err := s.cdc.Unmarshal(it.Value(), &policy); err == nil {
			policies = append(policies, policy)
		}
	}
	return policies
}

// DID documents
func (s Store) setDIDDocument(ctx context.Context, doc types.DIDDocument) {
	s.setBinary(ctx, types.DIDDocumentKey(doc.Did), &doc)
}

func (s Store) getDIDDocument(ctx context.Context, did string) (types.DIDDocument, bool) {
	var doc types.DIDDocument
	if !s.getBinary(ctx, types.DIDDocumentKey(did), &doc) {
		return types.DIDDocument{}, false
	}
	return doc, true
}

// DID address index (controller -> DIDs)
//
// ATOMICITY NOTE: appendAddressDID performs a read-modify-write on the address-to-DID index.
// The operation is atomic within a single transaction because Cosmos SDK processes
// transactions serially and the KVStore changes are committed atomically.
// However, if this function is called after storing the DID document and then fails,
// the DID exists but is not indexed. The RegisterDID function in keeper.go handles
// both operations together to maintain consistency.
func (s Store) appendAddressDID(ctx context.Context, addr, did string) {
	key := types.AddressToDIDIndexKey(addr)
	existing := s.kv(ctx).Get(key)
	buf := appendLine(existing, did)
	s.kv(ctx).Set(key, buf)
}

func (s Store) listAddressDIDs(ctx context.Context, addr string) []string {
	key := types.AddressToDIDIndexKey(addr)
	bz := s.kv(ctx).Get(key)
	if len(bz) == 0 {
		return nil
	}
	parts := bytes.Split(bz, []byte("\n"))
	res := make([]string, 0, len(parts))
	for _, p := range parts {
		if len(p) == 0 {
			continue
		}
		res = append(res, string(p))
	}
	return res
}

// iterateDIDDocuments returns all DID documents
func (s Store) iterateDIDDocuments(ctx context.Context) []types.DIDDocument {
	it := storetypes.KVStorePrefixIterator(s.kv(ctx), types.DIDDocumentKeyPrefix)
	defer it.Close()
	docs := make([]types.DIDDocument, 0, 64)
	for ; it.Valid(); it.Next() {
		var doc types.DIDDocument
		if err := s.cdc.Unmarshal(it.Value(), &doc); err == nil {
			docs = append(docs, doc)
		}
	}
	return docs
}

// Mint counts
func (s Store) setMintCount(ctx context.Context, address string, dayTimestamp int64, count uint64) {
	key := types.UserMintCountKey(address, dayTimestamp)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, count)
	s.kv(ctx).Set(key, buf)
}

func (s Store) getMintCount(ctx context.Context, address string, dayTimestamp int64) (uint64, bool) {
	key := types.UserMintCountKey(address, dayTimestamp)
	bz := s.kv(ctx).Get(key)
	if bz == nil {
		return 0, false
	}
	return binary.BigEndian.Uint64(bz), true
}

// iterateMintCounts returns map[address]map[day]count
// NOTE: This function returns a Go map which has non-deterministic iteration order.
// Do NOT use this for consensus-critical operations. Use cleanupOldMintCountsDeterministic instead.
func (s Store) iterateMintCounts(ctx context.Context) map[string]map[int64]uint64 {
	res := make(map[string]map[int64]uint64)
	it := storetypes.KVStorePrefixIterator(s.kv(ctx), types.UserMintCountKeyPrefix)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		addr, day, ok := parseMintKey(it.Key())
		if !ok {
			continue
		}
		count := binary.BigEndian.Uint64(it.Value())
		if res[addr] == nil {
			res[addr] = make(map[int64]uint64)
		}
		res[addr][day] = count
	}
	return res
}

// cleanupOldMintCountsDeterministic removes mint count entries older than the cutoff day.
// This function is consensus-safe: it uses the KVStore iterator which returns keys in
// deterministic lexicographic order, and collects keys to delete before modifying state
// (required because modifying during iteration is undefined behavior).
func (s Store) cleanupOldMintCountsDeterministic(ctx context.Context, cutoffDay int64) {
	// First pass: collect keys to delete in deterministic order
	keysToDelete := make([][]byte, 0)
	it := storetypes.KVStorePrefixIterator(s.kv(ctx), types.UserMintCountKeyPrefix)
	defer it.Close()

	for ; it.Valid(); it.Next() {
		_, day, ok := parseMintKey(it.Key())
		if !ok {
			continue
		}
		if day < cutoffDay {
			// Make a copy of the key since iterator keys are reused
			keyCopy := make([]byte, len(it.Key()))
			copy(keyCopy, it.Key())
			keysToDelete = append(keysToDelete, keyCopy)
		}
	}

	// Second pass: delete in the same deterministic order we collected
	store := s.kv(ctx)
	for _, key := range keysToDelete {
		store.Delete(key)
	}
}

// Metadata storage for generic key-value pairs
func (s Store) setMetadata(ctx context.Context, key string, value string) {
	fullKey := concat([]byte("metadata:"), []byte(key))
	s.kv(ctx).Set(fullKey, []byte(value))
}

func (s Store) getMetadata(ctx context.Context, key string) (string, bool) {
	fullKey := concat([]byte("metadata:"), []byte(key))
	bz := s.kv(ctx).Get(fullKey)
	if bz == nil {
		return "", false
	}
	return string(bz), true
}

// Remove a VC from user's list
func (s Store) removeUserVC(ctx context.Context, holderAddress string, vcID string) {
	existing := s.listUserVCs(ctx, holderAddress)
	if len(existing) == 0 {
		return
	}

	// Filter out the VC to remove
	filtered := make([]string, 0, len(existing))
	for _, id := range existing {
		if id != vcID {
			filtered = append(filtered, id)
		}
	}

	// Store filtered list
	key := types.UserVCIndexKey(holderAddress)
	if len(filtered) == 0 {
		s.kv(ctx).Delete(key)
	} else {
		var buf []byte
		for _, id := range filtered {
			buf = appendLine(buf, id)
		}
		s.kv(ctx).Set(key, buf)
	}
}
