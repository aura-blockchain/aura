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

export function registerStakingCommands(yargs) {
  yargs.command(
    'staking delegate <validator>',
    'Delegate stake to a validator',
    (cmd) =>
      cmd
        .positional('validator', {
          type: 'string',
          describe: 'Validator operator address',
        })
        .option('amount', {
          type: 'string',
          demandOption: true,
          describe: 'Amount in display units (e.g. 50.0 for AURA)',
        })
        .option('denom', {
          type: 'string',
          default: DEFAULTS.denom,
          describe: 'Denomination (defaults to uaura)',
        })
        .option('memo', {
          type: 'string',
          default: '',
          describe: 'Optional memo',
        })
        .option('mnemonic', {
          type: 'string',
          describe: 'BIP39 mnemonic (supercedes mnemonic-file)',
        })
        .option('mnemonic-file', {
          type: 'string',
          describe: 'Path to file containing mnemonic',
        })
        .option('gas', {
          type: 'number',
          default: 220000,
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
      const delegator = await getFirstAddress(wallet);
      const amountMicro = toMicroAmount(argv.amount, DEFAULTS.decimals);
      const fee = getFee(argv.gas, argv.gasPrice);
      const result = await client.delegateTokens(
        delegator,
        argv.validator,
        { denom: argv.denom, amount: amountMicro },
        fee,
        argv.memo || undefined,
      );
      console.log(`Delegated ${formatCoins(amountMicro, argv.denom, DEFAULTS.decimals, DEFAULTS.displayDenom)} to ${argv.validator}`);
      console.log('TxHash :', result.transactionHash);
      console.log('Height :', result.height);
    },
  );
}
