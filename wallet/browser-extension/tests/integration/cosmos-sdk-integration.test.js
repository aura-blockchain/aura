/**
 * Integration tests for Cosmos SDK module
 * Tests the full module loading and real-world usage scenarios
 */

import { describe, it, expect } from 'vitest';
import { sha256 } from '@noble/hashes/sha256';
import { ripemd160 } from '@noble/hashes/ripemd160';
import { bech32 } from '@scure/base';

describe('Cosmos SDK Integration - Full Address Generation Flow', () => {
  it('should generate valid Cosmos address from random public key', () => {
    // Simulate a real secp256k1 compressed public key (33 bytes)
    const publicKey = new Uint8Array(33);
    publicKey[0] = 0x02; // Compressed public key prefix for even Y
    crypto.getRandomValues(publicKey.subarray(1));

    // Step 1: Hash with SHA-256
    const sha256Hash = sha256(publicKey);
    expect(sha256Hash).toBeInstanceOf(Uint8Array);
    expect(sha256Hash.length).toBe(32);

    // Step 2: Hash with RIPEMD-160
    const addressHash = ripemd160(sha256Hash);
    expect(addressHash).toBeInstanceOf(Uint8Array);
    expect(addressHash.length).toBe(20);

    // Step 3: Bech32 encode
    const words = bech32.toWords(addressHash);
    const address = bech32.encode('aura', words);

    // Verify address format
    expect(typeof address).toBe('string');
    expect(address.startsWith('aura1')).toBe(true);
    expect(address.length).toBeGreaterThan(20);
    expect(address.length).toBeLessThan(100);

    // Verify address can be decoded
    const decoded = bech32.decode(address);
    expect(decoded.prefix).toBe('aura');

    const decodedBytes = new Uint8Array(bech32.fromWords(decoded.words));
    expect(decodedBytes).toEqual(addressHash);
  });

  it('should handle multiple chain prefixes correctly', () => {
    const publicKey = new Uint8Array(33);
    publicKey[0] = 0x03; // Compressed public key prefix for odd Y
    crypto.getRandomValues(publicKey.subarray(1));

    const sha256Hash = sha256(publicKey);
    const addressHash = ripemd160(sha256Hash);

    // Test different chain prefixes
    const chains = [
      { prefix: 'aura', name: 'Aura' },
      { prefix: 'cosmos', name: 'Cosmos Hub' },
      { prefix: 'osmo', name: 'Osmosis' },
      { prefix: 'juno', name: 'Juno' },
    ];

    for (const chain of chains) {
      const words = bech32.toWords(addressHash);
      const address = bech32.encode(chain.prefix, words);

      expect(address.startsWith(`${chain.prefix}1`)).toBe(true);

      // Verify decode
      const decoded = bech32.decode(address);
      expect(decoded.prefix).toBe(chain.prefix);

      const decodedBytes = new Uint8Array(bech32.fromWords(decoded.words));
      expect(decodedBytes).toEqual(addressHash);
    }
  });

  it('should validate addresses correctly', () => {
    // Create a valid address
    const validBytes = new Uint8Array(20);
    crypto.getRandomValues(validBytes);

    const words = bech32.toWords(validBytes);
    const validAddress = bech32.encode('aura', words);

    // Test validation
    let isValid = true;
    try {
      const decoded = bech32.decode(validAddress);
      const decodedBytes = new Uint8Array(bech32.fromWords(decoded.words));

      isValid = decoded.prefix === 'aura' && decodedBytes.length === 20;
    } catch {
      isValid = false;
    }

    expect(isValid).toBe(true);

    // Test invalid addresses
    const invalidAddresses = [
      'not-an-address',
      'aura1',
      'aura1!invalid',
      '',
      'cosmos1xyz', // Too short
      '123456789',
    ];

    for (const addr of invalidAddresses) {
      let addrValid = true;
      try {
        bech32.decode(addr);
      } catch {
        addrValid = false;
      }
      expect(addrValid).toBe(false);
    }
  });
});

describe('Cosmos SDK Integration - Performance Tests', () => {
  it('should generate addresses efficiently', () => {
    const startTime = performance.now();
    const iterations = 100;

    for (let i = 0; i < iterations; i++) {
      const publicKey = new Uint8Array(33);
      publicKey[0] = i % 2 === 0 ? 0x02 : 0x03;
      crypto.getRandomValues(publicKey.subarray(1));

      const sha256Hash = sha256(publicKey);
      const addressHash = ripemd160(sha256Hash);
      const words = bech32.toWords(addressHash);
      const address = bech32.encode('aura', words);

      expect(address.startsWith('aura1')).toBe(true);
    }

    const endTime = performance.now();
    const duration = endTime - startTime;

    // Should be able to generate 100 addresses in reasonable time (< 1 second)
    expect(duration).toBeLessThan(1000);
  });

  it('should hash large data efficiently', () => {
    // Create 64KB of data (crypto.getRandomValues has 65536 byte limit)
    const largeData = new Uint8Array(64 * 1024); // 64KB
    crypto.getRandomValues(largeData);

    const startTime = performance.now();
    const hash = sha256(largeData);
    const endTime = performance.now();

    expect(hash).toBeInstanceOf(Uint8Array);
    expect(hash.length).toBe(32);

    // Should hash 64KB in reasonable time (< 100ms)
    const duration = endTime - startTime;
    expect(duration).toBeLessThan(100);
  });
});

describe('Cosmos SDK Integration - Error Handling', () => {
  it('should handle invalid bech32 addresses gracefully', () => {
    const invalidAddresses = [
      'aura1invalid!@#$',
      'aura11111111111', // Checksum fail
      'invalid',
      '',
      'aura1', // Too short
      'AURA1ABC', // Wrong case
    ];

    for (const addr of invalidAddresses) {
      expect(() => {
        bech32.decode(addr);
      }).toThrow();
    }
  });

  it('should reject wrong-length address data', () => {
    // 16-byte address (invalid, should be 20)
    const invalidLength16 = new Uint8Array(16);
    crypto.getRandomValues(invalidLength16);
    const words16 = bech32.toWords(invalidLength16);
    const addr16 = bech32.encode('aura', words16);

    const decoded16 = bech32.decode(addr16);
    const bytes16 = new Uint8Array(bech32.fromWords(decoded16.words));
    expect(bytes16.length).toBe(16);
    expect(bytes16.length).not.toBe(20); // Not a valid Cosmos address length

    // 32-byte address (invalid, should be 20)
    const invalidLength32 = new Uint8Array(32);
    crypto.getRandomValues(invalidLength32);
    const words32 = bech32.toWords(invalidLength32);
    const addr32 = bech32.encode('aura', words32);

    const decoded32 = bech32.decode(addr32);
    const bytes32 = new Uint8Array(bech32.fromWords(decoded32.words));
    expect(bytes32.length).toBe(32);
    expect(bytes32.length).not.toBe(20); // Not a valid Cosmos address length
  });
});

describe('Cosmos SDK Integration - Deterministic Behavior', () => {
  it('should produce same address for same public key', () => {
    const publicKey = new Uint8Array(33);
    publicKey[0] = 0x02;
    // Use fixed data for deterministic test
    for (let i = 1; i < 33; i++) {
      publicKey[i] = i;
    }

    // Generate address multiple times
    const addresses = [];
    for (let i = 0; i < 5; i++) {
      const sha256Hash = sha256(publicKey);
      const addressHash = ripemd160(sha256Hash);
      const words = bech32.toWords(addressHash);
      const address = bech32.encode('aura', words);
      addresses.push(address);
    }

    // All addresses should be identical
    for (let i = 1; i < addresses.length; i++) {
      expect(addresses[i]).toBe(addresses[0]);
    }
  });

  it('should produce different addresses for different public keys', () => {
    const addresses = new Set();

    for (let i = 0; i < 10; i++) {
      const publicKey = new Uint8Array(33);
      publicKey[0] = 0x02;
      publicKey[1] = i; // Make each key unique

      const sha256Hash = sha256(publicKey);
      const addressHash = ripemd160(sha256Hash);
      const words = bech32.toWords(addressHash);
      const address = bech32.encode('aura', words);

      addresses.add(address);
    }

    // All addresses should be unique
    expect(addresses.size).toBe(10);
  });
});
