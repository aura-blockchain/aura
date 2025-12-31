import { assertBech32Prefix, validateFee, normalizePath } from '../src/services/hardware/guards';

describe('Hardware Wallet Guards', () => {
  describe('assertBech32Prefix', () => {
    it('should accept valid aura addresses', () => {
      const validAddress = 'aura1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqg4ha9y';
      expect(() => assertBech32Prefix(validAddress, 'aura')).not.toThrow();
    });

    it('should reject addresses with wrong prefix', () => {
      // cosmos address with valid checksum
      const cosmosAddress = 'cosmos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a';
      expect(() => assertBech32Prefix(cosmosAddress, 'aura')).toThrow('Address prefix mismatch');
    });

    it('should reject invalid bech32 addresses', () => {
      expect(() => assertBech32Prefix('invalid', 'aura')).toThrow();
    });
  });

  describe('validateFee', () => {
    it('should accept valid fee with uaura', () => {
      const fee = {
        amount: [{ denom: 'uaura', amount: '5000' }],
        gas: '200000',
      };
      expect(() => validateFee(fee, ['uaura'])).not.toThrow();
    });

    it('should reject fee with missing amount', () => {
      const fee = { gas: '200000' };
      expect(() => validateFee(fee as any, ['uaura'])).toThrow('Fee amount required');
    });

    it('should reject fee with empty amount array', () => {
      const fee = { amount: [], gas: '200000' };
      expect(() => validateFee(fee, ['uaura'])).toThrow('Fee amount required');
    });

    it('should reject fee with invalid denom', () => {
      const fee = {
        amount: [{ denom: 'uatom', amount: '5000' }],
        gas: '200000',
      };
      expect(() => validateFee(fee, ['uaura'])).toThrow('not permitted');
    });

    it('should reject fee with invalid gas', () => {
      const fee = {
        amount: [{ denom: 'uaura', amount: '5000' }],
        gas: '0',
      };
      expect(() => validateFee(fee, ['uaura'])).toThrow('Invalid gas');
    });

    it('should reject fee with NaN gas', () => {
      const fee = {
        amount: [{ denom: 'uaura', amount: '5000' }],
        gas: 'invalid',
      };
      expect(() => validateFee(fee, ['uaura'])).toThrow('Invalid gas');
    });

    it('should reject negative fee amount', () => {
      const fee = {
        amount: [{ denom: 'uaura', amount: '-100' }],
        gas: '200000',
      };
      expect(() => validateFee(fee, ['uaura'])).toThrow('non-negative');
    });
  });

  describe('normalizePath', () => {
    it('should normalize path with m/ prefix', () => {
      const result = normalizePath("m/44'/118'/0'/0/0", 4);
      expect(result).toBe("44'/118'/0'/0/0");
    });

    it('should normalize path without m/ prefix', () => {
      const result = normalizePath("44'/118'/0'/0/0", 4);
      expect(result).toBe("44'/118'/0'/0/0");
    });

    it('should reject path with too many segments', () => {
      expect(() => normalizePath("m/44'/118'/0'/0/0/0", 4)).toThrow('Invalid derivation path');
    });

    it('should reject path with too few segments', () => {
      expect(() => normalizePath("m/44'/118'/0'", 4)).toThrow('Invalid derivation path');
    });

    it('should reject account index exceeding max', () => {
      expect(() => normalizePath("m/44'/118'/5'/0/0", 4)).toThrow('exceeds maximum');
    });

    it('should accept account index at max', () => {
      const result = normalizePath("m/44'/118'/4'/0/0", 4);
      expect(result).toBe("44'/118'/4'/0/0");
    });

    it('should reject non-hardened first three segments', () => {
      expect(() => normalizePath("m/44/118'/0'/0/0", 4)).toThrow('must be hardened');
    });

    it('should reject invalid segment values', () => {
      expect(() => normalizePath("m/44'/abc'/0'/0/0", 4)).toThrow('Invalid path segment');
    });
  });
});
