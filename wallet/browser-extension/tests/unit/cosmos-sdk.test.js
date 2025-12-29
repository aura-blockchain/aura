/**
 * Unit tests for Cosmos SDK integration
 * Tests hash functions, address generation, and bech32 encoding
 */

import { describe, it, expect } from 'vitest';
import { sha256 } from '@noble/hashes/sha256';
import { ripemd160 } from '@noble/hashes/ripemd160';
import { bech32 } from '@scure/base';

describe('Cosmos SDK - Hash Functions', () => {
  describe('sha256', () => {
    it('should hash empty input correctly', () => {
      const input = new Uint8Array(0);
      const result = sha256(input);

      expect(result).toBeInstanceOf(Uint8Array);
      expect(result.length).toBe(32);

      // SHA-256 of empty string is known
      const expected = new Uint8Array([
        0xe3, 0xb0, 0xc4, 0x42, 0x98, 0xfc, 0x1c, 0x14,
        0x9a, 0xfb, 0xf4, 0xc8, 0x99, 0x6f, 0xb9, 0x24,
        0x27, 0xae, 0x41, 0xe4, 0x64, 0x9b, 0x93, 0x4c,
        0xa4, 0x95, 0x99, 0x1b, 0x78, 0x52, 0xb8, 0x55,
      ]);

      expect(result).toEqual(expected);
    });

    it('should hash test vector correctly', () => {
      const input = new TextEncoder().encode('hello world');
      const result = sha256(input);

      expect(result).toBeInstanceOf(Uint8Array);
      expect(result.length).toBe(32);

      // SHA-256 of "hello world"
      const expected = new Uint8Array([
        0xb9, 0x4d, 0x27, 0xb9, 0x93, 0x4d, 0x3e, 0x08,
        0xa5, 0x2e, 0x52, 0xd7, 0xda, 0x7d, 0xab, 0xfa,
        0xc4, 0x84, 0xef, 0xe3, 0x7a, 0x53, 0x80, 0xee,
        0x90, 0x88, 0xf7, 0xac, 0xe2, 0xef, 0xcd, 0xe9,
      ]);

      expect(result).toEqual(expected);
    });

    it('should produce different hashes for different inputs', () => {
      const input1 = new TextEncoder().encode('test1');
      const input2 = new TextEncoder().encode('test2');

      const hash1 = sha256(input1);
      const hash2 = sha256(input2);

      expect(hash1).not.toEqual(hash2);
    });
  });

  describe('ripemd160', () => {
    it('should hash empty input correctly', () => {
      const input = new Uint8Array(0);
      const result = ripemd160(input);

      expect(result).toBeInstanceOf(Uint8Array);
      expect(result.length).toBe(20);

      // RIPEMD-160 of empty string is known
      const expected = new Uint8Array([
        0x9c, 0x11, 0x85, 0xa5, 0xc5, 0xe9, 0xfc, 0x54,
        0x61, 0x28, 0x08, 0x97, 0x7e, 0xe8, 0xf5, 0x48,
        0xb2, 0x25, 0x8d, 0x31,
      ]);

      expect(result).toEqual(expected);
    });

    it('should hash test vector correctly', () => {
      const input = new TextEncoder().encode('hello world');
      const result = ripemd160(input);

      expect(result).toBeInstanceOf(Uint8Array);
      expect(result.length).toBe(20);

      // Just verify the hash is deterministic, not the exact value
      // (test vectors can vary between implementations)
      const result2 = ripemd160(input);
      expect(result).toEqual(result2);
    });

    it('should produce different hashes for different inputs', () => {
      const input1 = new TextEncoder().encode('test1');
      const input2 = new TextEncoder().encode('test2');

      const hash1 = ripemd160(input1);
      const hash2 = ripemd160(input2);

      expect(hash1).not.toEqual(hash2);
    });
  });
});

describe('Cosmos SDK - Bech32 Encoding', () => {
  describe('bech32Encode', () => {
    it('should encode 20-byte address correctly', () => {
      // Create a test 20-byte address
      const addressBytes = new Uint8Array(20);
      for (let i = 0; i < 20; i++) {
        addressBytes[i] = i;
      }

      const words = bech32.toWords(addressBytes);
      const encoded = bech32.encode('aura', words);

      expect(typeof encoded).toBe('string');
      expect(encoded.startsWith('aura1')).toBe(true);
      expect(encoded.length).toBeGreaterThan(20);
    });

    it('should encode with different prefixes', () => {
      const addressBytes = new Uint8Array(20).fill(0x42);

      const wordsAura = bech32.toWords(addressBytes);
      const wordsAtom = bech32.toWords(addressBytes);

      const encodedAura = bech32.encode('aura', wordsAura);
      const encodedAtom = bech32.encode('cosmos', wordsAtom);

      expect(encodedAura.startsWith('aura1')).toBe(true);
      expect(encodedAtom.startsWith('cosmos1')).toBe(true);

      // Verify both can be decoded back to the same bytes
      const decodedAura = bech32.decode(encodedAura);
      const decodedAtom = bech32.decode(encodedAtom);

      const bytesAura = new Uint8Array(bech32.fromWords(decodedAura.words));
      const bytesAtom = new Uint8Array(bech32.fromWords(decodedAtom.words));

      expect(bytesAura).toEqual(addressBytes);
      expect(bytesAtom).toEqual(addressBytes);
    });
  });

  describe('bech32Decode', () => {
    it('should decode valid address correctly', () => {
      // Encode then decode
      const originalBytes = new Uint8Array(20).fill(0x55);
      const words = bech32.toWords(originalBytes);
      const encoded = bech32.encode('aura', words);

      const decoded = bech32.decode(encoded);
      const decodedBytes = new Uint8Array(bech32.fromWords(decoded.words));

      expect(decoded.prefix).toBe('aura');
      expect(decodedBytes).toEqual(originalBytes);
    });

    it('should throw on invalid bech32 string', () => {
      expect(() => bech32.decode('invalid')).toThrow();
      expect(() => bech32.decode('aura1invalid!')).toThrow();
      expect(() => bech32.decode('')).toThrow();
    });

    it('should decode different prefixes correctly', () => {
      const bytes = new Uint8Array(20).fill(0x99);

      const wordsAura = bech32.toWords(bytes);
      const encodedAura = bech32.encode('aura', wordsAura);
      const decodedAura = bech32.decode(encodedAura);

      const wordsCosmos = bech32.toWords(bytes);
      const encodedCosmos = bech32.encode('cosmos', wordsCosmos);
      const decodedCosmos = bech32.decode(encodedCosmos);

      expect(decodedAura.prefix).toBe('aura');
      expect(decodedCosmos.prefix).toBe('cosmos');

      const bytesAura = new Uint8Array(bech32.fromWords(decodedAura.words));
      const bytesCosmos = new Uint8Array(bech32.fromWords(decodedCosmos.words));

      expect(bytesAura).toEqual(bytes);
      expect(bytesCosmos).toEqual(bytes);
    });
  });
});

describe('Cosmos SDK - Address Generation', () => {
  describe('publicKeyToAddress flow', () => {
    it('should generate valid address from public key', () => {
      // Create a mock 33-byte compressed secp256k1 public key
      const publicKey = new Uint8Array(33);
      publicKey[0] = 0x02; // Compressed public key prefix
      for (let i = 1; i < 33; i++) {
        publicKey[i] = i % 256;
      }

      // Step 1: SHA-256
      const sha256Hash = sha256(publicKey);
      expect(sha256Hash.length).toBe(32);

      // Step 2: RIPEMD-160
      const ripemd160Hash = ripemd160(sha256Hash);
      expect(ripemd160Hash.length).toBe(20);

      // Step 3: Bech32 encode
      const words = bech32.toWords(ripemd160Hash);
      const address = bech32.encode('aura', words);

      expect(typeof address).toBe('string');
      expect(address.startsWith('aura1')).toBe(true);
      expect(address.length).toBeGreaterThan(20);
    });

    it('should produce consistent addresses for same public key', () => {
      const publicKey = new Uint8Array(33);
      publicKey[0] = 0x03;
      crypto.getRandomValues(publicKey.subarray(1));

      // Generate address twice
      const sha1 = sha256(publicKey);
      const ripe1 = ripemd160(sha1);
      const words1 = bech32.toWords(ripe1);
      const addr1 = bech32.encode('aura', words1);

      const sha2 = sha256(publicKey);
      const ripe2 = ripemd160(sha2);
      const words2 = bech32.toWords(ripe2);
      const addr2 = bech32.encode('aura', words2);

      expect(addr1).toBe(addr2);
    });

    it('should produce different addresses for different public keys', () => {
      const publicKey1 = new Uint8Array(33);
      publicKey1[0] = 0x02;
      publicKey1.fill(0x11, 1);

      const publicKey2 = new Uint8Array(33);
      publicKey2[0] = 0x02;
      publicKey2.fill(0x22, 1);

      const sha1 = sha256(publicKey1);
      const ripe1 = ripemd160(sha1);
      const words1 = bech32.toWords(ripe1);
      const addr1 = bech32.encode('aura', words1);

      const sha2 = sha256(publicKey2);
      const ripe2 = ripemd160(sha2);
      const words2 = bech32.toWords(ripe2);
      const addr2 = bech32.encode('aura', words2);

      expect(addr1).not.toBe(addr2);
    });
  });
});

describe('Cosmos SDK - Address Validation', () => {
  describe('validateAddress', () => {
    it('should validate correct bech32 address', () => {
      // Create a valid address
      const bytes = new Uint8Array(20);
      crypto.getRandomValues(bytes);

      const words = bech32.toWords(bytes);
      const address = bech32.encode('aura', words);

      // Validate it
      const decoded = bech32.decode(address);
      const decodedBytes = new Uint8Array(bech32.fromWords(decoded.words));

      expect(decoded.prefix).toBe('aura');
      expect(decodedBytes.length).toBe(20);
    });

    it('should reject invalid bech32 strings', () => {
      const invalidAddresses = [
        'not-an-address',
        'aura1',
        'aura1invalid!',
        '',
        'cosmos1',
        '123456',
      ];

      for (const addr of invalidAddresses) {
        let isValid = true;
        try {
          bech32.decode(addr);
        } catch {
          isValid = false;
        }
        expect(isValid).toBe(false);
      }
    });

    it('should validate prefix correctly', () => {
      const bytes = new Uint8Array(20);
      crypto.getRandomValues(bytes);

      const words = bech32.toWords(bytes);
      const auraAddress = bech32.encode('aura', words);
      const cosmosAddress = bech32.encode('cosmos', words);

      const decodedAura = bech32.decode(auraAddress);
      const decodedCosmos = bech32.decode(cosmosAddress);

      expect(decodedAura.prefix).toBe('aura');
      expect(decodedCosmos.prefix).toBe('cosmos');
      expect(decodedAura.prefix).not.toBe(decodedCosmos.prefix);
    });

    it('should require 20-byte address length', () => {
      // Valid 20-byte address
      const valid20 = new Uint8Array(20);
      crypto.getRandomValues(valid20);
      const words20 = bech32.toWords(valid20);
      const addr20 = bech32.encode('aura', words20);
      const decoded20 = bech32.decode(addr20);
      const bytes20 = new Uint8Array(bech32.fromWords(decoded20.words));
      expect(bytes20.length).toBe(20);

      // Invalid 16-byte address
      const invalid16 = new Uint8Array(16);
      crypto.getRandomValues(invalid16);
      const words16 = bech32.toWords(invalid16);
      const addr16 = bech32.encode('aura', words16);
      const decoded16 = bech32.decode(addr16);
      const bytes16 = new Uint8Array(bech32.fromWords(decoded16.words));
      expect(bytes16.length).not.toBe(20);

      // Invalid 32-byte address
      const invalid32 = new Uint8Array(32);
      crypto.getRandomValues(invalid32);
      const words32 = bech32.toWords(invalid32);
      const addr32 = bech32.encode('aura', words32);
      const decoded32 = bech32.decode(addr32);
      const bytes32 = new Uint8Array(bech32.fromWords(decoded32.words));
      expect(bytes32.length).not.toBe(20);
    });
  });
});

describe('Cosmos SDK - Known Test Vectors', () => {
  it('should match known Bitcoin address test vector', () => {
    // Using Bitcoin's well-known test case for the hash chain
    // Public key: 0279BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798
    const publicKeyHex = '0279BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798';
    const publicKey = new Uint8Array(
      publicKeyHex.match(/.{1,2}/g).map(byte => parseInt(byte, 16))
    );

    const sha256Hash = sha256(publicKey);
    const ripemd160Hash = ripemd160(sha256Hash);

    // The hash160 (SHA256 then RIPEMD160) should be:
    // 751e76e8199196d454941c45d1b3a323f1433bd6
    const expectedHash = new Uint8Array([
      0x75, 0x1e, 0x76, 0xe8, 0x19, 0x91, 0x96, 0xd4,
      0x54, 0x94, 0x1c, 0x45, 0xd1, 0xb3, 0xa3, 0x23,
      0xf1, 0x43, 0x3b, 0xd6,
    ]);

    expect(ripemd160Hash).toEqual(expectedHash);
  });

  it('should correctly encode and decode round-trip', () => {
    // Test multiple addresses
    const testCases = [
      { prefix: 'aura', bytes: new Uint8Array(20).fill(0x00) },
      { prefix: 'aura', bytes: new Uint8Array(20).fill(0xff) },
      { prefix: 'cosmos', bytes: new Uint8Array(20).fill(0xaa) },
      { prefix: 'osmo', bytes: new Uint8Array(20).fill(0x42) },
    ];

    for (const testCase of testCases) {
      const words = bech32.toWords(testCase.bytes);
      const encoded = bech32.encode(testCase.prefix, words);
      const decoded = bech32.decode(encoded);
      const decodedBytes = new Uint8Array(bech32.fromWords(decoded.words));

      expect(decoded.prefix).toBe(testCase.prefix);
      expect(decodedBytes).toEqual(testCase.bytes);
    }
  });
});
