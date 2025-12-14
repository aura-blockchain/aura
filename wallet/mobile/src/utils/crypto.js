/**
 * Cryptographic Utilities
 * Implements BIP39 mnemonic generation, BIP32 HD wallet derivation, and BIP44 account discovery
 * Security: Uses industry-standard libraries for all cryptographic operations
 */

import * as bip39 from 'bip39';
import {ec as EC} from 'elliptic';
import {sha256} from 'js-sha256';
import RIPEMD160 from 'ripemd160';
import {bech32} from 'bech32';
import CryptoJS from 'crypto-js';

const secp256k1 = new EC('secp256k1');

/**
 * Generate a BIP39 mnemonic phrase
 * @param {number} strength - Entropy strength in bits (128 for 12 words, 256 for 24 words)
 * @returns {string} Mnemonic phrase
 */
export function generateMnemonic(strength = 256) {
  return bip39.generateMnemonic(strength);
}

/**
 * Validate a BIP39 mnemonic phrase
 * @param {string} mnemonic - Mnemonic phrase to validate
 * @returns {boolean} True if valid
 */
export function validateMnemonic(mnemonic) {
  return bip39.validateMnemonic(mnemonic);
}

/**
 * Derive a private key from a mnemonic using BIP32/BIP44
 * @param {string} mnemonic - BIP39 mnemonic phrase
 * @param {number} accountIndex - BIP44 account index (default: 0)
 * @param {number} addressIndex - BIP44 address index (default: 0)
 * @returns {string} Private key (hex)
 */
export function derivePrivateKeyFromMnemonic(
  mnemonic,
  accountIndex = 0,
  addressIndex = 0,
) {
  if (!validateMnemonic(mnemonic)) {
    throw new Error('Invalid mnemonic phrase');
  }

  // BIP39: mnemonic to seed
  const seed = bip39.mnemonicToSeedSync(mnemonic);

  // BIP32: derive master key from seed
  const masterKey = deriveMasterKey(seed);

  // BIP44 path: m/44'/118'/0'/0/0
  // 118 is the Cosmos coin type
  const path = `m/44'/118'/${accountIndex}'/0/${addressIndex}`;
  const childKey = deriveChildKey(masterKey, path);

  return childKey.privateKey.toString('hex');
}

/**
 * Derive master key from seed (BIP32)
 * @param {Buffer} seed - Seed bytes
 * @returns {Object} Master key with privateKey and chainCode
 */
function deriveMasterKey(seed) {
  const hmac = CryptoJS.HmacSHA512(
    CryptoJS.lib.WordArray.create(seed),
    'Bitcoin seed',
  );
  const hmacBytes = hexToBytes(hmac.toString());

  const privateKey = hmacBytes.slice(0, 32);
  const chainCode = hmacBytes.slice(32);

  return {
    privateKey: Buffer.from(privateKey),
    chainCode: Buffer.from(chainCode),
  };
}

/**
 * Derive child key from parent using BIP32 path
 * @param {Object} masterKey - Master key
 * @param {string} path - BIP32 derivation path
 * @returns {Object} Child key
 */
function deriveChildKey(masterKey, path) {
  const segments = path.split('/').slice(1); // Remove 'm'
  let key = masterKey;

  for (const segment of segments) {
    const hardened = segment.endsWith("'");
    const index = parseInt(segment.replace("'", ''), 10);
    const indexWithHardened = hardened ? index + 0x80000000 : index;

    key = deriveChild(key, indexWithHardened, hardened);
  }

  return key;
}

/**
 * Derive a single child key
 * @param {Object} parent - Parent key
 * @param {number} index - Child index
 * @param {boolean} hardened - Whether this is a hardened derivation
 * @returns {Object} Child key
 */
function deriveChild(parent, index, hardened) {
  const indexBuffer = Buffer.allocUnsafe(4);
  indexBuffer.writeUInt32BE(index, 0);

  let data;
  if (hardened) {
    // Hardened: HMAC-SHA512(chainCode, 0x00 || privateKey || index)
    data = Buffer.concat([
      Buffer.from([0x00]),
      parent.privateKey,
      indexBuffer,
    ]);
  } else {
    // Normal: HMAC-SHA512(chainCode, publicKey || index)
    const keyPair = secp256k1.keyFromPrivate(parent.privateKey);
    const publicKey = Buffer.from(keyPair.getPublic(true, 'array'));
    data = Buffer.concat([publicKey, indexBuffer]);
  }

  const hmac = CryptoJS.HmacSHA512(
    CryptoJS.lib.WordArray.create(data),
    CryptoJS.lib.WordArray.create(parent.chainCode),
  );
  const hmacBytes = hexToBytes(hmac.toString());

  const childPrivateKey = hmacBytes.slice(0, 32);
  const childChainCode = hmacBytes.slice(32);

  return {
    privateKey: Buffer.from(childPrivateKey),
    chainCode: Buffer.from(childChainCode),
  };
}

/**
 * Convert hex string to byte array
 * @param {string} hex - Hex string
 * @returns {Uint8Array} Byte array
 */
function hexToBytes(hex) {
  const bytes = new Uint8Array(hex.length / 2);
  for (let i = 0; i < hex.length; i += 2) {
    bytes[i / 2] = parseInt(hex.substr(i, 2), 16);
  }
  return bytes;
}

/**
 * Get public key from private key
 * @param {string} privateKey - Private key (hex)
 * @returns {string} Compressed public key (hex)
 */
export function getPublicKey(privateKey) {
  const keyPair = secp256k1.keyFromPrivate(privateKey, 'hex');
  return keyPair.getPublic(true, 'hex');
}

/**
 * Derive Bech32 address from public key
 * @param {string} publicKey - Compressed public key (hex)
 * @param {string} prefix - Bech32 prefix (default: 'aura')
 * @returns {string} Bech32 address
 */
export function deriveAddress(publicKey, prefix = 'aura') {
  // SHA256(publicKey)
  const sha256Hash = sha256.array(hexToBytes(publicKey));

  // RIPEMD160(SHA256(publicKey))
  const ripemd160Hash = new RIPEMD160().update(Buffer.from(sha256Hash)).digest();

  // Convert to 5-bit array for bech32
  const words = bech32.toWords(ripemd160Hash);

  // Encode as bech32
  return bech32.encode(prefix, words);
}

/**
 * Generate a complete wallet
 * @param {number} strength - Mnemonic strength (default: 256 for 24 words)
 * @returns {Object} Wallet with mnemonic, privateKey, publicKey, address
 */
export function generateWallet(strength = 256) {
  const mnemonic = generateMnemonic(strength);
  const privateKey = derivePrivateKeyFromMnemonic(mnemonic);
  const publicKey = getPublicKey(privateKey);
  const address = deriveAddress(publicKey);

  return {
    mnemonic,
    privateKey,
    publicKey,
    address,
  };
}

/**
 * Import wallet from mnemonic
 * @param {string} mnemonic - BIP39 mnemonic phrase
 * @param {number} accountIndex - Account index (default: 0)
 * @param {number} addressIndex - Address index (default: 0)
 * @returns {Object} Wallet with mnemonic, privateKey, publicKey, address
 */
export function importWalletFromMnemonic(
  mnemonic,
  accountIndex = 0,
  addressIndex = 0,
) {
  if (!validateMnemonic(mnemonic)) {
    throw new Error('Invalid mnemonic phrase');
  }

  const privateKey = derivePrivateKeyFromMnemonic(
    mnemonic,
    accountIndex,
    addressIndex,
  );
  const publicKey = getPublicKey(privateKey);
  const address = deriveAddress(publicKey);

  return {
    mnemonic,
    privateKey,
    publicKey,
    address,
  };
}

/**
 * Import wallet from private key
 * @param {string} privateKey - Private key (hex)
 * @returns {Object} Wallet with privateKey, publicKey, address (no mnemonic)
 */
export function importWalletFromPrivateKey(privateKey) {
  const publicKey = getPublicKey(privateKey);
  const address = deriveAddress(publicKey);

  return {
    privateKey,
    publicKey,
    address,
  };
}

/**
 * Sign a message with private key
 * @param {string} message - Message to sign
 * @param {string} privateKey - Private key (hex)
 * @returns {Object} Signature with r and s components
 */
export function signMessage(message, privateKey) {
  const messageHash = sha256(message);
  const keyPair = secp256k1.keyFromPrivate(privateKey, 'hex');
  const signature = keyPair.sign(messageHash, {canonical: true});

  return {
    r: signature.r.toString('hex'),
    s: signature.s.toString('hex'),
    recoveryParam: signature.recoveryParam,
  };
}

/**
 * Verify a signature
 * @param {string} message - Original message
 * @param {Object} signature - Signature object with r and s
 * @param {string} publicKey - Public key (hex)
 * @returns {boolean} True if signature is valid
 */
export function verifySignature(message, signature, publicKey) {
  try {
    const messageHash = sha256(message);
    const keyPair = secp256k1.keyFromPublic(publicKey, 'hex');

    return keyPair.verify(messageHash, {
      r: signature.r,
      s: signature.s,
    });
  } catch (error) {
    return false;
  }
}

/**
 * Encrypt data with password using AES-256
 * @param {string} data - Data to encrypt
 * @param {string} password - Encryption password
 * @returns {string} Encrypted data (base64)
 */
export function encrypt(data, password) {
  return CryptoJS.AES.encrypt(data, password).toString();
}

/**
 * Decrypt data with password
 * @param {string} encryptedData - Encrypted data (base64)
 * @param {string} password - Decryption password
 * @returns {string} Decrypted data
 */
export function decrypt(encryptedData, password) {
  try {
    const bytes = CryptoJS.AES.decrypt(encryptedData, password);
    return bytes.toString(CryptoJS.enc.Utf8);
  } catch (error) {
    return '';
  }
}
