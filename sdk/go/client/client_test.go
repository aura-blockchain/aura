package client

import (
	"context"
	"net"
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/go-bip39"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// --- test fixtures --------------------------------------------------------------------

type mockBankServer struct {
	banktypes.UnimplementedQueryServer
}

func (m mockBankServer) Balance(_ context.Context, req *banktypes.QueryBalanceRequest) (*banktypes.QueryBalanceResponse, error) {
	return &banktypes.QueryBalanceResponse{
		Balance: &sdk.Coin{Denom: "uaura", Amount: math.NewInt(42)},
	}, nil
}

func (m mockBankServer) AllBalances(_ context.Context, req *banktypes.QueryAllBalancesRequest) (*banktypes.QueryAllBalancesResponse, error) {
	return &banktypes.QueryAllBalancesResponse{
		Balances: sdk.NewCoins(
			sdk.NewCoin("uaura", math.NewInt(42)),
			sdk.NewCoin("stake", math.NewInt(10)),
		),
	}, nil
}

type mockAuthServer struct {
	authtypes.UnimplementedQueryServer
	account authtypes.AccountI
}

func (m mockAuthServer) Account(_ context.Context, req *authtypes.QueryAccountRequest) (*authtypes.QueryAccountResponse, error) {
	any, err := types.NewAnyWithValue(m.account)
	if err != nil {
		return nil, err
	}
	return &authtypes.QueryAccountResponse{Account: any}, nil
}

func startTestGRPCServer(t *testing.T) (string, func()) {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	server := grpc.NewServer()
	banktypes.RegisterQueryServer(server, &mockBankServer{})

	baseAccount := &authtypes.BaseAccount{
		Address:       "aura1testaccount",
		AccountNumber: 1,
		Sequence:      2,
	}
	authtypes.RegisterQueryServer(server, &mockAuthServer{account: baseAccount})

	go func() {
		_ = server.Serve(lis)
	}()

	return lis.Addr().String(), server.Stop
}

// --- tests ----------------------------------------------------------------------------

func TestNewClientDefaultsAndClose(t *testing.T) {
	addr, stop := startTestGRPCServer(t)
	defer stop()

	client, err := NewClient(Config{
		RPCEndpoint:  "http://localhost:26657",
		GRPCEndpoint: addr,
		ChainID:      "aura-mvp-1",
	})
	require.NoError(t, err)
	require.NotNil(t, client)

	// Defaults applied
	assert.Equal(t, "paw", client.config.Prefix)
	assert.Equal(t, "0.025uaura", client.config.GasPrice)
	assert.Equal(t, 1.5, client.config.GasAdjustment)
	assert.Equal(t, "aura-mvp-1", client.GetChainID())

	// Close should not error
	assert.NoError(t, client.Close())
}

func TestImportWalletFromMnemonic(t *testing.T) {
	addr, stop := startTestGRPCServer(t)
	defer stop()

	client, err := NewClient(Config{
		RPCEndpoint:  "http://localhost:26657",
		GRPCEndpoint: addr,
		ChainID:      "aura-mvp-1",
	})
	require.NoError(t, err)

	// Invalid mnemonic should fail fast
	_, err = client.ImportWalletFromMnemonic("bad", "not a mnemonic", "")
	assert.Error(t, err)

	// Valid mnemonic imports and returns address
	entropy, err := bip39.NewEntropy(128)
	require.NoError(t, err)
	validMnemonic, err := bip39.NewMnemonic(entropy)
	require.NoError(t, err)

	addrOut, err := client.ImportWalletFromMnemonic("alice", validMnemonic, "")
	require.NoError(t, err)
	assert.NotEmpty(t, addrOut.String())
}

func TestBalancesAndAccountQueries(t *testing.T) {
	addr, stop := startTestGRPCServer(t)
	defer stop()

	client, err := NewClient(Config{
		RPCEndpoint:  "http://localhost:26657",
		GRPCEndpoint: addr,
		ChainID:      "aura-mvp-1",
	})
	require.NoError(t, err)

	ctx := context.Background()

	bal, err := client.GetBalance(ctx, "aura1test", "uaura")
	require.NoError(t, err)
	assert.Equal(t, "42", bal.Amount.String())

	all, err := client.GetAllBalances(ctx, "aura1test")
	require.NoError(t, err)
	assert.Len(t, all, 2)

	account, err := client.GetAccount(ctx, "aura1testaccount")
	require.NoError(t, err)
	assert.Equal(t, uint64(1), account.GetAccountNumber())
	assert.Equal(t, uint64(2), account.GetSequence())
}

func TestSignAndBroadcastKeyLookupFailure(t *testing.T) {
	addr, stop := startTestGRPCServer(t)
	defer stop()

	client, err := NewClient(Config{
		RPCEndpoint:  "http://localhost:26657",
		GRPCEndpoint: addr,
		ChainID:      "aura-mvp-1",
	})
	require.NoError(t, err)

	// No keys imported yet; should surface keyring error
	_, err = client.SignAndBroadcast(context.Background(), "missing", &banktypes.MsgSend{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get key")
}
