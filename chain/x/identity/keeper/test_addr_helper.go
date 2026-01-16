package keeper

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// testAddr returns a deterministic, valid bech32 account address for tests.
func testAddr(index byte) string {
	return sdk.AccAddress(bytes.Repeat([]byte{index}, 20)).String()
}
