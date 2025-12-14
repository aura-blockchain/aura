/**
 * Test Setup File
 * Sets up global mocks and utilities for testing
 */

import { vi } from 'vitest';

// Mock Web Crypto API
if (!global.crypto) {
  global.crypto = {
    getRandomValues: (arr) => {
      for (let i = 0; i < arr.length; i++) {
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
