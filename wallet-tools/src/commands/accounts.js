import { DirectSecp256k1HdWallet } from '@cosmjs/proto-signing';
import { DEFAULTS } from '../config.js';
import { fetchJson } from '../utils/client.js';
import { formatCoins } from '../utils/amount.js';

export function registerAccountCommands(yargs) {
  yargs.command(
    'account new',
    'Create a new Aura account (writes mnemonic to stdout)',
    (cmd) =>
      cmd
        .option('prefix', {
          type: 'string',
          default: DEFAULTS.bech32Prefix,
          describe: 'Bech32 prefix for the generated address',
        })
        .option('length', {
          type: 'number',
          default: 24,
          describe: 'Mnemonic word count (12/24)',
          choices: [12, 24],
        })
        .option('json', {
          type: 'boolean',
          default: false,
          describe: 'Emit JSON (useful for scripting)',
        }),
    async (argv) => {
      const wallet = await DirectSecp256k1HdWallet.generate(argv.length, {
        prefix: argv.prefix,
      });
      const [account] = await wallet.getAccounts();
      const payload = {
        address: account.address,
        algo: account.algo,
        prefix: argv.prefix,
        mnemonic: wallet.mnemonic,
      };

      if (argv.json) {
        console.log(JSON.stringify(payload, null, 2));
      } else {
        console.log('Address :', payload.address);
        console.log('Prefix  :', payload.prefix);
        console.log('Mnemonic:', payload.mnemonic);
        console.log('\nStore this mnemonic securely; anyone with it can drain the account.');
      }
    },
  );

  yargs.command(
    'account balance <address>',
    'Query balances for an address',
    (cmd) =>
      cmd
        .positional('address', {
          type: 'string',
          describe: 'Bech32 address to inspect',
        })
        .option('denom', {
          type: 'string',
          default: DEFAULTS.denom,
          describe: 'Filter for a specific denom (defaults to uaura)',
        }),
    async (argv) => {
      const resp = await fetchJson(argv.rest, `/cosmos/bank/v1beta1/balances/${argv.address}`);
      const balances = resp.balances || [];
      const denomEntry = balances.find((b) => b.denom === argv.denom);
      if (!denomEntry) {
        console.log(`0 ${argv.denom}`);
        return;
      }
      console.log(formatCoins(denomEntry.amount, denomEntry.denom, DEFAULTS.decimals, DEFAULTS.displayDenom));
    },
  );
}
