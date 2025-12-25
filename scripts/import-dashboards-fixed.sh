#!/bin/bash
# Import Aura Grafana Dashboards - Fixed Version

set -e

GRAFANA_URL="http://localhost:3000"
GRAFANA_USER="admin"
GRAFANA_PASS="admin"
DASHBOARD_DIR="/home/decri/blockchain-projects/aura/grafana/dashboards"

echo "=================================="
echo "Importing Aura Grafana Dashboards"
echo "=================================="
echo ""

# Function to import a dashboard
import_dashboard() {
    local file=$1
    local name=$(basename "$file")

    echo -n "Importing $name... "

    # Import dashboard directly (files already have correct structure)
    RESPONSE=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d @"$file" \
        -u "$GRAFANA_USER:$GRAFANA_PASS" \
        "$GRAFANA_URL/api/dashboards/db")

    # Check if successful
    if echo "$RESPONSE" | grep -q '"status":"success"'; then
        echo "✓ SUCCESS"
    else
        ERROR=$(echo "$RESPONSE" | jq -r '.message // .error // "Unknown error"' 2>/dev/null || echo "Failed")
        echo "✗ FAILED: $ERROR"
    fi
}

# Import all dashboards
COUNT=0
for dashboard in "$DASHBOARD_DIR"/*.json; do
    if [ -f "$dashboard" ]; then
        import_dashboard "$dashboard"
        COUNT=$((COUNT + 1))
    fi
done

echo ""
echo "=================================="
echo "Imported $COUNT dashboards"
echo "=================================="
echo ""
echo "View dashboards at: $GRAFANA_URL/dashboards"
