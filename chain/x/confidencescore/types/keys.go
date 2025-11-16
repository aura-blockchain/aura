package types

const ModuleName = "confidencescore"

// Store key prefixes
const (
	// UserRecordStoreKeyPrefix is the prefix for storing user confidence records
	UserRecordStoreKeyPrefix = "user/"

	// IRCompletionStoreKeyPrefix is the prefix for storing IR completions
	IRCompletionStoreKeyPrefix = "completion/"

	// AnchorStoreKeyPrefix is the prefix for storing anchor info
	AnchorStoreKeyPrefix = "anchor/"

	// VerifiedUsersStoreKeyPrefix is the prefix for indexing verified users
	VerifiedUsersStoreKeyPrefix = "verified/"

	// ArenaCompletionStoreKeyPrefix is the prefix for indexing arena completions
	ArenaCompletionStoreKeyPrefix = "arena/"

	// ScoreHistoryStoreKeyPrefix is the prefix for storing score history
	ScoreHistoryStoreKeyPrefix = "history/"

	// RateLimitStoreKeyPrefix is the prefix for storing rate limit counters
	RateLimitStoreKeyPrefix = "ratelimit/"

	// SlashRecordStoreKeyPrefix is the prefix for storing slash records
	SlashRecordStoreKeyPrefix = "slash/"

	// ProofHashStoreKeyPrefix is the prefix for storing proof hashes (replay prevention)
	ProofHashStoreKeyPrefix = "proofhash/"
)

// UserRecordStoreKey returns the store key for a user's confidence record
func UserRecordStoreKey(walletAddr string) string {
	return UserRecordStoreKeyPrefix + walletAddr
}

// IRCompletionStoreKey returns the store key for a specific IR completion
// Format: completion/{wallet}/{ir_id}
func IRCompletionStoreKey(walletAddr, irID string) string {
	return IRCompletionStoreKeyPrefix + walletAddr + "/" + irID
}

// AnchorStoreKey returns the store key for a wallet's anchor info
func AnchorStoreKey(walletAddr string) string {
	return AnchorStoreKeyPrefix + walletAddr
}

// VerifiedUserStoreKey returns the store key for verified user index
func VerifiedUserStoreKey(walletAddr string) string {
	return VerifiedUsersStoreKeyPrefix + walletAddr
}

// ArenaCompletionStoreKey returns the store key for arena-based completion index
// Format: arena/{arena}/{wallet}/{ir_id}
func ArenaCompletionStoreKey(arena, walletAddr, irID string) string {
	return ArenaCompletionStoreKeyPrefix + arena + "/" + walletAddr + "/" + irID
}

// ScoreHistoryStoreKey returns the store key for score history
// Format: history/{wallet}/{block_height}
func ScoreHistoryStoreKey(walletAddr string, blockHeight uint64) string {
	return ScoreHistoryStoreKeyPrefix + walletAddr + "/" + string(rune(blockHeight))
}

// RateLimitStoreKey returns the store key for rate limit counters
// Format: ratelimit/{wallet}/{time_window}
func RateLimitStoreKey(walletAddr, timeWindow string) string {
	return RateLimitStoreKeyPrefix + walletAddr + "/" + timeWindow
}

// SlashRecordStoreKey returns the store key for slash records
// Format: slash/{wallet}/{slash_tx_hash}
func SlashRecordStoreKey(walletAddr, slashTxHash string) string {
	return SlashRecordStoreKeyPrefix + walletAddr + "/" + slashTxHash
}

// ProofHashStoreKey returns the store key for proof hash tracking (replay prevention)
// Format: proofhash/{wallet}/{proof_hash}
func ProofHashStoreKey(walletAddr string, proofHash []byte) string {
	return ProofHashStoreKeyPrefix + walletAddr + "/" + string(proofHash)
}
