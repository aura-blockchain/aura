/**
 * KeyStore Service
 * Secure storage for wallet credentials using react-native-keychain
 * Never stores private keys in AsyncStorage - only in secure OS keychain
 */

import * as Keychain from 'react-native-keychain';
import AsyncStorage from '@react-native-async-storage/async-storage';
import {encrypt, decrypt} from '../utils/crypto';

const KEYCHAIN_SERVICE = 'com.aura.wallet';
const STORAGE_PREFIX = '@Aura:';

class KeyStoreService {
  constructor() {
    this.isInitialized = false;
  }

  /**
   * Initialize the KeyStore service
   * @returns {Promise<boolean>} True if initialization successful
   */
  async initialize() {
    try {
      this.isInitialized = true;
      return true;
    } catch (error) {
      console.error('KeyStore initialization error:', error);
      return false;
    }
  }

  /**
   * Store wallet credentials securely in OS keychain
   * @param {string} privateKey - Private key (hex)
   * @param {string} mnemonic - Mnemonic phrase
   * @param {string} password - Encryption password
   * @returns {Promise<boolean>} True if storage successful
   */
  async storeWallet(privateKey, mnemonic, password) {
    try {
      // Encrypt sensitive data with password
      const encryptedPrivateKey = encrypt(privateKey, password);
      const encryptedMnemonic = encrypt(mnemonic, password);

      // Store encrypted data in OS keychain (never AsyncStorage!)
      const credentials = JSON.stringify({
        privateKey: encryptedPrivateKey,
        mnemonic: encryptedMnemonic,
      });

      await Keychain.setGenericPassword('aura_wallet', credentials, {
        service: KEYCHAIN_SERVICE,
        accessible: Keychain.ACCESSIBLE.WHEN_UNLOCKED_THIS_DEVICE_ONLY,
      });

      return true;
    } catch (error) {
      console.error('Error storing wallet:', error);
      throw new Error('Failed to store wallet credentials');
    }
  }

  /**
   * Retrieve wallet credentials from keychain
   * @param {string} password - Decryption password
   * @returns {Promise<Object>} {privateKey, mnemonic} or null
   */
  async retrieveWallet(password) {
    try {
      const credentials = await Keychain.getGenericPassword({
        service: KEYCHAIN_SERVICE,
      });

      if (!credentials) {
        return null;
      }

      const data = JSON.parse(credentials.password);

      // Decrypt with password
      const privateKey = decrypt(data.privateKey, password);
      const mnemonic = decrypt(data.mnemonic, password);

      // Verify decryption was successful
      if (!privateKey || !mnemonic) {
        throw new Error('Invalid password');
      }

      return {privateKey, mnemonic};
    } catch (error) {
      console.error('Error retrieving wallet:', error);
      throw new Error('Failed to retrieve wallet credentials');
    }
  }

  /**
   * Check if wallet exists
   * @returns {Promise<boolean>} True if wallet exists
   */
  async hasWallet() {
    try {
      const credentials = await Keychain.getGenericPassword({
        service: KEYCHAIN_SERVICE,
      });
      return !!credentials;
    } catch (error) {
      console.error('Error checking wallet existence:', error);
      return false;
    }
  }

  /**
   * Delete wallet from keychain
   * @returns {Promise<boolean>} True if deletion successful
   */
  async deleteWallet() {
    try {
      await Keychain.resetGenericPassword({service: KEYCHAIN_SERVICE});
      await AsyncStorage.removeItem(`${STORAGE_PREFIX}metadata`);
      await AsyncStorage.removeItem(`${STORAGE_PREFIX}address`);
      await AsyncStorage.removeItem(`${STORAGE_PREFIX}name`);
      return true;
    } catch (error) {
      console.error('Error deleting wallet:', error);
      throw new Error('Failed to delete wallet');
    }
  }

  /**
   * Store wallet metadata (non-sensitive data in AsyncStorage)
   * @param {Object} metadata - Wallet metadata
   * @returns {Promise<boolean>} True if storage successful
   */
  async storeMetadata(metadata) {
    try {
      await AsyncStorage.setItem(
        `${STORAGE_PREFIX}metadata`,
        JSON.stringify(metadata),
      );

      // Store commonly accessed fields separately for quick access
      if (metadata.address) {
        await AsyncStorage.setItem(`${STORAGE_PREFIX}address`, metadata.address);
      }
      if (metadata.name) {
        await AsyncStorage.setItem(`${STORAGE_PREFIX}name`, metadata.name);
      }

      return true;
    } catch (error) {
      console.error('Error storing metadata:', error);
      throw new Error('Failed to store wallet metadata');
    }
  }

  /**
   * Retrieve wallet metadata
   * @returns {Promise<Object>} Metadata object or null
   */
  async retrieveMetadata() {
    try {
      const metadataStr = await AsyncStorage.getItem(`${STORAGE_PREFIX}metadata`);
      return metadataStr ? JSON.parse(metadataStr) : null;
    } catch (error) {
      console.error('Error retrieving metadata:', error);
      return null;
    }
  }

  /**
   * Get wallet address
   * @returns {Promise<string>} Wallet address or null
   */
  async getAddress() {
    try {
      return await AsyncStorage.getItem(`${STORAGE_PREFIX}address`);
    } catch (error) {
      console.error('Error getting address:', error);
      return null;
    }
  }

  /**
   * Get wallet name
   * @returns {Promise<string>} Wallet name or default
   */
  async getName() {
    try {
      const name = await AsyncStorage.getItem(`${STORAGE_PREFIX}name`);
      return name || 'My Wallet';
    } catch (error) {
      console.error('Error getting name:', error);
      return 'My Wallet';
    }
  }

  /**
   * Store transactions (for caching)
   * @param {Array} transactions - Transaction array
   * @returns {Promise<boolean>} True if storage successful
   */
  async storeTransactions(transactions) {
    try {
      await AsyncStorage.setItem(
        `${STORAGE_PREFIX}transactions`,
        JSON.stringify(transactions),
      );
      return true;
    } catch (error) {
      console.error('Error storing transactions:', error);
      return false;
    }
  }

  /**
   * Retrieve stored transactions
   * @returns {Promise<Array>} Transaction array
   */
  async retrieveTransactions() {
    try {
      const txsStr = await AsyncStorage.getItem(`${STORAGE_PREFIX}transactions`);
      return txsStr ? JSON.parse(txsStr) : [];
    } catch (error) {
      console.error('Error retrieving transactions:', error);
      return [];
    }
  }

  /**
   * Clear all wallet data (keychain + AsyncStorage)
   * @returns {Promise<boolean>} True if clearing successful
   */
  async clearAll() {
    try {
      await Keychain.resetGenericPassword({service: KEYCHAIN_SERVICE});
      await AsyncStorage.clear();
      return true;
    } catch (error) {
      console.error('Error clearing all data:', error);
      throw new Error('Failed to clear wallet data');
    }
  }
}

// Export singleton instance
const KeyStore = new KeyStoreService();
export default KeyStore;
