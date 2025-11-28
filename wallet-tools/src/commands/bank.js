import { DEFAULTS, getFee } from '../config.js';
import { toMicroAmount, formatCoins } from '../utils/amount.js';
import { loadWallet } from '../utils/wallet.js';
import { getSigningClient, getFirstAddress } from '../utils/client.js';

async function ensureChainId(client, expected) {
  if (!expected) {
    return;
  }
  const actual = await client.getChainId();
  if (actual !== expected) {
    throw new Error(`Connected chain id ${actual} does not match expected ${expected}`);
  }
}

export function registerBankCommands(yargs) {
  yargs.command(
    'bank transfer <to>',
    'Send tokens from a mnemonic to a recipient',
    (cmd) =>
      cmd
        .positional('to', {
          type: 'string',
          describe: 'Recipient Bech32 address',
        })
        .option('amount', {
          type: 'string',
          demandOption: true,
          describe: 'Amount in display units (e.g. 1.23 for AURA)',
        })
        .option('denom', {
          type: 'string',
          default: DEFAULTS.denom,
          describe: 'Denomination (defaults to uaura)',
        })
        .option('memo', {
          type: 'string',
          default: '',
          describe: 'Optional memo to attach',
        })
        .option('mnemonic', {
          type: 'string',
          describe: 'BIP39 mnemonic (supercedes mnemonic-file)',
        })
        .option('mnemonic-file', {
          type: 'string',
          describe: 'Path to a file that stores the mnemonic',
        })
        .option('gas', {
          type: 'number',
          default: 180000,
          describe: 'Gas limit to apply',
        }),
    async (argv) => {
      const wallet = await loadWallet({
        mnemonic: argv.mnemonic,
        mnemonicFile: argv.mnemonicFile,
        prefix: DEFAULTS.bech32Prefix,
      });
      const client = await getSigningClient(argv.rpc, wallet);
      await ensureChainId(client, argv.chainId);

      const fromAddress = await getFirstAddress(wallet);
      const amountMicro = toMicroAmount(argv.amount, DEFAULTS.decimals);
      const fee = getFee(argv.gas, argv.gasPrice);

      const result = await client.sendTokens(
        fromAddress,
        argv.to,
        [{ denom: argv.denom, amount: amountMicro }],
        fee,
        argv.memo || undefined,
      );

      const display = formatCoins(amountMicro, argv.denom, DEFAULTS.decimals, DEFAULTS.displayDenom);
      console.log(`Sent ${display}`);
      console.log('TxHash :', result.transactionHash);
      console.log('Height :', result.height);
    },
  );
}
