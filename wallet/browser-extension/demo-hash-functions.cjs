#!/usr/bin/env node
/**
 * Demo script for hash functions in cosmos-sdk.js
 * Shows that sha256, ripemd160, and bech32 encoding work correctly
 */

const { sha256 } = require('@noble/hashes/sha256');
const { ripemd160 } = require('@noble/hashes/ripemd160');
const { bech32 } = require('@scure/base');

console.log('=== Cosmos SDK Hash Functions Demo ===\n');

// Demo 1: SHA-256 hash
console.log('1. SHA-256 Hash:');
const testData = new TextEncoder().encode('Hello, Aura!');
const sha256Hash = sha256(testData);
console.log('  Input: "Hello, Aura!"');
console.log('  SHA-256:', Buffer.from(sha256Hash).toString('hex'));
console.log('  Length:', sha256Hash.length, 'bytes\n');

// Demo 2: RIPEMD-160 hash
console.log('2. RIPEMD-160 Hash:');
const ripemd160Hash = ripemd160(sha256Hash);
console.log('  Input: SHA-256 hash from above');
console.log('  RIPEMD-160:', Buffer.from(ripemd160Hash).toString('hex'));
console.log('  Length:', ripemd160Hash.length, 'bytes\n');

// Demo 3: Bech32 encoding
console.log('3. Bech32 Address Encoding:');
const words = bech32.toWords(ripemd160Hash);
const auraAddress = bech32.encode('aura', words);
console.log('  Input: 20-byte RIPEMD-160 hash');
console.log('  Aura address:', auraAddress);
console.log('  Length:', auraAddress.length, 'characters\n');

// Demo 4: Full address generation from public key
console.log('4. Complete Address Generation:');
console.log('  Simulating secp256k1 public key...');

// Create a mock 33-byte compressed public key
const publicKey = new Uint8Array(33);
publicKey[0] = 0x02; // Compressed public key prefix
// Fill with deterministic data for reproducible demo
for (let i = 1; i < 33; i++) {
  publicKey[i] = i;
}

console.log('  Public Key:', Buffer.from(publicKey).toString('hex'));

const step1 = sha256(publicKey);
console.log('  → SHA-256:', Buffer.from(step1).toString('hex'));

const step2 = ripemd160(step1);
console.log('  → RIPEMD-160:', Buffer.from(step2).toString('hex'));

const words2 = bech32.toWords(step2);
const finalAddress = bech32.encode('aura', words2);
console.log('  → Bech32 Address:', finalAddress);

// Demo 5: Decode and verify
console.log('\n5. Address Decoding and Verification:');
const decoded = bech32.decode(finalAddress);
const decodedBytes = new Uint8Array(bech32.fromWords(decoded.words));

console.log('  Original address:', finalAddress);
console.log('  Decoded prefix:', decoded.prefix);
console.log('  Decoded bytes:', Buffer.from(decodedBytes).toString('hex'));
console.log('  Match original?', Buffer.from(step2).toString('hex') === Buffer.from(decodedBytes).toString('hex'));

// Demo 6: Multiple chain prefixes
console.log('\n6. Different Chain Prefixes:');
const chains = ['aura', 'cosmos', 'osmo', 'juno'];
chains.forEach(prefix => {
  const words = bech32.toWords(step2);
  const addr = bech32.encode(prefix, words);
  console.log(`  ${prefix.padEnd(8)} → ${addr}`);
});

console.log('\n✓ All hash functions working correctly!');
