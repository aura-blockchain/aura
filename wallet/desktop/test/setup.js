import '@testing-library/jest-dom';
import { TextEncoder, TextDecoder } from 'util';

// Polyfill TextEncoder/TextDecoder for Node.js
global.TextEncoder = TextEncoder;
global.TextDecoder = TextDecoder;

// Mock electron APIs
global.window = global.window || {};
global.window.electron = {
  store: {
    get: jest.fn((key) => {
      // Return null by default, tests can override
      if (key === 'apiEndpoint') {
        return Promise.resolve('http://localhost:1317');
      }
      return Promise.resolve(null);
    }),
    set: jest.fn(() => Promise.resolve()),
    delete: jest.fn(() => Promise.resolve()),
    clear: jest.fn(() => Promise.resolve())
  },
  dialog: {
    showOpenDialog: jest.fn(() => Promise.resolve({ canceled: false, filePaths: [] })),
    showSaveDialog: jest.fn(() => Promise.resolve({ canceled: false, filePath: '' })),
    showMessageBox: jest.fn(() => Promise.resolve({ response: 0 }))
  },
  app: {
    getVersion: jest.fn(() => Promise.resolve('1.0.0')),
    getPath: jest.fn((name) => Promise.resolve(`/tmp/test-${name}`))
  },
  onMenuAction: jest.fn(),
  removeMenuActionListener: jest.fn()
};

// Mock localStorage with jest.fn() methods
const localStorageMock = (() => {
  let store = {};
  return {
    getItem: jest.fn((key) => store[key] || null),
    setItem: jest.fn((key, value) => {
      store[key] = value.toString();
    }),
    removeItem: jest.fn((key) => {
      delete store[key];
    }),
    clear: jest.fn(() => {
      store = {};
    }),
    get length() {
      return Object.keys(store).length;
    },
    key: jest.fn((index) => {
      const keys = Object.keys(store);
      return keys[index] || null;
    })
  };
})();
global.localStorage = localStorageMock;

// Mock navigator.clipboard
Object.defineProperty(navigator, 'clipboard', {
  value: {
    writeText: jest.fn(() => Promise.resolve())
  },
  writable: true
});
