#!/bin/bash
#
# Background Test Runner
# Runs the complete test suite in the background using nohup
# This ensures tests continue even if the user or agent logs off
#
# Usage: ./run_tests_background.sh [cpu_limit_percent]
#

set -e

# Configuration
CPU_LIMIT="${1:-75}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
PID_FILE="/tmp/aura_test_suite.pid"
STATUS_FILE="/tmp/aura_test_suite.status"
NOHUP_LOG="/home/hudson/blockchain-projects/aura/chain/testing/test-logs/nohup_${TIMESTAMP}.log"

# Create log directory
mkdir -p "/home/hudson/blockchain-projects/aura/chain/testing/test-logs"

# Check if already running
if [ -f "$PID_FILE" ]; then
    OLD_PID=$(cat "$PID_FILE")
    if ps -p "$OLD_PID" > /dev/null 2>&1; then
        echo "ERROR: Test suite is already running (PID: $OLD_PID)"
        echo "Check status: cat /tmp/aura_test_suite.status"
        echo "Check logs: tail -f /home/hudson/blockchain-projects/aura/chain/testing/test-logs/nohup_*.log"
        exit 1
    else
        echo "Removing stale PID file"
        rm -f "$PID_FILE"
    fi
fi

echo "========================================="
echo "Aura Test Suite - Background Runner"
echo "========================================="
echo "CPU Limit: ${CPU_LIMIT}%"
echo "Nohup Log: $NOHUP_LOG"
echo ""
echo "Starting test suite in background..."

# Initialize status file
cat > "$STATUS_FILE" <<EOF
Status: RUNNING
Started: $(date)
CPU Limit: ${CPU_LIMIT}%
PID: Will be updated shortly
Log: $NOHUP_LOG
EOF

# Run in background with nohup
cd /home/hudson/blockchain-projects/aura/chain/testing
nohup ./run_complete_test_suite_throttled.sh "$CPU_LIMIT" > "$NOHUP_LOG" 2>&1 &
TEST_PID=$!

# Save PID
echo "$TEST_PID" > "$PID_FILE"

# Update status with PID
cat > "$STATUS_FILE" <<EOF
Status: RUNNING
Started: $(date)
CPU Limit: ${CPU_LIMIT}%
PID: $TEST_PID
Log: $NOHUP_LOG
EOF

echo "Test suite started successfully!"
echo ""
echo "PID: $TEST_PID"
echo "Status file: $STATUS_FILE"
echo "Log file: $NOHUP_LOG"
echo ""
echo "To check progress:"
echo "  tail -f $NOHUP_LOG"
echo "  cat $STATUS_FILE"
echo ""
echo "To check if running:"
echo "  ps -p $TEST_PID"
echo ""
echo "The tests will continue running even if you log off."
echo "========================================="

# Create completion monitor script
cat > /tmp/monitor_aura_tests.sh <<'MONITOR_EOF'
#!/bin/bash
PID_FILE="/tmp/aura_test_suite.pid"
STATUS_FILE="/tmp/aura_test_suite.status"

if [ ! -f "$PID_FILE" ]; then
    echo "No test suite is running"
    exit 1
fi

TEST_PID=$(cat "$PID_FILE")

if ps -p "$TEST_PID" > /dev/null 2>&1; then
    echo "Test suite is still RUNNING (PID: $TEST_PID)"
    echo ""
    cat "$STATUS_FILE"
else
    echo "Test suite has COMPLETED"
    echo ""
    cat "$STATUS_FILE"
    echo ""
    echo "Check the latest summary in:"
    ls -t /home/hudson/blockchain-projects/aura/chain/testing/test-logs/summary_*.log | head -1
fi
MONITOR_EOF

chmod +x /tmp/monitor_aura_tests.sh

echo "Monitor script created: /tmp/monitor_aura_tests.sh"
echo "Run it anytime to check status: /tmp/monitor_aura_tests.sh"
