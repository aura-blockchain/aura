/**
 * Send Component - Transaction UX Tests
 * Covers confirmation flow, happy path, and error handling.
 */

import React from 'react';
import { render, screen, waitFor, act } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import Send from '../src/components/Send';
import { ApiService } from '../src/services/api';
import { KeystoreService } from '../src/services/keystore';

const mockSendTokens = jest.fn();
const mockUnlockWallet = jest.fn();

jest.mock('../src/services/api', () => ({
  ApiService: jest.fn().mockImplementation(() => ({
    sendTokens: mockSendTokens
  }))
}));

jest.mock('../src/services/keystore', () => ({
  KeystoreService: jest.fn().mockImplementation(() => ({
    unlockWallet: mockUnlockWallet
  }))
}));

describe('Send Component - Transaction Flow', () => {
  const walletData = { address: 'aura1sender' };

  beforeEach(async () => {
    jest.clearAllMocks();
    mockSendTokens.mockReset();
    mockUnlockWallet.mockReset();
    await window.electron.store.clear();
  });

  test('shows confirmation screen when previewing a valid transfer', async () => {
    render(<Send walletData={walletData} />);

    await act(async () => {
      await userEvent.type(screen.getByPlaceholderText(/aura1\.\.\./i), 'aura1recipientxyz');
      await userEvent.type(screen.getByPlaceholderText(/0\.000000/), '1.25');
      await userEvent.type(screen.getByPlaceholderText(/enter your password/i), 'p@ssw0rd!');
    });

    await act(async () => {
      await userEvent.click(screen.getByText(/preview transaction/i));
    });

    await waitFor(() => {
      expect(screen.getByText(/confirm transaction/i)).toBeInTheDocument();
      expect(mockSendTokens).not.toHaveBeenCalled();
    });
  });

  test('sends tokens, surfaces success, and triggers refresh callback', async () => {
    const onSuccess = jest.fn();

    mockUnlockWallet.mockResolvedValue({
      privateKey: 'priv-key',
      address: 'aura1sender'
    });
    mockSendTokens.mockResolvedValue({
      transactionHash: 'HASH321',
      code: 0
    });

    render(<Send walletData={walletData} onSuccess={onSuccess} />);

    await act(async () => {
      await userEvent.type(screen.getByPlaceholderText(/aura1\.\.\./i), 'aura1recipientxyz');
      await userEvent.type(screen.getByPlaceholderText(/0\.000000/), '0.5');
      await userEvent.type(screen.getByPlaceholderText(/transaction note/i), 'for integration');
      await userEvent.type(screen.getByPlaceholderText(/enter your password/i), 'p@ssw0rd!');
    });

    await act(async () => {
      await userEvent.click(screen.getByText(/preview transaction/i));
    });

    await waitFor(() => {
      expect(screen.getByText(/confirm transaction/i)).toBeInTheDocument();
    });

    await act(async () => {
      await userEvent.click(screen.getByText(/confirm & send/i));
    });

    await waitFor(() => expect(mockSendTokens).toHaveBeenCalledTimes(1));
    expect(mockSendTokens).toHaveBeenCalledWith(
      'aura1sender',
      'aura1recipientxyz',
      500000,
      'uaura',
      'for integration',
      'priv-key'
    );

    await waitFor(() => {
      expect(screen.getByText(/transaction successful/i)).toBeInTheDocument();
      expect(screen.queryByText(/confirm transaction/i)).not.toBeInTheDocument();
      expect(screen.getByPlaceholderText(/aura1\.\.\./i)).toHaveValue('');
      expect(screen.getByPlaceholderText(/enter your password/i)).toHaveValue('');
    });

    await waitFor(() => expect(onSuccess).toHaveBeenCalled(), { timeout: 3000 });
  });

  test('rejects invalid password before broadcasting', async () => {
    mockUnlockWallet.mockResolvedValue(null);

    render(<Send walletData={walletData} />);

    await act(async () => {
      await userEvent.type(screen.getByPlaceholderText(/aura1\.\.\./i), 'aura1recipientxyz');
      await userEvent.type(screen.getByPlaceholderText(/0\.000000/), '2');
      await userEvent.type(screen.getByPlaceholderText(/enter your password/i), 'badpass');
    });

    await act(async () => {
      await userEvent.click(screen.getByText(/preview transaction/i));
    });

    await waitFor(() => {
      expect(screen.getByText(/confirm transaction/i)).toBeInTheDocument();
    });

    await act(async () => {
      await userEvent.click(screen.getByText(/confirm & send/i));
    });

    await waitFor(() => {
      expect(screen.getByText(/invalid password/i)).toBeInTheDocument();
      expect(mockSendTokens).not.toHaveBeenCalled();
    });
  });
});
