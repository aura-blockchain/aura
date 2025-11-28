package keeper

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/dataregistry/types"
)

type InvariantsTestSuite struct {
	KeeperTestSuite
}

func TestInvariantsTestSuite(t *testing.T) {
	suite.Run(t, new(InvariantsTestSuite))
}

func (suite *InvariantsTestSuite) TestAllInvariants() {
	// Test: All invariants on empty keeper
	inv := AllInvariants(suite.Keeper)
	msg, broken := inv()
	suite.False(broken, "all invariants should pass on empty keeper")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestRegisterInvariants() {
	// Register invariants - should not panic
	suite.NotPanics(func() {
		RegisterInvariants(suite.Keeper)
	})
}

func (suite *InvariantsTestSuite) TestParamsInvariant() {
	// Test: valid params pass
	inv := ParamsInvariant(suite.Keeper)
	msg, broken := inv()
	suite.False(broken, "valid params should pass")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestDataItemConsistencyInvariant() {
	// Test: valid data item passes
	ownerAddr := "cosmos1test"
	item := types.DataItem{
		DataId:       "test-data-1",
		OwnerAddress: ownerAddr,
		DataType:     types.DataItemType_DOCUMENT,
		CreatedAt:    1000,
		UpdatedAt:    1000,
		Status:       types.DataItemStatus_ACTIVE,
	}
	err := suite.Keeper.SetDataItem(item)
	suite.NoError(err)

	inv := DataItemConsistencyInvariant(suite.Keeper)
	msg, broken := inv()
	suite.False(broken, "valid data item should pass")
	suite.Empty(msg)

	// Test: empty owner address fails
	suite.SetupTest()
	emptyOwnerItem := types.DataItem{
		DataId:       "test-data-2",
		OwnerAddress: "",
		DataType:     types.DataItemType_DOCUMENT,
		CreatedAt:    1000,
		Status:       types.DataItemStatus_ACTIVE,
	}
	suite.Keeper.mu.Lock()
	suite.Keeper.dataItems["test-data-2"] = emptyOwnerItem
	suite.Keeper.mu.Unlock()

	msg, broken = inv()
	suite.True(broken, "empty owner address should break invariant")
	suite.Contains(msg, "empty owner address")

	// Test: non-positive created timestamp fails
	suite.SetupTest()
	zeroCreatedItem := types.DataItem{
		DataId:       "test-data-3",
		OwnerAddress: ownerAddr,
		DataType:     types.DataItemType_DOCUMENT,
		CreatedAt:    0,
		Status:       types.DataItemStatus_ACTIVE,
	}
	suite.Keeper.mu.Lock()
	suite.Keeper.dataItems["test-data-3"] = zeroCreatedItem
	suite.Keeper.mu.Unlock()

	msg, broken = inv()
	suite.True(broken, "non-positive created timestamp should break invariant")
	suite.Contains(msg, "non-positive created timestamp")

	// Test: unspecified data type fails
	suite.SetupTest()
	unspecifiedTypeItem := types.DataItem{
		DataId:       "test-data-4",
		OwnerAddress: ownerAddr,
		DataType:     types.DataItemType_UNSPECIFIED,
		CreatedAt:    1000,
		Status:       types.DataItemStatus_ACTIVE,
	}
	suite.Keeper.mu.Lock()
	suite.Keeper.dataItems["test-data-4"] = unspecifiedTypeItem
	suite.Keeper.mu.Unlock()

	msg, broken = inv()
	suite.True(broken, "unspecified data type should break invariant")
	suite.Contains(msg, "unspecified data type")

	// Test: data ID mismatch fails
	suite.SetupTest()
	mismatchedIDItem := types.DataItem{
		DataId:       "different-id",
		OwnerAddress: ownerAddr,
		DataType:     types.DataItemType_DOCUMENT,
		CreatedAt:    1000,
		Status:       types.DataItemStatus_ACTIVE,
	}
	suite.Keeper.mu.Lock()
	suite.Keeper.dataItems["test-data-5"] = mismatchedIDItem
	suite.Keeper.mu.Unlock()

	msg, broken = inv()
	suite.True(broken, "data ID mismatch should break invariant")
	suite.Contains(msg, "ID mismatch")
}

func (suite *InvariantsTestSuite) TestCIDValidityInvariant() {
	// Test: valid data ID format passes
	ownerAddr := "cosmos1test"
	item := types.DataItem{
		DataId:       "cosmos1test-DOCUMENT-hash123",
		OwnerAddress: ownerAddr,
		DataType:     types.DataItemType_DOCUMENT,
		CreatedAt:    1000,
		IpfsCid:      "QmTestCID1234567890",
		Status:       types.DataItemStatus_ACTIVE,
	}
	err := suite.Keeper.SetDataItem(item)
	suite.NoError(err)

	inv := CIDValidityInvariant(suite.Keeper)
	msg, broken := inv()
	suite.False(broken, "valid data ID format should pass")
	suite.Empty(msg)

	// Test: invalid data ID format (no hyphen) fails
	suite.SetupTest()
	invalidIDItem := types.DataItem{
		DataId:       "invalidid",
		OwnerAddress: ownerAddr,
		DataType:     types.DataItemType_DOCUMENT,
		CreatedAt:    1000,
		Status:       types.DataItemStatus_ACTIVE,
	}
	suite.Keeper.mu.Lock()
	suite.Keeper.dataItems["invalidid"] = invalidIDItem
	suite.Keeper.mu.Unlock()

	msg, broken = inv()
	suite.True(broken, "invalid data ID format should break invariant")
	suite.Contains(msg, "invalid ID format")

	// Test: short IPFS CID fails
	suite.SetupTest()
	shortCIDItem := types.DataItem{
		DataId:       "test-data-1",
		OwnerAddress: ownerAddr,
		DataType:     types.DataItemType_DOCUMENT,
		CreatedAt:    1000,
		IpfsCid:      "short",
		Status:       types.DataItemStatus_ACTIVE,
	}
	suite.Keeper.mu.Lock()
	suite.Keeper.dataItems["test-data-1"] = shortCIDItem
	suite.Keeper.mu.Unlock()

	msg, broken = inv()
	suite.True(broken, "short IPFS CID should break invariant")
	suite.Contains(msg, "invalid IPFS CID length")

	// Test: invalid IPFS CID prefix fails
	suite.SetupTest()
	invalidPrefixItem := types.DataItem{
		DataId:       "test-data-2",
		OwnerAddress: ownerAddr,
		DataType:     types.DataItemType_DOCUMENT,
		CreatedAt:    1000,
		IpfsCid:      "ZmInvalidPrefix123",
		Status:       types.DataItemStatus_ACTIVE,
	}
	suite.Keeper.mu.Lock()
	suite.Keeper.dataItems["test-data-2"] = invalidPrefixItem
	suite.Keeper.mu.Unlock()

	msg, broken = inv()
	suite.True(broken, "invalid IPFS CID prefix should break invariant")
	suite.Contains(msg, "invalid IPFS CID prefix")
}

func (suite *InvariantsTestSuite) TestOwnerIndexConsistencyInvariant() {
	// Test: valid owner index passes
	ownerAddr := "cosmos1test"
	item := types.DataItem{
		DataId:       "test-data-1",
		OwnerAddress: ownerAddr,
		DataType:     types.DataItemType_DOCUMENT,
		CreatedAt:    1000,
		Status:       types.DataItemStatus_ACTIVE,
	}
	err := suite.Keeper.SetDataItem(item)
	suite.NoError(err)

	inv := OwnerIndexConsistencyInvariant(suite.Keeper)
	msg, broken := inv()
	suite.False(broken, "valid owner index should pass")
	suite.Empty(msg)

	// Test: missing owner in userDataItems fails
	suite.SetupTest()
	orphanItem := types.DataItem{
		DataId:       "test-data-2",
		OwnerAddress: ownerAddr,
		DataType:     types.DataItemType_DOCUMENT,
		CreatedAt:    1000,
		Status:       types.DataItemStatus_ACTIVE,
	}
	suite.Keeper.mu.Lock()
	suite.Keeper.dataItems["test-data-2"] = orphanItem
	// Don't add to userDataItems - orphaned item
	suite.Keeper.mu.Unlock()

	msg, broken = inv()
	suite.True(broken, "missing owner in userDataItems should break invariant")
	suite.Contains(msg, "not found in userDataItems index")

	// Test: data ID not in owner's list fails
	suite.SetupTest()
	missingInListItem := types.DataItem{
		DataId:       "test-data-3",
		OwnerAddress: ownerAddr,
		DataType:     types.DataItemType_DOCUMENT,
		CreatedAt:    1000,
		Status:       types.DataItemStatus_ACTIVE,
	}
	suite.Keeper.mu.Lock()
	suite.Keeper.dataItems["test-data-3"] = missingInListItem
	suite.Keeper.userDataItems[ownerAddr] = []string{"other-data-id"}
	suite.Keeper.mu.Unlock()

	msg, broken = inv()
	suite.True(broken, "data ID not in owner's list should break invariant")
	suite.Contains(msg, "not found in owner")

	// Test: owner references non-existent data item fails
	suite.SetupTest()
	suite.Keeper.mu.Lock()
	suite.Keeper.userDataItems[ownerAddr] = []string{"non-existent-id"}
	suite.Keeper.mu.Unlock()

	msg, broken = inv()
	suite.True(broken, "owner referencing non-existent data item should break invariant")
	suite.Contains(msg, "non-existent data item")

	// Test: owner mismatch fails
	suite.SetupTest()
	mismatchItem := types.DataItem{
		DataId:       "test-data-4",
		OwnerAddress: "cosmos1different",
		DataType:     types.DataItemType_DOCUMENT,
		CreatedAt:    1000,
		Status:       types.DataItemStatus_ACTIVE,
	}
	suite.Keeper.mu.Lock()
	suite.Keeper.dataItems["test-data-4"] = mismatchItem
	suite.Keeper.userDataItems[ownerAddr] = []string{"test-data-4"}
	suite.Keeper.mu.Unlock()

	msg, broken = inv()
	suite.True(broken, "owner mismatch should break invariant")
	suite.Contains(msg, "owner mismatch")
}

func (suite *InvariantsTestSuite) TestMetadataIntegrityInvariant() {
	// Test: valid metadata passes
	ownerAddr := "cosmos1test"
	item := types.DataItem{
		DataId:       "test-data-1",
		OwnerAddress: ownerAddr,
		DataType:     types.DataItemType_DOCUMENT,
		CreatedAt:    1000,
		IsEncrypted:  true,
		EncryptionMethod: "aes-256-gcm",
		Metadata:     []byte("small metadata"),
		Tags:         []string{"tag1", "tag2"},
		Status:       types.DataItemStatus_ACTIVE,
	}
	err := suite.Keeper.SetDataItem(item)
	suite.NoError(err)

	inv := MetadataIntegrityInvariant(suite.Keeper)
	msg, broken := inv()
	suite.False(broken, "valid metadata should pass")
	suite.Empty(msg)

	// Test: encrypted without encryption method fails
	suite.SetupTest()
	noEncryptionMethod := types.DataItem{
		DataId:       "test-data-2",
		OwnerAddress: ownerAddr,
		DataType:     types.DataItemType_DOCUMENT,
		CreatedAt:    1000,
		IsEncrypted:  true,
		EncryptionMethod: "",
		Status:       types.DataItemStatus_ACTIVE,
	}
	suite.Keeper.mu.Lock()
	suite.Keeper.dataItems["test-data-2"] = noEncryptionMethod
	suite.Keeper.mu.Unlock()

	msg, broken = inv()
	suite.True(broken, "encrypted without encryption method should break invariant")
	suite.Contains(msg, "no encryption method")

	// Test: excessive metadata size fails
	suite.SetupTest()
	largeMetadata := make([]byte, 2000000) // 2MB
	excessiveMetadataItem := types.DataItem{
		DataId:       "test-data-3",
		OwnerAddress: ownerAddr,
		DataType:     types.DataItemType_DOCUMENT,
		CreatedAt:    1000,
		Metadata:     largeMetadata,
		Status:       types.DataItemStatus_ACTIVE,
	}
	suite.Keeper.mu.Lock()
	suite.Keeper.dataItems["test-data-3"] = excessiveMetadataItem
	suite.Keeper.mu.Unlock()

	msg, broken = inv()
	suite.True(broken, "excessive metadata size should break invariant")
	suite.Contains(msg, "excessive metadata size")

	// Test: too many tags fails
	suite.SetupTest()
	manyTags := make([]string, 150)
	for i := 0; i < 150; i++ {
		manyTags[i] = "tag"
	}
	tooManyTagsItem := types.DataItem{
		DataId:       "test-data-4",
		OwnerAddress: ownerAddr,
		DataType:     types.DataItemType_DOCUMENT,
		CreatedAt:    1000,
		Tags:         manyTags,
		Status:       types.DataItemStatus_ACTIVE,
	}
	suite.Keeper.mu.Lock()
	suite.Keeper.dataItems["test-data-4"] = tooManyTagsItem
	suite.Keeper.mu.Unlock()

	msg, broken = inv()
	suite.True(broken, "too many tags should break invariant")
	suite.Contains(msg, "too many tags")

	// Test: tag exceeds maximum length fails
	suite.SetupTest()
	longTag := make([]byte, 150)
	for i := 0; i < 150; i++ {
		longTag[i] = 'a'
	}
	longTagItem := types.DataItem{
		DataId:       "test-data-5",
		OwnerAddress: ownerAddr,
		DataType:     types.DataItemType_DOCUMENT,
		CreatedAt:    1000,
		Tags:         []string{string(longTag)},
		Status:       types.DataItemStatus_ACTIVE,
	}
	suite.Keeper.mu.Lock()
	suite.Keeper.dataItems["test-data-5"] = longTagItem
	suite.Keeper.mu.Unlock()

	msg, broken = inv()
	suite.True(broken, "tag exceeding maximum length should break invariant")
	suite.Contains(msg, "exceeds maximum length")
}
