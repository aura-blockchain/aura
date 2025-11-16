package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/aequitas/aura/chain/x/dataregistry/ipfs"
	"github.com/aequitas/aura/chain/x/dataregistry/params"
	"github.com/aequitas/aura/chain/x/dataregistry/types"
)

// Keeper manages the state of the dataregistry module
type Keeper struct {
	mu            sync.RWMutex
	dataItems     map[string]types.DataItem       // data_id -> DataItem
	userDataItems map[string][]string             // owner_address -> []data_id
	typeIndex     map[types.DataItemType][]string // data_type -> []data_id
	paramsStore   *params.Store
	ipfsClient    ipfs.IPFSClient // IPFS client for content storage
	currentHeight uint64
	currentTime   int64
	nextDataID    uint64
}

// NewKeeper creates a new Keeper instance
func NewKeeper(store *params.Store) *Keeper {
	if store == nil {
		store = params.NewStore(types.DefaultParams())
	}

	// Initialize with mock IPFS client by default
	// Can be replaced with real client via SetIPFSClient
	return &Keeper{
		dataItems:     make(map[string]types.DataItem),
		userDataItems: make(map[string][]string),
		typeIndex:     make(map[types.DataItemType][]string),
		paramsStore:   store,
		ipfsClient:    ipfs.NewMockClient(),
		currentTime:   time.Now().Unix(),
		nextDataID:    1,
	}
}

// NewKeeperWithIPFS creates a new Keeper instance with IPFS configuration
func NewKeeperWithIPFS(store *params.Store, ipfsConfig *ipfs.Config) (*Keeper, error) {
	keeper := NewKeeper(store)

	// Create real IPFS client
	client, err := ipfs.NewClient(ipfsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create IPFS client: %w", err)
	}

	keeper.ipfsClient = client
	return keeper, nil
}

// SetIPFSClient sets the IPFS client (useful for testing)
func (k *Keeper) SetIPFSClient(client ipfs.IPFSClient) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.ipfsClient = client
}

// GetIPFSClient returns the IPFS client
func (k *Keeper) GetIPFSClient() ipfs.IPFSClient {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.ipfsClient
}

// SetCurrentHeight sets the current block height
func (k *Keeper) SetCurrentHeight(height uint64) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.currentHeight = height
}

// SetCurrentTime sets the current time
func (k *Keeper) SetCurrentTime(t int64) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.currentTime = t
}

// GetParams returns the current module parameters
func (k *Keeper) GetParams() types.Params {
	if k.paramsStore != nil {
		return k.paramsStore.GetParams()
	}
	return types.DefaultParams()
}

// SetParams sets new module parameters
func (k *Keeper) SetParams(params types.Params) error {
	if k.paramsStore == nil {
		return types.ErrUnauthorized
	}
	return k.paramsStore.SetParams(params)
}

// ============================
// DATA ITEM MANAGEMENT
// ============================

// GetDataItem retrieves a data item by ID
func (k *Keeper) GetDataItem(dataID string) (types.DataItem, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	item, ok := k.dataItems[dataID]
	return item, ok
}

// SetDataItem stores a data item
func (k *Keeper) SetDataItem(item types.DataItem) error {
	if item.DataID == "" {
		return types.ErrInvalidDataID
	}
	if item.OwnerAddress == "" {
		return types.ErrInvalidOwner
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	k.dataItems[item.DataID] = item

	// Index by user
	k.userDataItems[item.OwnerAddress] = append(k.userDataItems[item.OwnerAddress], item.DataID)

	// Index by type
	k.typeIndex[item.DataType] = append(k.typeIndex[item.DataType], item.DataID)

	return nil
}

// DeleteDataItem removes a data item
func (k *Keeper) DeleteDataItem(dataID string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	item, ok := k.dataItems[dataID]
	if !ok {
		return types.ErrDataItemNotFound
	}

	// Unpin from IPFS if storage location is a valid CID
	if ipfs.IsValidCID(item.StorageLocation) {
		ctx := context.Background()
		if err := k.ipfsClient.Unpin(ctx, item.StorageLocation); err != nil {
			// Log error but continue with deletion
			// In production, this should be properly logged
			_ = err
		}
	}

	// Remove from main map
	delete(k.dataItems, dataID)

	// Remove from user index
	userItems := k.userDataItems[item.OwnerAddress]
	newUserItems := []string{}
	for _, id := range userItems {
		if id != dataID {
			newUserItems = append(newUserItems, id)
		}
	}
	k.userDataItems[item.OwnerAddress] = newUserItems

	// Remove from type index
	typeItems := k.typeIndex[item.DataType]
	newTypeItems := []string{}
	for _, id := range typeItems {
		if id != dataID {
			newTypeItems = append(newTypeItems, id)
		}
	}
	k.typeIndex[item.DataType] = newTypeItems

	return nil
}

// ListUserDataItems returns all data items for a user
func (k *Keeper) ListUserDataItems(ownerAddress string, typeFilter types.DataItemType, statusFilter types.DataItemStatus) []types.DataItem {
	k.mu.RLock()
	defer k.mu.RUnlock()

	dataIDs, ok := k.userDataItems[ownerAddress]
	if !ok {
		return []types.DataItem{}
	}

	items := []types.DataItem{}
	for _, dataID := range dataIDs {
		if item, ok := k.dataItems[dataID]; ok {
			// Apply filters
			if typeFilter != types.DataItemTypeUnspecified && item.DataType != typeFilter {
				continue
			}
			if statusFilter != types.DataItemStatusUnspecified && item.Status != statusFilter {
				continue
			}
			items = append(items, item)
		}
	}

	return items
}

// ============================
// ACCESS CONTROL
// ============================

// CheckAccess checks if requester has access to a data item
func (k *Keeper) CheckAccess(dataID string, requester string) bool {
	item, ok := k.GetDataItem(dataID)
	if !ok {
		return false
	}

	// Owner always has access
	if item.OwnerAddress == requester {
		return true
	}

	// Check access policy
	if item.AccessPolicy == nil {
		return false
	}

	switch item.AccessPolicy.Mode {
	case types.AccessModePrivate:
		return false
	case types.AccessModePublic:
		return true
	case types.AccessModeWhitelist:
		for _, addr := range item.AccessPolicy.AllowedAddresses {
			if addr == requester {
				return true
			}
		}
		return false
	case types.AccessModeVerifiedUsers:
		// In production, check if requester has AURA VC
		return true
	default:
		return false
	}
}

// ============================
// SEARCH FUNCTIONALITY
// ============================

// SearchDataItems searches for data items
func (k *Keeper) SearchDataItems(query string, tags []string, typeFilter types.DataItemType, geoLocation *types.GeoLocation, radiusKM float64, requester string) []types.DataItem {
	k.mu.RLock()
	defer k.mu.RUnlock()

	results := []types.DataItem{}

	// Simple search implementation
	for _, item := range k.dataItems {
		// Check access
		if !k.CheckAccess(item.DataID, requester) {
			continue
		}

		// Type filter
		if typeFilter != types.DataItemTypeUnspecified && item.DataType != typeFilter {
			continue
		}

		// Tag filter
		if len(tags) > 0 {
			hasTag := false
			for _, searchTag := range tags {
				for _, itemTag := range item.Tags {
					if itemTag == searchTag {
						hasTag = true
						break
					}
				}
				if hasTag {
					break
				}
			}
			if !hasTag {
				continue
			}
		}

		// Geo filter (simple distance check)
		if geoLocation != nil && item.GeoLocation != nil && radiusKM > 0 {
			distance := calculateDistance(
				geoLocation.Latitude, geoLocation.Longitude,
				item.GeoLocation.Latitude, item.GeoLocation.Longitude,
			)
			if distance > radiusKM {
				continue
			}
		}

		results = append(results, item)
	}

	return results
}

// calculateDistance calculates distance between two points (simplified)
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// Simplified distance calculation (in production, use Haversine formula)
	dlat := lat2 - lat1
	dlon := lon2 - lon1
	return dlat*dlat + dlon*dlon // Approximate distance squared
}

// ============================
// STATISTICS
// ============================

// GetStats returns registry statistics
func (k *Keeper) GetStats() types.RegistryStats {
	k.mu.RLock()
	defer k.mu.RUnlock()

	stats := types.RegistryStats{
		TotalDataItems: uint64(len(k.dataItems)),
		ItemsByType:    make(map[string]uint64),
	}

	for _, item := range k.dataItems {
		if item.Status == types.DataItemStatusVerified {
			stats.TotalVerifiedItems++
		}

		stats.TotalVerifications += uint64(len(item.Verifications))

		typeName := fmt.Sprintf("%d", item.DataType)
		stats.ItemsByType[typeName]++
	}

	return stats
}

// ============================
// GENESIS
// ============================

// InitGenesis initializes the keeper from genesis state
func (k *Keeper) InitGenesis(genesis types.GenesisState) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	// Set params
	if err := k.paramsStore.SetParams(genesis.Params); err != nil {
		return fmt.Errorf("failed to set params: %w", err)
	}

	// Load data items
	for _, item := range genesis.DataItems {
		k.dataItems[item.DataID] = item
		k.userDataItems[item.OwnerAddress] = append(k.userDataItems[item.OwnerAddress], item.DataID)
		k.typeIndex[item.DataType] = append(k.typeIndex[item.DataType], item.DataID)
	}

	// Set next data ID
	k.nextDataID = genesis.NextDataID

	return nil
}

// ExportGenesis exports the current state for genesis
func (k *Keeper) ExportGenesis() types.GenesisState {
	k.mu.RLock()
	defer k.mu.RUnlock()

	items := []types.DataItem{}
	for _, item := range k.dataItems {
		items = append(items, item)
	}

	return types.GenesisState{
		Params:     k.GetParams(),
		DataItems:  items,
		NextDataID: k.nextDataID,
	}
}

// GenerateDataID generates a unique data ID
func (k *Keeper) GenerateDataID(ownerAddress string, dataType types.DataItemType) string {
	h := sha256.New()
	h.Write([]byte(ownerAddress))
	h.Write([]byte(fmt.Sprintf("%d", dataType)))
	h.Write([]byte(fmt.Sprintf("%d", k.currentTime)))
	h.Write([]byte(fmt.Sprintf("%d", k.currentHeight)))
	k.nextDataID++
	h.Write([]byte(fmt.Sprintf("%d", k.nextDataID)))
	return "data:" + hex.EncodeToString(h.Sum(nil))[:32]
}
