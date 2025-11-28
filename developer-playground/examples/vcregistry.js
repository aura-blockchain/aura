// VCRegistry Module Examples - All 4 Languages

export const vcregistryExamples = {
    // ============ QUERY EXAMPLES ============

    'vcregistry-query-vc-js': {
        title: 'Query Verifiable Credential',
        description: 'Query a specific verifiable credential by ID',
        category: 'vcregistry',
        language: 'javascript',
        code: `// Query a specific Verifiable Credential
const vcId = '1'; // Replace with actual VC ID

try {
    // Get VC details
    const vc = await api.getVC(vcId);
    console.log('Verifiable Credential:', vc);

    // Display VC information
    console.log('VC ID:', vc.id);
    console.log('Issuer:', vc.issuer);
    console.log('Subject:', vc.subject);
    console.log('Type:', vc.type);
    console.log('Issue Date:', vc.issue_date);
    console.log('Expiration Date:', vc.expiration_date);
    console.log('Status:', vc.status);

    return vc;
} catch (error) {
    console.error('Error querying VC:', error);
    throw error;
}`
    },

    'vcregistry-query-vc-python': {
        title: 'Query Verifiable Credential (Python)',
        description: 'Query a specific verifiable credential by ID',
        category: 'vcregistry',
        language: 'python',
        code: `# Query a specific Verifiable Credential
import requests
import json

# API endpoint
api_url = 'https://api.aura.zone'
vc_id = '1'  # Replace with actual VC ID

try:
    # Query VC
    response = requests.get(
        f'{api_url}/aura/vcregistry/v1beta1/vc/{vc_id}',
        headers={'accept': 'application/json'}
    )
    response.raise_for_status()

    vc = response.json()
    print(f'Verifiable Credential: {json.dumps(vc, indent=2)}')

    # Display VC information
    print(f"VC ID: {vc['id']}")
    print(f"Issuer: {vc['issuer']}")
    print(f"Subject: {vc['subject']}")
    print(f"Type: {vc['type']}")
    print(f"Status: {vc['status']}")

except requests.exceptions.RequestException as e:
    print(f'Error querying VC: {e}')`
    },

    'vcregistry-query-vc-go': {
        title: 'Query Verifiable Credential (Go)',
        description: 'Query a specific verifiable credential by ID',
        category: 'vcregistry',
        language: 'go',
        code: `// Query a specific Verifiable Credential
package main

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"

    vctypes "github.com/aura/x/vcregistry/types"
)

func queryVC(vcID string) error {
    // API endpoint
    apiURL := "https://api.aura.zone"
    endpoint := fmt.Sprintf("%s/aura/vcregistry/v1beta1/vc/%s", apiURL, vcID)

    // Make request
    resp, err := http.Get(endpoint)
    if err != nil {
        return fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    // Read response
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return fmt.Errorf("failed to read response: %w", err)
    }

    // Parse VC
    var vc vctypes.VerifiableCredential
    if err := json.Unmarshal(body, &vc); err != nil {
        return fmt.Errorf("failed to parse VC: %w", err)
    }

    // Display VC information
    fmt.Printf("VC ID: %s\\n", vc.Id)
    fmt.Printf("Issuer: %s\\n", vc.Issuer)
    fmt.Printf("Subject: %s\\n", vc.Subject)
    fmt.Printf("Type: %s\\n", vc.Type)
    fmt.Printf("Status: %s\\n", vc.Status)

    return nil
}

func main() {
    vcID := "1" // Replace with actual VC ID
    if err := queryVC(vcID); err != nil {
        fmt.Printf("Error: %v\\n", err)
    }
}`
    },

    'vcregistry-query-vc-curl': {
        title: 'Query Verifiable Credential (cURL)',
        description: 'Query a specific verifiable credential by ID',
        category: 'vcregistry',
        language: 'shell',
        code: `# Query a specific Verifiable Credential
VC_ID="1"  # Replace with actual VC ID

curl -X GET "https://api.aura.zone/aura/vcregistry/v1beta1/vc/$VC_ID" \\
  -H "accept: application/json" \\
  | jq .

# Query all VCs for an address
ADDRESS="aura1..."  # Replace with actual address

curl -X GET "https://api.aura.zone/aura/vcregistry/v1beta1/vcs/$ADDRESS" \\
  -H "accept: application/json" \\
  | jq .`
    },

    // ============ MINT VC EXAMPLES ============

    'vcregistry-mint-vc-js': {
        title: 'Mint Verifiable Credential',
        description: 'Create and mint a new verifiable credential',
        category: 'vcregistry',
        language: 'javascript',
        code: `// Mint a new Verifiable Credential
if (!wallet.connected) {
    console.error('Please connect your wallet first');
    return;
}

// VC data
const vcData = {
    type: 'IdentityCredential',
    subject: 'aura1...', // Subject address
    claims: {
        name: 'John Doe',
        email: 'john@aura.network',
        verified: true
    },
    expirationDate: new Date(Date.now() + 365 * 24 * 60 * 60 * 1000).toISOString()
};

// Create MsgMintVC message
const msg = {
    type: 'aura/vcregistry/MsgMintVC',
    value: {
        issuer: wallet.address,
        subject: vcData.subject,
        vc_type: vcData.type,
        claims: JSON.stringify(vcData.claims),
        expiration_date: vcData.expirationDate
    }
};

console.log('Mint VC transaction:', msg);
console.log('Note: This will create a new verifiable credential on-chain');

return { transaction: msg, vcData };`
    },

    'vcregistry-mint-vc-python': {
        title: 'Mint Verifiable Credential (Python)',
        description: 'Create and mint a new verifiable credential',
        category: 'vcregistry',
        language: 'python',
        code: `# Mint a new Verifiable Credential
import json
from datetime import datetime, timedelta

# VC data
issuer = 'aura1...'  # Replace with issuer address
subject = 'aura1...'  # Replace with subject address

vc_data = {
    'type': 'IdentityCredential',
    'claims': {
        'name': 'John Doe',
        'email': 'john@aura.network',
        'verified': True
    }
}

# Calculate expiration (1 year from now)
expiration = (datetime.now() + timedelta(days=365)).isoformat()

# Build transaction message
msg = {
    '@type': '/aura.vcregistry.v1beta1.MsgMintVC',
    'issuer': issuer,
    'subject': subject,
    'vc_type': vc_data['type'],
    'claims': json.dumps(vc_data['claims']),
    'expiration_date': expiration
}

print(f'Mint VC Message: {json.dumps(msg, indent=2)}')
print('Note: Sign and broadcast this transaction using your wallet')`
    },

    'vcregistry-mint-vc-go': {
        title: 'Mint Verifiable Credential (Go)',
        description: 'Create and mint a new verifiable credential',
        category: 'vcregistry',
        language: 'go',
        code: `// Mint a new Verifiable Credential
package main

import (
    "encoding/json"
    "fmt"
    "time"

    sdk "github.com/cosmos/cosmos-sdk/types"
    vctypes "github.com/aura/x/vcregistry/types"
)

func mintVC(issuerAddr, subjectAddr string) (*vctypes.MsgMintVC, error) {
    // Parse addresses
    issuer, err := sdk.AccAddressFromBech32(issuerAddr)
    if err != nil {
        return nil, fmt.Errorf("invalid issuer address: %w", err)
    }

    subject, err := sdk.AccAddressFromBech32(subjectAddr)
    if err != nil {
        return nil, fmt.Errorf("invalid subject address: %w", err)
    }

    // Prepare claims
    claims := map[string]interface{}{
        "name":     "John Doe",
        "email":    "john@aura.network",
        "verified": true,
    }
    claimsJSON, _ := json.Marshal(claims)

    // Calculate expiration (1 year from now)
    expiration := time.Now().AddDate(1, 0, 0)

    // Create MsgMintVC
    msg := &vctypes.MsgMintVC{
        Issuer:         issuer.String(),
        Subject:        subject.String(),
        VcType:         "IdentityCredential",
        Claims:         string(claimsJSON),
        ExpirationDate: expiration.Unix(),
    }

    fmt.Printf("Mint VC Message: %+v\\n", msg)
    return msg, nil
}

func main() {
    issuer := "aura1..."  // Replace with issuer address
    subject := "aura1..." // Replace with subject address

    msg, err := mintVC(issuer, subject)
    if err != nil {
        fmt.Printf("Error: %v\\n", err)
        return
    }

    fmt.Println("Transaction ready to sign and broadcast")
}`
    },

    'vcregistry-mint-vc-curl': {
        title: 'Mint Verifiable Credential (cURL)',
        description: 'Create and mint a new verifiable credential',
        category: 'vcregistry',
        language: 'shell',
        code: `# Mint a new Verifiable Credential
# Note: This requires signing with aurad CLI or Keplr

# Prepare VC data
ISSUER="aura1..."  # Replace with issuer address
SUBJECT="aura1..." # Replace with subject address
VC_TYPE="IdentityCredential"
CLAIMS='{"name":"John Doe","email":"john@aura.network","verified":true}'
EXPIRATION_DATE=$(date -d "+1 year" +%s)

# Build transaction (using aurad CLI)
aurad tx vcregistry mint-vc \\
  $SUBJECT \\
  $VC_TYPE \\
  "$CLAIMS" \\
  $EXPIRATION_DATE \\
  --from $ISSUER \\
  --chain-id aura-1 \\
  --gas auto \\
  --gas-adjustment 1.5 \\
  --fees 5000uaura \\
  -y

# Or broadcast a pre-signed transaction via REST API
curl -X POST "https://api.aura.zone/cosmos/tx/v1beta1/txs" \\
  -H "Content-Type: application/json" \\
  -d '{
    "tx_bytes": "base64_encoded_signed_tx",
    "mode": "BROADCAST_MODE_SYNC"
  }'`
    },

    // ============ PRESENTATION EXAMPLES ============

    'vcregistry-create-presentation-js': {
        title: 'Create VC Presentation',
        description: 'Create a verifiable presentation from VCs',
        category: 'vcregistry',
        language: 'javascript',
        code: `// Create a Verifiable Presentation
if (!wallet.connected) {
    console.error('Please connect your wallet first');
    return;
}

// Get user's VCs
const vcs = await api.getVCs(wallet.address);
console.log('Available VCs:', vcs);

if (!vcs.vcs || vcs.vcs.length === 0) {
    console.error('No VCs available to create presentation');
    return;
}

// Select VCs for presentation
const vcIds = vcs.vcs.slice(0, 2).map(vc => vc.id);

// Create presentation message
const msg = {
    type: 'aura/vcregistry/MsgCreatePresentation',
    value: {
        holder: wallet.address,
        vc_ids: vcIds,
        purpose: 'Identity Verification',
        verifier: 'aura1...' // Verifier address
    }
};

console.log('Create Presentation transaction:', msg);
console.log('Selected VCs:', vcIds);

return { transaction: msg, vcIds };`
    },

    'vcregistry-create-presentation-python': {
        title: 'Create VC Presentation (Python)',
        description: 'Create a verifiable presentation from VCs',
        category: 'vcregistry',
        language: 'python',
        code: `# Create a Verifiable Presentation
import requests
import json

# Configuration
api_url = 'https://api.aura.zone'
holder = 'aura1...'  # Replace with holder address

# Query holder's VCs
response = requests.get(
    f'{api_url}/aura/vcregistry/v1beta1/vcs/{holder}',
    headers={'accept': 'application/json'}
)
vcs = response.json()

print(f'Available VCs: {len(vcs.get("vcs", []))}')

# Select VCs for presentation
vc_ids = [vc['id'] for vc in vcs.get('vcs', [])[:2]]

if not vc_ids:
    print('No VCs available to create presentation')
else:
    # Build presentation message
    msg = {
        '@type': '/aura.vcregistry.v1beta1.MsgCreatePresentation',
        'holder': holder,
        'vc_ids': vc_ids,
        'purpose': 'Identity Verification',
        'verifier': 'aura1...'  # Replace with verifier address
    }

    print(f'Create Presentation Message: {json.dumps(msg, indent=2)}')
    print(f'Selected VCs: {vc_ids}')`
    },

    'vcregistry-create-presentation-go': {
        title: 'Create VC Presentation (Go)',
        description: 'Create a verifiable presentation from VCs',
        category: 'vcregistry',
        language: 'go',
        code: `// Create a Verifiable Presentation
package main

import (
    "fmt"

    sdk "github.com/cosmos/cosmos-sdk/types"
    vctypes "github.com/aura/x/vcregistry/types"
)

func createPresentation(holderAddr string, vcIDs []string) (*vctypes.MsgCreatePresentation, error) {
    // Parse holder address
    holder, err := sdk.AccAddressFromBech32(holderAddr)
    if err != nil {
        return nil, fmt.Errorf("invalid holder address: %w", err)
    }

    // Verifier address
    verifier := "aura1..." // Replace with verifier address

    // Create MsgCreatePresentation
    msg := &vctypes.MsgCreatePresentation{
        Holder:   holder.String(),
        VcIds:    vcIDs,
        Purpose:  "Identity Verification",
        Verifier: verifier,
    }

    fmt.Printf("Create Presentation Message: %+v\\n", msg)
    fmt.Printf("Selected VCs: %v\\n", vcIDs)

    return msg, nil
}

func main() {
    holder := "aura1..." // Replace with holder address
    vcIDs := []string{"1", "2"} // Replace with actual VC IDs

    msg, err := createPresentation(holder, vcIDs)
    if err != nil {
        fmt.Printf("Error: %v\\n", err)
        return
    }

    fmt.Println("Presentation ready to create")
}`
    },

    'vcregistry-create-presentation-curl': {
        title: 'Create VC Presentation (cURL)',
        description: 'Create a verifiable presentation from VCs',
        category: 'vcregistry',
        language: 'shell',
        code: `# Create a Verifiable Presentation
# First, query available VCs
HOLDER="aura1..."  # Replace with holder address

curl -X GET "https://api.aura.zone/aura/vcregistry/v1beta1/vcs/$HOLDER" \\
  -H "accept: application/json" \\
  | jq '.vcs[].id'

# Create presentation (using aurad CLI)
VC_ID_1="1"
VC_ID_2="2"
VERIFIER="aura1..."  # Replace with verifier address

aurad tx vcregistry create-presentation \\
  $VC_ID_1,$VC_ID_2 \\
  "Identity Verification" \\
  $VERIFIER \\
  --from $HOLDER \\
  --chain-id aura-1 \\
  --gas auto \\
  --gas-adjustment 1.5 \\
  --fees 5000uaura \\
  -y`
    },

    // ============ REVOKE VC EXAMPLES ============

    'vcregistry-revoke-vc-js': {
        title: 'Revoke Verifiable Credential',
        description: 'Revoke an existing verifiable credential',
        category: 'vcregistry',
        language: 'javascript',
        code: `// Revoke a Verifiable Credential
if (!wallet.connected) {
    console.error('Please connect your wallet first');
    return;
}

const vcId = '1'; // Replace with actual VC ID
const reason = 'Credential no longer valid';

// Create MsgRevokeVC message
const msg = {
    type: 'aura/vcregistry/MsgRevokeVC',
    value: {
        issuer: wallet.address,
        vc_id: vcId,
        reason: reason
    }
};

console.log('Revoke VC transaction:', msg);
console.log('Note: Only the issuer can revoke a VC');
console.warn('This action is irreversible!');

return { transaction: msg };`
    },

    'vcregistry-revoke-vc-python': {
        title: 'Revoke Verifiable Credential (Python)',
        description: 'Revoke an existing verifiable credential',
        category: 'vcregistry',
        language: 'python',
        code: `# Revoke a Verifiable Credential
import json

# Configuration
issuer = 'aura1...'  # Replace with issuer address (must be VC issuer)
vc_id = '1'  # Replace with actual VC ID
reason = 'Credential no longer valid'

# Build revoke message
msg = {
    '@type': '/aura.vcregistry.v1beta1.MsgRevokeVC',
    'issuer': issuer,
    'vc_id': vc_id,
    'reason': reason
}

print(f'Revoke VC Message: {json.dumps(msg, indent=2)}')
print('WARNING: This action is irreversible!')
print('Note: Only the issuer can revoke a VC')`
    },

    'vcregistry-revoke-vc-go': {
        title: 'Revoke Verifiable Credential (Go)',
        description: 'Revoke an existing verifiable credential',
        category: 'vcregistry',
        language: 'go',
        code: `// Revoke a Verifiable Credential
package main

import (
    "fmt"

    sdk "github.com/cosmos/cosmos-sdk/types"
    vctypes "github.com/aura/x/vcregistry/types"
)

func revokeVC(issuerAddr, vcID string) (*vctypes.MsgRevokeVC, error) {
    // Parse issuer address
    issuer, err := sdk.AccAddressFromBech32(issuerAddr)
    if err != nil {
        return nil, fmt.Errorf("invalid issuer address: %w", err)
    }

    // Create MsgRevokeVC
    msg := &vctypes.MsgRevokeVC{
        Issuer: issuer.String(),
        VcId:   vcID,
        Reason: "Credential no longer valid",
    }

    fmt.Printf("Revoke VC Message: %+v\\n", msg)
    fmt.Println("WARNING: This action is irreversible!")
    fmt.Println("Note: Only the issuer can revoke a VC")

    return msg, nil
}

func main() {
    issuer := "aura1..." // Replace with issuer address
    vcID := "1"          // Replace with actual VC ID

    msg, err := revokeVC(issuer, vcID)
    if err != nil {
        fmt.Printf("Error: %v\\n", err)
        return
    }

    fmt.Println("Revoke transaction ready to sign and broadcast")
}`
    },

    'vcregistry-revoke-vc-curl': {
        title: 'Revoke Verifiable Credential (cURL)',
        description: 'Revoke an existing verifiable credential',
        category: 'vcregistry',
        language: 'shell',
        code: `# Revoke a Verifiable Credential
ISSUER="aura1..."  # Replace with issuer address (must be VC issuer)
VC_ID="1"          # Replace with actual VC ID
REASON="Credential no longer valid"

# Revoke VC (using aurad CLI)
aurad tx vcregistry revoke-vc \\
  $VC_ID \\
  "$REASON" \\
  --from $ISSUER \\
  --chain-id aura-1 \\
  --gas auto \\
  --gas-adjustment 1.5 \\
  --fees 5000uaura \\
  -y

echo "WARNING: This action is irreversible!"
echo "Note: Only the issuer can revoke a VC"`
    },

    // ============ STATS EXAMPLES ============

    'vcregistry-stats-js': {
        title: 'Query VC Registry Statistics',
        description: 'Get statistics about the VC registry',
        category: 'vcregistry',
        language: 'javascript',
        code: `// Query VC Registry Statistics
try {
    // Get registry stats
    const stats = await api.getVCStats();
    console.log('VC Registry Statistics:', stats);

    // Display stats
    console.log('Total VCs:', stats.total_vcs);
    console.log('Active VCs:', stats.active_vcs);
    console.log('Revoked VCs:', stats.revoked_vcs);
    console.log('Expired VCs:', stats.expired_vcs);
    console.log('Total Issuers:', stats.total_issuers);
    console.log('Total Presentations:', stats.total_presentations);

    // Get registry params
    const params = await api.getVCRegistryParams();
    console.log('Registry Parameters:', params);

    return { stats, params };
} catch (error) {
    console.error('Error querying registry stats:', error);
    throw error;
}`
    }
};
