/**
 * Aura Block Explorer Frontend
 */

class AuraExplorer {
  constructor(config = {}) {
    this.apiUrl = config.apiUrl || 'http://localhost:8000';
    this.wsUrl = config.wsUrl || 'ws://localhost:8000/ws';
    this.ws = null;
    this.cache = new Map();
    this.cacheTimeout = 5000; // 5 seconds
    this.pagination = {
      blocks: { page: 1, limit: 20 },
      txs: { page: 1, limit: 20 }
    };
    this.filters = {
      txType: null,
      txStatus: null
    };
    this.autoRefresh = true;
    this.refreshInterval = null;
  }

  async init() {
    this.attachEventListeners();
    this.initWebSocket();
    await this.loadInitialData();
    this.startAutoRefresh();
  }

  attachEventListeners() {
    // Search
    document.getElementById('searchBtn')?.addEventListener('click', () => this.search());
    document.getElementById('searchInput')?.addEventListener('keypress', (e) => {
      if (e.key === 'Enter') this.search();
    });

    // Blocks pagination
    document.getElementById('prevBlocks')?.addEventListener('click', () => this.prevBlocksPage());
    document.getElementById('nextBlocks')?.addEventListener('click', () => this.nextBlocksPage());
    document.getElementById('refreshBlocks')?.addEventListener('click', () => this.loadBlocks());

    // Transactions pagination and filters
    document.getElementById('prevTxs')?.addEventListener('click', () => this.prevTxsPage());
    document.getElementById('nextTxs')?.addEventListener('click', () => this.nextTxsPage());
    document.getElementById('txTypeFilter')?.addEventListener('change', (e) => {
      this.filters.txType = e.target.value || null;
      this.loadTransactions();
    });
    document.getElementById('txStatusFilter')?.addEventListener('change', (e) => {
      this.filters.txStatus = e.target.value || null;
      this.loadTransactions();
    });

    // Auto-refresh toggle
    document.getElementById('autoRefresh')?.addEventListener('change', (e) => {
      this.autoRefresh = e.target.checked;
      if (this.autoRefresh) {
        this.startAutoRefresh();
      } else {
        this.stopAutoRefresh();
      }
    });

    // Validator sort
    document.getElementById('validatorSort')?.addEventListener('change', () => {
      this.loadValidators();
    });
  }

  initWebSocket() {
    try {
      this.ws = new WebSocket(this.wsUrl);

      this.ws.onopen = () => {
        console.log('[Explorer WS] Connected');
        this.updateWSStatus('connected');
        this.ws.send(JSON.stringify({
          type: 'subscribe',
          data: { channel: 'blocks' }
        }));
        this.ws.send(JSON.stringify({
          type: 'subscribe',
          data: { channel: 'transactions' }
        }));
      };

      this.ws.onmessage = (event) => {
        const message = JSON.parse(event.data);
        this.handleWSMessage(message);
      };

      this.ws.onerror = (error) => {
        console.error('[Explorer WS] Error:', error);
        this.updateWSStatus('error');
      };

      this.ws.onclose = () => {
        console.log('[Explorer WS] Disconnected');
        this.updateWSStatus('disconnected');
        setTimeout(() => this.initWebSocket(), 5000);
      };
    } catch (err) {
      console.error('[Explorer WS] Failed to connect:', err);
      this.updateWSStatus('error');
    }
  }

  handleWSMessage(message) {
    switch (message.type) {
      case 'new_block':
        this.handleNewBlock(message.data);
        break;
      case 'new_transaction':
        this.handleNewTransaction(message.data);
        break;
      case 'subscribed':
        console.log('[Explorer WS] Subscribed to', message.data.channel);
        break;
      default:
        console.log('[Explorer WS] Unknown message type:', message.type);
    }
  }

  handleNewBlock(blockData) {
    console.log('[Explorer] New block:', blockData);
    this.cache.delete('blocks');
    if (this.pagination.blocks.page === 1) {
      this.loadBlocks();
    }
    this.updateQuickStats();
  }

  handleNewTransaction(txData) {
    console.log('[Explorer] New transaction:', txData);
    this.cache.delete('transactions');
    if (this.pagination.txs.page === 1) {
      this.loadTransactions();
    }
  }

  updateWSStatus(status) {
    const indicator = document.getElementById('wsStatus');
    if (!indicator) return;

    const dot = indicator.querySelector('.ws-dot');
    const text = indicator.querySelector('.ws-text');

    indicator.className = 'ws-indicator';
    switch (status) {
      case 'connected':
        indicator.classList.add('connected');
        text.textContent = 'Live';
        break;
      case 'disconnected':
        indicator.classList.add('disconnected');
        text.textContent = 'Disconnected';
        break;
      case 'error':
        indicator.classList.add('error');
        text.textContent = 'Error';
        break;
      default:
        text.textContent = 'Connecting...';
    }
  }

  async loadInitialData() {
    await Promise.all([
      this.loadBlocks(),
      this.loadTransactions(),
      this.loadValidators(),
      this.updateQuickStats()
    ]);
  }

  async loadBlocks() {
    try {
      const { page, limit } = this.pagination.blocks;
      const offset = (page - 1) * limit;

      const data = await this.fetchAPI(`/api/blocks?limit=${limit}&offset=${offset}`);
      this.renderBlocks(data.blocks || []);
      document.getElementById('blockPage').textContent = `Page ${page}`;
    } catch (err) {
      console.error('Failed to load blocks:', err);
      this.renderError('blocksTable', 'Failed to load blocks', 6);
    }
  }

  async loadTransactions() {
    try {
      const { page, limit } = this.pagination.txs;
      const offset = (page - 1) * limit;

      let url = `/api/transactions?limit=${limit}&offset=${offset}`;
      if (this.filters.txType) url += `&type=${this.filters.txType}`;
      if (this.filters.txStatus) url += `&status=${this.filters.txStatus}`;

      const data = await this.fetchAPI(url);
      this.renderTransactions(data.transactions || []);
      document.getElementById('txPage').textContent = `Page ${page}`;
    } catch (err) {
      console.error('Failed to load transactions:', err);
      this.renderError('txsTable', 'Failed to load transactions', 8);
    }
  }

  async loadValidators() {
    try {
      const sortBy = document.getElementById('validatorSort')?.value || 'voting_power';
      const data = await this.fetchAPI(`/api/validators?sort=${sortBy}`);
      this.renderValidators(data.validators || []);
    } catch (err) {
      console.error('Failed to load validators:', err);
      this.renderError('validatorsTable', 'Failed to load validators', 6);
    }
  }

  async updateQuickStats() {
    try {
      const data = await this.fetchAPI('/api/stats');
      document.getElementById('latestBlock').textContent = data.latest_block || '-';
      document.getElementById('avgBlockTime').textContent = data.avg_block_time
        ? `${data.avg_block_time.toFixed(2)}s`
        : '-';
      document.getElementById('totalTxs').textContent = this.formatNumber(data.total_txs || 0);
      document.getElementById('activeValidators').textContent = data.active_validators || '-';
    } catch (err) {
      console.error('Failed to update stats:', err);
    }
  }

  renderBlocks(blocks) {
    const tbody = document.getElementById('blocksTable');
    if (!tbody) return;

    if (blocks.length === 0) {
      tbody.innerHTML = '<tr><td colspan="6">No blocks found</td></tr>';
      return;
    }

    tbody.innerHTML = blocks.map(block => `
      <tr onclick="explorer.viewBlock('${block.height}')">
        <td><a href="#block/${block.height}">${block.height}</a></td>
        <td><code class="hash">${this.truncateHash(block.hash)}</code></td>
        <td>${this.formatTime(block.time)}</td>
        <td><code class="addr">${this.truncateAddr(block.proposer)}</code></td>
        <td>${block.num_txs || 0}</td>
        <td>${this.formatBytes(block.size)}</td>
      </tr>
    `).join('');
  }

  renderTransactions(txs) {
    const tbody = document.getElementById('txsTable');
    if (!tbody) return;

    if (txs.length === 0) {
      tbody.innerHTML = '<tr><td colspan="8">No transactions found</td></tr>';
      return;
    }

    tbody.innerHTML = txs.map(tx => `
      <tr onclick="explorer.viewTx('${tx.hash}')">
        <td><code class="hash">${this.truncateHash(tx.hash)}</code></td>
        <td><a href="#block/${tx.height}">${tx.height}</a></td>
        <td><span class="tx-type">${tx.type || 'Unknown'}</span></td>
        <td><code class="addr">${this.truncateAddr(tx.from)}</code></td>
        <td><code class="addr">${this.truncateAddr(tx.to)}</code></td>
        <td>${tx.amount || '-'}</td>
        <td><span class="status status-${tx.status}">${tx.status}</span></td>
        <td>${this.formatTime(tx.time)}</td>
      </tr>
    `).join('');
  }

  renderValidators(validators) {
    const tbody = document.getElementById('validatorsTable');
    if (!tbody) return;

    if (validators.length === 0) {
      tbody.innerHTML = '<tr><td colspan="6">No validators found</td></tr>';
      return;
    }

    tbody.innerHTML = validators.map((val, idx) => `
      <tr>
        <td>${idx + 1}</td>
        <td>
          <div class="validator-info">
            <strong>${val.moniker || 'Unknown'}</strong>
            <code class="addr">${this.truncateAddr(val.address)}</code>
          </div>
        </td>
        <td>${this.formatNumber(val.voting_power)}</td>
        <td>${(val.commission * 100).toFixed(2)}%</td>
        <td>
          <div class="uptime-bar">
            <div class="uptime-fill" style="width: ${val.uptime * 100}%"></div>
          </div>
          <span>${(val.uptime * 100).toFixed(2)}%</span>
        </td>
        <td><span class="val-status val-${val.status}">${val.status}</span></td>
      </tr>
    `).join('');
  }

  renderError(tableId, message, colspan) {
    const tbody = document.getElementById(tableId);
    if (tbody) {
      tbody.innerHTML = `<tr><td colspan="${colspan}" class="error">${message}</td></tr>`;
    }
  }

  async search() {
    const query = document.getElementById('searchInput')?.value.trim();
    if (!query) return;

    try {
      const data = await this.fetchAPI(`/api/search?q=${encodeURIComponent(query)}`);

      if (data.type === 'block') {
        this.viewBlock(data.result.height);
      } else if (data.type === 'transaction') {
        this.viewTx(data.result.hash);
      } else if (data.type === 'address') {
        this.viewAddress(data.result.address);
      } else {
        alert('No results found');
      }
    } catch (err) {
      console.error('Search failed:', err);
      alert('Search failed: ' + err.message);
    }
  }

  async viewBlock(height) {
    console.log('Viewing block:', height);
    // Implement block detail view
  }

  async viewTx(hash) {
    console.log('Viewing transaction:', hash);
    // Implement transaction detail view
  }

  async viewAddress(address) {
    console.log('Viewing address:', address);
    // Implement address detail view
  }

  prevBlocksPage() {
    if (this.pagination.blocks.page > 1) {
      this.pagination.blocks.page--;
      this.loadBlocks();
    }
  }

  nextBlocksPage() {
    this.pagination.blocks.page++;
    this.loadBlocks();
  }

  prevTxsPage() {
    if (this.pagination.txs.page > 1) {
      this.pagination.txs.page--;
      this.loadTransactions();
    }
  }

  nextTxsPage() {
    this.pagination.txs.page++;
    this.loadTransactions();
  }

  startAutoRefresh() {
    if (this.refreshInterval) return;
    this.refreshInterval = setInterval(() => {
      if (this.pagination.blocks.page === 1) {
        this.loadBlocks();
      }
      if (this.pagination.txs.page === 1) {
        this.loadTransactions();
      }
      this.updateQuickStats();
    }, 10000); // Refresh every 10 seconds
  }

  stopAutoRefresh() {
    if (this.refreshInterval) {
      clearInterval(this.refreshInterval);
      this.refreshInterval = null;
    }
  }

  async fetchAPI(endpoint) {
    const cacheKey = endpoint;

    // Check cache
    if (this.cache.has(cacheKey)) {
      const cached = this.cache.get(cacheKey);
      if (Date.now() - cached.timestamp < this.cacheTimeout) {
        return cached.data;
      }
    }

    const response = await fetch(this.apiUrl + endpoint);
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }

    const data = await response.json();

    // Cache result
    this.cache.set(cacheKey, {
      data,
      timestamp: Date.now()
    });

    return data;
  }

  truncateHash(hash) {
    if (!hash) return '-';
    return hash.length > 16 ? `${hash.slice(0, 8)}...${hash.slice(-8)}` : hash;
  }

  truncateAddr(addr) {
    if (!addr) return '-';
    return addr.length > 20 ? `${addr.slice(0, 10)}...${addr.slice(-8)}` : addr;
  }

  formatTime(timestamp) {
    if (!timestamp) return '-';
    const date = new Date(timestamp);
    const now = new Date();
    const diff = (now - date) / 1000;

    if (diff < 60) return `${Math.floor(diff)}s ago`;
    if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
    if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
    return date.toLocaleDateString();
  }

  formatBytes(bytes) {
    if (!bytes) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`;
  }

  formatNumber(num) {
    return new Intl.NumberFormat().format(num);
  }
}

// Initialize explorer
const explorer = new AuraExplorer();
document.addEventListener('DOMContentLoaded', () => explorer.init());
