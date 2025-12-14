/**
 * Biometric Authentication Service
 * Provides fingerprint and Face ID authentication
 * Uses react-native-biometrics for secure biometric operations
 */

import ReactNativeBiometrics from 'react-native-biometrics';

const rnBiometrics = new ReactNativeBiometrics();

class BiometricAuthService {
  constructor() {
    this.isAvailable = false;
    this.biometryType = null;
  }

  /**
   * Check if biometric authentication is available
   * @returns {Promise<Object>} {available: boolean, biometryType: string}
   */
  async checkAvailability() {
    try {
      const result = await rnBiometrics.isSensorAvailable();
      this.isAvailable = result.available;
      this.biometryType = result.biometryType;

      return {
        available: result.available,
        biometryType: result.biometryType,
      };
    } catch (error) {
      console.error('Error checking biometric availability:', error);
      this.isAvailable = false;
      this.biometryType = null;
      return {available: false, biometryType: null};
    }
  }

  /**
   * Get human-readable name for biometry type
   * @returns {string} Biometry type name
   */
  getBiometryTypeName() {
    switch (this.biometryType) {
      case 'TouchID':
        return 'Touch ID';
      case 'FaceID':
        return 'Face ID';
      case 'Biometrics':
        return 'Biometric';
      default:
        return 'Biometric';
    }
  }

  /**
   * Authenticate user with biometrics
   * @param {string} promptMessage - Message to display to user
   * @returns {Promise<boolean>} True if authentication successful
   */
  async authenticate(promptMessage = 'Authenticate to continue') {
    try {
      // Check availability first
      const availability = await this.checkAvailability();
      if (!availability.available) {
        throw new Error('Biometric authentication is not available');
      }

      const result = await rnBiometrics.simplePrompt({
        promptMessage,
        cancelButtonText: 'Cancel',
      });

      return result.success;
    } catch (error) {
      console.error('Biometric authentication error:', error);
      return false;
    }
  }

  /**
   * Create biometric keys for signature-based authentication
   * @returns {Promise<Object>} {publicKey: string}
   */
  async createKeys() {
    try {
      const result = await rnBiometrics.createKeys();
      return {publicKey: result.publicKey};
    } catch (error) {
      console.error('Error creating biometric keys:', error);
      throw new Error('Failed to create biometric keys');
    }
  }

  /**
   * Delete biometric keys
   * @returns {Promise<boolean>} True if deletion successful
   */
  async deleteKeys() {
    try {
      const result = await rnBiometrics.deleteKeys();
      return result.keysDeleted;
    } catch (error) {
      console.error('Error deleting biometric keys:', error);
      throw new Error('Failed to delete biometric keys');
    }
  }

  /**
   * Check if biometric keys exist
   * @returns {Promise<boolean>} True if keys exist
   */
  async keysExist() {
    try {
      const result = await rnBiometrics.biometricKeysExist();
      return result.keysExist;
    } catch (error) {
      console.error('Error checking biometric keys:', error);
      return false;
    }
  }

  /**
   * Create a signature with biometric authentication
   * @param {string} payload - Data to sign
   * @param {string} promptMessage - Message to display
   * @returns {Promise<Object>} {success: boolean, signature: string}
   */
  async createSignature(payload, promptMessage = 'Sign to authenticate') {
    try {
      const result = await rnBiometrics.createSignature({
        promptMessage,
        payload,
        cancelButtonText: 'Cancel',
      });

      if (result.success) {
        return {
          success: true,
          signature: result.signature,
        };
      } else {
        throw new Error('Signature creation cancelled');
      }
    } catch (error) {
      console.error('Error creating signature:', error);
      throw new Error('Failed to create biometric signature');
    }
  }
}

// Export singleton instance
const BiometricAuth = new BiometricAuthService();
export default BiometricAuth;
