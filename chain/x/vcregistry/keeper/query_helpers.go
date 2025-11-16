package keeper

// GetConfidenceScoreKeeper returns the confidence score keeper
func (k *Keeper) GetConfidenceScoreKeeper() ConfidenceScoreKeeper {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.csKeeper
}
