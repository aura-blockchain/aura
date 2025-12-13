package keeper

//lint:file-ignore SA1019 // invariants rely on deprecated SDK registry until upstream removal

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	storeprefix "cosmossdk.io/store/prefix"
	"github.com/aequitas/aura/chain/x/cryptography/types"
	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
)

// RegisterInvariants registers all cryptography module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "params-valid", ParamsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "key-rotation-validity", KeyRotationValidityInvariant(k))
	ir.RegisterRoute(types.ModuleName, "threshold-scheme-consistency", ThresholdSchemeConsistencyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "zk-proof-config-validity", ZKProofConfigValidityInvariant(k))
	ir.RegisterRoute(types.ModuleName, "secure-enclave-validity", SecureEnclaveValidityInvariant(k))
	ir.RegisterRoute(types.ModuleName, "quantum-key-validity", QuantumKeyValidityInvariant(k))
}

// AllInvariants runs all invariants of the cryptography module
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		invariants := []sdk.Invariant{
			ParamsInvariant(k),
			KeyRotationValidityInvariant(k),
			ThresholdSchemeConsistencyInvariant(k),
			ZKProofConfigValidityInvariant(k),
			SecureEnclaveValidityInvariant(k),
			QuantumKeyValidityInvariant(k),
		}

		for _, inv := range invariants {
			msg, broken := inv(ctx)
			if broken {
				return msg, broken
			}
		}

		return "", false
	}
}

// ParamsInvariant checks that module parameters are valid
func ParamsInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		params, err := k.GetParams(ctx)
		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"params-valid",
				fmt.Sprintf("failed to get params: %s", err.Error()),
			), true
		}

		if err := types.ValidateParams(params); err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"params-valid",
				fmt.Sprintf("invalid params: %s", err.Error()),
			), true
		}
		return "", false
	}
}

// KeyRotationValidityInvariant checks that key rotation schedules are valid
func KeyRotationValidityInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := k.getStore(ctx)
		prefixStore := storeprefix.NewStore(store, types.KeyRotationSchedulePrefix)

		iterator := prefixStore.Iterator(nil, nil)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var schedule types.KeyRotationSchedule
			if err := k.cdc.Unmarshal(iterator.Value(), &schedule); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"key-rotation-validity",
					fmt.Sprintf("failed to unmarshal key rotation schedule: %s", err.Error()),
				), true
			}

			// Check key ID is not empty
			if schedule.KeyId == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"key-rotation-validity",
					"key rotation schedule has empty key ID",
				), true
			}

			// Check rotation period is positive
			if schedule.RotationIntervalSeconds == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"key-rotation-validity",
					fmt.Sprintf("key %s has zero rotation period", schedule.KeyId),
				), true
			}

			// Check next rotation time
			if schedule.NextRotationTime.IsZero() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"key-rotation-validity",
					fmt.Sprintf("key %s has zero next rotation time", schedule.KeyId),
				), true
			}

			// Check last rotation time is before next rotation time
			if schedule.LastRotation != nil && !schedule.LastRotation.IsZero() &&
			   !schedule.NextRotationTime.IsZero() {
				if schedule.NextRotationTime.Before(*schedule.LastRotation) {
					return sdk.FormatInvariant(
						types.ModuleName,
						"key-rotation-validity",
						fmt.Sprintf("key %s next rotation is before last rotation", schedule.KeyId),
					), true
				}
			}
		}

		return "", false
	}
}

// ThresholdSchemeConsistencyInvariant checks threshold signature schemes
func ThresholdSchemeConsistencyInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := k.getStore(ctx)
		prefixStore := storeprefix.NewStore(store, types.ThresholdSchemePrefix)

		iterator := prefixStore.Iterator(nil, nil)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var scheme types.ThresholdSignatureScheme
			if err := k.cdc.Unmarshal(iterator.Value(), &scheme); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"threshold-scheme-consistency",
					fmt.Sprintf("failed to unmarshal threshold scheme: %s", err.Error()),
				), true
			}

			// Check scheme ID is not empty
			if scheme.SchemeId == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"threshold-scheme-consistency",
					"threshold scheme has empty ID",
				), true
			}

			// Check threshold <= total participants
			if scheme.Threshold == 0 || scheme.Threshold > scheme.TotalParticipants {
				return sdk.FormatInvariant(
					types.ModuleName,
					"threshold-scheme-consistency",
					fmt.Sprintf("scheme %s has invalid threshold: %d of %d",
						scheme.SchemeId, scheme.Threshold, scheme.TotalParticipants),
				), true
			}

			// Check participants count
			if scheme.TotalParticipants == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"threshold-scheme-consistency",
					fmt.Sprintf("scheme %s has zero participants", scheme.SchemeId),
				), true
			}

			// Check public key is not empty
			if len(scheme.PublicKey) == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"threshold-scheme-consistency",
					fmt.Sprintf("scheme %s has empty public key", scheme.SchemeId),
				), true
			}
		}

		return "", false
	}
}

// ZKProofConfigValidityInvariant checks ZK proof configurations
func ZKProofConfigValidityInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := k.getStore(ctx)
		prefixStore := storeprefix.NewStore(store, types.ZKProofConfigPrefix)

		iterator := prefixStore.Iterator(nil, nil)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var config types.ZKProofConfig
			if err := k.cdc.Unmarshal(iterator.Value(), &config); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"zk-proof-config-validity",
					fmt.Sprintf("failed to unmarshal ZK proof config: %s", err.Error()),
				), true
			}

			// Check config ID is not empty
			if config.CircuitId == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"zk-proof-config-validity",
					"ZK proof config has empty ID",
				), true
			}

			// Check proof type is valid (using enum comparison)
			validTypes := []cryptoproto.ZKProofType{
				cryptoproto.ZKProofType_ZK_PROOF_TYPE_GROTH16,
				cryptoproto.ZKProofType_ZK_PROOF_TYPE_PLONK,
				cryptoproto.ZKProofType_ZK_PROOF_TYPE_STARK,
				cryptoproto.ZKProofType_ZK_PROOF_TYPE_BULLETPROOFS,
			}
			typeValid := false
			for _, vt := range validTypes {
				if config.ProofType == vt {
					typeValid = true
					break
				}
			}
			if !typeValid {
				return sdk.FormatInvariant(
					types.ModuleName,
					"zk-proof-config-validity",
					fmt.Sprintf("config %s has invalid proof type: %s", config.CircuitId, config.ProofType.String()),
				), true
			}

			// Check public parameters exist (use PublicParameters not CircuitParameters)
			if len(config.PublicParameters) == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"zk-proof-config-validity",
					fmt.Sprintf("config %s has empty public parameters", config.CircuitId),
				), true
			}
		}

		return "", false
	}
}

// SecureEnclaveValidityInvariant checks secure enclave configurations
func SecureEnclaveValidityInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := k.getStore(ctx)
		prefixStore := storeprefix.NewStore(store, types.SecureEnclavePrefix)

		iterator := prefixStore.Iterator(nil, nil)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var enclave types.SecureEnclaveConfig
			if err := k.cdc.Unmarshal(iterator.Value(), &enclave); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"secure-enclave-validity",
					fmt.Sprintf("failed to unmarshal secure enclave: %s", err.Error()),
				), true
			}

			// Check enclave ID is not empty
			if enclave.EnclaveId == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"secure-enclave-validity",
					"secure enclave has empty ID",
				), true
			}

			// Check attestation is not empty (use AttestationData not Attestation)
			if len(enclave.AttestationData) == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"secure-enclave-validity",
					fmt.Sprintf("enclave %s has empty attestation data", enclave.EnclaveId),
				), true
			}

			// Check enclave type is valid (using enum comparison)
			validTypes := []cryptoproto.SecureEnclaveType{
				cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_SGX,
				cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_SEV,
				cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_TPM,
				cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_HSM,
			}
			typeValid := false
			for _, vt := range validTypes {
				if enclave.EnclaveType == vt {
					typeValid = true
					break
				}
			}
			if !typeValid {
				return sdk.FormatInvariant(
					types.ModuleName,
					"secure-enclave-validity",
					fmt.Sprintf("enclave %s has invalid type: %s", enclave.EnclaveId, enclave.EnclaveType.String()),
				), true
			}
		}

		return "", false
	}
}

// QuantumKeyValidityInvariant checks quantum-resistant keys
func QuantumKeyValidityInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := k.getStore(ctx)
		prefixStore := storeprefix.NewStore(store, types.QuantumResistantKeyPrefix)

		iterator := prefixStore.Iterator(nil, nil)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var qkey types.QuantumResistantKey
			if err := k.cdc.Unmarshal(iterator.Value(), &qkey); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"quantum-key-validity",
					fmt.Sprintf("failed to unmarshal quantum key: %s", err.Error()),
				), true
			}

			// Check key ID is not empty
			if qkey.KeyId == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"quantum-key-validity",
					"quantum key has empty ID",
				), true
			}

			// Check algorithm is valid (using enum comparison)
			validAlgs := []cryptoproto.QuantumResistantAlgorithm{
				cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_CRYSTALS_DILITHIUM,
				cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_CRYSTALS_KYBER,
				cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_FALCON,
				cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_SPHINCS_PLUS,
				cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_NTRU,
			}
			algValid := false
			for _, va := range validAlgs {
				if qkey.Algorithm == va {
					algValid = true
					break
				}
			}
			if !algValid {
				return sdk.FormatInvariant(
					types.ModuleName,
					"quantum-key-validity",
					fmt.Sprintf("key %s has invalid algorithm: %s", qkey.KeyId, qkey.Algorithm.String()),
				), true
			}

			// Check public key is not empty
			if len(qkey.PublicKey) == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"quantum-key-validity",
					fmt.Sprintf("key %s has empty public key", qkey.KeyId),
				), true
			}
		}

		return "", false
	}
}
