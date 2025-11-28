/**
 * Analytics Dashboard for Verifier Portal
 */

class VerifierAnalytics {
  constructor() {
    this.data = {
      assistants: [],
      scores: [],
      completions: [],
      alerts: []
    };
    this.charts = new Map();
  }

  async fetchAnalyticsData(baseRest) {
    const endpoints = [
      '/aura/aiassistant/v1beta1/assistants?pagination.limit=100',
      '/aura/confidencescore/v1beta1/analytics/summary',
      '/aura/confidencescore/v1beta1/analytics/trends'
    ];

    const results = await Promise.allSettled(
      endpoints.map(endpoint =>
        fetch(`${baseRest}${endpoint}`).then(r => r.json())
      )
    );

    return {
      assistants: results[0].status === 'fulfilled' ? results[0].value : null,
      summary: results[1].status === 'fulfilled' ? results[1].value : null,
      trends: results[2].status === 'fulfilled' ? results[2].value : null
    };
  }

  calculateMetrics(assistants) {
    const metrics = {
      total: assistants.length,
      active: 0,
      inactive: 0,
      totalStake: 0,
      totalSponsorship: 0,
      averageStake: 0,
      statusBreakdown: {},
      localeDistribution: {},
      misbehaviorRate: 0
    };

    let totalMisbehavior = 0;
    const now = Date.now() / 1000;
    const activeThreshold = 300; // 5 minutes

    for (const assistant of assistants) {
      // Status
      const status = assistant.status || 'unknown';
      metrics.statusBreakdown[status] = (metrics.statusBreakdown[status] || 0) + 1;

      // Active/Inactive based on heartbeat
      const lastHeartbeat = Number(assistant.last_heartbeat?.seconds || 0);
      if (now - lastHeartbeat < activeThreshold) {
        metrics.active++;
      } else {
        metrics.inactive++;
      }

      // Financial metrics
      const stake = Number(assistant.stake?.amount || 0);
      const sponsorship = Number(assistant.sponsorship_balance?.amount || 0);
      metrics.totalStake += stake;
      metrics.totalSponsorship += sponsorship;

      // Locale distribution
      for (const locale of assistant.locales || []) {
        metrics.localeDistribution[locale] = (metrics.localeDistribution[locale] || 0) + 1;
      }

      // Misbehavior
      totalMisbehavior += (assistant.misbehavior_reports || 0);
    }

    metrics.averageStake = metrics.total > 0 ? metrics.totalStake / metrics.total : 0;
    metrics.misbehaviorRate = metrics.total > 0 ? totalMisbehavior / metrics.total : 0;

    return metrics;
  }

  calculateScoreMetrics(completions) {
    const metrics = {
      totalCompletions: completions.length,
      arenaBreakdown: {},
      scoreDistribution: {
        positive: 0,
        negative: 0,
        neutral: 0
      },
      averageScore: 0,
      topArenas: []
    };

    let totalScore = 0;

    for (const completion of completions) {
      const arena = completion.arena || 'unknown';
      const scoreDelta = Number(completion.score_delta || 0);

      // Arena breakdown
      if (!metrics.arenaBreakdown[arena]) {
        metrics.arenaBreakdown[arena] = {
          count: 0,
          totalScore: 0
        };
      }
      metrics.arenaBreakdown[arena].count++;
      metrics.arenaBreakdown[arena].totalScore += scoreDelta;

      // Score distribution
      if (scoreDelta > 0) metrics.scoreDistribution.positive++;
      else if (scoreDelta < 0) metrics.scoreDistribution.negative++;
      else metrics.scoreDistribution.neutral++;

      totalScore += scoreDelta;
    }

    metrics.averageScore = metrics.totalCompletions > 0 ? totalScore / metrics.totalCompletions : 0;

    // Top arenas by activity
    metrics.topArenas = Object.entries(metrics.arenaBreakdown)
      .sort((a, b) => b[1].count - a[1].count)
      .slice(0, 5)
      .map(([arena, data]) => ({
        arena,
        count: data.count,
        avgScore: data.count > 0 ? data.totalScore / data.count : 0
      }));

    return metrics;
  }

  renderMetrics(containerId, metrics) {
    const container = document.getElementById(containerId);
    if (!container) return;

    container.innerHTML = `
      <div class="metrics-grid">
        <div class="metric-card">
          <div class="metric-label">Total Assistants</div>
          <div class="metric-value">${metrics.total}</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">Active</div>
          <div class="metric-value success">${metrics.active}</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">Inactive</div>
          <div class="metric-value warning">${metrics.inactive}</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">Total Stake</div>
          <div class="metric-value">${(metrics.totalStake / 1_000_000).toFixed(2)}</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">Avg Stake</div>
          <div class="metric-value">${(metrics.averageStake / 1_000_000).toFixed(2)}</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">Misbehavior Rate</div>
          <div class="metric-value ${metrics.misbehaviorRate > 0.1 ? 'danger' : 'success'}">
            ${(metrics.misbehaviorRate * 100).toFixed(2)}%
          </div>
        </div>
      </div>

      <div class="analytics-section">
        <h3>Status Distribution</h3>
        <div class="status-bars">
          ${Object.entries(metrics.statusBreakdown)
            .map(([status, count]) => `
              <div class="status-bar">
                <div class="status-label">${status}</div>
                <div class="status-bar-container">
                  <div class="status-bar-fill" style="width: ${(count / metrics.total * 100)}%"></div>
                </div>
                <div class="status-count">${count}</div>
              </div>
            `).join('')}
        </div>
      </div>

      <div class="analytics-section">
        <h3>Locale Distribution</h3>
        <div class="locale-grid">
          ${Object.entries(metrics.localeDistribution)
            .sort((a, b) => b[1] - a[1])
            .slice(0, 10)
            .map(([locale, count]) => `
              <div class="locale-card">
                <div class="locale-name">${locale}</div>
                <div class="locale-count">${count}</div>
              </div>
            `).join('')}
        </div>
      </div>
    `;
  }

  renderScoreAnalytics(containerId, metrics) {
    const container = document.getElementById(containerId);
    if (!container) return;

    container.innerHTML = `
      <div class="metrics-grid">
        <div class="metric-card">
          <div class="metric-label">Total Completions</div>
          <div class="metric-value">${metrics.totalCompletions}</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">Positive Scores</div>
          <div class="metric-value success">${metrics.scoreDistribution.positive}</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">Negative Scores</div>
          <div class="metric-value danger">${metrics.scoreDistribution.negative}</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">Average Score</div>
          <div class="metric-value">${metrics.averageScore.toFixed(2)}</div>
        </div>
      </div>

      <div class="analytics-section">
        <h3>Top Arenas by Activity</h3>
        <div class="arena-table">
          <table>
            <thead>
              <tr>
                <th>Arena</th>
                <th>Completions</th>
                <th>Avg Score</th>
              </tr>
            </thead>
            <tbody>
              ${metrics.topArenas.map(arena => `
                <tr>
                  <td>${arena.arena}</td>
                  <td>${arena.count}</td>
                  <td class="${arena.avgScore >= 0 ? 'success' : 'danger'}">
                    ${arena.avgScore.toFixed(2)}
                  </td>
                </tr>
              `).join('')}
            </tbody>
          </table>
        </div>
      </div>

      <div class="analytics-section">
        <h3>Arena Breakdown</h3>
        <div class="arena-bars">
          ${Object.entries(metrics.arenaBreakdown)
            .sort((a, b) => b[1].count - a[1].count)
            .map(([arena, data]) => `
              <div class="arena-bar">
                <div class="arena-label">${arena}</div>
                <div class="arena-bar-container">
                  <div class="arena-bar-fill" style="width: ${(data.count / metrics.totalCompletions * 100)}%"></div>
                </div>
                <div class="arena-stats">
                  ${data.count} (avg: ${(data.totalScore / data.count).toFixed(1)})
                </div>
              </div>
            `).join('')}
        </div>
      </div>
    `;
  }

  createTimeSeriesChart(containerId, data, options = {}) {
    const container = document.getElementById(containerId);
    if (!container) return;

    // Simple ASCII chart for demonstration
    // In production, use Chart.js or similar library
    const canvas = document.createElement('canvas');
    canvas.id = `${containerId}-canvas`;
    canvas.width = options.width || 600;
    canvas.height = options.height || 300;
    container.appendChild(canvas);

    // Store chart reference
    this.charts.set(containerId, {
      canvas,
      data,
      options
    });
  }

  updateChart(chartId, newData) {
    if (!this.charts.has(chartId)) return;

    const chart = this.charts.get(chartId);
    chart.data = newData;
    // Redraw logic here
  }

  exportData(format = 'json') {
    const exportData = {
      timestamp: new Date().toISOString(),
      data: this.data
    };

    if (format === 'json') {
      return JSON.stringify(exportData, null, 2);
    } else if (format === 'csv') {
      return this.convertToCSV(exportData);
    }
  }

  convertToCSV(data) {
    // Simple CSV conversion
    const rows = [];
    rows.push(['Timestamp', 'Type', 'Count', 'Details']);

    // Add data rows
    rows.push([data.timestamp, 'assistants', data.data.assistants.length, '']);
    rows.push([data.timestamp, 'scores', data.data.scores.length, '']);
    rows.push([data.timestamp, 'completions', data.data.completions.length, '']);

    return rows.map(row => row.join(',')).join('\n');
  }

  downloadExport(format = 'json') {
    const content = this.exportData(format);
    const blob = new Blob([content], { type: format === 'json' ? 'application/json' : 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `verifier-analytics-${Date.now()}.${format}`;
    a.click();
    URL.revokeObjectURL(url);
  }
}

if (typeof module !== 'undefined' && module.exports) {
  module.exports = VerifierAnalytics;
}
