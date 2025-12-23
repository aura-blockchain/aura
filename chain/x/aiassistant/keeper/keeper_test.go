package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/app"
	"github.com/aequitas/aura/chain/x/aiassistant/keeper"
	"github.com/aequitas/aura/chain/x/aiassistant/types"
)

func TestRegisterAssistant(t *testing.T) {
	k, ctx, bank := setupKeeper(t)
	owner := sdk.AccAddress("owner______________")
	bank.fundAccount(owner.String(), sdk.NewCoins(sdk.NewInt64Coin(types.DefaultStakeDenom, 20_000_000)))

	msg := &types.MsgRegisterAssistant{
		AssistantAddress:  sdk.AccAddress("assistant___________").String(),
		OwnerAddress:      owner.String(),
		Locales:           []string{"EN-us", "es-ES"},
		ModelHash:         "model-v1",
		ApiKeyFingerprint: "fp_123",
		Stake: types.Balance{
			Denom:  types.DefaultStakeDenom,
			Amount: math.NewInt(10_000_000),
		},
		Sponsorship: types.Balance{
			Denom:  types.DefaultStakeDenom,
			Amount: math.NewInt(1_000_000),
		},
	}

	asst, err := k.RegisterAssistant(ctx, msg)
	require.NoError(t, err)
	require.Equal(t, types.AssistantStatus_ACTIVE, asst.Status)
	require.ElementsMatch(t, []string{"en-us", "es-es"}, asst.Locales)

	// Module account should now hold stake+sponsorship
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin(types.DefaultStakeDenom, 11_000_000)), bank.moduleBalance(types.ModuleName))
}

func TestHeartbeatSlash(t *testing.T) {
	k, ctx, bank := setupKeeper(t)
	owner := sdk.AccAddress("owner______________")
	bank.fundAccount(owner.String(), sdk.NewCoins(sdk.NewInt64Coin(types.DefaultStakeDenom, 20_000_000)))

	msg := &types.MsgRegisterAssistant{
		AssistantAddress: sdk.AccAddress("assistant___________").String(),
		OwnerAddress:     owner.String(),
		Locales:          []string{"en-us"},
		ModelHash:        "model",
		Stake: types.Balance{
			Denom:  types.DefaultStakeDenom,
			Amount: math.NewInt(10_000_000),
		},
		Sponsorship: types.Balance{
			Denom:  types.DefaultStakeDenom,
			Amount: math.ZeroInt(),
		},
	}
	_, err := k.RegisterAssistant(ctx, msg)
	require.NoError(t, err)

	// Move time forward beyond heartbeat window+grace to trigger slash
	params, err := k.GetParams(ctx)
	require.NoError(t, err)
	advance := time.Duration(params.HeartbeatWindowSeconds+params.HeartbeatGraceSeconds+5) * time.Second
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(advance))

	_, err = k.Heartbeat(ctx, &types.MsgHeartbeat{
		AssistantAddress: msg.AssistantAddress,
		OperatorAddress:  owner.String(),
	})
	require.NoError(t, err)

	asst, ok := k.GetAssistant(ctx, msg.AssistantAddress)
	require.True(t, ok)
	currentStake := mustBalanceToCoin(t, asst.Stake)
	require.True(t, currentStake.Amount.LT(math.NewInt(10_000_000)))
	require.True(t, bank.burned.AmountOf(types.DefaultStakeDenom).GT(math.ZeroInt()))
}

func TestReportMisbehavior(t *testing.T) {
	k, ctx, bank := setupKeeper(t)
	owner := sdk.AccAddress("owner______________")
	bank.fundAccount(owner.String(), sdk.NewCoins(sdk.NewInt64Coin(types.DefaultStakeDenom, 20_000_000)))

	msg := &types.MsgRegisterAssistant{
		AssistantAddress: sdk.AccAddress("assistant___________").String(),
		OwnerAddress:     owner.String(),
		Locales:          []string{"en-us"},
		ModelHash:        "model",
		Stake: types.Balance{
			Denom:  types.DefaultStakeDenom,
			Amount: math.NewInt(10_000_000),
		},
		Sponsorship: types.Balance{
			Denom:  types.DefaultStakeDenom,
			Amount: math.ZeroInt(),
		},
	}
	_, err := k.RegisterAssistant(ctx, msg)
	require.NoError(t, err)

	resp, slash, err := k.ReportMisbehavior(ctx, &types.MsgReportMisbehavior{
		Reporter:         sdk.AccAddress("reporter___________").String(),
		AssistantAddress: msg.AssistantAddress,
		Infraction:       "double-sign",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, slash)
	asst, ok := k.GetAssistant(ctx, msg.AssistantAddress)
	require.True(t, ok)
	require.Equal(t, types.AssistantStatus_JAILED, asst.Status)
}

// ----------------------------------------------------------------------
// test helpers

func setupKeeper(t *testing.T) (*keeper.Keeper, sdk.Context, *mockBankKeeper) {
	t.Helper()
	return setupKeeperWithAuthority(t, "")
}

func setupKeeperWithAuthority(t *testing.T, authority string) (*keeper.Keeper, sdk.Context, *mockBankKeeper) {
	t.Helper()
	key := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		ChainID: "aiassistant-test",
		Time:    time.Now().UTC(),
	}, false, log.NewNopLogger())

	encoding := app.MakeEncodingConfig()
	bank := newMockBankKeeper()
	k := keeper.NewKeeper(encoding.Codec, key, authority, bank)
	return &k, ctx, bank
}

type mockBankKeeper struct {
	balances map[string]sdk.Coins
	burned   sdk.Coins
}

func newMockBankKeeper() *mockBankKeeper {
	return &mockBankKeeper{
		balances: make(map[string]sdk.Coins),
		burned:   sdk.NewCoins(),
	}
}

func (m *mockBankKeeper) fundAccount(addr string, coins sdk.Coins) {
	m.balances[addr] = coins
}

func (m *mockBankKeeper) balance(addr string) sdk.Coins {
	if bal, ok := m.balances[addr]; ok {
		return bal
	}
	return sdk.NewCoins()
}

func (m *mockBankKeeper) moduleBalance(module string) sdk.Coins {
	if bal, ok := m.balances[module]; ok {
		return bal
	}
	return sdk.NewCoins()
}

func (m *mockBankKeeper) SendCoinsFromAccountToModule(_ sdk.Context, sender sdk.AccAddress, module string, amt sdk.Coins) error {
	current := m.balances[sender.String()]
	var err error
	current, err = safeSubCoins(current, amt)
	if err != nil {
		return err
	}
	m.balances[sender.String()] = current
	m.balances[module] = m.balances[module].Add(amt...)
	return nil
}

func (m *mockBankKeeper) SendCoinsFromModuleToModule(_ sdk.Context, _, _ string, _ sdk.Coins) error {
	return nil
}

func (m *mockBankKeeper) BurnCoins(_ sdk.Context, module string, amt sdk.Coins) error {
	current := m.balances[module]
	var err error
	current, err = safeSubCoins(current, amt)
	if err != nil {
		return err
	}
	m.balances[module] = current
	m.burned = m.burned.Add(amt...)
	return nil
}

func safeSubCoins(balance sdk.Coins, amt sdk.Coins) (sdk.Coins, error) {
	if len(amt) == 0 {
		return balance, nil
	}
	if !balance.IsAllGTE(amt) {
		return balance, fmt.Errorf("insufficient funds")
	}
	return balance.Sub(amt...), nil
}

func mustBalanceToCoin(t *testing.T, balance types.Balance) sdk.Coin {
	t.Helper()
	require.False(t, balance.Amount.IsNil())
	return sdk.NewCoin(balance.Denom, balance.Amount)
}
