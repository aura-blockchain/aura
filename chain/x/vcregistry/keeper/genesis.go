// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"fmt"
	"sort"

	"github.com/aequitas/aura/chain/x/vcregistry/types"
)

// InitGenesis initializes the module state from genesis data
func (k *Keeper) InitGenesis(ctx context.Context, data types.GenesisState) error {
	// Validate genesis before importing
	if err := types.ValidateGenesisState(&data); err != nil {
		return fmt.Errorf("invalid genesis state: %w", err)
	}

	k.requireStore()

	// Set params
	if k.paramsStore != nil {
		if err := k.paramsStore.SetParams(data.Params); err != nil {
			return fmt.Errorf("failed to set params: %w", err)
		}
	}

	// Import VC records to KV store
	for _, vc := range data.VcRecords {
		if vc == nil {
			continue
		}
		k.store.setVCRecord(ctx, *vc)
		k.store.appendUserVC(ctx, vc.HolderAddress, vc.VcId)
	}

	// Import revocation records to KV store
	for _, revocation := range data.RevocationRecords {
		if revocation == nil {
			continue
		}
		k.store.setRevocationRecord(ctx, *revocation)
	}

	// Import revocation list to KV store
	if data.RevocationList != nil {
		k.store.setRevocationList(ctx, *data.RevocationList)
	}

	// Import DID documents to KV store
	for _, did := range data.DidDocuments {
		if did == nil {
			continue
		}
		k.store.setDIDDocument(ctx, *did)
		k.store.appendAddressDID(ctx, did.Controller, did.Did)
	}

	// Import VC policies to KV store
	for _, policy := range data.VcPolicies {
		if policy == nil {
			continue
		}
		k.store.setVCPolicy(ctx, *policy)
	}

	// Import mint counts for current day to KV store (deterministic ordering)
	dayTimestamp := k.getCurrentTime(ctx) / 86400
	if data.UserMintCounts != nil {
		mintAddrs := make([]string, 0, len(data.UserMintCounts))
		for addr := range data.UserMintCounts {
			if data.UserMintCounts[addr] != 0 {
				mintAddrs = append(mintAddrs, addr)
			}
		}
		sort.Strings(mintAddrs)
		for _, addr := range mintAddrs {
			count := data.UserMintCounts[addr]
			k.store.setMintCount(ctx, addr, dayTimestamp, count)
		}
	}

	// Import presentations to KV store
	// Indexes are built automatically when setPresentation is called,
	// so we don't need to import the exported indexes separately.
	for _, presentation := range data.Presentations {
		if presentation == nil {
			continue
		}
		k.store.setPresentation(ctx, *presentation)
		k.store.appendUserPresentation(ctx, presentation.HolderAddress, presentation.PresentationId)
	}

	// Validate that built user presentation indexes match exported indexes (deterministic ordering)
	// This ensures data integrity during genesis import
	if data.UserPresentationIndex != nil {
		indexAddrs := make([]string, 0, len(data.UserPresentationIndex))
		for addr := range data.UserPresentationIndex {
			indexAddrs = append(indexAddrs, addr)
		}
		sort.Strings(indexAddrs)
		for _, addr := range indexAddrs {
			exportedIDs := data.UserPresentationIndex[addr]
			if exportedIDs == nil || len(exportedIDs.Ids) == 0 {
				continue
			}

			// Get the index that was built from presentations
			builtIDs := k.store.listUserPresentations(ctx, addr)

			// Verify count matches
			if len(builtIDs) != len(exportedIDs.Ids) {
				return fmt.Errorf("user presentation index mismatch for %s: built %d entries, exported %d entries",
					addr, len(builtIDs), len(exportedIDs.Ids))
			}

			// Verify all exported IDs are present in built index
			builtMap := make(map[string]bool)
			for _, id := range builtIDs {
				builtMap[id] = true
			}
			for _, id := range exportedIDs.Ids {
				if !builtMap[id] {
					return fmt.Errorf("user presentation index mismatch for %s: exported ID %s not found in built index",
						addr, id)
				}
			}
		}
	}

	// Import attribute VCs to KV store
	// Indexes are built automatically when setAttributeVC is called,
	// so we don't need to import the exported indexes separately.
	for _, attrVC := range data.AttributeVcs {
		if attrVC == nil {
			continue
		}
		k.store.setAttributeVC(ctx, *attrVC)
		k.store.appendUserAttributeVC(ctx, attrVC.HolderAddress, attrVC.AttributeVcId)
	}

	// Validate that built user attribute indexes match exported indexes (deterministic ordering)
	// This ensures data integrity during genesis import
	if data.UserAttributeIndex != nil {
		attrAddrs := make([]string, 0, len(data.UserAttributeIndex))
		for addr := range data.UserAttributeIndex {
			attrAddrs = append(attrAddrs, addr)
		}
		sort.Strings(attrAddrs)
		for _, addr := range attrAddrs {
			exportedIDs := data.UserAttributeIndex[addr]
			if exportedIDs == nil || len(exportedIDs.Ids) == 0 {
				continue
			}

			// Get the index that was built from attribute VCs
			builtIDs := k.store.listUserAttributeVCs(ctx, addr)

			// Verify count matches
			if len(builtIDs) != len(exportedIDs.Ids) {
				return fmt.Errorf("user attribute index mismatch for %s: built %d entries, exported %d entries",
					addr, len(builtIDs), len(exportedIDs.Ids))
			}

			// Verify all exported IDs are present in built index
			builtMap := make(map[string]bool)
			for _, id := range builtIDs {
				builtMap[id] = true
			}
			for _, id := range exportedIDs.Ids {
				if !builtMap[id] {
					return fmt.Errorf("user attribute index mismatch for %s: exported ID %s not found in built index",
						addr, id)
				}
			}
		}
	}

	// Import disclosure policies to KV store
	for _, pol := range data.DisclosurePolicies {
		if pol == nil {
			continue
		}
		k.store.setDisclosurePolicy(ctx, *pol)
	}

	// Import disclosure requests to KV store
	for _, req := range data.DisclosureRequests {
		if req == nil {
			continue
		}
		k.store.setDisclosureRequest(ctx, *req)
	}

	// Import disclosure responses to KV store
	for _, resp := range data.DisclosureResponses {
		if resp == nil {
			continue
		}
		k.store.setDisclosureResponse(ctx, *resp)
	}

	// Import pending disclosure index to KV store (deterministic ordering)
	// Note: Unlike presentation and attribute indexes, the pending disclosure index
	// is NOT automatically built from disclosure requests, because not all requests
	// are pending (some have responses). Therefore, we must import it explicitly.
	if data.PendingDisclosureIndex != nil {
		pendingHolders := make([]string, 0, len(data.PendingDisclosureIndex))
		for holder := range data.PendingDisclosureIndex {
			pendingHolders = append(pendingHolders, holder)
		}
		sort.Strings(pendingHolders)
		for _, holder := range pendingHolders {
			exportedIDs := data.PendingDisclosureIndex[holder]
			if exportedIDs == nil || len(exportedIDs.Ids) == 0 {
				continue
			}

			// Import each pending disclosure entry
			for _, id := range exportedIDs.Ids {
				k.store.appendPendingDisclosure(ctx, holder, id)
			}

			// Validate that the imported index matches what we just built
			builtIDs := k.store.listPendingDisclosures(ctx, holder)
			if len(builtIDs) != len(exportedIDs.Ids) {
				return fmt.Errorf("pending disclosure index mismatch for %s: built %d entries, exported %d entries",
					holder, len(builtIDs), len(exportedIDs.Ids))
			}
		}
	}

	return nil
}

// ExportGenesis exports the current module state to genesis
func (k *Keeper) ExportGenesis(ctx context.Context) types.GenesisState {
	k.requireStore()

	// Export params
	params, _ := k.GetParams(ctx)

	// Export VC records from KV store
	vcRecords := make([]*types.VCRecord, 0)
	for _, rec := range k.store.iterateVCRecords(ctx) {
		recCopy := rec
		vcRecords = append(vcRecords, &recCopy)
	}

	// Export revocation records from KV store (deterministic ordering)
	// iterateRevocationRecords returns a map, so we must sort keys before iterating
	revocationRecords := make([]*types.RevocationRecord, 0)
	revocationData := k.store.iterateRevocationRecords(ctx)
	revocationVCIDs := make([]string, 0, len(revocationData))
	for vcID := range revocationData {
		revocationVCIDs = append(revocationVCIDs, vcID)
	}
	sort.Strings(revocationVCIDs)
	for _, vcID := range revocationVCIDs {
		rec := revocationData[vcID]
		revocationRecords = append(revocationRecords, &rec)
	}

	// Export revocation list from KV store
	revocationList := &types.RevocationList{
		MerkleRoot:        []byte{},
		TotalRevocations:  0,
		LastUpdatedHeight: 0,
		LastUpdated:       nil,
	}
	if list, ok := k.store.getRevocationList(ctx); ok {
		revocationList = &list
	}

	// Export DID documents from KV store
	didDocuments := make([]*types.DIDDocument, 0)
	for _, doc := range k.store.iterateDIDDocuments(ctx) {
		docCopy := doc
		didDocuments = append(didDocuments, &docCopy)
	}

	// Export VC policies from KV store
	vcPolicies := make([]*types.VCPolicy, 0)
	for _, policy := range k.store.iterateVCPolicies(ctx) {
		policyCopy := policy
		vcPolicies = append(vcPolicies, &policyCopy)
	}

	// Export user mint counts from KV store (current day only, deterministic ordering)
	userMintCounts := make(map[string]uint64)
	dayTimestamp := k.getCurrentTime(ctx) / 86400
	mintCountData := k.store.iterateMintCounts(ctx)
	mintCountAddrs := make([]string, 0, len(mintCountData))
	for addr := range mintCountData {
		mintCountAddrs = append(mintCountAddrs, addr)
	}
	sort.Strings(mintCountAddrs)
	for _, addr := range mintCountAddrs {
		counts := mintCountData[addr]
		if count, ok := counts[dayTimestamp]; ok {
			userMintCounts[addr] = count
		}
	}

	// Export presentations from KV store
	presentations := make([]*types.VCPresentation, 0)
	userPresentationIndex := make(map[string]*types.PresentationIds)
	for _, pres := range k.store.iteratePresentations(ctx) {
		presCopy := pres
		presentations = append(presentations, &presCopy)

		// Build user presentation index
		if userPresentationIndex[pres.HolderAddress] == nil {
			userPresentationIndex[pres.HolderAddress] = &types.PresentationIds{Ids: []string{}}
		}
		userPresentationIndex[pres.HolderAddress].Ids = append(
			userPresentationIndex[pres.HolderAddress].Ids,
			pres.PresentationId,
		)
	}

	// Export attribute VCs from KV store
	attributeVCs := make([]*types.AttributeVC, 0)
	userAttributeIndex := make(map[string]*types.AttributeVcIds)
	for _, avc := range k.store.iterateAttributeVCs(ctx) {
		avcCopy := avc
		attributeVCs = append(attributeVCs, &avcCopy)

		// Build user attribute index
		if userAttributeIndex[avc.HolderAddress] == nil {
			userAttributeIndex[avc.HolderAddress] = &types.AttributeVcIds{Ids: []string{}}
		}
		userAttributeIndex[avc.HolderAddress].Ids = append(
			userAttributeIndex[avc.HolderAddress].Ids,
			avc.AttributeVcId,
		)
	}

	// Export disclosure policies from KV store
	disclosurePolicies := make([]*types.DisclosurePolicy, 0)
	for _, pol := range k.store.iterateDisclosurePolicies(ctx) {
		polCopy := pol
		disclosurePolicies = append(disclosurePolicies, &polCopy)
	}

	// Export disclosure requests from KV store
	disclosureRequests := make([]*types.DisclosureRequest, 0)
	for _, req := range k.store.iterateDisclosureRequests(ctx) {
		reqCopy := req
		disclosureRequests = append(disclosureRequests, &reqCopy)
	}

	// Export disclosure responses from KV store
	disclosureResponses := make([]*types.DisclosureResponse, 0)
	for _, resp := range k.store.iterateDisclosureResponses(ctx) {
		respCopy := resp
		disclosureResponses = append(disclosureResponses, &respCopy)
	}

	// Export pending disclosure index from KV store (deterministic ordering)
	pendingDisclosures := make(map[string]*types.RequestIds)
	pendingData := k.store.iteratePendingDisclosures(ctx)
	pendingHolders := make([]string, 0, len(pendingData))
	for holder := range pendingData {
		pendingHolders = append(pendingHolders, holder)
	}
	sort.Strings(pendingHolders)
	for _, holder := range pendingHolders {
		ids := pendingData[holder]
		if len(ids) > 0 {
			pendingDisclosures[holder] = &types.RequestIds{Ids: ids}
		}
	}

	return types.GenesisState{
		Params:                 params,
		VcRecords:              vcRecords,
		RevocationRecords:      revocationRecords,
		RevocationList:         revocationList,
		DidDocuments:           didDocuments,
		VcPolicies:             vcPolicies,
		UserMintCounts:         userMintCounts,
		Presentations:          presentations,
		UserPresentationIndex:  userPresentationIndex,
		AttributeVcs:           attributeVCs,
		UserAttributeIndex:     userAttributeIndex,
		DisclosurePolicies:     disclosurePolicies,
		DisclosureRequests:     disclosureRequests,
		DisclosureResponses:    disclosureResponses,
		PendingDisclosureIndex: pendingDisclosures,
	}
}
