import { DEFAULTS } from '../config.js';
import { fetchJson } from '../utils/client.js';

export function registerDexCommands(yargs) {
  yargs.command(
    'dex spot-price <poolId>',
    'Query instantaneous spot price inside a pool',
    (cmd) =>
      cmd
        .positional('poolId', {
          type: 'string',
          describe: 'Pool identifier',
        })
        .option('base', {
          type: 'string',
          demandOption: true,
          describe: 'Base denom (sold asset)',
        })
        .option('quote', {
          type: 'string',
          demandOption: true,
          describe: 'Quote denom (received asset)',
        }),
    async (argv) => {
      const path = `/aura/dex/v1beta1/spot/${argv.poolId}/${argv.base}/${argv.quote}`;
      const resp = await fetchJson(argv.rest, path);
      console.log(`Spot price (${argv.base} -> ${argv.quote}): ${resp.price}`);
    },
  );

  yargs.command(
    'dex market-price <coin>',
    'Fetch the consolidated market price snapshot for a coin',
    (cmd) =>
      cmd.positional('coin', {
        type: 'string',
        describe: 'Coin symbol (e.g. usdt, btc)',
      }),
    async (argv) => {
      const resp = await fetchJson(argv.rest, `/aura/dex/v1beta1/price/${argv.coin}`);
      const price = resp.price;
      console.log(`Price (${argv.coin}): ${price.price_aura} AURA | ${price.price_usd} USD (sample size ${price.sample_size})`);
    },
  );
}
