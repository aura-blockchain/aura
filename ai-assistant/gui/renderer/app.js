const { DateTime } = luxon;

const el = {
  restEndpoint: document.getElementById('restEndpoint'),
  rpcEndpoint: document.getElementById('rpcEndpoint'),
  auradPath: document.getElementById('auradPath'),
  keyName: document.getElementById('keyName'),
  assistantAddress: document.getElementById('assistantAddress'),
  chainId: document.getElementById('chainId'),
  gasPrice: document.getElementById('gasPrice'),
  voucherBinary: document.getElementById('voucherBinary'),
  pushGateway: document.getElementById('pushGateway'),
  voucherPassphrase: document.getElementById('voucherPassphrase'),
  logOutput: document.getElementById('logOutput'),
  status: {
    stake: document.getElementById('stakeValue'),
    sponsor: document.getElementById('sponsorValue'),
    locales: document.getElementById('localesValue'),
    heartbeat: document.getElementById('heartbeatValue'),
    status: document.getElementById('statusValue'),
    reports: document.getElementById('reportsValue'),
  },
  voucherPayload: document.getElementById('voucherPayload'),
  voucherRedeemNotes: document.getElementById('voucherRedeemNotes'),
  issueAmount: document.getElementById('issueAmount'),
  issueDenom: document.getElementById('issueDenom'),
  issueSponsor: document.getElementById('issueSponsor'),
  issueExpiry: document.getElementById('issueExpiry'),
  issueNotes: document.getElementById('issueNotes'),
  extraHeartbeatFlags: document.getElementById('extraHeartbeatFlags'),
};

const defaultSettings = {
  restEndpoint: 'http://localhost:1317',
  rpcEndpoint: 'http://localhost:26657',
  auradPath: '/usr/local/bin/aurad',
  voucherBinary: '/usr/local/bin/aura-voucher',
  gasPrice: '0.025uaura',
  chainId: 'aura-localnet',
};

const SETTINGS_KEY = 'auraAssistantSettings';

function loadSettings() {
  try {
    const data = localStorage.getItem(SETTINGS_KEY);
    if (!data) return;
    const parsed = JSON.parse(data);
    Object.entries(parsed).forEach(([key, value]) => {
      if (el[key]) {
        el[key].value = value;
      }
    });
  } catch (err) {
    log(`Failed to load settings: ${err.message}`, 'error');
  }
}

function saveSettings() {
  const payload = {
    restEndpoint: el.restEndpoint.value || defaultSettings.restEndpoint,
    rpcEndpoint: el.rpcEndpoint.value || defaultSettings.rpcEndpoint,
    auradPath: el.auradPath.value || defaultSettings.auradPath,
    keyName: el.keyName.value,
    assistantAddress: el.assistantAddress.value,
    chainId: el.chainId.value || defaultSettings.chainId,
    gasPrice: el.gasPrice.value || defaultSettings.gasPrice,
    voucherBinary: el.voucherBinary.value || defaultSettings.voucherBinary,
    pushGateway: el.pushGateway.value,
  };
  localStorage.setItem(SETTINGS_KEY, JSON.stringify(payload));
  log('Settings saved.');
}

function log(message, level = 'info') {
  const timestamp = DateTime.now().toFormat('HH:mm:ss');
  const line = `[${timestamp}] ${message}`;
  el.logOutput.textContent = `${line}\n${el.logOutput.textContent}`.slice(0, 8000);
  if (level === 'error') {
    console.error(message);
  } else {
    console.log(message);
  }
}

async function refreshStatus() {
  const settings = currentSettings();
  if (!settings.assistantAddress) {
    log('Set assistant address before refreshing status', 'error');
    return;
  }
  try {
    const resp = await fetch(
      `${settings.restEndpoint}/aura/aiassistant/v1beta1/assistants/${settings.assistantAddress}`,
    );
    if (!resp.ok) {
      throw new Error(`LCD returned ${resp.status}`);
    }
    const data = await resp.json();
    const assistant = data.assistant;
    el.status.stake.textContent = formatBalance(assistant.stake);
    el.status.sponsor.textContent = formatBalance(assistant.sponsorship_balance);
    el.status.locales.textContent = assistant.locales?.join(', ') || '—';
    el.status.status.textContent = assistant.status || '—';
    el.status.reports.textContent = assistant.misbehavior_reports ?? '0';
    const hb = assistant.last_heartbeat?.seconds
      ? DateTime.fromSeconds(Number(assistant.last_heartbeat.seconds))
      : null;
    el.status.heartbeat.textContent = hb ? hb.toRelative() : '—';
    log('Assistant status refreshed.');
  } catch (err) {
    log(`Status fetch failed: ${err.message}`, 'error');
  }
}

function formatBalance(balance) {
  if (!balance || !balance.amount) {
    return '0';
  }
  const amount = Number(balance.amount) / 1_000_000;
  return `${amount.toFixed(2)} ${balance.denom?.toUpperCase() || ''}`;
}

function currentSettings() {
  return {
    restEndpoint: el.restEndpoint.value || defaultSettings.restEndpoint,
    rpcEndpoint: el.rpcEndpoint.value || defaultSettings.rpcEndpoint,
    auradPath: el.auradPath.value || defaultSettings.auradPath,
    keyName: el.keyName.value,
    assistantAddress: el.assistantAddress.value,
    chainId: el.chainId.value || defaultSettings.chainId,
    gasPrice: el.gasPrice.value || defaultSettings.gasPrice,
    voucherBinary: el.voucherBinary.value || defaultSettings.voucherBinary,
    pushGateway: el.pushGateway.value,
  };
}

async function runCLI(command, args, env) {
  log(`$ ${command} ${args.join(' ')}`);
  try {
    const result = await window.assistantBridge.runCommand(command, args, {
      env,
    });
    if (result.stdout) {
      log(result.stdout.trim());
    }
    if (result.stderr) {
      log(result.stderr.trim(), 'error');
    }
    if (result.code !== 0) {
      log(`Command exited with ${result.code}`, 'error');
    }
    return result;
  } catch (err) {
    log(`Command failed: ${err.message}`, 'error');
    throw err;
  }
}

async function voucherEnv() {
  const secret = await window.assistantBridge.readSecret('voucher_passphrase');
  if (secret) {
    return { AURA_VOUCHER_PASSPHRASE: secret };
  }
  return {};
}

async function sendHeartbeat() {
  const settings = currentSettings();
  if (!settings.auradPath || !settings.keyName || !settings.assistantAddress) {
    log('Set aurad path, key name, and assistant address first.', 'error');
    return;
  }
  const args = [
    'tx',
    'aiassistant',
    'heartbeat',
    settings.assistantAddress,
    '--from',
    settings.keyName,
    '--chain-id',
    settings.chainId,
    '--gas-prices',
    settings.gasPrice,
    '--node',
    settings.rpcEndpoint,
    '--yes',
  ];
  const extra = el.extraHeartbeatFlags.value.trim();
  if (extra) {
    args.push(...extra.split(/\s+/).filter(Boolean));
  }
  await runCLI(settings.auradPath, args);
  log('Heartbeat command dispatched.');
}

async function redeemVoucher() {
  const settings = currentSettings();
  const payload = el.voucherPayload.value.trim();
  if (!payload) {
    log('Paste a voucher payload first.', 'error');
    return;
  }
  const args = ['redeem', '--voucher', payload];
  if (settings.assistantAddress) {
    args.push('--assistant', settings.assistantAddress);
  }
  if (settings.pushGateway) {
    args.push('--pushgateway', settings.pushGateway);
  }
  const notes = el.voucherRedeemNotes.value.trim();
  if (notes) {
    args.push('--notes', notes);
  }
  const env = await voucherEnv();
  await runCLI(settings.voucherBinary, args, env);
  log('Voucher redemption logged.');
}

async function issueVoucher() {
  const settings = currentSettings();
  const amount = el.issueAmount.value;
  const denom = el.issueDenom.value || 'uaura';
  const sponsor = el.issueSponsor.value || settings.assistantAddress;
  if (!amount || !sponsor) {
    log('Provide amount and sponsor address to issue vouchers.', 'error');
    return;
  }
  const args = ['issue', '--amount', amount, '--denom', denom, '--sponsor', sponsor];
  const expiryRaw = el.issueExpiry.value;
  if (expiryRaw) {
    const iso = DateTime.fromISO(expiryRaw);
    if (iso.isValid) {
      args.push('--expires', iso.toUTC().toISO());
    }
  }
  const notes = el.issueNotes.value.trim();
  if (notes) {
    args.push('--notes', notes);
  }
  if (settings.pushGateway) {
    args.push('--pushgateway', settings.pushGateway);
  }
  const env = await voucherEnv();
  const result = await runCLI(settings.voucherBinary, args, env);
  if (result.stdout) {
    navigator.clipboard?.writeText(result.stdout.trim()).catch(() => {});
  }
  log('Voucher issued. Payload copied to clipboard (if allowed).');
}

document.getElementById('btnSaveSettings').addEventListener('click', saveSettings);
document.getElementById('btnRefreshStatus').addEventListener('click', refreshStatus);
document.getElementById('btnHeartbeat').addEventListener('click', sendHeartbeat);
document.getElementById('btnRedeemVoucher').addEventListener('click', redeemVoucher);
document.getElementById('btnIssueVoucher').addEventListener('click', issueVoucher);

document.getElementById('btnBrowseAurad').addEventListener('click', async () => {
  const file = await window.assistantBridge.openFileDialog();
  if (file) {
    el.auradPath.value = file;
  }
});

document.getElementById('btnBrowseVoucher').addEventListener('click', async () => {
  const file = await window.assistantBridge.openFileDialog();
  if (file) {
    el.voucherBinary.value = file;
  }
});

document.getElementById('btnStorePass').addEventListener('click', storeVoucherPassphrase);
document.getElementById('btnClearPass').addEventListener('click', clearVoucherPassphrase);

loadSettings();
Object.entries(defaultSettings).forEach(([key, value]) => {
  if (!el[key].value) {
    el[key].value = value;
  }
});
refreshStatus();
refreshPassphraseStatus();

async function storeVoucherPassphrase() {
  const value = el.voucherPassphrase.value.trim();
  if (!value) {
    log('Enter a passphrase before storing.', 'error');
    return;
  }
  await window.assistantBridge.saveSecret('voucher_passphrase', value);
  el.voucherPassphrase.value = '';
  log('Voucher passphrase stored in OS keychain.');
  refreshPassphraseStatus();
}

async function clearVoucherPassphrase() {
  await window.assistantBridge.deleteSecret('voucher_passphrase');
  log('Stored voucher passphrase removed.');
  refreshPassphraseStatus();
}

async function refreshPassphraseStatus() {
  const secret = await window.assistantBridge.readSecret('voucher_passphrase');
  if (secret) {
    el.voucherPassphrase.placeholder = 'Passphrase stored in keychain';
  } else {
    el.voucherPassphrase.placeholder = 'Store passphrase securely via keychain';
  }
}
