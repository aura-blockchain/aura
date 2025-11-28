/**
 * Advanced Filtering and Search for Verifier Portal
 */

class AdvancedFilters {
  constructor() {
    this.filters = {
      status: null,
      minStake: null,
      maxStake: null,
      locales: [],
      minMisbehavior: null,
      maxMisbehavior: null,
      heartbeatAge: null,
      ownerAddress: null,
      assistantAddress: null
    };
    this.sortConfig = {
      field: null,
      direction: 'asc'
    };
  }

  applyFilters(assistants) {
    let filtered = [...assistants];

    // Status filter
    if (this.filters.status) {
      filtered = filtered.filter(a => a.status === this.filters.status);
    }

    // Stake range
    if (this.filters.minStake !== null) {
      filtered = filtered.filter(a => {
        const stake = Number(a.stake?.amount || 0);
        return stake >= this.filters.minStake;
      });
    }

    if (this.filters.maxStake !== null) {
      filtered = filtered.filter(a => {
        const stake = Number(a.stake?.amount || 0);
        return stake <= this.filters.maxStake;
      });
    }

    // Locale filter
    if (this.filters.locales.length > 0) {
      filtered = filtered.filter(a => {
        const assistantLocales = a.locales || [];
        return this.filters.locales.some(locale => assistantLocales.includes(locale));
      });
    }

    // Misbehavior range
    if (this.filters.minMisbehavior !== null) {
      filtered = filtered.filter(a =>
        (a.misbehavior_reports || 0) >= this.filters.minMisbehavior
      );
    }

    if (this.filters.maxMisbehavior !== null) {
      filtered = filtered.filter(a =>
        (a.misbehavior_reports || 0) <= this.filters.maxMisbehavior
      );
    }

    // Heartbeat age
    if (this.filters.heartbeatAge !== null) {
      const now = Date.now() / 1000;
      filtered = filtered.filter(a => {
        const lastHeartbeat = Number(a.last_heartbeat?.seconds || 0);
        return (now - lastHeartbeat) <= this.filters.heartbeatAge;
      });
    }

    // Owner address search
    if (this.filters.ownerAddress) {
      const search = this.filters.ownerAddress.toLowerCase();
      filtered = filtered.filter(a =>
        (a.owner_address || '').toLowerCase().includes(search)
      );
    }

    // Assistant address search
    if (this.filters.assistantAddress) {
      const search = this.filters.assistantAddress.toLowerCase();
      filtered = filtered.filter(a =>
        (a.assistant_address || '').toLowerCase().includes(search)
      );
    }

    return filtered;
  }

  applySort(assistants) {
    if (!this.sortConfig.field) return assistants;

    const sorted = [...assistants];
    const multiplier = this.sortConfig.direction === 'asc' ? 1 : -1;

    sorted.sort((a, b) => {
      let aVal, bVal;

      switch (this.sortConfig.field) {
        case 'stake':
          aVal = Number(a.stake?.amount || 0);
          bVal = Number(b.stake?.amount || 0);
          break;
        case 'sponsorship':
          aVal = Number(a.sponsorship_balance?.amount || 0);
          bVal = Number(b.sponsorship_balance?.amount || 0);
          break;
        case 'misbehavior':
          aVal = a.misbehavior_reports || 0;
          bVal = b.misbehavior_reports || 0;
          break;
        case 'heartbeat':
          aVal = Number(a.last_heartbeat?.seconds || 0);
          bVal = Number(b.last_heartbeat?.seconds || 0);
          break;
        case 'status':
          aVal = a.status || '';
          bVal = b.status || '';
          return aVal.localeCompare(bVal) * multiplier;
        case 'owner':
          aVal = a.owner_address || '';
          bVal = b.owner_address || '';
          return aVal.localeCompare(bVal) * multiplier;
        case 'assistant':
          aVal = a.assistant_address || '';
          bVal = b.assistant_address || '';
          return aVal.localeCompare(bVal) * multiplier;
        default:
          return 0;
      }

      return (aVal - bVal) * multiplier;
    });

    return sorted;
  }

  setFilter(key, value) {
    if (key in this.filters) {
      this.filters[key] = value;
    }
  }

  setSort(field, direction = 'asc') {
    this.sortConfig = { field, direction };
  }

  toggleSort(field) {
    if (this.sortConfig.field === field) {
      this.sortConfig.direction = this.sortConfig.direction === 'asc' ? 'desc' : 'asc';
    } else {
      this.sortConfig = { field, direction: 'asc' };
    }
  }

  clearFilters() {
    this.filters = {
      status: null,
      minStake: null,
      maxStake: null,
      locales: [],
      minMisbehavior: null,
      maxMisbehavior: null,
      heartbeatAge: null,
      ownerAddress: null,
      assistantAddress: null
    };
  }

  getActiveFilters() {
    const active = [];
    for (const [key, value] of Object.entries(this.filters)) {
      if (value !== null && value !== '' && (Array.isArray(value) ? value.length > 0 : true)) {
        active.push({ key, value });
      }
    }
    return active;
  }

  renderFilterUI(containerId) {
    const container = document.getElementById(containerId);
    if (!container) return;

    container.innerHTML = `
      <div class="advanced-filters">
        <div class="filter-section">
          <h3>Filters</h3>

          <div class="filter-group">
            <label>Status
              <select id="filter-status">
                <option value="">All</option>
                <option value="active">Active</option>
                <option value="inactive">Inactive</option>
                <option value="suspended">Suspended</option>
                <option value="jailed">Jailed</option>
              </select>
            </label>
          </div>

          <div class="filter-group">
            <label>Min Stake (micro)
              <input type="number" id="filter-min-stake" placeholder="0">
            </label>
            <label>Max Stake (micro)
              <input type="number" id="filter-max-stake" placeholder="∞">
            </label>
          </div>

          <div class="filter-group">
            <label>Locales (comma-separated)
              <input type="text" id="filter-locales" placeholder="en,es,fr">
            </label>
          </div>

          <div class="filter-group">
            <label>Min Misbehavior Reports
              <input type="number" id="filter-min-misbehavior" placeholder="0" min="0">
            </label>
            <label>Max Misbehavior Reports
              <input type="number" id="filter-max-misbehavior" placeholder="∞" min="0">
            </label>
          </div>

          <div class="filter-group">
            <label>Max Heartbeat Age (seconds)
              <input type="number" id="filter-heartbeat-age" placeholder="300">
            </label>
          </div>

          <div class="filter-group">
            <label>Owner Address
              <input type="text" id="filter-owner" placeholder="Search owner...">
            </label>
          </div>

          <div class="filter-group">
            <label>Assistant Address
              <input type="text" id="filter-assistant" placeholder="Search assistant...">
            </label>
          </div>

          <div class="filter-actions">
            <button id="apply-filters" class="primary">Apply Filters</button>
            <button id="clear-filters">Clear All</button>
          </div>
        </div>

        <div class="filter-section">
          <h3>Active Filters</h3>
          <div id="active-filters" class="active-filters-list">
            <p class="muted">No active filters</p>
          </div>
        </div>
      </div>
    `;

    this.attachEventListeners();
  }

  attachEventListeners() {
    const applyBtn = document.getElementById('apply-filters');
    const clearBtn = document.getElementById('clear-filters');

    if (applyBtn) {
      applyBtn.addEventListener('click', () => {
        this.readFiltersFromUI();
        this.onFilterChange();
      });
    }

    if (clearBtn) {
      clearBtn.addEventListener('click', () => {
        this.clearFilters();
        this.clearFilterUI();
        this.onFilterChange();
      });
    }
  }

  readFiltersFromUI() {
    this.filters.status = document.getElementById('filter-status')?.value || null;

    const minStake = document.getElementById('filter-min-stake')?.value;
    this.filters.minStake = minStake ? Number(minStake) : null;

    const maxStake = document.getElementById('filter-max-stake')?.value;
    this.filters.maxStake = maxStake ? Number(maxStake) : null;

    const locales = document.getElementById('filter-locales')?.value || '';
    this.filters.locales = locales ? locales.split(',').map(l => l.trim()).filter(l => l) : [];

    const minMisbehavior = document.getElementById('filter-min-misbehavior')?.value;
    this.filters.minMisbehavior = minMisbehavior ? Number(minMisbehavior) : null;

    const maxMisbehavior = document.getElementById('filter-max-misbehavior')?.value;
    this.filters.maxMisbehavior = maxMisbehavior ? Number(maxMisbehavior) : null;

    const heartbeatAge = document.getElementById('filter-heartbeat-age')?.value;
    this.filters.heartbeatAge = heartbeatAge ? Number(heartbeatAge) : null;

    this.filters.ownerAddress = document.getElementById('filter-owner')?.value || null;
    this.filters.assistantAddress = document.getElementById('filter-assistant')?.value || null;

    this.updateActiveFiltersDisplay();
  }

  clearFilterUI() {
    const inputs = [
      'filter-status', 'filter-min-stake', 'filter-max-stake',
      'filter-locales', 'filter-min-misbehavior', 'filter-max-misbehavior',
      'filter-heartbeat-age', 'filter-owner', 'filter-assistant'
    ];

    inputs.forEach(id => {
      const el = document.getElementById(id);
      if (el) el.value = '';
    });

    this.updateActiveFiltersDisplay();
  }

  updateActiveFiltersDisplay() {
    const container = document.getElementById('active-filters');
    if (!container) return;

    const active = this.getActiveFilters();

    if (active.length === 0) {
      container.innerHTML = '<p class="muted">No active filters</p>';
      return;
    }

    container.innerHTML = active.map(({ key, value }) => `
      <div class="filter-tag">
        <span class="filter-key">${key}:</span>
        <span class="filter-value">${Array.isArray(value) ? value.join(', ') : value}</span>
        <button class="filter-remove" data-filter="${key}">×</button>
      </div>
    `).join('');

    // Attach remove handlers
    container.querySelectorAll('.filter-remove').forEach(btn => {
      btn.addEventListener('click', (e) => {
        const filterKey = e.target.dataset.filter;
        this.setFilter(filterKey, null);
        this.clearFilterUI();
        this.onFilterChange();
      });
    });
  }

  onFilterChange() {
    // Override this method to handle filter changes
    console.log('Filters changed:', this.filters);
  }
}

if (typeof module !== 'undefined' && module.exports) {
  module.exports = AdvancedFilters;
}
