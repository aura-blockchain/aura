# Data Registry Module - Quick Reference

**Module:** `chain/x/dataregistry`
**Status:** ✅ COMPLETE
**Tests:** ✅ 5/5 Passing

---

## Quick Start

### Build & Test
```bash
cd chain
go build ./x/dataregistry/...
go test ./x/dataregistry/... -v
```

### Import in Go Code
```go
import (
    "github.com/aequitas/aura/chain/x/dataregistry"
    drkeeper "github.com/aequitas/aura/chain/x/dataregistry/keeper"
    drparams "github.com/aequitas/aura/chain/x/dataregistry/params"
    drtypes "github.com/aequitas/aura/chain/x/dataregistry/types"
)
```

---

## API Quick Reference

### Store Data
```go
dataID, err := keeper.StoreDataItem(
    ownerAddress,      // "aura1..."
    dataType,          // types.DataItemTypePhoto
    title,             // "My Photo"
    description,       // "Description"
    contentHash,       // []byte{...}
    storageLocation,   // "ipfs://Qm..."
    isEncrypted,       // false
    geoLocation,       // &types.GeoLocation{...} or nil
    metadata,          // map[string]string{...}
    accessPolicy,      // &types.AccessPolicy{...}
    tags,              // []string{"photo", "vacation"}
)
```

### Update Data
```go
err := keeper.UpdateDataItem(
    dataID,
    ownerAddress,
    newTitle,
    newDescription,
    newMetadata,
    newAccessPolicy,
    newTags,
)
```

### Verify Data
```go
err := keeper.VerifyDataItem(
    dataID,
    verifierAddress,
    verificationLevel,  // types.VerificationLevelPeerVerified
    confidenceScore,    // 85
    notes,              // "Verified"
    method,             // "Manual review"
    proof,              // []byte{...} or nil
)
```

### Query Data
```go
// Get single item
item, exists := keeper.GetDataItem(dataID)

// Get user's items
items := keeper.ListUserDataItems(
    ownerAddress,
    typeFilter,   // types.DataItemTypeGolfScore or 0
    statusFilter, // types.DataItemStatusVerified or 0
)

// Search
results := keeper.SearchDataItems(
    query,        // ""
    tags,         // []string{"golf"}
    typeFilter,   // types.DataItemTypeGolfScore
    geoLocation,  // &types.GeoLocation{...} or nil
    radiusKM,     // 50.0
    requester,    // "aura1..."
)
```

### Check Access
```go
hasAccess := keeper.CheckAccess(dataID, requesterAddress)
```

---

## Data Types

### Documents
- `DataItemTypeVehicleRegistration`
- `DataItemTypeVehicleInsurance`
- `DataItemTypePropertyDeed`
- `DataItemTypeLeaseAgreement`
- `DataItemTypeContract`
- `DataItemTypeReceipt`
- `DataItemTypeWarranty`

### Media
- `DataItemTypePhoto`
- `DataItemTypeVideo`
- `DataItemTypeAudio`
- `DataItemTypeDocumentPDF`

### Scores & Achievements
- `DataItemTypeGolfScore`
- `DataItemTypeTestScore`
- `DataItemTypeCertification`
- `DataItemTypeAchievement`

### Digital Assets
- `DataItemTypeNFT`
- `DataItemTypeDigitalArt`
- `DataItemTypeMusicLicense`

### Health
- `DataItemTypeVaccinationRecord`
- `DataItemTypeMedicalRecord`
- `DataItemTypePrescription`

### Custom
- `DataItemTypeCustom`

---

## Verification Levels

```go
VerificationLevelSelfAttested       // Level 1
VerificationLevelPeerVerified       // Level 2
VerificationLevelAIVerified         // Level 3
VerificationLevelAuthorityVerified  // Level 4
VerificationLevelBlockchainAnchored // Level 5
```

---

## Access Modes

```go
AccessModePrivate       // Owner only
AccessModeWhitelist     // Specific addresses
AccessModePublic        // Anyone
AccessModeVerifiedUsers // AURA-verified users
```

---

## Status Values

```go
DataItemStatusPendingVerification
DataItemStatusVerified
DataItemStatusRejected
DataItemStatusExpired
DataItemStatusRevoked
```

---

## Geo Location

```go
geoLocation := &types.GeoLocation{
    Latitude:       36.5697,
    Longitude:      -121.9489,
    Altitude:       0.0,
    AccuracyMeters: 10.0,
    Timestamp:      time.Now(),
    LocationName:   "Pebble Beach",
}
```

---

## Access Policy

```go
// Private
policy := &types.AccessPolicy{
    Mode: types.AccessModePrivate,
}

// Public
policy := &types.AccessPolicy{
    Mode: types.AccessModePublic,
}

// Whitelist
policy := &types.AccessPolicy{
    Mode: types.AccessModeWhitelist,
    AllowedAddresses: []string{"aura1...", "aura2..."},
}

// Verified users with CS threshold
policy := &types.AccessPolicy{
    Mode: types.AccessModeVerifiedUsers,
    RequireVerifiedIdentity: true,
    MinConfidenceScore: 80,
}
```

---

## Example Use Cases

### Store Golf Score
```go
golfScore := map[string]string{
    "course":      "Pebble Beach",
    "total_score": "78",
    "handicap":    "12",
    "date":        "2025-11-13",
}

dataID, _ := keeper.StoreDataItem(
    "aura1golfer",
    types.DataItemTypeGolfScore,
    "Pebble Beach Round",
    "Great round at Pebble Beach!",
    sha256("scorecard.jpg"),
    "ipfs://Qm...",
    false,
    pebbleBeachLocation,
    golfScore,
    &types.AccessPolicy{Mode: types.AccessModePublic},
    []string{"golf", "pebble-beach", "score"},
)
```

### Store Vehicle Registration
```go
vehicleData := map[string]string{
    "vin":    "5YJ3E1EA1KF123456",
    "make":   "Tesla",
    "model":  "Model 3",
    "year":   "2024",
    "state":  "CA",
    "plate":  "ABC1234",
}

dataID, _ := keeper.StoreDataItem(
    "aura1seller",
    types.DataItemTypeVehicleRegistration,
    "2024 Tesla Model 3",
    "California registration",
    sha256("registration.pdf"),
    "ipfs://Qm...",
    false,
    nil,
    vehicleData,
    &types.AccessPolicy{Mode: types.AccessModePublic},
    []string{"vehicle", "tesla", "registration"},
)
```

### Store Medical Record (Encrypted)
```go
medicalData := map[string]string{
    "vaccine":  "Pfizer",
    "dose":     "Booster",
    "date":     "2025-10-15",
    "provider": "Kaiser",
}

dataID, _ := keeper.StoreDataItem(
    "aura1patient",
    types.DataItemTypeVaccinationRecord,
    "COVID-19 Booster",
    "Booster vaccination",
    sha256("vax-record.pdf"),
    "ipfs://Qm...",
    true, // ENCRYPTED
    nil,
    medicalData,
    &types.AccessPolicy{Mode: types.AccessModePrivate},
    []string{"vaccination", "covid", "health"},
)
```

---

## Common Patterns

### Search Golf Scores Near Location
```go
pebbleBeach := &types.GeoLocation{
    Latitude:  36.5697,
    Longitude: -121.9489,
}

scores := keeper.SearchDataItems(
    "",                            // No text query
    []string{"golf"},              // Tag filter
    types.DataItemTypeGolfScore,   // Type filter
    pebbleBeach,                   // Near this location
    50.0,                          // Within 50km
    "aura1requester",              // Requester
)
```

### Get All Public Photos
```go
photos := keeper.SearchDataItems(
    "",
    []string{"photo"},
    types.DataItemTypePhoto,
    nil,
    0,
    "aura1requester",
)

// Filter for public access
publicPhotos := []types.DataItem{}
for _, photo := range photos {
    if photo.AccessPolicy != nil &&
       photo.AccessPolicy.Mode == types.AccessModePublic {
        publicPhotos = append(publicPhotos, photo)
    }
}
```

### Verify with Multiple Verifiers
```go
// First verifier
keeper.VerifyDataItem(
    dataID,
    "aura1verifier1",
    types.VerificationLevelPeerVerified,
    90,
    "Looks good",
    "Manual review",
    nil,
)

// Second verifier
keeper.VerifyDataItem(
    dataID,
    "aura1verifier2",
    types.VerificationLevelPeerVerified,
    95,
    "Confirmed",
    "Manual review",
    nil,
)

// Get all verifications
verifications, _ := keeper.GetDataItemVerifications(dataID)
// len(verifications) == 2
```

---

## Module Parameters

```go
params := keeper.GetParams()
// params.MaxDataItemsPerUser = 1000
// params.MaxStorageBytes = 104857600 (100MB)
// params.StorageFee = "100000uaura"
// params.VerificationReward = 1000
// params.AuthorizedVerifiers = []string{...}
```

---

## Testing

```bash
# Run all tests
go test ./x/dataregistry/... -v

# Run specific test
go test ./x/dataregistry/keeper -v -run TestStoreDataItem

# Run with coverage
go test ./x/dataregistry/... -cover
```

---

## File Locations

### Module Code
- `chain/x/dataregistry/keeper/keeper.go` - Core keeper
- `chain/x/dataregistry/keeper/data_item.go` - CRUD operations
- `chain/x/dataregistry/types/*.go` - Type definitions
- `chain/x/dataregistry/module.go` - Module interface

### Proto Files
- `proto/aura/dataregistry/v1beta1/data_registry.proto` - Core types
- `proto/aura/dataregistry/v1beta1/tx.proto` - Transactions
- `proto/aura/dataregistry/v1beta1/query.proto` - Queries

### Tests
- `chain/x/dataregistry/keeper/keeper_test.go` - Unit tests

### Documentation
- `DATAREGISTRY_IMPLEMENTATION_SUMMARY.md` - Full implementation details
- `DATAREGISTRY_COMPLETION_REPORT.md` - Completion report
- `DATAREGISTRY_QUICK_REFERENCE.md` - This file

---

## Need Help?

See full documentation in:
- `DATAREGISTRY_IMPLEMENTATION_SUMMARY.md`
- `DATAREGISTRY_COMPLETION_REPORT.md`
- `docs/modules/FEATURE_SPECIFICATIONS.md` (Feature #3)

**Status:** ✅ READY FOR USE
**Tests:** ✅ ALL PASSING
**Build:** ✅ NO ERRORS
