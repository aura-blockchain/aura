package types

const (
	// ModuleName defines the module name
	ModuleName = "economics"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

// KVStore key prefixes
var (
	// Fee management keys
	DynamicFeeConfigKey    = []byte{0x01}
	TransferTaxConfigKey   = []byte{0x02}
	FeeMultiplierKey       = []byte{0x03}
	UtilizationHistoryKey  = []byte{0x04}

	// Vesting keys
	VestingSchedulePrefix = []byte{0x10}
	UserVestingIndexPrefix = []byte{0x11}
	VoteLockPrefix        = []byte{0x12}
	UserVoteLockIndexPrefix = []byte{0x13}

	// Treasury keys
	TreasuryMultisigKey       = []byte{0x20}
	PendingTreasuryTxPrefix   = []byte{0x21}
	TreasuryBalanceKey        = []byte{0x22}
	TreasuryTransactionCounter = []byte{0x23}

	// Governance keys
	ProposalPrefix         = []byte{0x30}
	VotePrefix             = []byte{0x31}
	DepositPrefix          = []byte{0x32}
	NextProposalIDKey      = []byte{0x33}
	VoteDelegationPrefix   = []byte{0x34}
	SnapshotVotePrefix     = []byte{0x35}
	VoteCommitmentPrefix   = []byte{0x36}
	VetoRequestPrefix      = []byte{0x37}
	TokenLockPrefix        = []byte{0x38}

	// Economic monitoring keys
	InflationAlertPrefix   = []byte{0x40}
	LargeTxRecordPrefix    = []byte{0x41}
	LastLargeTxTimePrefix  = []byte{0x42}
	AddressHoldingPrefix   = []byte{0x43}
	PreviousInflationKey   = []byte{0x44}

	// MEV keys
	UserMEVBalancePrefix   = []byte{0x50}
	TotalMEVPendingKey     = []byte{0x51}
	TotalBurnedKey         = []byte{0x52}

	// State tracking keys
	CurrentHeightKey = []byte{0x60}
	CurrentTimeKey   = []byte{0x61}
	ParamsKey        = []byte{0x62}
)

// GetVestingScheduleKey returns the store key for a vesting schedule
func GetVestingScheduleKey(scheduleID string) []byte {
	return append(VestingSchedulePrefix, []byte(scheduleID)...)
}

// GetUserVestingIndexKey returns the store key for a user's vesting index
func GetUserVestingIndexKey(userAddress string) []byte {
	return append(UserVestingIndexPrefix, []byte(userAddress)...)
}

// GetVoteLockKey returns the store key for a vote lock
func GetVoteLockKey(lockID string) []byte {
	return append(VoteLockPrefix, []byte(lockID)...)
}

// GetUserVoteLockIndexKey returns the store key for a user's vote lock index
func GetUserVoteLockIndexKey(userAddress string) []byte {
	return append(UserVoteLockIndexPrefix, []byte(userAddress)...)
}

// GetPendingTreasuryTxKey returns the store key for a pending treasury transaction
func GetPendingTreasuryTxKey(txID string) []byte {
	return append(PendingTreasuryTxPrefix, []byte(txID)...)
}

// GetProposalKey returns the store key for a proposal
func GetProposalKey(proposalID uint64) []byte {
	bz := make([]byte, 8)
	// Use big-endian encoding for consistent ordering
	bz[0] = byte(proposalID >> 56)
	bz[1] = byte(proposalID >> 48)
	bz[2] = byte(proposalID >> 40)
	bz[3] = byte(proposalID >> 32)
	bz[4] = byte(proposalID >> 24)
	bz[5] = byte(proposalID >> 16)
	bz[6] = byte(proposalID >> 8)
	bz[7] = byte(proposalID)
	return append(ProposalPrefix, bz...)
}

// GetVoteKey returns the store key for a vote
func GetVoteKey(proposalID uint64, voter string) []byte {
	bz := make([]byte, 8)
	bz[0] = byte(proposalID >> 56)
	bz[1] = byte(proposalID >> 48)
	bz[2] = byte(proposalID >> 40)
	bz[3] = byte(proposalID >> 32)
	bz[4] = byte(proposalID >> 24)
	bz[5] = byte(proposalID >> 16)
	bz[6] = byte(proposalID >> 8)
	bz[7] = byte(proposalID)
	key := append(VotePrefix, bz...)
	return append(key, []byte(voter)...)
}

// GetDepositKey returns the store key for a deposit
func GetDepositKey(proposalID uint64, depositor string) []byte {
	bz := make([]byte, 8)
	bz[0] = byte(proposalID >> 56)
	bz[1] = byte(proposalID >> 48)
	bz[2] = byte(proposalID >> 40)
	bz[3] = byte(proposalID >> 32)
	bz[4] = byte(proposalID >> 24)
	bz[5] = byte(proposalID >> 16)
	bz[6] = byte(proposalID >> 8)
	bz[7] = byte(proposalID)
	key := append(DepositPrefix, bz...)
	return append(key, []byte(depositor)...)
}

// GetVoteDelegationKey returns the store key for a vote delegation
func GetVoteDelegationKey(delegator, delegate string) []byte {
	key := append(VoteDelegationPrefix, []byte(delegator)...)
	key = append(key, byte(0x00)) // separator
	return append(key, []byte(delegate)...)
}

// GetSnapshotVoteKey returns the store key for a snapshot vote
func GetSnapshotVoteKey(proposalID uint64, voter string) []byte {
	bz := make([]byte, 8)
	bz[0] = byte(proposalID >> 56)
	bz[1] = byte(proposalID >> 48)
	bz[2] = byte(proposalID >> 40)
	bz[3] = byte(proposalID >> 32)
	bz[4] = byte(proposalID >> 24)
	bz[5] = byte(proposalID >> 16)
	bz[6] = byte(proposalID >> 8)
	bz[7] = byte(proposalID)
	key := append(SnapshotVotePrefix, bz...)
	return append(key, []byte(voter)...)
}

// GetVoteCommitmentKey returns the store key for a vote commitment
func GetVoteCommitmentKey(proposalID uint64, voter string) []byte {
	bz := make([]byte, 8)
	bz[0] = byte(proposalID >> 56)
	bz[1] = byte(proposalID >> 48)
	bz[2] = byte(proposalID >> 40)
	bz[3] = byte(proposalID >> 32)
	bz[4] = byte(proposalID >> 24)
	bz[5] = byte(proposalID >> 16)
	bz[6] = byte(proposalID >> 8)
	bz[7] = byte(proposalID)
	key := append(VoteCommitmentPrefix, bz...)
	return append(key, []byte(voter)...)
}

// GetVetoRequestKey returns the store key for a veto request
func GetVetoRequestKey(proposalID uint64) []byte {
	bz := make([]byte, 8)
	bz[0] = byte(proposalID >> 56)
	bz[1] = byte(proposalID >> 48)
	bz[2] = byte(proposalID >> 40)
	bz[3] = byte(proposalID >> 32)
	bz[4] = byte(proposalID >> 24)
	bz[5] = byte(proposalID >> 16)
	bz[6] = byte(proposalID >> 8)
	bz[7] = byte(proposalID)
	return append(VetoRequestPrefix, bz...)
}

// GetTokenLockKey returns the store key for a token lock
func GetTokenLockKey(owner string, lockID string) []byte {
	key := append(TokenLockPrefix, []byte(owner)...)
	key = append(key, byte(0x00)) // separator
	return append(key, []byte(lockID)...)
}

// GetInflationAlertKey returns the store key for an inflation alert
func GetInflationAlertKey(alertID string) []byte {
	return append(InflationAlertPrefix, []byte(alertID)...)
}

// GetLargeTxRecordKey returns the store key for a large transaction record
func GetLargeTxRecordKey(txHash string) []byte {
	return append(LargeTxRecordPrefix, []byte(txHash)...)
}

// GetLastLargeTxTimeKey returns the store key for last large tx time
func GetLastLargeTxTimeKey(address string) []byte {
	return append(LastLargeTxTimePrefix, []byte(address)...)
}

// GetAddressHoldingKey returns the store key for address holdings
func GetAddressHoldingKey(address string) []byte {
	return append(AddressHoldingPrefix, []byte(address)...)
}

// GetUserMEVBalanceKey returns the store key for user MEV balance
func GetUserMEVBalanceKey(address string) []byte {
	return append(UserMEVBalancePrefix, []byte(address)...)
}
