#!/bin/bash
# AURA Inclusion Routine Manager
# Simple interface for adding, updating, deleting IRs without touching core code

set -e

CHAIN_ID="aura-1"
NODE="http://localhost:26657"
AUTHORITY="aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"  # Governance module address
KEYRING_BACKEND="test"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Display banner
echo -e "${BLUE}"
echo "╔════════════════════════════════════════════════╗"
echo "║   AURA Inclusion Routine Manager v1.0          ║"
echo "║   Manage IRs without touching core code        ║"
echo "╚════════════════════════════════════════════════╝"
echo -e "${NC}"

# Function to display menu
show_menu() {
    echo -e "\n${GREEN}Available Operations:${NC}"
    echo "  1) List all IRs"
    echo "  2) Get IR details"
    echo "  3) Create new IR"
    echo "  4) Update existing IR"
    echo "  5) Delete IR"
    echo "  6) Bulk import IRs from JSON"
    echo "  7) Export all IRs to JSON"
    echo "  8) Check IR statistics"
    echo "  9) Validate IR definitions"
    echo "  0) Exit"
    echo ""
}

# Function to list all IRs
list_irs() {
    echo -e "\n${BLUE}Querying all IRs...${NC}"
    aurad query inclusionroutines list-irs \
        --node "$NODE" \
        --output json | jq '.irs[] | {id: .id, name: .name, score: .score, status: .status}'
}

# Function to get IR details
get_ir() {
    read -p "Enter IR ID (e.g., IR-001): " ir_id
    echo -e "\n${BLUE}Querying IR: $ir_id${NC}"
    aurad query inclusionroutines ir "$ir_id" \
        --node "$NODE" \
        --output json | jq '.'
}

# Function to create new IR
create_ir() {
    echo -e "\n${GREEN}=== Create New IR ===${NC}"

    # Collect IR details
    read -p "IR ID (e.g., IR-301): " ir_id
    read -p "IR Name: " ir_name
    read -p "Description: " ir_description
    read -p "Score (10-30): " ir_score
    read -p "POI Reward: " poi_reward

    echo -e "\n${YELLOW}Select Arena:${NC}"
    echo "  1) ARENA_ANCHOR"
    echo "  2) ARENA_BIOMETRIC"
    echo "  3) ARENA_POSSESSION"
    echo "  4) ARENA_KNOWLEDGE"
    echo "  5) ARENA_SOCIAL"
    echo "  6) ARENA_GEOLOCATION"
    echo "  7) ARENA_HIGH_ASSURANCE"
    echo "  8) ARENA_PERSISTENCE"
    echo "  9) ARENA_SPECIALIZED"
    read -p "Arena (1-9): " arena_choice

    case $arena_choice in
        1) arena="ARENA_ANCHOR";;
        2) arena="ARENA_BIOMETRIC";;
        3) arena="ARENA_POSSESSION";;
        4) arena="ARENA_KNOWLEDGE";;
        5) arena="ARENA_SOCIAL";;
        6) arena="ARENA_GEOLOCATION";;
        7) arena="ARENA_HIGH_ASSURANCE";;
        8) arena="ARENA_PERSISTENCE";;
        9) arena="ARENA_SPECIALIZED";;
        *) echo -e "${RED}Invalid arena${NC}"; return;;
    esac

    echo -e "\n${YELLOW}Select Privacy Tier:${NC}"
    echo "  1) PRIVACY_TIER_LOW"
    echo "  2) PRIVACY_TIER_MEDIUM"
    echo "  3) PRIVACY_TIER_HIGH"
    read -p "Privacy Tier (1-3): " privacy_choice

    case $privacy_choice in
        1) privacy="PRIVACY_TIER_LOW";;
        2) privacy="PRIVACY_TIER_MEDIUM";;
        3) privacy="PRIVACY_TIER_HIGH";;
        *) echo -e "${RED}Invalid privacy tier${NC}"; return;;
    esac

    read -p "Locale tags (comma-separated, e.g., us,global): " locale_tags
    read -p "Version (e.g., 1.0): " version

    # Create the transaction
    echo -e "\n${BLUE}Creating IR transaction...${NC}"

    aurad tx inclusionroutines create-ir \
        "$ir_id" \
        "$ir_name" \
        "$arena" \
        "$ir_description" \
        "$ir_score" \
        "$poi_reward" \
        "$locale_tags" \
        "$privacy" \
        "$version" \
        "" \
        0 \
        0 \
        --from "$AUTHORITY" \
        --chain-id "$CHAIN_ID" \
        --node "$NODE" \
        --keyring-backend "$KEYRING_BACKEND" \
        --yes

    echo -e "${GREEN}✓ IR created successfully!${NC}"
}

# Function to update IR
update_ir() {
    echo -e "\n${GREEN}=== Update Existing IR ===${NC}"

    read -p "IR ID to update: " ir_id

    # Get current IR details
    echo -e "${BLUE}Fetching current IR details...${NC}"
    current_ir=$(aurad query inclusionroutines ir "$ir_id" --node "$NODE" --output json)

    if [ -z "$current_ir" ]; then
        echo -e "${RED}IR not found!${NC}"
        return
    fi

    echo -e "${YELLOW}Current IR details:${NC}"
    echo "$current_ir" | jq '.'

    read -p "New name (press Enter to keep current): " new_name
    read -p "New description (press Enter to keep current): " new_description
    read -p "New score (press Enter to keep current): " new_score
    read -p "New POI reward (press Enter to keep current): " new_poi

    # Build update command with non-empty values
    update_cmd="aurad tx inclusionroutines update-ir \"$ir_id\""

    if [ ! -z "$new_name" ]; then
        update_cmd="$update_cmd --name=\"$new_name\""
    fi

    if [ ! -z "$new_description" ]; then
        update_cmd="$update_cmd --description=\"$new_description\""
    fi

    if [ ! -z "$new_score" ]; then
        update_cmd="$update_cmd --score=$new_score"
    fi

    if [ ! -z "$new_poi" ]; then
        update_cmd="$update_cmd --poi-reward=$new_poi"
    fi

    update_cmd="$update_cmd --from=\"$AUTHORITY\" --chain-id=\"$CHAIN_ID\" --node=\"$NODE\" --keyring-backend=\"$KEYRING_BACKEND\" --yes"

    echo -e "\n${BLUE}Updating IR...${NC}"
    eval $update_cmd

    echo -e "${GREEN}✓ IR updated successfully!${NC}"
}

# Function to delete IR
delete_ir() {
    echo -e "\n${RED}=== Delete IR ===${NC}"

    read -p "IR ID to delete: " ir_id

    # Get IR details first
    ir_details=$(aurad query inclusionroutines ir "$ir_id" --node "$NODE" --output json 2>/dev/null || echo "")

    if [ -z "$ir_details" ]; then
        echo -e "${RED}IR not found!${NC}"
        return
    fi

    echo -e "${YELLOW}IR to be deleted:${NC}"
    echo "$ir_details" | jq '{id: .id, name: .name, score: .score}'

    read -p "Are you sure you want to delete this IR? (yes/no): " confirm

    if [ "$confirm" != "yes" ]; then
        echo -e "${YELLOW}Deletion cancelled${NC}"
        return
    fi

    echo -e "\n${BLUE}Deleting IR...${NC}"
    aurad tx inclusionroutines delete-ir "$ir_id" \
        --from "$AUTHORITY" \
        --chain-id "$CHAIN_ID" \
        --node "$NODE" \
        --keyring-backend "$KEYRING_BACKEND" \
        --yes

    echo -e "${GREEN}✓ IR deleted successfully!${NC}"
}

# Function to bulk import IRs from JSON
bulk_import() {
    echo -e "\n${GREEN}=== Bulk Import IRs ===${NC}"

    read -p "Path to JSON file: " json_file

    if [ ! -f "$json_file" ]; then
        echo -e "${RED}File not found!${NC}"
        return
    fi

    echo -e "${BLUE}Reading JSON file...${NC}"

    # Count IRs in file
    ir_count=$(jq '.irs | length' "$json_file")
    echo -e "${YELLOW}Found $ir_count IRs to import${NC}"

    read -p "Proceed with import? (yes/no): " confirm

    if [ "$confirm" != "yes" ]; then
        echo -e "${YELLOW}Import cancelled${NC}"
        return
    fi

    # Import each IR
    for i in $(seq 0 $((ir_count - 1))); do
        ir_id=$(jq -r ".irs[$i].id" "$json_file")
        ir_name=$(jq -r ".irs[$i].name" "$json_file")
        arena=$(jq -r ".irs[$i].arena" "$json_file")
        description=$(jq -r ".irs[$i].description" "$json_file")
        score=$(jq -r ".irs[$i].score" "$json_file")
        poi_reward=$(jq -r ".irs[$i].poi_reward" "$json_file")
        locale_tags=$(jq -r ".irs[$i].locale_tags | join(\",\")" "$json_file")
        privacy_tier=$(jq -r ".irs[$i].privacy_tier" "$json_file")
        version=$(jq -r ".irs[$i].version" "$json_file")

        echo -e "\n${BLUE}Importing $ir_id...${NC}"

        aurad tx inclusionroutines create-ir \
            "$ir_id" \
            "$ir_name" \
            "$arena" \
            "$description" \
            "$score" \
            "$poi_reward" \
            "$locale_tags" \
            "$privacy_tier" \
            "$version" \
            "" \
            0 \
            0 \
            --from "$AUTHORITY" \
            --chain-id "$CHAIN_ID" \
            --node "$NODE" \
            --keyring-backend "$KEYRING_BACKEND" \
            --yes \
            2>/dev/null || echo -e "${YELLOW}  (IR may already exist, skipping)${NC}"

        sleep 1  # Avoid overwhelming the node
    done

    echo -e "\n${GREEN}✓ Bulk import completed!${NC}"
}

# Function to export all IRs to JSON
export_irs() {
    echo -e "\n${GREEN}=== Export All IRs ===${NC}"

    read -p "Output file path (e.g., ./exported_irs.json): " output_file

    echo -e "${BLUE}Exporting IRs...${NC}"

    aurad query inclusionroutines list-irs \
        --node "$NODE" \
        --output json > "$output_file"

    echo -e "${GREEN}✓ IRs exported to $output_file${NC}"
}

# Function to show IR statistics
show_statistics() {
    echo -e "\n${GREEN}=== IR Statistics ===${NC}"

    all_irs=$(aurad query inclusionroutines list-irs --node "$NODE" --output json)

    total=$(echo "$all_irs" | jq '.irs | length')
    active=$(echo "$all_irs" | jq '[.irs[] | select(.status == "IR_STATUS_ACTIVE")] | length')
    suspended=$(echo "$all_irs" | jq '[.irs[] | select(.status == "IR_STATUS_SUSPENDED")] | length')

    echo -e "${BLUE}Total IRs:${NC} $total"
    echo -e "${GREEN}Active IRs:${NC} $active"
    echo -e "${RED}Suspended IRs:${NC} $suspended"

    echo -e "\n${BLUE}IRs by Arena:${NC}"
    echo "$all_irs" | jq -r '.irs | group_by(.arena) | .[] | "\(.[0].arena): \(length)"'

    echo -e "\n${BLUE}Point Distribution:${NC}"
    total_points=$(echo "$all_irs" | jq '[.irs[].score] | add')
    avg_points=$(echo "$all_irs" | jq '[.irs[].score] | add / length | floor')
    echo -e "Total possible points: $total_points"
    echo -e "Average points per IR: $avg_points"
}

# Function to validate IR definitions
validate_irs() {
    echo -e "\n${GREEN}=== Validate IR Definitions ===${NC}"

    all_irs=$(aurad query inclusionroutines list-irs --node "$NODE" --output json)

    echo -e "${BLUE}Running validation checks...${NC}\n"

    # Check 1: Point range (10-30)
    invalid_points=$(echo "$all_irs" | jq '[.irs[] | select(.score < 10 or .score > 30)] | length')
    if [ "$invalid_points" -gt 0 ]; then
        echo -e "${RED}✗ Found $invalid_points IRs with invalid point values (must be 10-30)${NC}"
        echo "$all_irs" | jq -r '.irs[] | select(.score < 10 or .score > 30) | "  - \(.id): \(.score) points"'
    else
        echo -e "${GREEN}✓ All IRs have valid point values (10-30)${NC}"
    fi

    # Check 2: Trinity categories
    doc_count=$(echo "$all_irs" | jq '[.irs[] | select(.trinity_category == "official_document")] | length')
    bio_count=$(echo "$all_irs" | jq '[.irs[] | select(.trinity_category == "biometric")] | length')
    activity_count=$(echo "$all_irs" | jq '[.irs[] | select(.trinity_category == "witnessed_activity")] | length')

    echo -e "\n${BLUE}Trinity Category Distribution:${NC}"
    echo -e "  Official Documents: $doc_count"
    echo -e "  Biometric: $bio_count"
    echo -e "  Witnessed Activity: $activity_count"

    if [ "$doc_count" -eq 0 ] || [ "$bio_count" -eq 0 ] || [ "$activity_count" -eq 0 ]; then
        echo -e "${RED}✗ Missing IRs in one or more trinity categories!${NC}"
    else
        echo -e "${GREEN}✓ All trinity categories have IRs${NC}"
    fi

    # Check 3: Duplicate IDs
    duplicate_ids=$(echo "$all_irs" | jq -r '.irs[].id' | sort | uniq -d)
    if [ ! -z "$duplicate_ids" ]; then
        echo -e "\n${RED}✗ Found duplicate IR IDs:${NC}"
        echo "$duplicate_ids"
    else
        echo -e "\n${GREEN}✓ No duplicate IR IDs${NC}"
    fi

    echo -e "\n${GREEN}Validation complete!${NC}"
}

# Main loop
while true; do
    show_menu
    read -p "Select operation (0-9): " choice

    case $choice in
        1) list_irs;;
        2) get_ir;;
        3) create_ir;;
        4) update_ir;;
        5) delete_ir;;
        6) bulk_import;;
        7) export_irs;;
        8) show_statistics;;
        9) validate_irs;;
        0) echo -e "\n${GREEN}Goodbye!${NC}\n"; exit 0;;
        *) echo -e "${RED}Invalid option. Please try again.${NC}";;
    esac

    read -p "Press Enter to continue..."
done
