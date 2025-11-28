"""Comprehensive tests for all AURA modules."""

import pytest
from unittest.mock import AsyncMock, MagicMock, patch
from datetime import datetime
from aura import AuraClient, AuraWallet
from aura.types import (
    ChainConfig, VCParams, VCType, BridgeTransferParams, ComplianceCheckParams,
    ComplianceLevel, ConfidenceScoreParams, KeyRotationParams, KeyType,
    DataItemParams, DataItemType, AccessLevel, DynamicFeeParams,
    IdentityChangeParams, IdentityChangeType, InclusionRoutineParams,
    RoutinePriority, PrevalidationParams, ValidationLevel, PrivacyParams,
    PrivacyLevel, MixingStrategy, RingSignatureParams, JailingParams,
    SessionParams, WalletSecurityParams, SecurityFeature
)


@pytest.fixture
def mock_client():
    """Create a mock AuraClient."""
    config = ChainConfig(
        chain_id="aura-testnet-1",
        rpc_endpoint="http://localhost:26657",
        rest_endpoint="http://localhost:1317"
    )
    client = AuraClient(config)
    client.get = AsyncMock()
    client.post = AsyncMock()
    return client


@pytest.fixture
def mock_wallet():
    """Create a mock wallet."""
    wallet = MagicMock(spec=AuraWallet)
    wallet.address = "aura1test123"
    return wallet


class TestVCRegistryModule:
    """Tests for VCRegistry module."""

    @pytest.mark.asyncio
    async def test_mint_vc(self, mock_client, mock_wallet):
        """Test minting a verifiable credential."""
        await mock_client.connect_wallet(mock_wallet)
        mock_tx_builder = MagicMock()
        mock_tx_builder.sign_and_broadcast = AsyncMock(return_value={"tx_hash": "test"})
        mock_client._tx_builder = mock_tx_builder

        params = VCParams(
            issuer="aura1issuer",
            subject="aura1subject",
            vc_type=VCType.IDENTITY,
            claims={"name": "John Doe"}
        )

        result = await mock_client.vcregistry.mint_vc(params)
        assert result["tx_hash"] == "test"

    @pytest.mark.asyncio
    async def test_verify_vc(self, mock_client):
        """Test verifying a credential."""
        mock_client.get.return_value = {
            "vc_id": "vc123",
            "valid": True,
            "status": "active",
            "issuer_verified": True,
            "signature_valid": True,
            "not_expired": True,
            "not_revoked": True,
            "verified_at": datetime.now().isoformat()
        }

        result = await mock_client.vcregistry.verify_vc("vc123")
        assert result.valid is True
        assert result.vc_id == "vc123"


class TestBridgeModule:
    """Tests for Bridge module."""

    @pytest.mark.asyncio
    async def test_lock_tokens(self, mock_client, mock_wallet):
        """Test locking tokens for bridge transfer."""
        await mock_client.connect_wallet(mock_wallet)
        mock_tx_builder = MagicMock()
        mock_tx_builder.sign_and_broadcast = AsyncMock(return_value={"tx_hash": "test"})
        mock_client._tx_builder = mock_tx_builder

        params = BridgeTransferParams(
            source_chain="aura",
            destination_chain="ethereum",
            token="uaura",
            amount="1000000",
            recipient="0x123"
        )

        result = await mock_client.bridge.lock_tokens(params)
        assert result["tx_hash"] == "test"

    @pytest.mark.asyncio
    async def test_get_bridge_stats(self, mock_client):
        """Test getting bridge statistics."""
        mock_client.get.return_value = {
            "total_transfers": 100,
            "pending_transfers": 5,
            "completed_transfers": 90,
            "failed_transfers": 5,
            "total_volume": "1000000",
            "total_fees": "10000",
            "active_relayers": 3
        }

        stats = await mock_client.bridge.get_bridge_stats()
        assert stats["total_transfers"] == 100
        assert stats["active_relayers"] == 3


class TestComplianceModule:
    """Tests for Compliance module."""

    @pytest.mark.asyncio
    async def test_get_kyc_status(self, mock_client):
        """Test getting KYC status."""
        mock_client.get.return_value = {
            "kyc_status": {
                "address": "aura1test",
                "kyc_verified": True,
                "aml_cleared": True,
                "verification_date": datetime.now().isoformat()
            }
        }

        status = await mock_client.compliance.get_kyc_status("aura1test")
        assert status.kyc_verified is True
        assert status.aml_cleared is True

    @pytest.mark.asyncio
    async def test_check_sanctions(self, mock_client):
        """Test sanctions check."""
        mock_client.get.return_value = {
            "is_sanctioned": False,
            "lists_matched": [],
            "checked_at": datetime.now().isoformat()
        }

        result = await mock_client.compliance.check_sanctions("aura1test")
        assert result.is_sanctioned is False


class TestConfidenceScoreModule:
    """Tests for ConfidenceScore module."""

    @pytest.mark.asyncio
    async def test_get_confidence_score(self, mock_client):
        """Test getting confidence score."""
        mock_client.get.return_value = {
            "score": {
                "address": "aura1test",
                "score": 95.5,
                "rank": 10,
                "total_participants": 1000,
                "last_updated": datetime.now().isoformat()
            }
        }

        score = await mock_client.confidencescore.get_confidence_score("aura1test")
        assert score.score == 95.5
        assert score.rank == 10

    @pytest.mark.asyncio
    async def test_get_leaderboard(self, mock_client):
        """Test getting leaderboard."""
        mock_client.get.return_value = {
            "leaderboard": [
                {
                    "address": "aura1top",
                    "score": 99.0,
                    "rank": 1,
                    "total_participants": 1000,
                    "last_updated": datetime.now().isoformat()
                }
            ]
        }

        leaderboard = await mock_client.confidencescore.get_leaderboard(limit=10)
        assert len(leaderboard) == 1
        assert leaderboard[0].rank == 1


class TestCryptographyModule:
    """Tests for Cryptography module."""

    @pytest.mark.asyncio
    async def test_generate_keypair(self, mock_client):
        """Test generating keypair."""
        mock_client.get.return_value = {
            "keypair": {
                "public_key": "pubkey123",
                "private_key": "privkey123",
                "key_type": "ed25519",
                "created_at": datetime.now().isoformat()
            }
        }

        keypair = await mock_client.cryptography.generate_keypair()
        assert keypair.public_key == "pubkey123"
        assert keypair.key_type == KeyType.ED25519

    @pytest.mark.asyncio
    async def test_encrypt_decrypt(self, mock_client):
        """Test encryption and decryption."""
        from aura.types import EncryptionParams, EncryptionAlgorithm, DecryptionParams

        # Test encryption
        mock_client.post.return_value = {
            "encrypted": {
                "ciphertext": "encrypted_data",
                "algorithm": "aes256",
                "nonce": "nonce123"
            }
        }

        enc_params = EncryptionParams(
            data="test data",
            algorithm=EncryptionAlgorithm.AES256
        )
        encrypted = await mock_client.cryptography.encrypt_data(enc_params)
        assert encrypted.ciphertext == "encrypted_data"

        # Test decryption
        mock_client.post.return_value = {"decrypted_data": "test data"}

        dec_params = DecryptionParams(
            encrypted_data="encrypted_data",
            algorithm=EncryptionAlgorithm.AES256,
            private_key="privkey123"
        )
        decrypted = await mock_client.cryptography.decrypt_data(dec_params)
        assert decrypted == "test data"


class TestDataRegistryModule:
    """Tests for DataRegistry module."""

    @pytest.mark.asyncio
    async def test_register_data(self, mock_client, mock_wallet):
        """Test registering data."""
        await mock_client.connect_wallet(mock_wallet)
        mock_tx_builder = MagicMock()
        mock_tx_builder.sign_and_broadcast = AsyncMock(return_value={"tx_hash": "test"})
        mock_client._tx_builder = mock_tx_builder

        params = DataItemParams(
            owner="aura1owner",
            name="test_data",
            data_type=DataItemType.TEXT,
            content="test content",
            access_level=AccessLevel.PUBLIC
        )

        result = await mock_client.dataregistry.register_data(params)
        assert result["tx_hash"] == "test"

    @pytest.mark.asyncio
    async def test_search_data(self, mock_client):
        """Test searching data."""
        mock_client.get.return_value = {
            "results": [
                {
                    "id": "data123",
                    "owner": "aura1owner",
                    "name": "test",
                    "data_type": "text",
                    "content_hash": "hash123",
                    "access_level": "public",
                    "created_at": datetime.now().isoformat(),
                    "updated_at": datetime.now().isoformat(),
                    "size": 100,
                    "version": 1
                }
            ]
        }

        results = await mock_client.dataregistry.search_data("test")
        assert len(results) == 1
        assert results[0].name == "test"


class TestEconomicSecurityModule:
    """Tests for EconomicSecurity module."""

    @pytest.mark.asyncio
    async def test_get_dynamic_fees(self, mock_client):
        """Test getting dynamic fees."""
        mock_client.post.return_value = {
            "fee_structure": {
                "base_fee": "100",
                "priority_fee": "50",
                "total_fee": "150",
                "gas_price": "0.01",
                "estimated_time": 5
            }
        }

        params = DynamicFeeParams(
            transaction_size=1000,
            priority=1,
            network_congestion=0.5
        )

        fees = await mock_client.economicsecurity.get_dynamic_fees(params)
        assert fees.total_fee == "150"

    @pytest.mark.asyncio
    async def test_get_economic_stats(self, mock_client):
        """Test getting economic stats."""
        mock_client.get.return_value = {
            "metrics": {
                "total_value_locked": "1000000",
                "circulating_supply": "500000",
                "inflation_rate": 0.05,
                "staking_ratio": 0.6,
                "average_fee": "100",
                "mev_prevented": "5000",
                "timestamp": datetime.now().isoformat()
            }
        }

        stats = await mock_client.economicsecurity.get_economic_stats()
        assert stats.inflation_rate == 0.05


class TestIdentityChangeModule:
    """Tests for IdentityChange module."""

    @pytest.mark.asyncio
    async def test_create_identity(self, mock_client, mock_wallet):
        """Test creating identity."""
        await mock_client.connect_wallet(mock_wallet)
        mock_tx_builder = MagicMock()
        mock_tx_builder.sign_and_broadcast = AsyncMock(return_value={"tx_hash": "test"})
        mock_client._tx_builder = mock_tx_builder

        result = await mock_client.identitychange.create_identity("aura1test")
        assert result["tx_hash"] == "test"

    @pytest.mark.asyncio
    async def test_get_identity(self, mock_client):
        """Test getting identity."""
        mock_client.get.return_value = {
            "identity": {
                "address": "aura1test",
                "verified": True,
                "verification_level": 2,
                "created_at": datetime.now().isoformat(),
                "updated_at": datetime.now().isoformat(),
                "attestations": []
            }
        }

        identity = await mock_client.identitychange.get_identity("aura1test")
        assert identity.verified is True


class TestInclusionRoutinesModule:
    """Tests for InclusionRoutines module."""

    @pytest.mark.asyncio
    async def test_create_routine(self, mock_client, mock_wallet):
        """Test creating routine."""
        await mock_client.connect_wallet(mock_wallet)
        mock_tx_builder = MagicMock()
        mock_tx_builder.sign_and_broadcast = AsyncMock(return_value={"tx_hash": "test"})
        mock_client._tx_builder = mock_tx_builder

        params = InclusionRoutineParams(
            creator="aura1creator",
            name="test_routine",
            description="Test routine",
            code="function() { return true; }",
            priority=RoutinePriority.MEDIUM,
            max_gas=100000
        )

        result = await mock_client.inclusionroutines.create_routine(params)
        assert result["tx_hash"] == "test"

    @pytest.mark.asyncio
    async def test_get_registry_stats(self, mock_client):
        """Test getting registry stats."""
        mock_client.get.return_value = {
            "stats": {
                "total_routines": 50,
                "active_routines": 40,
                "total_executions": 1000,
                "average_success_rate": 0.95,
                "top_routines": []
            }
        }

        stats = await mock_client.inclusionroutines.get_registry_stats()
        assert stats.total_routines == 50


class TestMonitoringModule:
    """Tests for Monitoring module."""

    @pytest.mark.asyncio
    async def test_get_system_metrics(self, mock_client):
        """Test getting system metrics."""
        mock_client.get.return_value = {
            "health": {
                "overall_status": "healthy",
                "components": {},
                "uptime": 3600,
                "last_check": datetime.now().isoformat(),
                "active_alerts": 0,
                "cpu_usage": 50.0,
                "memory_usage": 60.0,
                "disk_usage": 40.0,
                "network_latency": 10.0
            }
        }

        health = await mock_client.monitoring.get_system_metrics()
        assert health.cpu_usage == 50.0

    @pytest.mark.asyncio
    async def test_get_performance_stats(self, mock_client):
        """Test getting performance stats."""
        mock_client.get.return_value = {
            "metrics": {
                "average_block_time": 6.0,
                "transactions_per_second": 100.0,
                "average_gas_price": "0.01",
                "network_hashrate": 1000.0,
                "active_validators": 100,
                "timestamp": datetime.now().isoformat()
            }
        }

        stats = await mock_client.monitoring.get_performance_stats()
        assert stats.transactions_per_second == 100.0


class TestNetworkSecurityModule:
    """Tests for NetworkSecurity module."""

    @pytest.mark.asyncio
    async def test_get_reputation_score(self, mock_client):
        """Test getting reputation score."""
        mock_client.get.return_value = {
            "reputation": {
                "peer_id": "peer123",
                "ip_address": "1.2.3.4",
                "reputation_score": 95.0,
                "successful_interactions": 100,
                "failed_interactions": 5,
                "last_seen": datetime.now().isoformat(),
                "is_trusted": True,
                "is_blacklisted": False
            }
        }

        rep = await mock_client.networksecurity.get_reputation_score("peer123")
        assert rep.reputation_score == 95.0

    @pytest.mark.asyncio
    async def test_check_fork_detection(self, mock_client):
        """Test fork detection."""
        mock_client.get.return_value = {
            "fork": {
                "fork_detected": False,
                "resolution_status": "none"
            }
        }

        fork = await mock_client.networksecurity.check_fork_detection()
        assert fork.fork_detected is False


class TestPrevalidationModule:
    """Tests for Prevalidation module."""

    @pytest.mark.asyncio
    async def test_prevalidate_transaction(self, mock_client):
        """Test prevalidating transaction."""
        mock_client.post.return_value = {
            "result": {
                "status": "valid",
                "valid": True,
                "errors": [],
                "warnings": [],
                "estimated_gas": 21000,
                "estimated_fee": "100",
                "validated_at": datetime.now().isoformat(),
                "checks_performed": ["balance", "nonce", "signature"]
            }
        }

        params = PrevalidationParams(
            transaction_data="tx_data",
            validation_level=ValidationLevel.STANDARD
        )

        result = await mock_client.prevalidation.prevalidate_transaction(params)
        assert result.valid is True

    @pytest.mark.asyncio
    async def test_estimate_gas(self, mock_client):
        """Test gas estimation."""
        mock_client.post.return_value = {
            "estimation": {
                "base_gas": 21000,
                "computation_gas": 5000,
                "storage_gas": 2000,
                "total_gas": 28000,
                "gas_price": "0.01",
                "total_fee": "280",
                "confidence": 0.9
            }
        }

        estimation = await mock_client.prevalidation.estimate_gas("tx_data")
        assert estimation.total_gas == 28000


class TestPrivacyModule:
    """Tests for Privacy module."""

    @pytest.mark.asyncio
    async def test_generate_ring_signature(self, mock_client):
        """Test generating ring signature."""
        mock_client.post.return_value = {
            "signature": {
                "signature": "sig123",
                "ring_members": ["addr1", "addr2", "addr3"],
                "key_image": "key123",
                "ring_size": 3,
                "created_at": datetime.now().isoformat(),
                "verified": False
            }
        }

        params = RingSignatureParams(
            message="test message",
            ring_members=["addr1", "addr2", "addr3"],
            key_image="key123",
            ring_size=3
        )

        sig = await mock_client.privacy.generate_ring_signature(params)
        assert sig.ring_size == 3

    @pytest.mark.asyncio
    async def test_verify_ring_signature(self, mock_client):
        """Test verifying ring signature."""
        mock_client.post.return_value = {"valid": True}

        valid = await mock_client.privacy.verify_ring_signature(
            "sig123", "message", ["addr1", "addr2"], "key123"
        )
        assert valid is True


class TestValidatorSecurityModule:
    """Tests for ValidatorSecurity module."""

    @pytest.mark.asyncio
    async def test_get_validator_status(self, mock_client):
        """Test getting validator status."""
        mock_client.get.return_value = {
            "validator": {
                "operator_address": "auravaloper1test",
                "consensus_address": "auravalcons1test",
                "status": "active",
                "jailed": False,
                "tombstoned": False,
                "commission_rate": "0.1",
                "voting_power": 1000,
                "uptime": 99.9,
                "last_active": datetime.now().isoformat()
            }
        }

        status = await mock_client.validatorsecurity.get_validator_status("auravaloper1test")
        assert status.uptime == 99.9

    @pytest.mark.asyncio
    async def test_get_slashing_history(self, mock_client):
        """Test getting slashing history."""
        mock_client.get.return_value = {
            "events": []
        }

        history = await mock_client.validatorsecurity.get_slashing_history("auravaloper1test")
        assert isinstance(history, list)


class TestWalletSecurityModule:
    """Tests for WalletSecurity module."""

    @pytest.mark.asyncio
    async def test_enable_multisig(self, mock_client, mock_wallet):
        """Test enabling multisig."""
        await mock_client.connect_wallet(mock_wallet)
        mock_tx_builder = MagicMock()
        mock_tx_builder.sign_and_broadcast = AsyncMock(return_value={"tx_hash": "test"})
        mock_client._tx_builder = mock_tx_builder

        result = await mock_client.walletsecurity.enable_multisig(
            "aura1test", 2, ["aura1signer1", "aura1signer2"]
        )
        assert result["tx_hash"] == "test"

    @pytest.mark.asyncio
    async def test_get_security_settings(self, mock_client):
        """Test getting security settings."""
        mock_client.get.return_value = {
            "settings": {
                "address": "aura1test",
                "security_level": 3,
                "session_timeout": 3600,
                "require_biometric": True,
                "require_2fa": True,
                "require_confirmation": True,
                "whitelist_enabled": False,
                "whitelisted_addresses": []
            }
        }

        settings = await mock_client.walletsecurity.get_security_settings("aura1test")
        assert settings.security_level == 3


# Run all tests
if __name__ == "__main__":
    pytest.main([__file__, "-v", "--asyncio-mode=auto"])
