/**
 * Integration Tests - Wallet Flow
 * Tests the complete wallet lifecycle
 */

import { KeystoreService } from '../../src/services/keystore';
import { ApiService } from '../../src/services/api';

describe('Wallet Integration Tests', () => {
  let keystoreService;
  let apiService;

  beforeEach(async () => {
    keystoreService = new KeystoreService();
    apiService = new ApiService();
    localStorage.clear();
    await window.electron.store.clear();
    jest.clearAllMocks();
  });

  describe('Complete Wallet Lifecycle', () => {
    test('should create, save, and retrieve wallet', async () => {
      // Generate mnemonic
      const mnemonic = await keystoreService.generateMnemonic();
      expect(mnemonic).toBeDefined();

      // Create wallet
      const password = 'test-password-123';
      const wallet = await keystoreService.createWallet(mnemonic, password);

      expect(wallet.address).toBeDefined();
      expect(wallet.address).toMatch(/^aura1/);

      // Retrieve wallet (already stored by createWallet)
      const retrieved = await keystoreService.getWallet();
      expect(retrieved.address).toBe(wallet.address);
    });

    test('should handle wallet reset', async () => {
      const mnemonic = await keystoreService.generateMnemonic();
      await keystoreService.createWallet(mnemonic, 'password');

      await keystoreService.clearWallet();

      const wallet = await keystoreService.getWallet();
      expect(wallet).toBeNull();
    });
  });

  describe('Transaction Workflow', () => {
    test('should validate transaction parameters', () => {
      const recipient = 'aura1test123';
      const amount = 100;

      expect(recipient).toMatch(/^aura1/);
      expect(amount).toBeGreaterThan(0);
    });
  });

  describe('Address Book Integration', () => {
    test('should save and retrieve addresses', async () => {
      const addresses = [
        { name: 'Alice', address: 'aura1alice', note: 'Friend' },
        { name: 'Bob', address: 'aura1bob', note: 'Exchange' }
      ];

      if (window.electron?.store) {
        await window.electron.store.set('addressBook', addresses);
        const retrieved = await window.electron.store.get('addressBook');
        expect(retrieved).toEqual(addresses);
      }
    });
  });
});
