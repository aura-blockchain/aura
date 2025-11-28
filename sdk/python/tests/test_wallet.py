"""Tests for wallet functionality."""

import pytest
from aura import AuraWallet


def test_generate_mnemonic():
    """Test mnemonic generation."""
    mnemonic = AuraWallet.generate_mnemonic()
    words = mnemonic.split()
    assert len(words) == 24
    assert AuraWallet.validate_mnemonic(mnemonic)


def test_validate_mnemonic():
    """Test mnemonic validation."""
    valid = AuraWallet.generate_mnemonic()
    assert AuraWallet.validate_mnemonic(valid)

    assert not AuraWallet.validate_mnemonic("invalid mnemonic")
    assert not AuraWallet.validate_mnemonic("")


def test_from_mnemonic():
    """Test wallet creation from mnemonic."""
    mnemonic = AuraWallet.generate_mnemonic()
    wallet = AuraWallet("paw")
    wallet.from_mnemonic(mnemonic)

    assert wallet.address.startswith("paw")
    assert len(wallet.public_key) > 0


def test_invalid_mnemonic():
    """Test wallet creation with invalid mnemonic."""
    wallet = AuraWallet("paw")
    with pytest.raises(ValueError):
        wallet.from_mnemonic("invalid mnemonic phrase")


def test_sign_message():
    """Test message signing."""
    mnemonic = AuraWallet.generate_mnemonic()
    wallet = AuraWallet("paw")
    wallet.from_mnemonic(mnemonic)

    message = b"test message"
    signature = wallet.sign(message)
    assert len(signature) > 0


def test_export_mnemonic():
    """Test mnemonic export."""
    mnemonic = AuraWallet.generate_mnemonic()
    wallet = AuraWallet("paw")
    wallet.from_mnemonic(mnemonic)

    exported = wallet.export_mnemonic()
    assert exported == mnemonic
