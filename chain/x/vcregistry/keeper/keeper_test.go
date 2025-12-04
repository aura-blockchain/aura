package keeper

import (
	"context"
	"testing"
	"time"

	"github.com/aequitas/aura/chain/x/vcregistry/params"
	"github.com/aequitas/aura/chain/x/vcregistry/types"
	vcregistrypb "github.com/aequitas/aura/proto/aura/vcregistry/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ============================
// TEST: NewKeeper
// ============================

func TestNewKeeper(t *testing.T) {
	tests := []struct {
		name        string
		paramsStore *params.Store
	}{
		{
			name:        "with nil params store",
			paramsStore: nil,
		},
		{
			name:        "with initialized params store",
			paramsStore: params.NewStore(*types.DefaultParams()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keeper := NewKeeper(tt.paramsStore, "authority")

			if keeper == nil {
				t.Fatal("expected non-nil keeper")
			}

			// If no params store provided, should have default
			if tt.paramsStore == nil && keeper.paramsStore == nil {
				t.Error("expected paramsStore to be initialized with defaults")
			}

			// Keeper uses KV store for all state (no in-memory fields)
			// KV persistence is tested in keeper_kv_persistence_test.go
			if keeper.authority != "authority" {
				t.Error("expected authority to be set correctly")
			}
		})
	}
}

// ============================
// TEST: SetGetVCRecord
// ============================

func TestSetGetVCRecord(t *testing.T) {
	t.Skip("This test requires refactoring to use proper SDK context - see keeper_kv_persistence_test.go for examples")
	tests := []struct {
		name      string
		vcRecord  *types.VCRecord
		shouldErr bool
		errType   error
	}{
		{
			name: "valid VC record",
			vcRecord: &types.VCRecord{
				VcId:          "vc:test123",
				VcType:        vcregistrypb.VCType_VC_TYPE_CUSTOM,
				HolderAddress: "aura1holder123",
				HolderDid:     "did:aura:test123",
				Status:        types.VCStatus_VC_STATUS_ACTIVE,
				IssuedAt:      timestamppb.Now(),
				IssuedHeight:  100,
			},
			shouldErr: false,
		},
		{
			name: "missing VC ID",
			vcRecord: &types.VCRecord{
				VcId:          "",
				VcType:        vcregistrypb.VCType_VC_TYPE_CUSTOM,
				HolderAddress: "aura1holder123",
				HolderDid:     "did:aura:test123",
				Status:        types.VCStatus_VC_STATUS_ACTIVE,
				IssuedAt:      timestamppb.Now(),
			},
			shouldErr: true,
			errType:   types.ErrInvalidVCID,
		},
		{
			name: "missing holder address",
			vcRecord: &types.VCRecord{
				VcId:          "vc:test456",
				VcType:        vcregistrypb.VCType_VC_TYPE_CUSTOM,
				HolderAddress: "",
				HolderDid:     "did:aura:test456",
				Status:        types.VCStatus_VC_STATUS_ACTIVE,
				IssuedAt:      timestamppb.Now(),
			},
			shouldErr: true,
			errType:   types.ErrInvalidHolderAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keeper := NewKeeper(nil, "authority")

			// Test SetVCRecord
			err := keeper.SetVCRecord(context.Background(), *tt.vcRecord)
			if (err != nil) != tt.shouldErr {
				t.Fatalf("SetVCRecord error: expected err=%v, got %v", tt.shouldErr, err)
			}

			if tt.shouldErr {
				if err != tt.errType {
					t.Errorf("expected error %v, got %v", tt.errType, err)
				}
				return
			}

			// Test GetVCRecord
			retrieved, ok := keeper.GetVCRecord(context.Background(), tt.vcRecord.VcId)
			if !ok {
				t.Fatal("expected to find VC record")
			}

			if retrieved.VcId != tt.vcRecord.VcId {
				t.Errorf("expected VC ID %s, got %s", tt.vcRecord.VcId, retrieved.VcId)
			}

			if retrieved.HolderAddress != tt.vcRecord.HolderAddress {
				t.Errorf("expected holder %s, got %s", tt.vcRecord.HolderAddress, retrieved.HolderAddress)
			}

			if retrieved.Status != tt.vcRecord.Status {
				t.Errorf("expected status %v, got %v", tt.vcRecord.Status, retrieved.Status)
			}
		})
	}
}

// ============================
// TEST: ListUserVCs
// ============================

func TestListUserVCs(t *testing.T) {
	t.Skip("This test requires refactoring to use proper SDK context - see keeper_kv_persistence_test.go for examples")
	keeper := NewKeeper(nil, "authority")
	holderAddr := "aura1testuser"

	// Create multiple VCs with different types and statuses
	vcs := []*types.VCRecord{
		{
			VcId:          "vc:active1",
			VcType:        vcregistrypb.VCType_VC_TYPE_CUSTOM,
			HolderAddress: holderAddr,
			HolderDid:     "did:aura:holder1",
			Status:        types.VCStatus_VC_STATUS_ACTIVE,
			IssuedAt:      timestamppb.Now(),
		},
		{
			VcId:          "vc:active2",
			VcType:        vcregistrypb.VCType_VC_TYPE_CUSTOM,
			HolderAddress: holderAddr,
			HolderDid:     "did:aura:holder1",
			Status:        types.VCStatus_VC_STATUS_ACTIVE,
			IssuedAt:      timestamppb.Now(),
		},
		{
			VcId:          "vc:revoked1",
			VcType:        vcregistrypb.VCType_VC_TYPE_CUSTOM,
			HolderAddress: holderAddr,
			HolderDid:     "did:aura:holder1",
			Status:        types.VCStatus_VC_STATUS_REVOKED,
			IssuedAt:      timestamppb.Now(),
		},
		{
			VcId:          "vc:expired1",
			VcType:        vcregistrypb.VCType_VC_TYPE_CUSTOM,
			HolderAddress: holderAddr,
			HolderDid:     "did:aura:holder1",
			Status:        types.VCStatus_VC_STATUS_EXPIRED,
			IssuedAt:      timestamppb.Now(),
		},
	}

	// Store all VCs
	for _, vc := range vcs {
		err := keeper.SetVCRecord(context.Background(), *vc)
		if err != nil {
			t.Fatalf("failed to set VC record: %v", err)
		}
	}

	tests := []struct {
		name       string
		statusFilt types.VCStatus
		typeFilt   types.VCType
		expected   int
	}{
		{
			name:       "list all VCs (no filters)",
			statusFilt: types.VCStatus_VC_STATUS_UNSPECIFIED,
			typeFilt:   types.VCType_VC_TYPE_UNSPECIFIED,
			expected:   4,
		},
		{
			name:       "filter by active status",
			statusFilt: types.VCStatus_VC_STATUS_ACTIVE,
			typeFilt:   types.VCType_VC_TYPE_UNSPECIFIED,
			expected:   2,
		},
		{
			name:       "filter by revoked status",
			statusFilt: types.VCStatus_VC_STATUS_REVOKED,
			typeFilt:   types.VCType_VC_TYPE_UNSPECIFIED,
			expected:   1,
		},
		{
			name:       "filter by expired status",
			statusFilt: types.VCStatus_VC_STATUS_EXPIRED,
			typeFilt:   types.VCType_VC_TYPE_UNSPECIFIED,
			expected:   1,
		},
		{
			name:       "list VCs for unknown user",
			statusFilt: types.VCStatus_VC_STATUS_UNSPECIFIED,
			typeFilt:   types.VCType_VC_TYPE_UNSPECIFIED,
			expected:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr := holderAddr
			if tt.name == "list VCs for unknown user" {
				addr = "aura1unknown"
			}

			result := keeper.ListUserVCs(context.Background(), addr, tt.statusFilt, tt.typeFilt)
			if len(result) != tt.expected {
				t.Errorf("expected %d VCs, got %d", tt.expected, len(result))
			}
		})
	}
}

// ============================
// TEST: CheckVCStatus
// ============================

func TestCheckVCStatus(t *testing.T) {
	t.Skip("This test requires refactoring to use proper SDK context - see keeper_kv_persistence_test.go for examples")
	keeper := NewKeeper(nil, "authority")
	currentTime := time.Now().Unix()
	keeper.SetCurrentTime(currentTime)

	tests := []struct {
		name           string
		vcID           string
		vcRecord       *vcregistrypb.VCRecord
		shouldErr      bool
		expectedValid  bool
		expectedStatus types.VCStatus
	}{
		{
			name:      "VC not found",
			vcID:      "vc:nonexistent",
			vcRecord:  nil,
			shouldErr: true,
		},
		{
			name: "active VC",
			vcID: "vc:active",
			vcRecord: &vcregistrypb.VCRecord{
				VcId:          "vc:active",
				HolderAddress: "aura1holder",
				HolderDid:     "did:aura:holder",
				Status:        types.VCStatus_VC_STATUS_ACTIVE,
				IssuedAt:      timestamppb.New(time.Unix(currentTime-3600, 0)),
				ExpiresAt:     timestamppb.New(time.Unix(currentTime+86400, 0)), // Expires tomorrow
			},
			shouldErr:      false,
			expectedValid:  true,
			expectedStatus: types.VCStatus_VC_STATUS_ACTIVE,
		},
		{
			name: "expired VC",
			vcID: "vc:expired",
			vcRecord: &vcregistrypb.VCRecord{
				VcId:          "vc:expired",
				HolderAddress: "aura1holder",
				HolderDid:     "did:aura:holder",
				Status:        types.VCStatus_VC_STATUS_ACTIVE,
				IssuedAt:      timestamppb.New(time.Unix(currentTime-86400*2, 0)),
				ExpiresAt:     timestamppb.New(time.Unix(currentTime-3600, 0)), // Expired 1 hour ago
			},
			shouldErr:      false,
			expectedValid:  false,
			expectedStatus: types.VCStatus_VC_STATUS_EXPIRED,
		},
		{
			name: "revoked VC",
			vcID: "vc:revoked",
			vcRecord: &vcregistrypb.VCRecord{
				VcId:          "vc:revoked",
				HolderAddress: "aura1holder",
				HolderDid:     "did:aura:holder",
				Status:        types.VCStatus_VC_STATUS_REVOKED,
				IssuedAt:      timestamppb.New(time.Unix(currentTime-3600, 0)),
			},
			shouldErr:      false,
			expectedValid:  false,
			expectedStatus: types.VCStatus_VC_STATUS_REVOKED,
		},
		{
			name: "suspended VC",
			vcID: "vc:suspended",
			vcRecord: &vcregistrypb.VCRecord{
				VcId:          "vc:suspended",
				HolderAddress: "aura1holder",
				HolderDid:     "did:aura:holder",
				Status:        types.VCStatus_VC_STATUS_SUSPENDED,
				IssuedAt:      timestamppb.New(time.Unix(currentTime-3600, 0)),
			},
			shouldErr:      false,
			expectedValid:  false,
			expectedStatus: types.VCStatus_VC_STATUS_SUSPENDED,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.vcRecord != nil {
				// Convert vcregistrypb.VCRecord to types.VCRecord
				typesRecord := types.VCRecord{
					VcId:            tt.vcRecord.VcId,
					VcType:          tt.vcRecord.VcType,
					VcTypeCustom:    tt.vcRecord.VcTypeCustom,
					HolderAddress:   tt.vcRecord.HolderAddress,
					HolderDid:       tt.vcRecord.HolderDid,
					IssuerAssistant: tt.vcRecord.IssuerAssistant,
					Status:          tt.vcRecord.Status,
					IssuedAt:        tt.vcRecord.IssuedAt,
					ExpiresAt:       tt.vcRecord.ExpiresAt,
					IssuedHeight:    tt.vcRecord.IssuedHeight,
				}
				keeper.SetVCRecord(context.Background(), typesRecord)
			}

			status, valid, err := keeper.CheckVCStatus(context.Background(), tt.vcID)

			if (err != nil) != tt.shouldErr {
				t.Fatalf("expected err=%v, got %v", tt.shouldErr, err)
			}

			if !tt.shouldErr {
				if status != tt.expectedStatus {
					t.Errorf("expected status %v, got %v", tt.expectedStatus, status)
				}
				if valid != tt.expectedValid {
					t.Errorf("expected valid=%v, got %v", tt.expectedValid, valid)
				}
			}
		})
	}
}

// ============================
// TEST: RevokeVC
// ============================

func TestRevokeVC(t *testing.T) {
	t.Skip("This test requires refactoring to use proper SDK context - see keeper_kv_persistence_test.go for examples")
	keeper := NewKeeper(nil, "authority")
	currentTime := time.Now().Unix()
	keeper.SetCurrentTime(currentTime)
	keeper.SetCurrentHeight(1000)

	tests := []struct {
		name      string
		vcID      string
		vcRecord  *vcregistrypb.VCRecord
		reason    types.RevocationReason
		revoker   string
		evidence  string
		shouldErr bool
		errType   error
	}{
		{
			name:      "revoke non-existent VC",
			vcID:      "vc:nonexistent",
			shouldErr: true,
			errType:   types.ErrVCNotFound,
		},
		{
			name: "successfully revoke active VC",
			vcID: "vc:active",
			vcRecord: &vcregistrypb.VCRecord{
				VcId:          "vc:active",
				HolderAddress: "aura1holder",
				HolderDid:     "did:aura:holder",
				Status:        types.VCStatus_VC_STATUS_ACTIVE,
				IssuedAt:      timestamppb.New(time.Unix(currentTime-3600, 0)),
			},
			reason:    types.RevocationReason_REVOCATION_REASON_USER_REQUEST,
			revoker:   "aura1revoker",
			evidence:  "user requested revocation",
			shouldErr: false,
		},
		{
			name: "revoke already revoked VC",
			vcID: "vc:revoked",
			vcRecord: &vcregistrypb.VCRecord{
				VcId:          "vc:revoked",
				HolderAddress: "aura1holder",
				HolderDid:     "did:aura:holder",
				Status:        types.VCStatus_VC_STATUS_REVOKED,
				IssuedAt:      timestamppb.New(time.Unix(currentTime-3600, 0)),
			},
			reason:    types.RevocationReason_REVOCATION_REASON_USER_REQUEST,
			revoker:   "aura1revoker",
			evidence:  "duplicate revocation attempt",
			shouldErr: true,
			errType:   types.ErrVCAlreadyRevoked,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset keeper for each test
			keeper := NewKeeper(nil, "authority")
			keeper.SetCurrentTime(currentTime)
			keeper.SetCurrentHeight(1000)

			if tt.vcRecord != nil {
				// Convert vcregistrypb.VCRecord to types.VCRecord
				typesRecord := types.VCRecord{
					VcId:            tt.vcRecord.VcId,
					VcType:          tt.vcRecord.VcType,
					VcTypeCustom:    tt.vcRecord.VcTypeCustom,
					HolderAddress:   tt.vcRecord.HolderAddress,
					HolderDid:       tt.vcRecord.HolderDid,
					IssuerAssistant: tt.vcRecord.IssuerAssistant,
					Status:          tt.vcRecord.Status,
					IssuedAt:        tt.vcRecord.IssuedAt,
					ExpiresAt:       tt.vcRecord.ExpiresAt,
					IssuedHeight:    tt.vcRecord.IssuedHeight,
				}
				keeper.SetVCRecord(context.Background(), typesRecord)
			}

			err := keeper.RevokeVC(context.Background(), tt.vcID, tt.reason, tt.revoker, tt.evidence)

			if (err != nil) != tt.shouldErr {
				t.Fatalf("expected err=%v, got %v", tt.shouldErr, err)
			}

			if tt.shouldErr {
				if err != tt.errType {
					t.Errorf("expected error %v, got %v", tt.errType, err)
				}
				return
			}

			// Verify revocation was recorded
			revRecord, ok := keeper.GetRevocationRecord(context.Background(), tt.vcID)
			if !ok {
				t.Fatal("expected to find revocation record")
			}

			if revRecord.VcId != tt.vcID {
				t.Errorf("expected VC ID %s, got %s", tt.vcID, revRecord.VcId)
			}

			if revRecord.Revoker != tt.revoker {
				t.Errorf("expected revoker %s, got %s", tt.revoker, revRecord.Revoker)
			}

			// Verify VC status was updated
			vc, ok := keeper.GetVCRecord(context.Background(), tt.vcID)
			if !ok {
				t.Fatal("expected to find VC record")
			}

			if vc.Status != types.VCStatus_VC_STATUS_REVOKED {
				t.Errorf("expected status REVOKED, got %v", vc.Status)
			}

			// Verify revocation list was updated
			revList := keeper.GetRevocationList(context.Background())
			if revList.TotalRevocations == 0 {
				t.Error("expected revocation list to be updated")
			}
		})
	}
}

// ============================
// TEST: DIDManagement
// ============================

func TestDIDManagement(t *testing.T) {
	t.Skip("This test requires refactoring to use proper SDK context - see keeper_kv_persistence_test.go for examples")
	keeper := NewKeeper(nil, "authority")
	currentTime := time.Now().Unix()
	keeper.SetCurrentTime(currentTime)

	t.Run("RegisterDID", func(t *testing.T) {
		tests := []struct {
			name        string
			did         string
			controller  string
			methods     []*vcregistrypb.VerificationMethod
			metadataURI string
			shouldErr   bool
			errType     error
		}{
			{
				name:        "valid DID registration",
				did:         "did:aura:test123",
				controller:  "aura1controller",
				methods:     []*vcregistrypb.VerificationMethod{},
				metadataURI: "https://example.com/metadata",
				shouldErr:   false,
			},
			{
				name:       "missing DID",
				did:        "",
				controller: "aura1controller",
				methods:    []*vcregistrypb.VerificationMethod{},
				shouldErr:  true,
				errType:    types.ErrInvalidDID,
			},
			{
				name:       "missing controller",
				did:        "did:aura:test456",
				controller: "",
				methods:    []*vcregistrypb.VerificationMethod{},
				shouldErr:  true,
				errType:    types.ErrInvalidHolderAddress,
			},
			{
				name:       "duplicate DID",
				did:        "did:aura:duplicate",
				controller: "aura1controller",
				methods:    []*vcregistrypb.VerificationMethod{},
				shouldErr:  true,
				errType:    types.ErrDIDAlreadyExists,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				keeper := NewKeeper(nil, "authority")
				keeper.SetCurrentTime(currentTime)

				// If testing duplicate, first registration should succeed
				if tt.name == "duplicate DID" {
					keeper.RegisterDID(context.Background(), tt.did, tt.controller, tt.methods, tt.metadataURI)
				}

				err := keeper.RegisterDID(context.Background(), tt.did, tt.controller, tt.methods, tt.metadataURI)

				if (err != nil) != tt.shouldErr {
					t.Fatalf("expected err=%v, got %v", tt.shouldErr, err)
				}

				if tt.shouldErr {
					if err != tt.errType {
						t.Errorf("expected error %v, got %v", tt.errType, err)
					}
					return
				}

				// Verify DID was registered
				doc, ok := keeper.GetDIDDocument(context.Background(), tt.did)
				if !ok {
					t.Fatal("expected to find DID document")
				}

				if doc.Did != tt.did {
					t.Errorf("expected DID %s, got %s", tt.did, doc.Did)
				}

				if doc.Controller != tt.controller {
					t.Errorf("expected controller %s, got %s", tt.controller, doc.Controller)
				}
			})
		}
	})

	t.Run("GetDIDDocument", func(t *testing.T) {
		keeper := NewKeeper(nil, "authority")
		keeper.SetCurrentTime(currentTime)

		did := "did:aura:gettest"
		controller := "aura1gettest"

		// Register DID
		keeper.RegisterDID(context.Background(), did, controller, []*vcregistrypb.VerificationMethod{}, "")

		// Retrieve DID
		doc, ok := keeper.GetDIDDocument(context.Background(), did)
		if !ok {
			t.Fatal("expected to find DID document")
		}

		if doc.Did != did {
			t.Errorf("expected DID %s, got %s", did, doc.Did)
		}

		if doc.Controller != controller {
			t.Errorf("expected controller %s, got %s", controller, doc.Controller)
		}

		// Try to get non-existent DID
		_, ok = keeper.GetDIDDocument(context.Background(), "did:aura:nonexistent")
		if ok {
			t.Error("expected not to find non-existent DID")
		}
	})

	t.Run("UpdateDIDDocument", func(t *testing.T) {
		keeper := NewKeeper(nil, "authority")
		keeper.SetCurrentTime(currentTime)

		ctx := context.Background()

		did := "did:aura:updatetest"
		controller := "aura1updatetest"

		// Register DID
		keeper.RegisterDID(ctx, did, controller, []*vcregistrypb.VerificationMethod{}, "old_metadata")

		// Update DID
		newMethods := []*vcregistrypb.VerificationMethod{
			{
				Id:        "key1",
				Type:      "Ed25519VerificationKey2020",
				PublicKey: []byte("abc123"),
			},
		}
		err := keeper.UpdateDIDDocument(ctx, did, newMethods, "new_metadata")
		if err != nil {
			t.Fatalf("failed to update DID: %v", err)
		}

		// Verify update
		doc, ok := keeper.GetDIDDocument(ctx, did)
		if !ok {
			t.Fatal("expected to find updated DID document")
		}

		if doc.MetadataUri != "new_metadata" {
			t.Errorf("expected metadata URI 'new_metadata', got '%s'", doc.MetadataUri)
		}

		if len(doc.VerificationMethods) != 1 {
			t.Errorf("expected 1 verification method, got %d", len(doc.VerificationMethods))
		}

		// Try to update non-existent DID
		err = keeper.UpdateDIDDocument(ctx, "did:aura:nonexistent", []*vcregistrypb.VerificationMethod{}, "")
		if err != types.ErrDIDNotFound {
			t.Errorf("expected ErrDIDNotFound, got %v", err)
		}
	})

	t.Run("GetDIDsByAddress", func(t *testing.T) {
		keeper := NewKeeper(nil, "authority")
		keeper.SetCurrentTime(currentTime)

		controller := "aura1multi"

		// Register multiple DIDs for same controller
		dids := []string{"did:aura:multi1", "did:aura:multi2", "did:aura:multi3"}
		for _, did := range dids {
			keeper.RegisterDID(context.Background(), did, controller, []*vcregistrypb.VerificationMethod{}, "")
		}

		// Retrieve DIDs by address
		result := keeper.GetDIDsByAddress(context.Background(), controller)
		if len(result) != len(dids) {
			t.Errorf("expected %d DIDs, got %d", len(dids), len(result))
		}

		// Retrieve DIDs for unknown address
		result = keeper.GetDIDsByAddress(context.Background(), "aura1unknown")
		if len(result) != 0 {
			t.Errorf("expected 0 DIDs for unknown address, got %d", len(result))
		}
	})

	t.Run("AddCredentialToDID", func(t *testing.T) {
		keeper := NewKeeper(nil, "authority")
		keeper.SetCurrentTime(currentTime)

		ctx := context.Background()

		did := "did:aura:credtest"
		controller := "aura1credtest"

		// Register DID
		keeper.RegisterDID(ctx, did, controller, []*vcregistrypb.VerificationMethod{}, "")

		// Add credentials
		vcID1 := "vc:cred1"
		vcID2 := "vc:cred2"

		err := keeper.AddCredentialToDID(ctx, did, vcID1)
		if err != nil {
			t.Fatalf("failed to add credential: %v", err)
		}

		err = keeper.AddCredentialToDID(ctx, did, vcID2)
		if err != nil {
			t.Fatalf("failed to add credential: %v", err)
		}

		// Verify credentials were added
		doc, ok := keeper.GetDIDDocument(ctx, did)
		if !ok {
			t.Fatal("expected to find DID document")
		}

		if len(doc.CredentialIds) != 2 {
			t.Errorf("expected 2 credentials, got %d", len(doc.CredentialIds))
		}

		// Try to add credential to non-existent DID
		err = keeper.AddCredentialToDID(ctx, "did:aura:nonexistent", "vc:test")
		if err != types.ErrDIDNotFound {
			t.Errorf("expected ErrDIDNotFound, got %v", err)
		}
	})

	t.Run("RemoveCredentialFromDID", func(t *testing.T) {
		keeper := NewKeeper(nil, "authority")
		keeper.SetCurrentTime(currentTime)

		ctx := context.Background()

		did := "did:aura:removaltest"
		controller := "aura1removaltest"

		// Register DID and add credentials
		keeper.RegisterDID(ctx, did, controller, []*vcregistrypb.VerificationMethod{}, "")
		vcID1 := "vc:remove1"
		vcID2 := "vc:remove2"
		keeper.AddCredentialToDID(ctx, did, vcID1)
		keeper.AddCredentialToDID(ctx, did, vcID2)

		// Remove one credential
		err := keeper.RemoveCredentialFromDID(ctx, did, vcID1)
		if err != nil {
			t.Fatalf("failed to remove credential: %v", err)
		}

		// Verify credential was removed
		doc, ok := keeper.GetDIDDocument(ctx, did)
		if !ok {
			t.Fatal("expected to find DID document")
		}

		if len(doc.CredentialIds) != 1 {
			t.Errorf("expected 1 credential after removal, got %d", len(doc.CredentialIds))
		}

		if doc.CredentialIds[0] != vcID2 {
			t.Errorf("expected remaining credential to be %s, got %s", vcID2, doc.CredentialIds[0])
		}

		// Try to remove credential from non-existent DID
		err = keeper.RemoveCredentialFromDID(ctx, "did:aura:nonexistent", "vc:test")
		if err != types.ErrDIDNotFound {
			t.Errorf("expected ErrDIDNotFound, got %v", err)
		}
	})
}

// ============================
// TEST: VCPolicyManagement
// ============================

func TestVCPolicyManagement(t *testing.T) {
	t.Skip("This test requires refactoring to use proper SDK context - see keeper_kv_persistence_test.go for examples")
	keeper := NewKeeper(nil, "authority")
	currentTime := time.Now().Unix()
	keeper.SetCurrentTime(currentTime)

	t.Run("SetGetVCPolicy", func(t *testing.T) {
		tests := []struct {
			name      string
			policy    *vcregistrypb.VCPolicy
			shouldErr bool
		}{
			{
				name: "valid policy",
				policy: &vcregistrypb.VCPolicy{
					VcTypeName:         "TestVC",
					Status:             vcregistrypb.VCPolicyStatus_VC_POLICY_STATUS_ACTIVE,
					CsThreshold:        1000,
					ExpiryDurationDays: 365,
					CreatedAt:          timestamppb.New(time.Unix(currentTime, 0)),
				},
				shouldErr: false,
			},
			{
				name: "missing VC type name",
				policy: &vcregistrypb.VCPolicy{
					VcTypeName: "",
					Status:     vcregistrypb.VCPolicyStatus_VC_POLICY_STATUS_ACTIVE,
					CreatedAt:  timestamppb.New(time.Unix(currentTime, 0)),
				},
				shouldErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				keeper := NewKeeper(nil, "authority")
				keeper.SetCurrentTime(currentTime)

				err := keeper.SetVCPolicy(context.Background(), *tt.policy)

				if (err != nil) != tt.shouldErr {
					t.Fatalf("expected err=%v, got %v", tt.shouldErr, err)
				}

				if tt.shouldErr {
					return
				}

				// Get policy
				retrieved, ok := keeper.GetVCPolicy(context.Background(), tt.policy.VcTypeName)
				if !ok {
					t.Fatal("expected to find policy")
				}

				if retrieved.VcTypeName != tt.policy.VcTypeName {
					t.Errorf("expected type name %s, got %s", tt.policy.VcTypeName, retrieved.VcTypeName)
				}

				if retrieved.CsThreshold != tt.policy.CsThreshold {
					t.Errorf("expected CS threshold %d, got %d", tt.policy.CsThreshold, retrieved.CsThreshold)
				}
			})
		}
	})

	t.Run("ListVCPolicies", func(t *testing.T) {
		keeper := NewKeeper(nil, "authority")
		keeper.SetCurrentTime(currentTime)

		// Create policies with different statuses
		policies := []*vcregistrypb.VCPolicy{
			{
				VcTypeName: "ActiveVC1",
				Status:     vcregistrypb.VCPolicyStatus_VC_POLICY_STATUS_ACTIVE,
				CreatedAt:  timestamppb.Now(),
			},
			{
				VcTypeName: "ActiveVC2",
				Status:     vcregistrypb.VCPolicyStatus_VC_POLICY_STATUS_ACTIVE,
				CreatedAt:  timestamppb.Now(),
			},
			{
				VcTypeName: "DraftVC",
				Status:     vcregistrypb.VCPolicyStatus_VC_POLICY_STATUS_DRAFT,
				CreatedAt:  timestamppb.Now(),
			},
			{
				VcTypeName: "DeprecatedVC",
				Status:     vcregistrypb.VCPolicyStatus_VC_POLICY_STATUS_DEPRECATED,
				CreatedAt:  timestamppb.Now(),
			},
		}

		for _, policy := range policies {
			keeper.SetVCPolicy(context.Background(), *policy)
		}

		// Test listing all policies
		all := keeper.ListVcPolicies(context.Background(), types.VCPolicyStatusUnspecified)
		if len(all) != 4 {
			t.Errorf("expected 4 policies, got %d", len(all))
		}

		// Test filtering by active status
		active := keeper.ListVcPolicies(context.Background(), types.VCPolicyStatusActive)
		if len(active) != 2 {
			t.Errorf("expected 2 active policies, got %d", len(active))
		}

		// Test filtering by draft status
		draft := keeper.ListVcPolicies(context.Background(), types.VCPolicyStatusDraft)
		if len(draft) != 1 {
			t.Errorf("expected 1 draft policy, got %d", len(draft))
		}
	})
}

// ============================
// TEST: RateLimiting
// ============================

func TestRateLimiting(t *testing.T) {
	t.Skip("This test requires refactoring to use proper SDK context - see keeper_kv_persistence_test.go for examples")
	t.Run("CheckMintRateLimit", func(t *testing.T) {
		keeper := NewKeeper(nil, "authority")
		currentTime := time.Now().Unix()
		keeper.SetCurrentTime(currentTime)

		params := types.DefaultParams()
		params.RateLimitingEnabled = true
		params.MaxMintPerDay = 5
		keeper.SetParams(*params)

		holderAddr := "aura1ratelimituser"

		// Test: no mints yet - should pass
		err := keeper.CheckMintRateLimit(context.Background(), holderAddr)
		if err != nil {
			t.Errorf("first rate limit check should pass, got %v", err)
		}

		// Increment mint count to max
		for i := 0; i < int(params.MaxMintPerDay); i++ {
			keeper.IncrementMintCount(context.Background(), holderAddr)
		}

		// Test: should fail when exceeding limit
		err = keeper.CheckMintRateLimit(context.Background(), holderAddr)
		if err != types.ErrRateLimitExceeded {
			t.Errorf("expected ErrRateLimitExceeded, got %v", err)
		}

		// Test: with rate limiting disabled
		params.RateLimitingEnabled = false
		keeper.SetParams(*params)

		err = keeper.CheckMintRateLimit(context.Background(), holderAddr)
		if err != nil {
			t.Errorf("rate limit check should pass when disabled, got %v", err)
		}
	})

	t.Run("IncrementMintCount", func(t *testing.T) {
		keeper := NewKeeper(nil, "authority")
		currentTime := time.Now().Unix()
		keeper.SetCurrentTime(currentTime)

		holderAddr := "aura1incrementuser"

		// Verify initial state
		err := keeper.CheckMintRateLimit(context.Background(), holderAddr)
		if err != nil {
			t.Errorf("should have no mints yet, got %v", err)
		}

		// Increment multiple times
		for i := 0; i < 3; i++ {
			keeper.IncrementMintCount(context.Background(), holderAddr)
		}

		// Verify count was incremented
		params := types.DefaultParams()
		params.RateLimitingEnabled = true
		params.MaxMintPerDay = 2
		keeper.SetParams(*params)

		// Should now fail (count=3, limit=2)
		err = keeper.CheckMintRateLimit(context.Background(), holderAddr)
		if err != types.ErrRateLimitExceeded {
			t.Errorf("expected rate limit exceeded after 3 increments, got %v", err)
		}
	})

	t.Run("CleanupOldMintCounts", func(t *testing.T) {
		keeper := NewKeeper(nil, "authority")
		currentTime := time.Now().Unix()
		keeper.SetCurrentTime(currentTime)

		holderAddr := "aura1cleanupuser"

		// Add mint count for current day
		keeper.IncrementMintCount(context.Background(), holderAddr)

		// Simulate time passing (8 days later)
		newTime := currentTime + (8 * 86400)
		keeper.SetCurrentTime(newTime)

		// Add mint count for new day
		keeper.IncrementMintCount(context.Background(), holderAddr)

		// Run cleanup
		keeper.CleanupOldMintCounts(context.Background())

		// Old entries should be removed (older than 7 days)
		// We can't directly verify the internal state, but we can test that
		// the function doesn't error
	})
}

// ============================
// TEST: GetStats
// ============================

func TestGetStats(t *testing.T) {
	t.Skip("This test requires refactoring to use proper SDK context - see keeper_kv_persistence_test.go for examples")
	keeper := NewKeeper(nil, "authority")
	currentTime := time.Now().Unix()
	keeper.SetCurrentTime(currentTime)

	holderAddr := "aura1statsuser"

	// Create VCs with different statuses
	activeVC := &vcregistrypb.VCRecord{
		VcId:          "vc:stat_active",
		VcType:        vcregistrypb.VCType_VC_TYPE_CUSTOM,
		VcTypeCustom:  "TestVC",
		HolderAddress: holderAddr,
		HolderDid:     "did:aura:stats",
		Status:        types.VCStatus_VC_STATUS_ACTIVE,
		IssuedAt:      timestamppb.Now(),
	}

	revokedVC := &vcregistrypb.VCRecord{
		VcId:          "vc:stat_revoked",
		VcType:        vcregistrypb.VCType_VC_TYPE_CUSTOM,
		VcTypeCustom:  "TestVC",
		HolderAddress: holderAddr,
		HolderDid:     "did:aura:stats",
		Status:        types.VCStatus_VC_STATUS_REVOKED,
		IssuedAt:      timestamppb.Now(),
	}

	expiredVC := &vcregistrypb.VCRecord{
		VcId:          "vc:stat_expired",
		VcType:        vcregistrypb.VCType_VC_TYPE_CUSTOM,
		VcTypeCustom:  "TestVC",
		HolderAddress: holderAddr,
		HolderDid:     "did:aura:stats",
		Status:        types.VCStatus_VC_STATUS_EXPIRED,
		IssuedAt:      timestamppb.Now(),
	}

	// Convert vcregistrypb.VCRecord to types.VCRecord
	keeper.SetVCRecord(context.Background(), types.VCRecord{
		VcId:            activeVC.VcId,
		VcType:          activeVC.VcType,
		VcTypeCustom:    activeVC.VcTypeCustom,
		HolderAddress:   activeVC.HolderAddress,
		HolderDid:       activeVC.HolderDid,
		IssuerAssistant: activeVC.IssuerAssistant,
		Status:          activeVC.Status,
		IssuedAt:        activeVC.IssuedAt,
		IssuedHeight:    activeVC.IssuedHeight,
	})
	keeper.SetVCRecord(context.Background(), types.VCRecord{
		VcId:            revokedVC.VcId,
		VcType:          revokedVC.VcType,
		VcTypeCustom:    revokedVC.VcTypeCustom,
		HolderAddress:   revokedVC.HolderAddress,
		HolderDid:       revokedVC.HolderDid,
		IssuerAssistant: revokedVC.IssuerAssistant,
		Status:          revokedVC.Status,
		IssuedAt:        revokedVC.IssuedAt,
		IssuedHeight:    revokedVC.IssuedHeight,
	})
	keeper.SetVCRecord(context.Background(), types.VCRecord{
		VcId:            expiredVC.VcId,
		VcType:          expiredVC.VcType,
		VcTypeCustom:    expiredVC.VcTypeCustom,
		HolderAddress:   expiredVC.HolderAddress,
		HolderDid:       expiredVC.HolderDid,
		IssuerAssistant: expiredVC.IssuerAssistant,
		Status:          expiredVC.Status,
		IssuedAt:        expiredVC.IssuedAt,
		IssuedHeight:    expiredVC.IssuedHeight,
	})

	// Add DIDs and policies
	keeper.RegisterDID(context.Background(), "did:aura:stats", holderAddr, []*vcregistrypb.VerificationMethod{}, "")
	keeper.SetVCPolicy(context.Background(), vcregistrypb.VCPolicy{
		VcTypeName: "TestVC",
		Status:     vcregistrypb.VCPolicyStatus_VC_POLICY_STATUS_ACTIVE,
		CreatedAt:  timestamppb.Now(),
	})

	// Get stats
	stats := keeper.GetStats(context.Background())

	if stats.TotalVCs != 3 {
		t.Errorf("expected 3 total VCs, got %d", stats.TotalVCs)
	}

	if stats.ActiveVCs != 1 {
		t.Errorf("expected 1 active VC, got %d", stats.ActiveVCs)
	}

	if stats.RevokedVCs != 1 {
		t.Errorf("expected 1 revoked VC, got %d", stats.RevokedVCs)
	}

	if stats.ExpiredVCs != 1 {
		t.Errorf("expected 1 expired VC, got %d", stats.ExpiredVCs)
	}

	if stats.TotalDIDs != 1 {
		t.Errorf("expected 1 DID, got %d", stats.TotalDIDs)
	}

	if stats.TotalPolicies != 1 {
		t.Errorf("expected 1 policy, got %d", stats.TotalPolicies)
	}

	// Check VC by type
	count, ok := stats.VCsByType[vcregistrypb.VCType_VC_TYPE_CUSTOM]
	if !ok || count != 3 {
		t.Errorf("expected 3 VCs of type CUSTOM, got %d", count)
	}
}

// ============================
// TEST: Genesis
// ============================

func TestInitExportGenesis(t *testing.T) {
	t.Skip("This test requires refactoring to use proper SDK context - see keeper_kv_persistence_test.go for examples")
	keeper := NewKeeper(nil, "authority")
	currentTime := time.Now().Unix()
	keeper.SetCurrentTime(currentTime)

	// Create test data
	vcRecord := &vcregistrypb.VCRecord{
		VcId:          "vc:genesis1",
		HolderAddress: "aura1genesisuser",
		HolderDid:     "did:aura:genesis",
		Status:        types.VCStatus_VC_STATUS_ACTIVE,
		IssuedAt:      timestamppb.Now(),
	}

	didDoc := &vcregistrypb.DIDDocument{
		Did:        "did:aura:genesis",
		Controller: "aura1genesisuser",
		Created:    timestamppb.Now(),
		Updated:    timestamppb.Now(),
	}

	policy := &vcregistrypb.VCPolicy{
		VcTypeName: "GenesisVC",
		Status:     vcregistrypb.VCPolicyStatus_VC_POLICY_STATUS_ACTIVE,
		CreatedAt:  timestamppb.Now(),
	}

	// Convert vcregistrypb.VCRecord to types.VCRecord
	keeper.SetVCRecord(context.Background(), types.VCRecord{
		VcId:            vcRecord.VcId,
		VcType:          vcRecord.VcType,
		VcTypeCustom:    vcRecord.VcTypeCustom,
		HolderAddress:   vcRecord.HolderAddress,
		HolderDid:       vcRecord.HolderDid,
		IssuerAssistant: vcRecord.IssuerAssistant,
		Status:          vcRecord.Status,
		IssuedAt:        vcRecord.IssuedAt,
		IssuedHeight:    vcRecord.IssuedHeight,
	})
	keeper.RegisterDID(context.Background(), didDoc.Did, didDoc.Controller, []*vcregistrypb.VerificationMethod{}, "")
	keeper.SetVCPolicy(context.Background(), *policy)

	// Export genesis
	genesis := keeper.ExportGenesis(context.Background())

	if len(genesis.VcRecords) != 1 {
		t.Errorf("expected 1 VC record in genesis, got %d", len(genesis.VcRecords))
	}

	if len(genesis.DidDocuments) != 1 {
		t.Errorf("expected 1 DID document in genesis, got %d", len(genesis.DidDocuments))
	}

	if len(genesis.VcPolicies) != 1 {
		t.Errorf("expected 1 policy in genesis, got %d", len(genesis.VcPolicies))
	}

	// Re-import and verify
	keeper2 := NewKeeper(nil, "authority")
	err := keeper2.InitGenesis(context.Background(), genesis)
	if err != nil {
		t.Fatalf("failed to init genesis: %v", err)
	}

	// Verify data was loaded
	retrievedVC, ok := keeper2.GetVCRecord(context.Background(), vcRecord.VcId)
	if !ok {
		t.Fatal("expected to find VC after genesis init")
	}

	if retrievedVC.VcId != vcRecord.VcId {
		t.Errorf("expected VC ID %s, got %s", vcRecord.VcId, retrievedVC.VcId)
	}
}

// ============================
// Helper functions
// ============================

// createMockConfidenceScoreKeeper creates a mock implementation for testing
func createMockConfidenceScoreKeeper(userScore uint64) *mockCSKeeper {
	return &mockCSKeeper{
		userScore: userScore,
	}
}

type mockCSKeeper struct {
	userScore uint64
}

func (m *mockCSKeeper) GetUserScore(walletAddr string) (uint64, bool) {
	return m.userScore, true
}

func (m *mockCSKeeper) HasCompletedIR(walletAddr, irID string) bool {
	return true
}

func (m *mockCSKeeper) GetArenaScore(walletAddr, arena string) (uint64, error) {
	return 5000, nil
}

func (m *mockCSKeeper) GetAnchorInfo(walletAddr string) (interface{}, bool) {
	return nil, true
}

func (m *mockCSKeeper) IsVerified(walletAddr string) bool {
	return true
}
