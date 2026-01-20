#!/bin/bash
# AURA Testnet Automation Setup Script
# Deploys and configures all automation scripts on testnet servers
#
# Run from bcpc: ./setup-automation.sh
#
# This script will:
# 1. Copy automation scripts to both testnet servers
# 2. Set up cron jobs for health checks, recovery, and backups
# 3. Configure systemd services for Prometheus metrics
# 4. Enable log rotation

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVERS=("aura-testnet" "services-testnet")
REMOTE_SCRIPTS_DIR="/opt/aura/scripts"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"
}

# Deploy scripts to a server
deploy_scripts() {
    local server="$1"
    log "Deploying scripts to $server..."

    # Create directory (requires sudo for /opt)
    ssh "$server" "sudo mkdir -p $REMOTE_SCRIPTS_DIR && sudo chown ubuntu:ubuntu $REMOTE_SCRIPTS_DIR"

    # Copy local scripts (these run directly on the server without SSH)
    scp -q "$SCRIPT_DIR/aura-health-monitor.sh" "$server:$REMOTE_SCRIPTS_DIR/"
    scp -q "$SCRIPT_DIR/health-check-local.sh" "$server:$REMOTE_SCRIPTS_DIR/"
    scp -q "$SCRIPT_DIR/auto-recovery-local.sh" "$server:$REMOTE_SCRIPTS_DIR/"
    scp -q "$SCRIPT_DIR/backup.sh" "$server:$REMOTE_SCRIPTS_DIR/"
    scp -q "$SCRIPT_DIR/disk-maintenance.sh" "$server:$REMOTE_SCRIPTS_DIR/"
    scp -q "$SCRIPT_DIR/prometheus-alerts.yml" "$server:$REMOTE_SCRIPTS_DIR/"

    # Make executable
    ssh "$server" "chmod +x $REMOTE_SCRIPTS_DIR/*.sh"

    log "Scripts deployed to $server"
}

# Set up cron jobs on a server
setup_cron() {
    local server="$1"
    log "Setting up cron jobs on $server..."

    # First create the log directory with proper permissions
    ssh "$server" "sudo mkdir -p /var/log/aura && sudo chown ubuntu:ubuntu /var/log/aura"

    ssh "$server" 'cat > /tmp/aura-cron << "CRON_EOF"
# AURA Testnet Automation Cron Jobs
# Managed by setup-automation.sh - manual edits may be overwritten

# Health check every 2 minutes (local script, no SSH)
*/2 * * * * /opt/aura/scripts/aura-health-monitor.sh --json >> /var/log/aura/health.log 2>&1

# Auto-recovery every 5 minutes (local script, no SSH)
*/5 * * * * /opt/aura/scripts/auto-recovery-local.sh >> /var/log/aura/recovery.log 2>&1

# Daily config backup at 2 AM UTC
0 2 * * * /opt/aura/scripts/backup.sh >> /var/log/aura/backup.log 2>&1

# Weekly full backup on Sunday at 3 AM UTC
0 3 * * 0 /opt/aura/scripts/backup.sh --full >> /var/log/aura/backup.log 2>&1

# Log rotation check daily at 1 AM
0 1 * * * /usr/sbin/logrotate /etc/logrotate.d/aura-testnet 2>/dev/null || true

# Cleanup old tmp files daily at 4 AM
0 4 * * * find /tmp -name "aura-*" -mtime +7 -delete 2>/dev/null || true
CRON_EOF
crontab /tmp/aura-cron
rm /tmp/aura-cron'

    log "Cron jobs configured on $server"
}

# Set up log rotation
setup_logrotate() {
    local server="$1"
    log "Setting up log rotation on $server..."

    ssh "$server" 'sudo tee /etc/logrotate.d/aura-testnet > /dev/null << "LOGROTATE_EOF"
/var/log/aura/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0644 ubuntu ubuntu
}

/home/ubuntu/.aura-mvp-*/logs/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    copytruncate
}
LOGROTATE_EOF'

    log "Log rotation configured on $server"
}

# Enable Prometheus metrics on nodes
setup_prometheus() {
    local server="$1"
    log "Enabling Prometheus metrics on $server..."

    # Update config.toml for all nodes to enable metrics
    ssh "$server" '
for home in ~/.aura-mvp-*; do
    if [[ -f "$home/config/config.toml" ]]; then
        node=$(basename "$home")
        # Enable prometheus metrics
        sed -i "s/prometheus = false/prometheus = true/" "$home/config/config.toml" 2>/dev/null || true
        # Set unique port per node to avoid conflicts
        case "$node" in
            .aura-mvp-val1) port=26660 ;;
            .aura-mvp-val2) port=26760 ;;
            .aura-mvp-val3) port=26660 ;;
            .aura-mvp-val4) port=26760 ;;
            .aura-mvp-sentry1) port=26860 ;;
            .aura-mvp-sentry2) port=26860 ;;
            *) port=26660 ;;
        esac
        sed -i "s/prometheus_listen_addr = .*/prometheus_listen_addr = \":$port\"/" "$home/config/config.toml" 2>/dev/null || true
        echo "Configured metrics for $node on port $port"
    fi
done'

    log "Prometheus metrics enabled on $server (restart nodes to apply)"
}

# Create systemd drop-in for better restart handling
setup_systemd_improvements() {
    local server="$1"
    log "Improving systemd service configuration on $server..."

    ssh "$server" 'for service in aurad-mvp-val1 aurad-mvp-val2 aurad-mvp-val3 aurad-mvp-val4 aurad-mvp-sentry1 aurad-mvp-sentry2; do
    if systemctl list-unit-files | grep -q "$service"; then
        sudo mkdir -p "/etc/systemd/system/${service}.service.d"
        sudo tee "/etc/systemd/system/${service}.service.d/override.conf" > /dev/null << EOF
[Unit]
# Rate limiting for restarts (must be in [Unit] section)
StartLimitIntervalSec=300
StartLimitBurst=5

[Service]
# Restart behavior
Restart=always
RestartSec=5

# Resource limits
LimitNOFILE=65535
LimitNPROC=65535

# OOM handling
OOMScoreAdjust=-500

# Watchdog (optional - requires app support)
# WatchdogSec=60
EOF
        echo "Configured override for $service"
    fi
done
sudo systemctl daemon-reload'

    log "Systemd improvements applied on $server"
}

# Verify deployment
verify_deployment() {
    local server="$1"
    log "Verifying deployment on $server..."

    local checks_passed=0
    local checks_failed=0

    # Check scripts exist
    if ssh "$server" "test -x $REMOTE_SCRIPTS_DIR/aura-health-monitor.sh"; then
        ((checks_passed++))
    else
        log "FAIL: aura-health-monitor.sh not executable"
        ((checks_failed++))
    fi

    # Check cron is set up
    if ssh "$server" "crontab -l 2>/dev/null | grep -q aura-health"; then
        ((checks_passed++))
    else
        log "FAIL: cron not configured"
        ((checks_failed++))
    fi

    # Check logrotate
    if ssh "$server" "test -f /etc/logrotate.d/aura-testnet"; then
        ((checks_passed++))
    else
        log "FAIL: logrotate not configured"
        ((checks_failed++))
    fi

    log "Verification on $server: $checks_passed passed, $checks_failed failed"
    return $checks_failed
}

# Run a test health check
test_health_check() {
    local server="$1"
    log "Running test health check on $server..."

    ssh "$server" "$REMOTE_SCRIPTS_DIR/aura-health-monitor.sh" || {
        log "WARNING: Health check reported issues (this may be expected)"
    }
}

# Main
main() {
    log "=========================================="
    log "AURA Testnet Automation Setup"
    log "=========================================="

    local total_failures=0

    for server in "${SERVERS[@]}"; do
        log ""
        log "Configuring $server..."
        log "------------------------------------------"

        deploy_scripts "$server"
        setup_cron "$server"
        setup_logrotate "$server"
        setup_prometheus "$server"
        setup_systemd_improvements "$server"

        if ! verify_deployment "$server"; then
            ((total_failures++))
        fi

        test_health_check "$server"
    done

    log ""
    log "=========================================="
    log "Setup Complete"
    log "=========================================="
    log ""
    log "Automation is now active on both servers:"
    log "  - Health checks: every 2 minutes"
    log "  - Auto-recovery: every 5 minutes"
    log "  - Config backups: daily at 2 AM UTC"
    log "  - Full backups: weekly on Sunday at 3 AM UTC"
    log ""
    log "Logs are available at:"
    log "  - /var/log/aura-health.log"
    log "  - /var/log/aura-recovery.log"
    log "  - /var/log/aura-backup.log"
    log ""
    log "Prometheus metrics available on nodes (after restart):"
    log "  - val1/val3: port 26660"
    log "  - val2/val4: port 26760"
    log "  - sentries: port 26860"
    log ""

    if [[ $total_failures -gt 0 ]]; then
        log "WARNING: $total_failures servers had verification failures"
        exit 1
    fi

    log "To apply Prometheus config, restart nodes:"
    log "  ssh aura-testnet 'sudo systemctl restart aurad-mvp-val1 aurad-mvp-val2 aurad-mvp-sentry1'"
    log "  ssh services-testnet 'sudo systemctl restart aurad-mvp-val3 aurad-mvp-val4 aurad-mvp-sentry2'"
}

main "$@"
