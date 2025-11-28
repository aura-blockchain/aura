import { readFile } from 'fs/promises';
import { DirectSecp256k1HdWallet } from '@cosmjs/proto-signing';

export async function loadWallet({ mnemonic, mnemonicFile, prefix }) {
  if (!mnemonic && mnemonicFile) {
    mnemonic = (await readFile(mnemonicFile, 'utf8')).trim();
  }

  if (!mnemonic) {
    throw new Error('mnemonic or mnemonic-file flag is required for signed operations');
  }

  return DirectSecp256k1HdWallet.fromMnemonic(mnemonic, {
    prefix,
  });
}
