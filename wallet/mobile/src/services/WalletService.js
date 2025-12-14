/**
 * Wallet Service
 * High-level wallet operations: create, import, balance, transactions
 * Orchestrates KeyStore, BiometricAuth, and PawAPI services
 */

import {generateWallet, importWalletFromMnemonic, validateMnemonic} from '../utils/crypto';
import KeyStore from './KeyStore';
import BiometricAuth from './BiometricAuth';
import PawAPI from './PawAPI';

class WalletServiceClass {
  /**
   * Create a new wallet
   * @param {Object} options - {walletName, password, useBiometric}
   * @returns {Promise<Object>} {address, publicKey, mnemonic}
   */
  async createWallet({walletName, password, useBiometric = false}) {
    try {
      // Validate password
      if (!password || password.length < 8) {
        throw new Error('Password must be at least 8 characters');
      }

      // Generate new wallet
      const wallet = generateWallet();

      // Store wallet credentials securely
      await KeyStore.storeWallet(wallet.privateKey, wallet.mnemonic, password);

      // Store metadata
      const metadata = {
        name: walletName || 'My Wallet',
        address: wallet.address,
        createdAt: new Date().toISOString(),
        biometricEnabled: false,
      };

      // Setup biometric if requested
      if (useBiometric) {
        const availability = await BiometricAuth.checkAvailability();
        if (availability.available) {
          await BiometricAuth.createKeys();
          metadata.biometricEnabled = true;
        }
      }

      await KeyStore.storeMetadata(metadata);

      return {
        address: wallet.address,
        publicKey: wallet.publicKey,
        mnemonic: wallet.mnemonic,
      };
    } catch (error) {
      console.error('Error creating wallet:', error);
      throw error;
    }
  }

  /**
   * Import existing wallet from mnemonic
   * @param {Object} options - {mnemonic, walletName, password, useBiometric}
   * @returns {Promise<Object>} {address, publicKey}
   */
  async importWallet({mnemonic, walletName, password, useBiometric = false}) {
    try {
      // Validate inputs
      if (!validateMnemonic(mnemonic)) {
        throw new Error('Invalid mnemonic phrase');
      }

      if (!password || password.length < 8) {
        throw new Error('Password must be at least 8 characters');
      }

      // Import wallet from mnemonic
      const wallet = importWalletFromMnemonic(mnemonic);

      // Store wallet credentials
      await KeyStore.storeWallet(wallet.privateKey, wallet.mnemonic, password);

      // Store metadata
      const metadata = {
        name: walletName || 'Imported Wallet',
        address: wallet.address,
        createdAt: new Date().toISOString(),
        imported: true,
        biometricEnabled: false,
      };

      // Setup biometric if requested
      if (useBiometric) {
        const availability = await BiometricAuth.checkAvailability();
        if (availability.available) {
          await BiometricAuth.createKeys();
          metadata.biometricEnabled = true;
        }
      }

      await KeyStore.storeMetadata(metadata);

      return {
        address: wallet.address,
        publicKey: wallet.publicKey,
      };
    } catch (error) {
      console.error('Error importing wallet:', error);
      throw error;
    }
  }

  /**
   * Get wallet balance
   * @returns {Promise<Object>} {amount, formatted, denom}
   */
  async getBalance() {
    try {
      const address = await KeyStore.getAddress();
      if (!address) {
        throw new Error('No wallet found');
      }

      const balanceData = await PawAPI.getBalance(address);
      const balances = balanceData.balances || [];

      // Find uaura balance (micro-aura)
      const auraBalance = balances.find(b => b.denom === 'uaura');
      const amount = auraBalance ? parseInt(auraBalance.amount, 10) : 0;

      // Convert micro-aura to aura (1 aura = 1,000,000 uaura)
      const formatted = (amount / 1000000).toFixed(6);

      return {
        amount,
        formatted,
        denom: 'Aura',
      };
    } catch (error) {
      console.error('Error getting balance:', error);
      throw error;
    }
  }

  /**
   * Get transaction history
   * @param {number} limit - Max transactions to fetch
   * @returns {Promise<Array>} Transaction array
   */
  async getTransactionHistory(limit = 20) {
    try {
      const address = await KeyStore.getAddress();
      if (!address) {
        throw new Error('No wallet found');
      }

      const transactions = await PawAPI.getTransactionsByAddress(address, limit);

      // Cache transactions locally
      await KeyStore.storeTransactions(transactions);

      return transactions;
    } catch (error) {
      console.error('Error getting transaction history:', error);

      // Return cached transactions on network error
      const cached = await KeyStore.retrieveTransactions();
      return cached;
    }
  }

  /**
   * Check if wallet exists
   * @returns {Promise<boolean>} True if wallet exists
   */
  async hasWallet() {
    return await KeyStore.hasWallet();
  }

  /**
   * Get wallet information
   * @returns {Promise<Object>} Wallet info
   */
  async getWalletInfo() {
    try {
      const metadata = await KeyStore.retrieveMetadata();
      const address = await KeyStore.getAddress();
      const name = await KeyStore.getName();

      return {
        address,
        name,
        createdAt: metadata?.createdAt,
        biometricEnabled: metadata?.biometricEnabled || false,
        imported: metadata?.imported || false,
      };
    } catch (error) {
      console.error('Error getting wallet info:', error);
      throw error;
    }
  }

  /**
   * Delete wallet
   * @returns {Promise<boolean>} True if deletion successful
   */
  async deleteWallet() {
    try {
      // Delete biometric keys if they exist
      const keysExist = await BiometricAuth.keysExist();
      if (keysExist) {
        await BiometricAuth.deleteKeys();
      }

      // Clear all wallet data
      await KeyStore.clearAll();

      return true;
    } catch (error) {
      console.error('Error deleting wallet:', error);
      throw error;
    }
  }

  /**
   * Unlock wallet with password
   * @param {string} password - Wallet password
   * @returns {Promise<Object>} {privateKey, mnemonic}
   */
  async unlockWallet(password) {
    try {
      const wallet = await KeyStore.retrieveWallet(password);
      if (!wallet) {
        throw new Error('Invalid password');
      }
      return wallet;
    } catch (error) {
      console.error('Error unlocking wallet:', error);
      throw error;
    }
  }

  /**
   * Unlock wallet with biometric
   * @returns {Promise<Object>} {privateKey, mnemonic}
   */
  async unlockWithBiometric() {
    try {
      const metadata = await KeyStore.retrieveMetadata();
      if (!metadata?.biometricEnabled) {
        throw new Error('Biometric authentication not enabled');
      }

      const authenticated = await BiometricAuth.authenticate(
        'Unlock your wallet',
      );

      if (!authenticated) {
        throw new Error('Biometric authentication failed');
      }

      // For biometric, we still need the password stored somewhere
      // In a real implementation, you'd use the biometric signature
      // to decrypt a stored password. For this demo, we'll throw an error.
      throw new Error(
        'Biometric unlock requires password for full security',
      );
    } catch (error) {
      console.error('Error unlocking with biometric:', error);
      throw error;
    }
  }
}

// Export singleton instance
const WalletService = new WalletServiceClass();
export default WalletService;
