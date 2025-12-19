/**
 * Component Tests
 * Tests for React components
 */

import React from 'react';
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import Wallet from '../src/components/Wallet';
import Send from '../src/components/Send';
import Receive from '../src/components/Receive';
import History from '../src/components/History';
import AddressBook from '../src/components/AddressBook';
import Settings from '../src/components/Settings';

// Mock ApiService
jest.mock('../src/services/api', () => ({
  ApiService: jest.fn().mockImplementation(() => ({
    getApiEndpoint: jest.fn(() => Promise.resolve('http://localhost:1317')),
    getEndpoint: jest.fn(() => Promise.resolve('http://localhost:1317')),
    getBalance: jest.fn(() => Promise.resolve({ balances: [] })),
    getTransactions: jest.fn(() => Promise.resolve([]))
  }))
}));

describe('Component Tests', () => {
  beforeEach(async () => {
    jest.clearAllMocks();
    await window.electron.store.clear();
  });

  describe('Wallet Component', () => {
    test('should render wallet balance', async () => {
      const mockWalletData = {
        address: 'aura1test123',
        publicKey: '0x123456'
      };

      await act(async () => {
        render(<Wallet walletData={mockWalletData} />);
        await flushAsync();
      });

      await waitFor(() => {
        expect(screen.getByRole('heading', { name: /balance/i })).toBeInTheDocument();
        expect(screen.getByText(/no balance found/i)).toBeInTheDocument();
      });
    });

    test('should display loading state', async () => {
      await act(async () => {
        render(<Wallet />);
      });

      expect(screen.getByText(/loading balance/i)).toBeInTheDocument();

      await act(async () => {
        await flushAsync();
      });
    });
  });

  describe('Send Component', () => {
    test('should render send form', () => {
      const mockWalletData = {
        address: 'aura1sender'
      };

      render(<Send walletData={mockWalletData} />);

      expect(screen.getByPlaceholderText(/aura1.../i)).toBeInTheDocument();
      expect(screen.getByText(/amount/i)).toBeInTheDocument();
    });

    test('should validate recipient address', async () => {
      const mockWalletData = {
        address: 'aura1sender'
      };

      render(<Send walletData={mockWalletData} />);

      const recipientInput = screen.getByPlaceholderText(/aura1.../i);
      const previewButton = screen.getByText(/preview/i);

      await act(async () => {
        await userEvent.type(recipientInput, 'invalid-address');
      });
      await act(async () => {
        fireEvent.click(previewButton);
      });

      await waitFor(() => {
        expect(screen.getByText(/invalid/i)).toBeInTheDocument();
      });
    });
  });

  describe('Receive Component', () => {
    test('should display wallet address', () => {
      const mockWalletData = {
        address: 'aura1receiver123'
      };

      render(<Receive walletData={mockWalletData} />);

      expect(screen.getByText(/aura1receiver123/i)).toBeInTheDocument();
    });

    test('should copy address to clipboard', async () => {
      const mockWalletData = {
        address: 'aura1receiver123'
      };

      render(<Receive walletData={mockWalletData} />);

      const copyButton = screen.getByText(/copy/i);
      fireEvent.click(copyButton);

      await waitFor(() => {
        expect(navigator.clipboard.writeText).toHaveBeenCalledWith('aura1receiver123');
      });
    });
  });

  describe('History Component', () => {
    test('should render transaction list', async () => {
      const mockWalletData = {
        address: 'aura1test'
      };

      render(<History walletData={mockWalletData} />);

      // Component renders either loading state or history
      await waitFor(() => {
        const loadingText = screen.queryByText(/loading/i);
        const historyText = screen.queryByText(/transaction history/i);
        expect(loadingText || historyText).toBeTruthy();
      });
    });

    test('should display empty state', async () => {
      const mockWalletData = {
        address: 'aura1test'
      };

      render(<History walletData={mockWalletData} />);

      await waitFor(() => {
        expect(screen.getByText(/no transactions/i)).toBeInTheDocument();
      });
    });
  });

  describe('AddressBook Component', () => {
    test('should render address book', () => {
      render(<AddressBook />);

      expect(screen.getByText(/address book/i)).toBeInTheDocument();
    });

    test('should show add address form', () => {
      render(<AddressBook />);

      const addButton = screen.getByText(/add address/i);
      fireEvent.click(addButton);

      expect(screen.getByPlaceholderText(/alice's wallet/i)).toBeInTheDocument();
    });
  });

  describe('Settings Component', () => {
    test('should render settings', async () => {
      render(<Settings />);

      await waitFor(() => {
        expect(screen.getByText(/network settings/i)).toBeInTheDocument();
      });
    });

    test('should save settings', async () => {
      render(<Settings />);

      const saveButton = screen.getByText(/save settings/i);
      await act(async () => {
        fireEvent.click(saveButton);
      });

      await waitFor(() => {
        expect(screen.getByText(/saved/i)).toBeInTheDocument();
      });
    });
  });
});
const flushAsync = () => new Promise((resolve) => setTimeout(resolve, 0));
