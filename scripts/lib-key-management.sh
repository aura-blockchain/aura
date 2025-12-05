#!/bin/bash
# ============================================================================
# Key Management Library for AURA Blockchain
# ============================================================================
# This library provides secure key management utilities following blockchain
# industry best practices. It supports multiple keyring backends, deterministic
# key generation for testing, and comprehensive validation.
#
# Security Principles:
# - Never hardcode secrets in scripts
# - Use environment variables for sensitive data
# - Support multiple keyring backends (file, os, test, memory)
# - Validate key availability before operations
# - Generate deterministic keys for reproducible testing
#
# Usage:
#   source scripts/lib-key-management.sh
#   setup_keyring_backend "test"
#   generate_key_deterministic "validator1" "$HOME/.aura" "test-mnemonic-seed-1"
# ============================================================================

# Colors for output
KEY_MGMT_RED='\033[0;31m'
KEY_MGMT_GREEN='\033[0;32m'
KEY_MGMT_YELLOW='\033[1;33m'
KEY_MGMT_BLUE='\033[0;34m'
KEY_MGMT_NC='\033[0m'

# Default keyring backend
AURA_KEYRING_BACKEND="${AURA_KEYRING_BACKEND:-test}"

# Keyring backend descriptions
declare -A KEYRING_BACKEND_DOCS=(
    ["os"]="Use OS native keyring (Keychain on macOS, Windows Credential Vault, Secret Service on Linux)"
    ["file"]="Encrypted file-based keyring (requires passphrase)"
    ["test"]="Unencrypted file-based keyring (for testing only, NOT for production)"
    ["memory"]="In-memory keyring (session only, ideal for CI/CD)"
)

# ============================================================================
# Keyring Backend Management
# ============================================================================

# setup_keyring_backend configures the keyring backend for aurad operations
# Args:
#   $1 - backend name (os|file|test|memory)
# Returns:
#   0 on success, 1 on error
setup_keyring_backend() {
    local backend="$1"

    if [ -z "$backend" ]; then
        echo -e "${KEY_MGMT_RED}✗ Error: backend not specified${KEY_MGMT_NC}" >&2
        return 1
    fi

    # Validate backend
    case "$backend" in
        os|file|test|memory)
            export AURA_KEYRING_BACKEND="$backend"
            echo -e "${KEY_MGMT_GREEN}✓ Keyring backend set to: $backend${KEY_MGMT_NC}"
            echo -e "${KEY_MGMT_BLUE}ℹ ${KEYRING_BACKEND_DOCS[$backend]}${KEY_MGMT_NC}"
            return 0
            ;;
        *)
            echo -e "${KEY_MGMT_RED}✗ Error: invalid backend: $backend${KEY_MGMT_NC}" >&2
            echo -e "${KEY_MGMT_YELLOW}Valid backends: os, file, test, memory${KEY_MGMT_NC}" >&2
            return 1
            ;;
    esac
}

# show_keyring_backends displays all available keyring backends and their descriptions
show_keyring_backends() {
    echo -e "${KEY_MGMT_BLUE}=== Available Keyring Backends ===${KEY_MGMT_NC}"
    echo ""

    for backend in os file test memory; do
        local current=""
        if [ "$backend" = "$AURA_KEYRING_BACKEND" ]; then
            current=" ${KEY_MGMT_GREEN}(current)${KEY_MGMT_NC}"
        fi

        echo -e "${KEY_MGMT_BLUE}${backend}${KEY_MGMT_NC}${current}"
        echo -e "  ${KEYRING_BACKEND_DOCS[$backend]}"
        echo ""
    done

    echo -e "${KEY_MGMT_YELLOW}Security Recommendations:${KEY_MGMT_NC}"
    echo -e "  ${KEY_MGMT_GREEN}✓${KEY_MGMT_NC} Production/Mainnet: Use 'os' backend with HSM support"
    echo -e "  ${KEY_MGMT_GREEN}✓${KEY_MGMT_NC} Testnet: Use 'file' backend with strong passphrases"
    echo -e "  ${KEY_MGMT_YELLOW}⚠${KEY_MGMT_NC} Local dev: Use 'test' backend (unencrypted)"
    echo -e "  ${KEY_MGMT_GREEN}✓${KEY_MGMT_NC} CI/CD: Use 'memory' backend (ephemeral)"
    echo ""
}

# ============================================================================
# Key Generation
# ============================================================================

# generate_key_deterministic creates a deterministic key from a seed
# This is useful for reproducible testing environments
# Args:
#   $1 - key name
#   $2 - home directory
#   $3 - seed string (deterministic source)
#   $4 - keyring backend (optional, defaults to AURA_KEYRING_BACKEND)
# Returns:
#   0 on success, 1 on error
generate_key_deterministic() {
    local key_name="$1"
    local home_dir="$2"
    local seed="$3"
    local backend="${4:-$AURA_KEYRING_BACKEND}"

    if [ -z "$key_name" ] || [ -z "$home_dir" ] || [ -z "$seed" ]; then
        echo -e "${KEY_MGMT_RED}✗ Error: missing required arguments${KEY_MGMT_NC}" >&2
        echo -e "Usage: generate_key_deterministic <key-name> <home-dir> <seed> [backend]" >&2
        return 1
    fi

    # Generate deterministic mnemonic from seed
    # Using sha256 of seed to generate consistent 256-bit entropy
    local entropy=$(echo -n "$seed" | sha256sum | cut -d' ' -f1)

    # Generate 24-word mnemonic from entropy
    # Note: This is a simplified approach for testing. Production should use
    # proper BIP39 mnemonic generation from a CSPRNG.
    local mnemonic=$(generate_bip39_mnemonic "$entropy")

    if [ -z "$mnemonic" ]; then
        echo -e "${KEY_MGMT_RED}✗ Error: failed to generate mnemonic${KEY_MGMT_NC}" >&2
        return 1
    fi

    # Check if aurad binary exists
    local aurad="${home_dir}/../chain/aurad"
    if [ ! -f "$aurad" ]; then
        # Try alternative paths
        aurad="./aurad"
        if [ ! -f "$aurad" ]; then
            aurad="$(which aurad 2>/dev/null)"
            if [ -z "$aurad" ]; then
                echo -e "${KEY_MGMT_RED}✗ Error: aurad binary not found${KEY_MGMT_NC}" >&2
                return 1
            fi
        fi
    fi

    # Add key from mnemonic
    echo -e "${KEY_MGMT_BLUE}Generating deterministic key: $key_name${KEY_MGMT_NC}"

    echo "$mnemonic" | "$aurad" keys add "$key_name" \
        --recover \
        --keyring-backend "$backend" \
        --home "$home_dir" \
        --output json > /dev/null 2>&1

    if [ $? -ne 0 ]; then
        echo -e "${KEY_MGMT_RED}✗ Failed to generate key${KEY_MGMT_NC}" >&2
        return 1
    fi

    echo -e "${KEY_MGMT_GREEN}✓ Key generated: $key_name${KEY_MGMT_NC}"

    # Save mnemonic to file for reference (only for test/memory backends)
    if [ "$backend" = "test" ] || [ "$backend" = "memory" ]; then
        local mnemonic_file="${home_dir}/${key_name}.mnemonic"
        echo "$mnemonic" > "$mnemonic_file"
        chmod 600 "$mnemonic_file"
        echo -e "${KEY_MGMT_YELLOW}ℹ Mnemonic saved to: $mnemonic_file${KEY_MGMT_NC}"
    fi

    return 0
}

# generate_bip39_mnemonic generates a BIP39 mnemonic from hex entropy
# Args:
#   $1 - hex entropy (64 characters for 24 words)
# Returns:
#   Outputs mnemonic to stdout
generate_bip39_mnemonic() {
    local entropy="$1"

    # For simplicity in testing, we generate a pseudo-mnemonic
    # Production code should use proper BIP39 wordlist and checksum
    # This generates deterministic but not BIP39-compliant mnemonics

    # Split entropy into 24 chunks and map to simple words
    local words=()
    for i in {0..23}; do
        local offset=$((i * 2))
        local chunk="${entropy:$offset:2}"
        local word_index=$((16#$chunk % 2048))
        # Use simple deterministic words for testing
        words+=("word${word_index}")
    done

    echo "${words[*]}"
}

# generate_key_random creates a random key
# Args:
#   $1 - key name
#   $2 - home directory
#   $3 - keyring backend (optional, defaults to AURA_KEYRING_BACKEND)
# Returns:
#   0 on success, 1 on error
generate_key_random() {
    local key_name="$1"
    local home_dir="$2"
    local backend="${3:-$AURA_KEYRING_BACKEND}"

    if [ -z "$key_name" ] || [ -z "$home_dir" ]; then
        echo -e "${KEY_MGMT_RED}✗ Error: missing required arguments${KEY_MGMT_NC}" >&2
        return 1
    fi

    local aurad="${home_dir}/../chain/aurad"
    if [ ! -f "$aurad" ]; then
        aurad="./aurad"
        if [ ! -f "$aurad" ]; then
            aurad="$(which aurad 2>/dev/null)"
            if [ -z "$aurad" ]; then
                echo -e "${KEY_MGMT_RED}✗ Error: aurad binary not found${KEY_MGMT_NC}" >&2
                return 1
            fi
        fi
    fi

    echo -e "${KEY_MGMT_BLUE}Generating random key: $key_name${KEY_MGMT_NC}"

    "$aurad" keys add "$key_name" \
        --keyring-backend "$backend" \
        --home "$home_dir" \
        --output json > "${home_dir}/${key_name}.json" 2>&1

    if [ $? -ne 0 ]; then
        echo -e "${KEY_MGMT_RED}✗ Failed to generate key${KEY_MGMT_NC}" >&2
        return 1
    fi

    echo -e "${KEY_MGMT_GREEN}✓ Key generated: $key_name${KEY_MGMT_NC}"
    echo -e "${KEY_MGMT_YELLOW}ℹ Key info saved to: ${home_dir}/${key_name}.json${KEY_MGMT_NC}"

    return 0
}

# ============================================================================
# Key Validation
# ============================================================================

# validate_key_exists checks if a key exists in the keyring
# Args:
#   $1 - key name
#   $2 - home directory
#   $3 - keyring backend (optional, defaults to AURA_KEYRING_BACKEND)
# Returns:
#   0 if key exists, 1 otherwise
validate_key_exists() {
    local key_name="$1"
    local home_dir="$2"
    local backend="${3:-$AURA_KEYRING_BACKEND}"

    if [ -z "$key_name" ] || [ -z "$home_dir" ]; then
        echo -e "${KEY_MGMT_RED}✗ Error: missing required arguments${KEY_MGMT_NC}" >&2
        return 1
    fi

    local aurad="${home_dir}/../chain/aurad"
    if [ ! -f "$aurad" ]; then
        aurad="./aurad"
        if [ ! -f "$aurad" ]; then
            aurad="$(which aurad 2>/dev/null)"
            if [ -z "$aurad" ]; then
                echo -e "${KEY_MGMT_RED}✗ Error: aurad binary not found${KEY_MGMT_NC}" >&2
                return 1
            fi
        fi
    fi

    "$aurad" keys show "$key_name" \
        --keyring-backend "$backend" \
        --home "$home_dir" \
        --output json > /dev/null 2>&1

    return $?
}

# validate_keys_batch validates multiple keys exist
# Args:
#   $1 - home directory
#   $2 - keyring backend
#   $@ - key names (remaining arguments)
# Returns:
#   0 if all keys exist, 1 otherwise
validate_keys_batch() {
    local home_dir="$1"
    local backend="$2"
    shift 2
    local key_names=("$@")

    if [ -z "$home_dir" ] || [ ${#key_names[@]} -eq 0 ]; then
        echo -e "${KEY_MGMT_RED}✗ Error: missing required arguments${KEY_MGMT_NC}" >&2
        return 1
    fi

    echo -e "${KEY_MGMT_BLUE}=== Validating Keys ===${KEY_MGMT_NC}"
    echo -e "${KEY_MGMT_BLUE}Home: $home_dir${KEY_MGMT_NC}"
    echo -e "${KEY_MGMT_BLUE}Backend: $backend${KEY_MGMT_NC}"
    echo ""

    local all_valid=true
    for key_name in "${key_names[@]}"; do
        if validate_key_exists "$key_name" "$home_dir" "$backend"; then
            echo -e "${KEY_MGMT_GREEN}✓${KEY_MGMT_NC} Key exists: $key_name"
        else
            echo -e "${KEY_MGMT_RED}✗${KEY_MGMT_NC} Key missing: $key_name"
            all_valid=false
        fi
    done

    echo ""

    if [ "$all_valid" = true ]; then
        echo -e "${KEY_MGMT_GREEN}✓ All keys validated successfully${KEY_MGMT_NC}"
        return 0
    else
        echo -e "${KEY_MGMT_RED}✗ One or more keys missing${KEY_MGMT_NC}"
        echo -e "${KEY_MGMT_YELLOW}Remediation:${KEY_MGMT_NC}"
        echo -e "  1. Generate missing keys using generate_key_deterministic or generate_key_random"
        echo -e "  2. Or import keys from mnemonic/private key"
        return 1
    fi
}

# get_key_address retrieves the address for a key
# Args:
#   $1 - key name
#   $2 - home directory
#   $3 - keyring backend (optional, defaults to AURA_KEYRING_BACKEND)
# Returns:
#   Outputs address to stdout, returns 1 on error
get_key_address() {
    local key_name="$1"
    local home_dir="$2"
    local backend="${3:-$AURA_KEYRING_BACKEND}"

    if [ -z "$key_name" ] || [ -z "$home_dir" ]; then
        echo "Error: missing required arguments" >&2
        return 1
    fi

    local aurad="${home_dir}/../chain/aurad"
    if [ ! -f "$aurad" ]; then
        aurad="./aurad"
        if [ ! -f "$aurad" ]; then
            aurad="$(which aurad 2>/dev/null)"
            if [ -z "$aurad" ]; then
                echo "Error: aurad binary not found" >&2
                return 1
            fi
        fi
    fi

    local address=$("$aurad" keys show "$key_name" \
        --keyring-backend "$backend" \
        --home "$home_dir" \
        --address 2>/dev/null)

    if [ -z "$address" ]; then
        echo "Error: failed to get address for key: $key_name" >&2
        return 1
    fi

    echo "$address"
    return 0
}

# ============================================================================
# Environment Variable Management
# ============================================================================

# load_secrets_from_env loads secrets from environment variables
# This function demonstrates the pattern for loading secrets securely
# Environment variables expected:
#   AURA_VALIDATOR_MNEMONIC - Validator key mnemonic
#   AURA_OPERATOR_MNEMONIC - Operator key mnemonic
#   AURA_KEYRING_BACKEND - Keyring backend to use
# Returns:
#   0 on success, 1 on error
load_secrets_from_env() {
    echo -e "${KEY_MGMT_BLUE}=== Loading Secrets from Environment ===${KEY_MGMT_NC}"

    # Check for keyring backend
    if [ -n "$AURA_KEYRING_BACKEND" ]; then
        setup_keyring_backend "$AURA_KEYRING_BACKEND"
    else
        echo -e "${KEY_MGMT_YELLOW}⚠ AURA_KEYRING_BACKEND not set, using default: test${KEY_MGMT_NC}"
        export AURA_KEYRING_BACKEND="test"
    fi

    # Validate required secrets exist
    local secrets_missing=false

    if [ -z "$AURA_VALIDATOR_MNEMONIC" ]; then
        echo -e "${KEY_MGMT_YELLOW}⚠ AURA_VALIDATOR_MNEMONIC not set${KEY_MGMT_NC}"
        secrets_missing=true
    else
        echo -e "${KEY_MGMT_GREEN}✓ AURA_VALIDATOR_MNEMONIC loaded${KEY_MGMT_NC}"
    fi

    if [ -z "$AURA_OPERATOR_MNEMONIC" ]; then
        echo -e "${KEY_MGMT_YELLOW}⚠ AURA_OPERATOR_MNEMONIC not set${KEY_MGMT_NC}"
        secrets_missing=true
    else
        echo -e "${KEY_MGMT_GREEN}✓ AURA_OPERATOR_MNEMONIC loaded${KEY_MGMT_NC}"
    fi

    if [ "$secrets_missing" = true ]; then
        echo ""
        echo -e "${KEY_MGMT_YELLOW}To set secrets:${KEY_MGMT_NC}"
        echo -e "  export AURA_VALIDATOR_MNEMONIC='your 24 word mnemonic here'"
        echo -e "  export AURA_OPERATOR_MNEMONIC='your 24 word mnemonic here'"
        echo -e "  export AURA_KEYRING_BACKEND='test'"
        echo ""
        return 1
    fi

    echo ""
    echo -e "${KEY_MGMT_GREEN}✓ All secrets loaded successfully${KEY_MGMT_NC}"
    return 0
}

# ============================================================================
# Security Utilities
# ============================================================================

# secure_delete securely deletes a file containing sensitive data
# Args:
#   $1 - file path
# Returns:
#   0 on success, 1 on error
secure_delete() {
    local file_path="$1"

    if [ -z "$file_path" ]; then
        echo -e "${KEY_MGMT_RED}✗ Error: file path not specified${KEY_MGMT_NC}" >&2
        return 1
    fi

    if [ ! -f "$file_path" ]; then
        echo -e "${KEY_MGMT_YELLOW}⚠ File does not exist: $file_path${KEY_MGMT_NC}" >&2
        return 0
    fi

    # Overwrite file with random data before deletion
    if command -v shred &> /dev/null; then
        shred -vfz -n 3 "$file_path" 2>/dev/null
    else
        # Fallback: overwrite with zeros
        dd if=/dev/zero of="$file_path" bs=1k count=$(stat -f%z "$file_path" 2>/dev/null || stat -c%s "$file_path") conv=notrunc 2>/dev/null
        rm -f "$file_path"
    fi

    echo -e "${KEY_MGMT_GREEN}✓ Securely deleted: $file_path${KEY_MGMT_NC}"
    return 0
}

# Export functions for use in other scripts
export -f setup_keyring_backend
export -f show_keyring_backends
export -f generate_key_deterministic
export -f generate_key_random
export -f validate_key_exists
export -f validate_keys_batch
export -f get_key_address
export -f load_secrets_from_env
export -f secure_delete
