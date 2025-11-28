package types

const (
	// ModuleName defines the module name
	ModuleName = "economicsecurity"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key (for transient data)
	MemStoreKey = "mem_" + ModuleName
)

// Store key prefixes for KV store
var (
	// ParamsKey is the key for module parameters
	ParamsKey = []byte{0x00}

	// DynamicFeeConfigPrefix is the prefix for dynamic fee configurations
	DynamicFeeConfigPrefix = []byte{0x01}

	// MEVConfigPrefix is the prefix for MEV protection configurations
	MEVConfigPrefix = []byte{0x02}

	// WhaleProtectionPrefix is the prefix for whale protection state
	WhaleProtectionPrefix = []byte{0x03}

	// VestingSchedulePrefix is the prefix for vesting schedules
	VestingSchedulePrefix = []byte{0x04}

	// VoteLockPrefix is the prefix for vote locks
	VoteLockPrefix = []byte{0x05}

	// PendingTreasuryTxPrefix is the prefix for pending treasury transactions
	PendingTreasuryTxPrefix = []byte{0x06}

	// RewardDistributionPrefix is the prefix for reward distribution records
	RewardDistributionPrefix = []byte{0x07}

	// UserVestingIndexPrefix is the prefix for user vesting index
	UserVestingIndexPrefix = []byte{0x08}

	// UserVoteLockIndexPrefix is the prefix for user vote lock index
	UserVoteLockIndexPrefix = []byte{0x09}

	// WhaleActivityPrefix is the prefix for whale activity tracking
	WhaleActivityPrefix = []byte{0x0A}

	// InflationAlertPrefix is the prefix for inflation alerts
	InflationAlertPrefix = []byte{0x0B}

	// LargeTxRecordPrefix is the prefix for large transaction records
	LargeTxRecordPrefix = []byte{0x0C}

	// LastLargeTxTimePrefix is the prefix for last large tx time by address
	LastLargeTxTimePrefix = []byte{0x0D}

	// AddressHoldingPrefix is the prefix for address holdings
	AddressHoldingPrefix = []byte{0x0E}

	// UserMEVBalancePrefix is the prefix for user MEV balances
	UserMEVBalancePrefix = []byte{0x0F}

	// TotalMEVPendingKey is the singleton key for total pending MEV
	TotalMEVPendingKey = []byte{0x10}

	// TotalBurnedKey is the singleton key for total burned amount
	TotalBurnedKey = []byte{0x11}

	// PreviousInflationKey is the singleton key for previous inflation rate
	PreviousInflationKey = []byte{0x12}

	// CurrentHeightKey is the singleton key for current block height
	CurrentHeightKey = []byte{0x13}

	// CurrentTimeKey is the singleton key for current block time
	CurrentTimeKey = []byte{0x14}

	// InflationAlertCounterKey is the singleton key for inflation alert counter
	InflationAlertCounterKey = []byte{0x15}

	// LargeTxRecordCounterKey is the singleton key for large tx record counter
	LargeTxRecordCounterKey = []byte{0x16}
)

// GetDynamicFeeConfigKey returns the store key for a dynamic fee config
func GetDynamicFeeConfigKey(id string) []byte {
	return append(DynamicFeeConfigPrefix, []byte(id)...)
}

// GetMEVConfigKey returns the store key for MEV configuration
func GetMEVConfigKey(id string) []byte {
	return append(MEVConfigPrefix, []byte(id)...)
}

// GetWhaleProtectionKey returns the store key for whale protection by address
func GetWhaleProtectionKey(address string) []byte {
	return append(WhaleProtectionPrefix, []byte(address)...)
}

// GetVestingScheduleKey returns the store key for a vesting schedule
func GetVestingScheduleKey(scheduleID string) []byte {
	return append(VestingSchedulePrefix, []byte(scheduleID)...)
}

// GetVoteLockKey returns the store key for a vote lock
func GetVoteLockKey(lockID string) []byte {
	return append(VoteLockPrefix, []byte(lockID)...)
}

// GetPendingTreasuryTxKey returns the store key for a pending treasury tx
func GetPendingTreasuryTxKey(txID string) []byte {
	return append(PendingTreasuryTxPrefix, []byte(txID)...)
}

// GetRewardDistributionKey returns the store key for a reward distribution record
func GetRewardDistributionKey(distributionID string) []byte {
	return append(RewardDistributionPrefix, []byte(distributionID)...)
}

// GetUserVestingIndexKey returns the store key for user vesting index
func GetUserVestingIndexKey(userAddress string) []byte {
	return append(UserVestingIndexPrefix, []byte(userAddress)...)
}

// GetUserVoteLockIndexKey returns the store key for user vote lock index
func GetUserVoteLockIndexKey(userAddress string) []byte {
	return append(UserVoteLockIndexPrefix, []byte(userAddress)...)
}

// GetWhaleActivityKey returns the store key for whale activity tracking
func GetWhaleActivityKey(address string) []byte {
	return append(WhaleActivityPrefix, []byte(address)...)
}

// GetInflationAlertKey returns the store key for an inflation alert
func GetInflationAlertKey(alertID string) []byte {
	return append(InflationAlertPrefix, []byte(alertID)...)
}

// GetLargeTxRecordKey returns the store key for a large tx record
func GetLargeTxRecordKey(recordID string) []byte {
	return append(LargeTxRecordPrefix, []byte(recordID)...)
}

// GetLastLargeTxTimeKey returns the store key for last large tx time
func GetLastLargeTxTimeKey(address string) []byte {
	return append(LastLargeTxTimePrefix, []byte(address)...)
}

// GetAddressHoldingKey returns the store key for address holding
func GetAddressHoldingKey(address string) []byte {
	return append(AddressHoldingPrefix, []byte(address)...)
}

// GetUserMEVBalanceKey returns the store key for user MEV balance
func GetUserMEVBalanceKey(address string) []byte {
	return append(UserMEVBalancePrefix, []byte(address)...)
}
