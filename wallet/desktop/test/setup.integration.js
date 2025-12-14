// Integration test setup
import { TextEncoder, TextDecoder } from 'util';

process.env.NODE_ENV = 'test';

// Setup test timeouts
jest.setTimeout(30000);

// Polyfill TextEncoder/TextDecoder for Node.js
global.TextEncoder = TextEncoder;
global.TextDecoder = TextDecoder;

// Mock electron store with actual storage
const electronStoreData = {};

// Mock electron APIs for integration tests
global.window = global.window || {};
global.window.electron = {
  store: {
    get: jest.fn((key) => {
      if (key === 'apiEndpoint' && !electronStoreData[key]) {
        return Promise.resolve('http://localhost:1317');
      }
      return Promise.resolve(electronStoreData[key] || null);
    }),
    set: jest.fn((key, value) => {
      electronStoreData[key] = value;
      return Promise.resolve();
    }),
    delete: jest.fn((key) => {
      delete electronStoreData[key];
      return Promise.resolve();
    }),
    clear: jest.fn(() => {
      Object.keys(electronStoreData).forEach(key => delete electronStoreData[key]);
      return Promise.resolve();
    })
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

// Mock localStorage
class LocalStorageMock {
  constructor() {
    this.store = {};
  }

  getItem(key) {
    return this.store[key] || null;
  }

  setItem(key, value) {
    this.store[key] = value.toString();
  }

  removeItem(key) {
    delete this.store[key];
  }

  clear() {
    this.store = {};
  }

  get length() {
    return Object.keys(this.store).length;
  }

  key(index) {
    const keys = Object.keys(this.store);
    return keys[index] || null;
  }
}

global.localStorage = new LocalStorageMock();
