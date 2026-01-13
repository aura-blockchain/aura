/**
 * Storage Service
 * General-purpose storage for app settings and non-sensitive data
 * Uses AsyncStorage for persistent data (NOT for private keys - use KeyStore for that)
 */

import AsyncStorage from '@react-native-async-storage/async-storage';
const {CHAIN_CONFIG, REST_ENDPOINTS} = require('../../../config/chain');

const STORAGE_PREFIX = '@Aura:';

class StorageServiceClass {
  /**
   * Store an item
   * @param {string} key - Storage key
   * @param {*} value - Value to store (will be JSON stringified)
   * @returns {Promise<boolean>} True if successful
   */
  async setItem(key, value) {
    try {
      const jsonValue = JSON.stringify(value);
      await AsyncStorage.setItem(key, jsonValue);
      return true;
    } catch (error) {
      console.error('Error storing item:', error);
      return false;
    }
  }

  /**
   * Get an item
   * @param {string} key - Storage key
   * @param {*} defaultValue - Default value if not found
   * @returns {Promise<*>} Stored value or default
   */
  async getItem(key, defaultValue = null) {
    try {
      const jsonValue = await AsyncStorage.getItem(key);
      return jsonValue != null ? JSON.parse(jsonValue) : defaultValue;
    } catch (error) {
      console.error('Error retrieving item:', error);
      return defaultValue;
    }
  }

  /**
   * Remove an item
   * @param {string} key - Storage key
   * @returns {Promise<boolean>} True if successful
   */
  async removeItem(key) {
    try {
      await AsyncStorage.removeItem(key);
      return true;
    } catch (error) {
      console.error('Error removing item:', error);
      return false;
    }
  }

  /**
   * Get app theme
   * @returns {Promise<string>} Theme name (dark/light)
   */
  async getTheme() {
    return await this.getItem(`${STORAGE_PREFIX}theme`, 'dark');
  }

  /**
   * Set app theme
   * @param {string} theme - Theme name
   * @returns {Promise<boolean>} True if successful
   */
  async setTheme(theme) {
    return await this.setItem(`${STORAGE_PREFIX}theme`, theme);
  }

  /**
   * Get network configuration
   * @returns {Promise<Object>} Network config
   */
  async getNetwork() {
    return await this.getItem(`${STORAGE_PREFIX}network`, {
      name: CHAIN_CONFIG.chainName || 'Aura',
      rpcUrl: REST_ENDPOINTS[0]?.address || 'http://localhost:1317',
      chainId: CHAIN_CONFIG.chainId || 'aura-mvp-1',
    });
  }

  /**
   * Set network configuration
   * @param {Object} network - Network config
   * @returns {Promise<boolean>} True if successful
   */
  async setNetwork(network) {
    return await this.setItem(`${STORAGE_PREFIX}network`, network);
  }

  /**
   * Get address book
   * @returns {Promise<Array>} Address book entries
   */
  async getAddressBook() {
    return await this.getItem(`${STORAGE_PREFIX}addressBook`, []);
  }

  /**
   * Add address to address book
   * @param {Object} entry - {name, address, note}
   * @returns {Promise<Object>} Added entry with id
   */
  async addAddress(entry) {
    try {
      const addressBook = await this.getAddressBook();
      const newEntry = {
        ...entry,
        id: Date.now().toString(),
        createdAt: new Date().toISOString(),
      };
      addressBook.push(newEntry);
      await this.setItem(`${STORAGE_PREFIX}addressBook`, addressBook);
      return newEntry;
    } catch (error) {
      console.error('Error adding address:', error);
      throw error;
    }
  }

  /**
   * Remove address from address book
   * @param {string} id - Entry id
   * @returns {Promise<boolean>} True if successful
   */
  async removeAddress(id) {
    try {
      const addressBook = await this.getAddressBook();
      const filtered = addressBook.filter(entry => entry.id !== id);
      return await this.setItem(`${STORAGE_PREFIX}addressBook`, filtered);
    } catch (error) {
      console.error('Error removing address:', error);
      return false;
    }
  }

  /**
   * Update address in address book
   * @param {string} id - Entry id
   * @param {Object} updates - Fields to update
   * @returns {Promise<Object>} Updated entry
   */
  async updateAddress(id, updates) {
    try {
      const addressBook = await this.getAddressBook();
      const index = addressBook.findIndex(entry => entry.id === id);
      if (index >= 0) {
        addressBook[index] = {...addressBook[index], ...updates};
        await this.setItem(`${STORAGE_PREFIX}addressBook`, addressBook);
        return addressBook[index];
      }
      throw new Error('Address not found');
    } catch (error) {
      console.error('Error updating address:', error);
      throw error;
    }
  }

  /**
   * Get recent addresses
   * @returns {Promise<Array>} Recent address list
   */
  async getRecentAddresses() {
    return await this.getItem(`${STORAGE_PREFIX}recentAddresses`, []);
  }

  /**
   * Add recent address
   * @param {string} address - Address to add
   * @returns {Promise<boolean>} True if successful
   */
  async addRecentAddress(address) {
    try {
      const recent = await this.getRecentAddresses();
      const filtered = recent.filter(addr => addr !== address);
      filtered.unshift(address);
      const limited = filtered.slice(0, 10); // Keep only 10 most recent
      return await this.setItem(`${STORAGE_PREFIX}recentAddresses`, limited);
    } catch (error) {
      console.error('Error adding recent address:', error);
      return false;
    }
  }

  /**
   * Get price alerts
   * @returns {Promise<Array>} Price alert list
   */
  async getPriceAlerts() {
    return await this.getItem(`${STORAGE_PREFIX}priceAlerts`, []);
  }

  /**
   * Add price alert
   * @param {Object} alert - {type: 'above'|'below', price: number}
   * @returns {Promise<Object>} Added alert with id
   */
  async addPriceAlert(alert) {
    try {
      const alerts = await this.getPriceAlerts();
      const newAlert = {
        ...alert,
        id: Date.now().toString(),
        enabled: true,
        createdAt: new Date().toISOString(),
      };
      alerts.push(newAlert);
      await this.setItem(`${STORAGE_PREFIX}priceAlerts`, alerts);
      return newAlert;
    } catch (error) {
      console.error('Error adding price alert:', error);
      throw error;
    }
  }

  /**
   * Remove price alert
   * @param {string} id - Alert id
   * @returns {Promise<boolean>} True if successful
   */
  async removePriceAlert(id) {
    try {
      const alerts = await this.getPriceAlerts();
      const filtered = alerts.filter(alert => alert.id !== id);
      return await this.setItem(`${STORAGE_PREFIX}priceAlerts`, filtered);
    } catch (error) {
      console.error('Error removing price alert:', error);
      return false;
    }
  }

  /**
   * Toggle price alert enabled state
   * @param {string} id - Alert id
   * @returns {Promise<Object>} Updated alert
   */
  async togglePriceAlert(id) {
    try {
      const alerts = await this.getPriceAlerts();
      const index = alerts.findIndex(alert => alert.id === id);
      if (index >= 0) {
        alerts[index].enabled = !alerts[index].enabled;
        await this.setItem(`${STORAGE_PREFIX}priceAlerts`, alerts);
        return alerts[index];
      }
      throw new Error('Alert not found');
    } catch (error) {
      console.error('Error toggling price alert:', error);
      throw error;
    }
  }

  /**
   * Get notification settings
   * @returns {Promise<Object>} Notification settings
   */
  async getNotificationSettings() {
    return await this.getItem(`${STORAGE_PREFIX}notifications`, {
      enabled: true,
      transactions: true,
      priceAlerts: true,
      security: true,
    });
  }

  /**
   * Set notification settings
   * @param {Object} settings - Notification settings
   * @returns {Promise<boolean>} True if successful
   */
  async setNotificationSettings(settings) {
    return await this.setItem(`${STORAGE_PREFIX}notifications`, settings);
  }

  /**
   * Get security settings
   * @returns {Promise<Object>} Security settings
   */
  async getSecuritySettings() {
    return await this.getItem(`${STORAGE_PREFIX}security`, {
      biometricEnabled: false,
      autoLockTimeout: 300, // 5 minutes
      requirePasswordForSend: true,
    });
  }

  /**
   * Set security settings
   * @param {Object} settings - Security settings
   * @returns {Promise<boolean>} True if successful
   */
  async setSecuritySettings(settings) {
    return await this.setItem(`${STORAGE_PREFIX}security`, settings);
  }
}

// Export singleton instance
const StorageService = new StorageServiceClass();
export default StorageService;
