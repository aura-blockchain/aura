#!/usr/bin/env node
import yargs from 'yargs';
import { hideBin } from 'yargs/helpers';
import { DEFAULTS } from './config.js';
import { registerAccountCommands } from './commands/accounts.js';
import { registerBankCommands } from './commands/bank.js';
import { registerStakingCommands } from './commands/staking.js';
import { registerDexCommands } from './commands/dex.js';

process.on('unhandledRejection', (err) => {
  console.error('[fatal] Unhandled rejection:', err instanceof Error ? err.message : err);
  if (err?.stack) {
    console.error(err.stack);
  }
  process.exit(1);
});

const cli = yargs(hideBin(process.argv))
  .scriptName('aura-wallet')
  .usage('$0 <command> [options]')
  .strict()
  .option('rest', {
    type: 'string',
    default: DEFAULTS.restEndpoint,
    describe: 'REST endpoint (LCD) for read-only operations',
  })
  .option('rpc', {
    type: 'string',
    default: DEFAULTS.rpcEndpoint,
    describe: 'RPC endpoint for signing/broadcasting',
  })
  .option('chain-id', {
    type: 'string',
    default: DEFAULTS.chainId,
    describe: 'Expected chain-id (set to empty string to skip verification)',
  })
  .option('gas-price', {
    type: 'string',
    default: DEFAULTS.gasPrice,
    describe: 'Gas price used when calculating fees',
  })
  .group(['rest', 'rpc', 'chain-id', 'gas-price'], 'Network Options:')
  .demandCommand(1, 'Choose a command from --help')
  .recommendCommands()
  .help();

registerAccountCommands(cli);
registerBankCommands(cli);
registerStakingCommands(cli);
registerDexCommands(cli);

cli.parse();
