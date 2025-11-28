const settingsKey = 'auraVerifierPortal';

const els = {
  rest: document.getElementById('portalRest'),
  assistantFilter: document.getElementById('assistantFilter'),
  walletAddress: document.getElementById('walletAddress'),
  assistantTable: document.getElementById('assistantTable'),
  assistantStats: document.getElementById('assistantStats'),
  walletDetails: document.getElementById('walletDetails'),
  walletStatus: document.getElementById('walletStatus'),
  completionTable: document.getElementById('completionTable'),
};

loadSettings();
refreshAssistants();

document.getElementById('btnSavePortal').addEventListener('click', () => {
  const payload = {
    rest: els.rest.value || 'http://localhost:1317',
    assistantFilter: els.assistantFilter.value || '.*',
    wallet: els.walletAddress.value.trim(),
  };
  localStorage.setItem(settingsKey, JSON.stringify(payload));
  notify('Settings saved.');
});

document.getElementById('btnRefreshAll').addEventListener('click', () => {
  refreshAssistants();
  if (els.walletAddress.value.trim()) {
    lookupWallet();
  }
});

document.getElementById('btnLookupWallet').addEventListener('click', lookupWallet);

function loadSettings() {
  try {
    const data = JSON.parse(localStorage.getItem(settingsKey)) || {};
    if (data.rest) els.rest.value = data.rest;
    if (data.assistantFilter) els.assistantFilter.value = data.assistantFilter;
    if (data.wallet) els.walletAddress.value = data.wallet;
  } catch (err) {
    console.warn('Failed to load settings', err);
  }
}

async function refreshAssistants() {
  try {
    const resp = await fetch(`${baseRest()}/aura/aiassistant/v1beta1/assistants?pagination.limit=50`);
    if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
    const data = await resp.json();
    const assistants = data.assistants || [];
    const filter = new RegExp(els.assistantFilter.value || '.*');
    const filtered = assistants.filter((a) => filter.test(a.assistant_address));
    els.assistantStats.textContent = `${filtered.length} / ${assistants.length}`;
    if (filtered.length === 0) {
      els.assistantTable.innerHTML = '<tr><td colspan="8">No assistants matched this filter.</td></tr>';
      return;
    }
    els.assistantTable.innerHTML = filtered
      .map(
        (a) => `
        <tr>
          <td>${shorten(a.assistant_address)}</td>
          <td>${shorten(a.owner_address)}</td>
          <td>${formatMicro(a.stake)}</td>
          <td>${formatMicro(a.sponsorship_balance)}</td>
          <td>${a.locales?.join(', ') || '—'}</td>
          <td>${a.status}</td>
          <td>${formatRelative(a.last_heartbeat)}</td>
          <td>${a.misbehavior_reports ?? 0}</td>
        </tr>`,
      )
      .join('');
  } catch (err) {
    els.assistantTable.innerHTML = `<tr><td colspan="8">Failed to load assistants: ${err.message}</td></tr>`;
  }
}

async function lookupWallet() {
  const wallet = els.walletAddress.value.trim();
  if (!wallet) {
    notify('Enter a wallet address first.');
    return;
  }
  els.walletStatus.textContent = 'Loading...';
  try {
    const scoreResp = await fetch(`${baseRest()}/aura/confidencescore/v1beta1/user_score/${wallet}`);
    if (!scoreResp.ok) throw new Error(`HTTP ${scoreResp.status}`);
    const scoreData = await scoreResp.json();
    renderScore(scoreData, wallet);
    await renderCompletions(wallet);
    els.walletStatus.textContent = 'Ready';
  } catch (err) {
    els.walletStatus.textContent = 'Error';
    els.walletDetails.innerHTML = `<p class="muted">Lookup failed: ${err.message}</p>`;
    els.completionTable.innerHTML = '<tr><td colspan="6">Unable to load IR completions.</td></tr>';
  }
}

function renderScore(data, wallet) {
  const resp = data || {};
  const total = resp.total_score ?? 0;
  const verified = resp.is_verified ? 'Verified' : 'Provisional';
  const arenaRows = Object.entries(resp.arena_scores || {}).map(
    ([arena, value]) => `
      <div class="score-card">
        <span class="muted">${arena}</span>
        <strong>${value?.score ?? 0}</strong>
        <small class="muted">${value?.ir_count ?? 0} IRs</small>
      </div>`,
  );
  els.walletDetails.innerHTML = `
    <div class="score-card">
      <span class="muted">Wallet</span>
      <strong>${shorten(wallet)}</strong>
      <small class="muted">${verified}</small>
    </div>
    <div class="score-card">
      <span class="muted">Total Score</span>
      <strong>${total}</strong>
      <small class="muted">${resp.ir_count ?? 0} IRs</small>
    </div>
    ${arenaRows.join('')}
  `;
}

async function renderCompletions(wallet) {
  try {
    const resp = await fetch(`${baseRest()}/aura/confidencescore/v1beta1/user_completions/${wallet}?pagination.limit=5`);
    if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
    const data = await resp.json();
    const results = data.completions || [];
    if (!results.length) {
      els.completionTable.innerHTML = '<tr><td colspan="6">No IR completions recorded.</td></tr>';
      return;
    }
    els.completionTable.innerHTML = results
      .map(
        (c) => `
        <tr>
          <td>${c.ir_id}</td>
          <td>${c.arena}</td>
          <td>${shorten(c.assistant_address)}</td>
          <td>${c.score_delta ?? 0}</td>
          <td>${formatRelative(c.completed_at)}</td>
          <td>${c.notes || '—'}</td>
        </tr>`,
      )
      .join('');
  } catch (err) {
    els.completionTable.innerHTML = `<tr><td colspan="6">Failed to load completions: ${err.message}</td></tr>`;
  }
}

function formatMicro(balance) {
  if (!balance || !balance.amount) return '0';
  const value = Number(balance.amount) / 1_000_000;
  return `${value.toFixed(2)} ${balance.denom || ''}`.trim();
}

function formatRelative(ts) {
  if (!ts) return '—';
  const seconds = Number(ts.seconds ?? 0);
  if (!seconds) return '—';
  const diff = Date.now() / 1000 - seconds;
  if (diff < 60) return `${Math.floor(diff)}s ago`;
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
  return new Date(seconds * 1000).toLocaleString();
}

function shorten(value = '') {
  if (value.length <= 12) return value;
  return `${value.slice(0, 6)}…${value.slice(-4)}`;
}

function baseRest() {
  return els.rest.value || 'http://localhost:1317';
}

function notify(message) {
  console.log(`[portal] ${message}`);
}
