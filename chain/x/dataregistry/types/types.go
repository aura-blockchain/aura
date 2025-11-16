package types

import "time"

// DataItemType categorizes stored data
type DataItemType int32

const (
	DataItemTypeUnspecified DataItemType = 0

	// Documents
	DataItemTypeVehicleRegistration DataItemType = 1
	DataItemTypeVehicleInsurance    DataItemType = 2
	DataItemTypePropertyDeed        DataItemType = 3
	DataItemTypeLeaseAgreement      DataItemType = 4
	DataItemTypeContract            DataItemType = 5
	DataItemTypeReceipt             DataItemType = 6
	DataItemTypeWarranty            DataItemType = 7

	// Media
	DataItemTypePhoto       DataItemType = 10
	DataItemTypeVideo       DataItemType = 11
	DataItemTypeAudio       DataItemType = 12
	DataItemTypeDocumentPDF DataItemType = 13
	DataItemTypeDocument    DataItemType = 14
	DataItemTypeText        DataItemType = 15
	DataItemTypeCode        DataItemType = 16

	// Scores & Achievements
	DataItemTypeGolfScore     DataItemType = 20
	DataItemTypeTestScore     DataItemType = 21
	DataItemTypeCertification DataItemType = 22
	DataItemTypeAchievement   DataItemType = 23

	// Digital Assets
	DataItemTypeNFT          DataItemType = 30
	DataItemTypeDigitalArt   DataItemType = 31
	DataItemTypeMusicLicense DataItemType = 32

	// Health
	DataItemTypeVaccinationRecord DataItemType = 40
	DataItemTypeMedicalRecord     DataItemType = 41
	DataItemTypePrescription      DataItemType = 42
	DataItemTypeHealthRecord      DataItemType = 43
	DataItemTypeBiometric         DataItemType = 44

	// Location & Sensor
	DataItemTypeLocation   DataItemType = 50
	DataItemTypeSensorData DataItemType = 51

	// Credentials
	DataItemTypeCertificate DataItemType = 60
	DataItemTypeCredential  DataItemType = 61

	// Social
	DataItemTypeSocialPost DataItemType = 70
	DataItemTypeMessage    DataItemType = 71

	// Custom
	DataItemTypeCustom DataItemType = 100
)

// DataItemStatus represents lifecycle
type DataItemStatus int32

const (
	DataItemStatusUnspecified         DataItemStatus = 0
	DataItemStatusPendingVerification DataItemStatus = 1
	DataItemStatusVerified            DataItemStatus = 2
	DataItemStatusRejected            DataItemStatus = 3
	DataItemStatusExpired             DataItemStatus = 4
	DataItemStatusRevoked             DataItemStatus = 5
)

// VerificationLevel indicates trust level
type VerificationLevel int32

const (
	VerificationLevelUnspecified        VerificationLevel = 0
	VerificationLevelSelfAttested       VerificationLevel = 1
	VerificationLevelPeerVerified       VerificationLevel = 2
	VerificationLevelAIVerified         VerificationLevel = 3
	VerificationLevelExpertVerified     VerificationLevel = 3 // Alias for AI
	VerificationLevelAuthorityVerified  VerificationLevel = 4
	VerificationLevelBlockchainAnchored VerificationLevel = 5
)

// AccessMode defines visibility policy
type AccessMode int32

const (
	AccessModePrivate       AccessMode = 0
	AccessModeWhitelist     AccessMode = 1
	AccessModePublic        AccessMode = 2
	AccessModeVerifiedUsers AccessMode = 3
)

// GeoLocation for geotagged data
type GeoLocation struct {
	Latitude       float64
	Longitude      float64
	Altitude       float64
	AccuracyMeters float64
	Timestamp      time.Time
	LocationName   string
}

// Verification record
type Verification struct {
	VerifierAddress    string
	Level              VerificationLevel
	VerifiedAt         time.Time
	VerificationMethod string
	ConfidenceScore    uint64
	Notes              string
	Proof              []byte
}

// AccessPolicy controls who can view data
type AccessPolicy struct {
	Mode                    AccessMode
	AllowedAddresses        []string
	DeniedAddresses         []string
	RequireVerifiedIdentity bool
	MinConfidenceScore      uint64
}

// DataItem represents a stored data item
type DataItem struct {
	DataID            string
	OwnerAddress      string
	DataType          DataItemType
	DataTypeCustom    string
	Status            DataItemStatus
	VerificationLevel VerificationLevel
	ContentHash       []byte
	StorageLocation   string
	IsEncrypted       bool
	EncryptionKeyHash []byte
	Title             string
	Description       string
	Metadata          map[string]string
	Tags              []string
	CreatedAt         time.Time
	VerifiedAt        time.Time
	ExpiresAt         time.Time
	GeoLocation       *GeoLocation
	Verifications     []Verification
	VerifiedBy        string
	AccessPolicy      *AccessPolicy
	PreviousDataID    string
	Version           uint64
}

// GolfScoreData represents golf score with geotagging
type GolfScoreData struct {
	CourseName        string
	Location          *GeoLocation
	PlayedAt          time.Time
	TotalScore        uint32
	HoleScores        []uint32
	Handicap          uint32
	ScorecardImageCID string
	PlayingPartners   []string
}

// VehicleRegistrationData represents vehicle registration
type VehicleRegistrationData struct {
	VIN                  string
	Make                 string
	Model                string
	Year                 uint32
	LicensePlate         string
	State                string
	RegisteredAt         time.Time
	ExpiresAt            time.Time
	RegistrationImageCID string
}

// PhotoData represents photo with metadata
type PhotoData struct {
	PhotoCID     string
	Location     *GeoLocation
	TakenAt      time.Time
	CameraModel  string
	Description  string
	PeopleTagged []string
	Width        uint32
	Height       uint32
}

// RegistryStats represents registry statistics
type RegistryStats struct {
	TotalDataItems     uint64
	TotalVerifiedItems uint64
	TotalStorageBytes  uint64
	ItemsByType        map[string]uint64
	TotalVerifications uint64
}
