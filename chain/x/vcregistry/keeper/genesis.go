package keeper

import (
	"context"
	"fmt"

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
	if data.Params != nil && k.paramsStore != nil {
		if err := k.paramsStore.SetParams(*data.Params); err != nil {
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

	// Import mint counts for current day to KV store
	dayTimestamp := k.getCurrentTime(ctx) / 86400
	if data.UserMintCounts != nil {
		for addr, count := range data.UserMintCounts {
			if count == 0 {
				continue
			}
			k.store.setMintCount(ctx, addr, dayTimestamp, count)
		}
	}

	// Import presentations to KV store
	for _, presentation := range data.Presentations {
		if presentation == nil {
			continue
		}
		k.store.setPresentation(ctx, *presentation)
		k.store.appendUserPresentation(ctx, presentation.HolderAddress, presentation.PresentationId)
	}

	// Import user presentation index to KV store
	if data.UserPresentationIndex != nil {
		for addr, presIDs := range data.UserPresentationIndex {
			if presIDs != nil {
				for _, id := range presIDs.Ids {
					k.store.appendUserPresentation(ctx, addr, id)
				}
			}
		}
	}

	// Import attribute VCs to KV store
	for _, attrVC := range data.AttributeVcs {
		if attrVC == nil {
			continue
		}
		k.store.setAttributeVC(ctx, *attrVC)
		k.store.appendUserAttributeVC(ctx, attrVC.HolderAddress, attrVC.AttributeVcId)
	}

	// Import user attribute index to KV store
	if data.UserAttributeIndex != nil {
		for addr, attrIDs := range data.UserAttributeIndex {
			if attrIDs != nil {
				for _, id := range attrIDs.Ids {
					k.store.appendUserAttributeVC(ctx, addr, id)
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

	// Import pending disclosure index to KV store
	if data.PendingDisclosureIndex != nil {
		for holder, ids := range data.PendingDisclosureIndex {
			if ids != nil {
				for _, id := range ids.Ids {
					k.store.appendPendingDisclosure(ctx, holder, id)
				}
			}
		}
	}

	return nil
}

// ExportGenesis exports the current module state to genesis
func (k *Keeper) ExportGenesis(ctx context.Context) types.GenesisState {
	k.requireStore()

	// Export params
	params := k.GetParams()

	// Export VC records from KV store
	vcRecords := make([]*types.VCRecord, 0)
	for _, rec := range k.store.iterateVCRecords(ctx) {
		recCopy := rec
		vcRecords = append(vcRecords, &recCopy)
	}

	// Export revocation records from KV store
	revocationRecords := make([]*types.RevocationRecord, 0)
	for _, rec := range k.store.iterateRevocationRecords(ctx) {
		recCopy := rec
		revocationRecords = append(revocationRecords, &recCopy)
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

	// Export user mint counts from KV store (current day only)
	userMintCounts := make(map[string]uint64)
	dayTimestamp := k.getCurrentTime(ctx) / 86400
	for addr, counts := range k.store.iterateMintCounts(ctx) {
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

	// Export pending disclosure index from KV store
	pendingDisclosures := make(map[string]*types.RequestIds)
	for holder, ids := range k.store.iteratePendingDisclosures(ctx) {
		if len(ids) > 0 {
			pendingDisclosures[holder] = &types.RequestIds{Ids: ids}
		}
	}

	return types.GenesisState{
		Params:                 &params,
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
