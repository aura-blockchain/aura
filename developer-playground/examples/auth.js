// Auth Module Examples - All 4 Languages

export const authExamples = {
    'auth-query-account-js': {
        title: 'Query Account',
        description: 'Query account information and authentication details',
        category: 'auth',
        language: 'javascript',
        code: `// Query account information
const address = 'aura1...'; // Replace with actual address

try {
    // Get account details
    const account = await api.getAccount(address);
    console.log('Account:', account);

    console.log('Address:', account.address);
    console.log('Account Number:', account.account_number);
    console.log('Sequence:', account.sequence);
    console.log('Public Key:', account.pub_key);

    // Get auth params
    const params = await api.getAuthParams();
    console.log('Auth Parameters:', params);

    return { account, params };
} catch (error) {
    console.error('Error querying account:', error);
    throw error;
}`
    },

    'auth-query-account-python': {
        title: 'Query Account (Python)',
        description: 'Query account information and authentication details',
        category: 'auth',
        language: 'python',
        code: `# Query account information
import requests
import json

api_url = 'https://api.aura.zone'
address = 'aura1...'  # Replace with actual address

try:
    # Query account
    response = requests.get(
        f'{api_url}/cosmos/auth/v1beta1/accounts/{address}',
        headers={'accept': 'application/json'}
    )
    response.raise_for_status()

    account = response.json()['account']
    print(f'Account: {json.dumps(account, indent=2)}')

    # Query auth params
    params_response = requests.get(
        f'{api_url}/cosmos/auth/v1beta1/params'
    )
    params = params_response.json()['params']
    print(f'Auth Parameters: {json.dumps(params, indent=2)}')

except requests.exceptions.RequestException as e:
    print(f'Error: {e}')`
    },

    'auth-query-account-go': {
        title: 'Query Account (Go)',
        description: 'Query account information and authentication details',
        category: 'auth',
        language: 'go',
        code: `// Query account information
package main

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"

    authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func queryAccount(address string) error {
    apiURL := "https://api.aura.zone"
    endpoint := fmt.Sprintf("%s/cosmos/auth/v1beta1/accounts/%s", apiURL, address)

    resp, err := http.Get(endpoint)
    if err != nil {
        return fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)

    var result struct {
        Account json.RawMessage \`json:"account"\`
    }
    json.Unmarshal(body, &result)

    fmt.Printf("Account: %s\\n", string(result.Account))

    return nil
}

func main() {
    address := "aura1..." // Replace with actual address
    if err := queryAccount(address); err != nil {
        fmt.Printf("Error: %v\\n", err)
    }
}`
    },

    'auth-query-account-curl': {
        title: 'Query Account (cURL)',
        description: 'Query account information and authentication details',
        category: 'auth',
        language: 'shell',
        code: `# Query account information
ADDRESS="aura1..."  # Replace with actual address

# Query account
curl -X GET "https://api.aura.zone/cosmos/auth/v1beta1/accounts/$ADDRESS" \\
  -H "accept: application/json" \\
  | jq .

# Query auth params
curl -X GET "https://api.aura.zone/cosmos/auth/v1beta1/params" \\
  -H "accept: application/json" \\
  | jq .`
    }
};
