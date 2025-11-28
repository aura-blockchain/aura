package params_test

import (
	"testing"

	"github.com/aequitas/aura/chain/x/identitychange/params"
	"github.com/aequitas/aura/chain/x/identitychange/types"
)

func TestStoreSetGet(t *testing.T) {
	store := params.NewStore(types.DefaultParams())
	got := store.GetParams()
	if got.MaxRequestsPerWalletPerMonth != types.DefaultParams().MaxRequestsPerWalletPerMonth {
		t.Fatalf("unexpected default params: %v", got)
	}
	update := got
	update.MaxRequestsPerWalletPerMonth = 5
	if err := store.SetParams(update); err != nil {
		t.Fatalf("set params: %v", err)
	}
	after := store.GetParams()
	if after.MaxRequestsPerWalletPerMonth != 5 {
		t.Fatalf("params not updated: %v", after)
	}
}
