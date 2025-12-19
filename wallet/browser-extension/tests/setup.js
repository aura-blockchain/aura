/**
 * Test Setup File
 * Sets up global mocks and utilities for testing
 */

import { vi } from 'vitest';
import { webcrypto as nodeWebcrypto } from 'node:crypto';

// Prefer the real WebCrypto implementation when running under Node/Vitest.
const resolvedCrypto = global.crypto || nodeWebcrypto;

if (!global.crypto && resolvedCrypto) {
  global.crypto = resolvedCrypto;
} else if (!resolvedCrypto) {
  // Minimal fallback mock (should not run in CI; kept for safety).
  global.crypto = {
    getRandomValues: (arr) => {
      for (let i = 0; i < arr.length; i += 1) {
        arr[i] = Math.floor(Math.random() * 256);
      }
      return arr;
    },
    subtle: {
      importKey: vi.fn(),
      exportKey: vi.fn(),
      generateKey: vi.fn(),
      sign: vi.fn(),
      verify: vi.fn(),
      encrypt: vi.fn(),
      decrypt: vi.fn(),
      deriveKey: vi.fn(),
      deriveBits: vi.fn(),
      digest: vi.fn(),
    },
  };
}

if (nodeWebcrypto?.subtle && !global.crypto.subtle) {
  global.crypto.subtle = nodeWebcrypto.subtle;
}

if (nodeWebcrypto?.getRandomValues && !global.crypto.getRandomValues) {
  global.crypto.getRandomValues = nodeWebcrypto.getRandomValues.bind(nodeWebcrypto);
}

// Mock TextEncoder/TextDecoder
if (!global.TextEncoder) {
  global.TextEncoder = class {
    encode(str) {
      const arr = new Uint8Array(str.length);
      for (let i = 0; i < str.length; i++) {
        arr[i] = str.charCodeAt(i);
      }
      return arr;
    }
  };
}

if (!global.TextDecoder) {
  global.TextDecoder = class {
    decode(arr) {
      return String.fromCharCode(...arr);
    }
  };
}

// Mock btoa/atob for base64
if (!global.btoa) {
  global.btoa = (str) => Buffer.from(str, 'binary').toString('base64');
}

if (!global.atob) {
  global.atob = (b64) => Buffer.from(b64, 'base64').toString('binary');
}

// Mock console methods to reduce noise in tests
global.console = {
  ...console,
  log: vi.fn(),
  debug: vi.fn(),
  info: vi.fn(),
  warn: vi.fn(),
  error: vi.fn(),
};
