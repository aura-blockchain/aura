package types

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/assert"
)

func TestCoinStruct(t *testing.T) {
	c := Coin{Denom: "uaura", Amount: math.NewInt(123)}
	assert.Equal(t, "uaura", c.Denom)
	assert.True(t, c.Amount.Equal(math.NewInt(123)))
}

func TestPaginationDefaults(t *testing.T) {
	req := PageRequest{Offset: 10, Limit: 50, CountTotal: true}
	resp := PageResponse{NextKey: []byte{0x01}, Total: 100}

	assert.Equal(t, uint64(10), req.Offset)
	assert.Equal(t, uint64(50), req.Limit)
	assert.True(t, req.CountTotal)

	assert.Equal(t, uint64(100), resp.Total)
	assert.NotEmpty(t, resp.NextKey)
}

func TestTxResponseFields(t *testing.T) {
	tx := TxResponse{
		TxHash:    "0xabc",
		Height:    10,
		Code:      0,
		Data:      "ok",
		RawLog:    "[]",
		GasWanted: 200000,
		GasUsed:   180000,
	}

	assert.Equal(t, "0xabc", tx.TxHash)
	assert.Equal(t, int64(10), tx.Height)
	assert.Equal(t, uint32(0), tx.Code)
	assert.Equal(t, int64(200000), tx.GasWanted)
	assert.Equal(t, int64(180000), tx.GasUsed)
}
