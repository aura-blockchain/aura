package types

const (
	// ModuleName defines the module name
	ModuleName = "confidencescore"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

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

	// VerificationProofHashStoreKeyPrefix is the prefix for storing verification proof hashes
	VerificationProofHashStoreKeyPrefix = "verifyproof/"

	// DelegationStoreKeyPrefix is the prefix for storing score delegations
	DelegationStoreKeyPrefix = "delegation/"

	// DelegationExpirationIndexPrefix is the prefix for indexing delegations by expiration height
	DelegationExpirationIndexPrefix = "delexp/"

	// MarketplaceListingStoreKeyPrefix is the prefix for storing marketplace listings
	MarketplaceListingStoreKeyPrefix = "listing/"

	// MarketplacePurchaseStoreKeyPrefix is the prefix for storing marketplace purchases
	MarketplacePurchaseStoreKeyPrefix = "purchase/"
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

// VerificationProofHashStoreKey returns the store key for verification proof hash tracking
// Format: verifyproof/{wallet}/{proof_hash}
func VerificationProofHashStoreKey(walletAddr, proofHash string) string {
	return VerificationProofHashStoreKeyPrefix + walletAddr + "/" + proofHash
}

// DelegationStoreKey returns the store key for score delegations
// Format: delegation/{delegation_id}
func DelegationStoreKey(delegationID string) string {
	return DelegationStoreKeyPrefix + delegationID
}

// ExpirationIndexKey returns the store key for delegation expiration index
// Format: delexp/{height}/{delegation_id}
// Height is encoded as string for proper sorting
func ExpirationIndexKey(height uint64, delegationID string) string {
	return DelegationExpirationIndexPrefix + fmt.Sprintf("%020d", height) + "/" + delegationID
}

// ExpirationIndexPrefix returns the prefix for all delegations expiring at a specific height
// Format: delexp/{height}/
// Height is encoded as string for proper sorting
func ExpirationIndexPrefix(height uint64) string {
	return DelegationExpirationIndexPrefix + fmt.Sprintf("%020d", height) + "/"
}

// MarketplaceListingStoreKey returns the store key for marketplace listings
// Format: listing/{listing_id}
func MarketplaceListingStoreKey(listingID string) string {
	return MarketplaceListingStoreKeyPrefix + listingID
}

// MarketplacePurchaseStoreKey returns the store key for marketplace purchases
// Format: purchase/{purchase_id}
func MarketplacePurchaseStoreKey(purchaseID string) string {
	return MarketplacePurchaseStoreKeyPrefix + purchaseID
}
