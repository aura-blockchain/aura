#!/bin/bash
# Fix Aura Grafana dashboards by adding titles

DASHBOARD_DIR="/home/decri/blockchain-projects/aura/grafana/dashboards"

for file in "$DASHBOARD_DIR"/*.json; do
    filename=$(basename "$file" .json)
    title=$(echo "$filename" | sed 's/-/ /g' | sed 's/\b\(.\)/\u\1/g')

    echo "Fixing $filename... adding title: $title"

    # Add title if missing
    if ! grep -q '"title"' "$file"; then
        # Insert title after first {
        jq ". + {title: \"$title\"}" "$file" > "${file}.tmp" && mv "${file}.tmp" "$file"
    fi
done

echo "Done fixing dashboards"
