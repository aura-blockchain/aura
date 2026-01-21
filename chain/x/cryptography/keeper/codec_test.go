// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"encoding/hex"
	"testing"

	txsigning "cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
)

func TestCodecMarshalUnmarshal(t *testing.T) {
	// Create interface registry like the app does
	addrCodec := address.NewBech32Codec("aura")
	valCodec := address.NewBech32Codec("auravaloper")
	interfaceRegistry, err := codectypes.NewInterfaceRegistryWithOptions(codectypes.InterfaceRegistryOptions{
		ProtoFiles: proto.HybridResolver,
		SigningOptions: txsigning.Options{
			AddressCodec:          addrCodec,
			ValidatorAddressCodec: valCodec,
		},
	})
	require.NoError(t, err)

	cdc := codec.NewProtoCodec(interfaceRegistry)

	// Create params with same values as genesis
	params := &cryptoproto.Params{
		DefaultRotationIntervalDays: 90,
		EnableAutoRotation:          true,
		MinThresholdParticipants:    2,
		MaxThresholdParticipants:    100,
		MinEntropyBits:              256,
		MinPbkdf2Iterations:         100000,
		MinArgon2MemoryKb:           65536,
		MinArgon2Iterations:         3,
		EnforceCertificatePinning:   true,
		CertificatePinValidityDays:  365,
		MinSaltLengthBytes:          16,
		MinKeyLengthBits:            256,
	}

	t.Logf("Original params: %+v", params)

	// Marshal
	bz, err := cdc.Marshal(params)
	require.NoError(t, err, "marshal should succeed")
	t.Logf("Marshaled bytes (%d): %s", len(bz), hex.EncodeToString(bz))

	// Unmarshal
	var params2 cryptoproto.Params
	err = cdc.Unmarshal(bz, &params2)
	require.NoError(t, err, "unmarshal should succeed")
	t.Logf("Unmarshaled params: %+v", &params2)

	// Verify values
	require.Equal(t, params.DefaultRotationIntervalDays, params2.DefaultRotationIntervalDays)
	require.Equal(t, params.EnableAutoRotation, params2.EnableAutoRotation)
}
