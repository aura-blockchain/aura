#!/bin/bash
# Import Aura Grafana Dashboards
# Simple script to import all dashboards without needing API tokens

set -e

GRAFANA_URL="http://localhost:3000"
GRAFANA_USER="admin"
GRAFANA_PASS="admin"
DASHBOARD_DIR="/home/decri/blockchain-projects/aura/grafana/dashboards"

echo "=================================="
echo "Importing Aura Grafana Dashboards"
echo "=================================="
echo ""

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo "Installing jq..."
    echo "2885" | sudo -S apt install -y jq
fi

# Function to import a dashboard
import_dashboard() {
    local file=$1
    local name=$(basename "$file")

    echo -n "Importing $name... "

    # Read dashboard JSON and wrap it
    DASHBOARD_JSON=$(cat "$file")

    # Create import payload
    IMPORT_PAYLOAD=$(jq -n \
        --argjson dashboard "$DASHBOARD_JSON" \
        '{
            dashboard: $dashboard,
            overwrite: true,
            inputs: [],
            folderId: 0
        }')

    # Import dashboard
    RESPONSE=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$IMPORT_PAYLOAD" \
        -u "$GRAFANA_USER:$GRAFANA_PASS" \
        "$GRAFANA_URL/api/dashboards/db")

    # Check if successful
    if echo "$RESPONSE" | jq -e '.status == "success"' > /dev/null 2>&1; then
        echo "✓ SUCCESS"
    elif echo "$RESPONSE" | jq -e '.message' > /dev/null 2>&1; then
        ERROR=$(echo "$RESPONSE" | jq -r '.message')
        echo "✗ FAILED: $ERROR"
    else
        echo "✗ FAILED"
    fi
}

# Import all dashboards
for dashboard in "$DASHBOARD_DIR"/*.json; do
    if [ -f "$dashboard" ]; then
        import_dashboard "$dashboard"
    fi
done

echo ""
echo "=================================="
echo "Dashboard Import Complete"
echo "=================================="
echo ""
echo "View dashboards at: $GRAFANA_URL/dashboards"
