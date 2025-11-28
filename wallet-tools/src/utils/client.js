import fetch from 'node-fetch';
import { SigningStargateClient } from '@cosmjs/stargate';

export async function getSigningClient(rpcEndpoint, wallet) {
  return SigningStargateClient.connectWithSigner(rpcEndpoint, wallet, {
    gasPrice: undefined,
  });
}

export async function getFirstAddress(wallet) {
  const [account] = await wallet.getAccounts();
  if (!account) {
    throw new Error('wallet returned no accounts');
  }
  return account.address;
}

export async function fetchJson(baseUrl, path, abortMs = 5000) {
  const url = new URL(path, baseUrl);
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), abortMs);

  try {
    const response = await fetch(url, { signal: controller.signal });
    if (!response.ok) {
      const text = await response.text();
      throw new Error(`Request failed (${response.status}): ${text}`);
    }
    return response.json();
  } finally {
    clearTimeout(timeout);
  }
}
