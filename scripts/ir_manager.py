#!/usr/bin/env python3
"""
AURA Inclusion Routine Manager (Python Version)
Simple interface for managing IRs without touching core code
"""

import json
import subprocess
import sys
from typing import Dict, List, Optional

# Configuration
CHAIN_ID = "aura-1"
NODE = "http://localhost:26657"
AUTHORITY = "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"
KEYRING_BACKEND = "test"

# ANSI color codes
class Colors:
    RED = '\033[0;31m'
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    BLUE = '\033[0;34m'
    NC = '\033[0m'  # No Color

def print_banner():
    """Display welcome banner"""
    print(f"{Colors.BLUE}")
    print("╔════════════════════════════════════════════════╗")
    print("║   AURA Inclusion Routine Manager v1.0          ║")
    print("║   Python Edition - Manage IRs Easily           ║")
    print("╚════════════════════════════════════════════════╝")
    print(f"{Colors.NC}")

def show_menu():
    """Display main menu"""
    print(f"\n{Colors.GREEN}Available Operations:{Colors.NC}")
    print("  1) List all IRs")
    print("  2) Get IR details")
    print("  3) Create new IR")
    print("  4) Update existing IR")
    print("  5) Delete IR")
    print("  6) Bulk import IRs from JSON")
    print("  7) Export all IRs to JSON")
    print("  8) Check IR statistics")
    print("  9) Validate IR definitions")
    print("  0) Exit")
    print("")

def run_command(cmd: List[str], capture_output: bool = True) -> Optional[str]:
    """Execute shell command"""
    try:
        if capture_output:
            result = subprocess.run(cmd, capture_output=True, text=True, check=True)
            return result.stdout
        else:
            subprocess.run(cmd, check=True)
            return None
    except subprocess.CalledProcessError as e:
        print(f"{Colors.RED}Error executing command: {e}{Colors.NC}")
        return None

def list_irs():
    """List all IRs"""
    print(f"\n{Colors.BLUE}Querying all IRs...{Colors.NC}")

    cmd = [
        "aurad", "query", "inclusionroutines", "list-irs",
        "--node", NODE,
        "--output", "json"
    ]

    output = run_command(cmd)
    if output:
        try:
            data = json.loads(output)
            irs = data.get('irs', [])

            print(f"\n{Colors.GREEN}Found {len(irs)} IRs:{Colors.NC}")
            for ir in irs:
                print(f"  {ir['id']}: {ir['name']} ({ir['score']} pts) - {ir['status']}")
        except json.JSONDecodeError:
            print(f"{Colors.RED}Error parsing JSON response{Colors.NC}")

def get_ir_details():
    """Get details of a specific IR"""
    ir_id = input("Enter IR ID (e.g., IR-001): ").strip()

    print(f"\n{Colors.BLUE}Querying IR: {ir_id}{Colors.NC}")

    cmd = [
        "aurad", "query", "inclusionroutines", "ir", ir_id,
        "--node", NODE,
        "--output", "json"
    ]

    output = run_command(cmd)
    if output:
        try:
            data = json.loads(output)
            print(json.dumps(data, indent=2))
        except json.JSONDecodeError:
            print(f"{Colors.RED}Error parsing JSON response{Colors.NC}")

def create_ir():
    """Create a new IR interactively"""
    print(f"\n{Colors.GREEN}=== Create New IR ==={Colors.NC}")

    # Collect IR details
    ir_id = input("IR ID (e.g., IR-301): ").strip()
    ir_name = input("IR Name: ").strip()
    ir_description = input("Description: ").strip()
    ir_score = input("Score (10-30): ").strip()
    poi_reward = input("POI Reward: ").strip()

    # Arena selection
    print(f"\n{Colors.YELLOW}Select Arena:{Colors.NC}")
    arenas = [
        "ARENA_ANCHOR",
        "ARENA_BIOMETRIC",
        "ARENA_POSSESSION",
        "ARENA_KNOWLEDGE",
        "ARENA_SOCIAL",
        "ARENA_GEOLOCATION",
        "ARENA_HIGH_ASSURANCE",
        "ARENA_PERSISTENCE",
        "ARENA_SPECIALIZED"
    ]

    for i, arena in enumerate(arenas, 1):
        print(f"  {i}) {arena}")

    arena_choice = int(input("Arena (1-9): ").strip())
    if 1 <= arena_choice <= 9:
        arena = arenas[arena_choice - 1]
    else:
        print(f"{Colors.RED}Invalid arena{Colors.NC}")
        return

    # Privacy tier selection
    print(f"\n{Colors.YELLOW}Select Privacy Tier:{Colors.NC}")
    privacy_tiers = [
        "PRIVACY_TIER_LOW",
        "PRIVACY_TIER_MEDIUM",
        "PRIVACY_TIER_HIGH"
    ]

    for i, tier in enumerate(privacy_tiers, 1):
        print(f"  {i}) {tier}")

    privacy_choice = int(input("Privacy Tier (1-3): ").strip())
    if 1 <= privacy_choice <= 3:
        privacy = privacy_tiers[privacy_choice - 1]
    else:
        print(f"{Colors.RED}Invalid privacy tier{Colors.NC}")
        return

    locale_tags = input("Locale tags (comma-separated, e.g., us,global): ").strip()
    version = input("Version (e.g., 1.0): ").strip()

    # Create the transaction
    print(f"\n{Colors.BLUE}Creating IR transaction...{Colors.NC}")

    cmd = [
        "aurad", "tx", "inclusionroutines", "create-ir",
        ir_id, ir_name, arena, ir_description,
        ir_score, poi_reward, locale_tags, privacy, version,
        "", "0", "0",
        "--from", AUTHORITY,
        "--chain-id", CHAIN_ID,
        "--node", NODE,
        "--keyring-backend", KEYRING_BACKEND,
        "--yes"
    ]

    if run_command(cmd, capture_output=False) is not None:
        print(f"{Colors.GREEN}✓ IR created successfully!{Colors.NC}")

def update_ir():
    """Update an existing IR"""
    print(f"\n{Colors.GREEN}=== Update Existing IR ==={Colors.NC}")

    ir_id = input("IR ID to update: ").strip()

    # Get current IR details
    print(f"{Colors.BLUE}Fetching current IR details...{Colors.NC}")
    cmd = ["aurad", "query", "inclusionroutines", "ir", ir_id, "--node", NODE, "--output", "json"]
    output = run_command(cmd)

    if not output:
        print(f"{Colors.RED}IR not found!{Colors.NC}")
        return

    print(f"{Colors.YELLOW}Current IR details:{Colors.NC}")
    print(output)

    new_name = input("New name (press Enter to keep current): ").strip()
    new_description = input("New description (press Enter to keep current): ").strip()
    new_score = input("New score (press Enter to keep current): ").strip()
    new_poi = input("New POI reward (press Enter to keep current): ").strip()

    # Build update command
    cmd = [
        "aurad", "tx", "inclusionroutines", "update-ir", ir_id,
        "--from", AUTHORITY,
        "--chain-id", CHAIN_ID,
        "--node", NODE,
        "--keyring-backend", KEYRING_BACKEND,
        "--yes"
    ]

    if new_name:
        cmd.extend(["--name", new_name])
    if new_description:
        cmd.extend(["--description", new_description])
    if new_score:
        cmd.extend(["--score", new_score])
    if new_poi:
        cmd.extend(["--poi-reward", new_poi])

    print(f"\n{Colors.BLUE}Updating IR...{Colors.NC}")
    if run_command(cmd, capture_output=False) is not None:
        print(f"{Colors.GREEN}✓ IR updated successfully!{Colors.NC}")

def delete_ir():
    """Delete an IR"""
    print(f"\n{Colors.RED}=== Delete IR ==={Colors.NC}")

    ir_id = input("IR ID to delete: ").strip()

    # Get IR details first
    cmd = ["aurad", "query", "inclusionroutines", "ir", ir_id, "--node", NODE, "--output", "json"]
    output = run_command(cmd)

    if not output:
        print(f"{Colors.RED}IR not found!{Colors.NC}")
        return

    print(f"{Colors.YELLOW}IR to be deleted:{Colors.NC}")
    try:
        ir_data = json.loads(output)
        print(f"  ID: {ir_data.get('id')}")
        print(f"  Name: {ir_data.get('name')}")
        print(f"  Score: {ir_data.get('score')}")
    except json.JSONDecodeError:
        print(output)

    confirm = input("Are you sure you want to delete this IR? (yes/no): ").strip().lower()

    if confirm != "yes":
        print(f"{Colors.YELLOW}Deletion cancelled{Colors.NC}")
        return

    print(f"\n{Colors.BLUE}Deleting IR...{Colors.NC}")

    cmd = [
        "aurad", "tx", "inclusionroutines", "delete-ir", ir_id,
        "--from", AUTHORITY,
        "--chain-id", CHAIN_ID,
        "--node", NODE,
        "--keyring-backend", KEYRING_BACKEND,
        "--yes"
    ]

    if run_command(cmd, capture_output=False) is not None:
        print(f"{Colors.GREEN}✓ IR deleted successfully!{Colors.NC}")

def bulk_import():
    """Import IRs from JSON file"""
    print(f"\n{Colors.GREEN}=== Bulk Import IRs ==={Colors.NC}")

    json_file = input("Path to JSON file: ").strip()

    try:
        with open(json_file, 'r') as f:
            data = json.load(f)
    except FileNotFoundError:
        print(f"{Colors.RED}File not found!{Colors.NC}")
        return
    except json.JSONDecodeError:
        print(f"{Colors.RED}Invalid JSON file!{Colors.NC}")
        return

    irs = data.get('irs', [])
    print(f"{Colors.YELLOW}Found {len(irs)} IRs to import{Colors.NC}")

    confirm = input("Proceed with import? (yes/no): ").strip().lower()

    if confirm != "yes":
        print(f"{Colors.YELLOW}Import cancelled{Colors.NC}")
        return

    # Import each IR
    success_count = 0
    for ir in irs:
        ir_id = ir.get('id')
        print(f"\n{Colors.BLUE}Importing {ir_id}...{Colors.NC}")

        cmd = [
            "aurad", "tx", "inclusionroutines", "create-ir",
            ir.get('id'),
            ir.get('name'),
            ir.get('arena'),
            ir.get('description'),
            str(ir.get('score')),
            str(ir.get('poi_reward')),
            ','.join(ir.get('locale_tags', [])),
            ir.get('privacy_tier'),
            ir.get('version'),
            "",
            "0",
            "0",
            "--from", AUTHORITY,
            "--chain-id", CHAIN_ID,
            "--node", NODE,
            "--keyring-backend", KEYRING_BACKEND,
            "--yes"
        ]

        if run_command(cmd, capture_output=False) is not None:
            success_count += 1

        # Rate limit to avoid overwhelming the node
        import time
        time.sleep(1)

    print(f"\n{Colors.GREEN}✓ Imported {success_count}/{len(irs)} IRs successfully!{Colors.NC}")

def export_irs():
    """Export all IRs to JSON file"""
    print(f"\n{Colors.GREEN}=== Export All IRs ==={Colors.NC}")

    output_file = input("Output file path (e.g., ./exported_irs.json): ").strip()

    print(f"{Colors.BLUE}Exporting IRs...{Colors.NC}")

    cmd = [
        "aurad", "query", "inclusionroutines", "list-irs",
        "--node", NODE,
        "--output", "json"
    ]

    output = run_command(cmd)
    if output:
        try:
            with open(output_file, 'w') as f:
                f.write(output)
            print(f"{Colors.GREEN}✓ IRs exported to {output_file}{Colors.NC}")
        except IOError as e:
            print(f"{Colors.RED}Error writing file: {e}{Colors.NC}")

def show_statistics():
    """Show IR statistics"""
    print(f"\n{Colors.GREEN}=== IR Statistics ==={Colors.NC}")

    cmd = ["aurad", "query", "inclusionroutines", "list-irs", "--node", NODE, "--output", "json"]
    output = run_command(cmd)

    if not output:
        return

    try:
        data = json.loads(output)
        irs = data.get('irs', [])

        total = len(irs)
        active = len([ir for ir in irs if ir.get('status') == 'IR_STATUS_ACTIVE'])
        suspended = len([ir for ir in irs if ir.get('status') == 'IR_STATUS_SUSPENDED'])

        print(f"{Colors.BLUE}Total IRs:{Colors.NC} {total}")
        print(f"{Colors.GREEN}Active IRs:{Colors.NC} {active}")
        print(f"{Colors.RED}Suspended IRs:{Colors.NC} {suspended}")

        # IRs by arena
        print(f"\n{Colors.BLUE}IRs by Arena:{Colors.NC}")
        arenas = {}
        for ir in irs:
            arena = ir.get('arena', 'Unknown')
            arenas[arena] = arenas.get(arena, 0) + 1

        for arena, count in sorted(arenas.items()):
            print(f"  {arena}: {count}")

        # Point distribution
        print(f"\n{Colors.BLUE}Point Distribution:{Colors.NC}")
        scores = [ir.get('score', 0) for ir in irs]
        total_points = sum(scores)
        avg_points = total_points // len(scores) if scores else 0

        print(f"Total possible points: {total_points}")
        print(f"Average points per IR: {avg_points}")

    except json.JSONDecodeError:
        print(f"{Colors.RED}Error parsing JSON response{Colors.NC}")

def validate_irs():
    """Validate IR definitions"""
    print(f"\n{Colors.GREEN}=== Validate IR Definitions ==={Colors.NC}")

    cmd = ["aurad", "query", "inclusionroutines", "list-irs", "--node", NODE, "--output", "json"]
    output = run_command(cmd)

    if not output:
        return

    try:
        data = json.loads(output)
        irs = data.get('irs', [])

        print(f"{Colors.BLUE}Running validation checks...{Colors.NC}\n")

        # Check 1: Point range (10-30)
        invalid_points = [ir for ir in irs if ir.get('score', 0) < 10 or ir.get('score', 0) > 30]
        if invalid_points:
            print(f"{Colors.RED}✗ Found {len(invalid_points)} IRs with invalid point values (must be 10-30){Colors.NC}")
            for ir in invalid_points:
                print(f"  - {ir['id']}: {ir['score']} points")
        else:
            print(f"{Colors.GREEN}✓ All IRs have valid point values (10-30){Colors.NC}")

        # Check 2: Trinity categories
        trinity_cats = {}
        for ir in irs:
            cat = ir.get('trinity_category', '')
            if cat:
                trinity_cats[cat] = trinity_cats.get(cat, 0) + 1

        print(f"\n{Colors.BLUE}Trinity Category Distribution:{Colors.NC}")
        print(f"  Official Documents: {trinity_cats.get('official_document', 0)}")
        print(f"  Biometric: {trinity_cats.get('biometric', 0)}")
        print(f"  Witnessed Activity: {trinity_cats.get('witnessed_activity', 0)}")

        if all(trinity_cats.get(cat, 0) > 0 for cat in ['official_document', 'biometric', 'witnessed_activity']):
            print(f"{Colors.GREEN}✓ All trinity categories have IRs{Colors.NC}")
        else:
            print(f"{Colors.RED}✗ Missing IRs in one or more trinity categories!{Colors.NC}")

        # Check 3: Duplicate IDs
        ids = [ir.get('id') for ir in irs]
        duplicates = [id for id in ids if ids.count(id) > 1]

        if duplicates:
            print(f"\n{Colors.RED}✗ Found duplicate IR IDs:{Colors.NC}")
            for dup in set(duplicates):
                print(f"  - {dup}")
        else:
            print(f"\n{Colors.GREEN}✓ No duplicate IR IDs{Colors.NC}")

        print(f"\n{Colors.GREEN}Validation complete!{Colors.NC}")

    except json.JSONDecodeError:
        print(f"{Colors.RED}Error parsing JSON response{Colors.NC}")

def main():
    """Main program loop"""
    print_banner()

    while True:
        show_menu()
        choice = input("Select operation (0-9): ").strip()

        if choice == '1':
            list_irs()
        elif choice == '2':
            get_ir_details()
        elif choice == '3':
            create_ir()
        elif choice == '4':
            update_ir()
        elif choice == '5':
            delete_ir()
        elif choice == '6':
            bulk_import()
        elif choice == '7':
            export_irs()
        elif choice == '8':
            show_statistics()
        elif choice == '9':
            validate_irs()
        elif choice == '0':
            print(f"\n{Colors.GREEN}Goodbye!{Colors.NC}\n")
            sys.exit(0)
        else:
            print(f"{Colors.RED}Invalid option. Please try again.{Colors.NC}")

        input("\nPress Enter to continue...")

if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print(f"\n\n{Colors.YELLOW}Interrupted by user{Colors.NC}")
        sys.exit(0)
