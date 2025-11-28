# AURA Network Dashboards

This directory contains web-based dashboards for monitoring and interacting with the AURA blockchain network. These dashboards were adapted from the PAW blockchain implementation and updated for AURA's specific features and modules.

## Available Dashboards

### 1. Validator Dashboard
**Location:** `./validator/`

A comprehensive monitoring dashboard for AURA validators, providing real-time insights into validator performance, rewards, delegations, and network health.

**Features:**
- Real-time validator metrics (uptime, voting power, commission)
- Delegation tracking and management
- Reward distribution monitoring
- Performance analytics and historical charts
- Slash event tracking
- Alert system for critical events

**Quick Start:**
```bash
cd validator
# Open index.html in a web browser
```

### 2. Staking Dashboard
**Location:** `./staking/`

A user-friendly interface for delegators to manage their staked AURA tokens, track rewards, and compare validators.

**Features:**
- Validator comparison and rankings
- Delegation portfolio view
- Real-time rewards tracking
- Staking calculator (APY estimates)
- Unbonding period tracker
- Network staking statistics

**Quick Start:**
```bash
cd staking
# Open index.html in a web browser
```

### 3. Governance Dashboard
**Location:** `./governance/`

A complete governance portal for AURA community members to participate in on-chain decision making.

**Features:**
- View all proposals (active, passed, rejected)
- Detailed proposal information with vote breakdown
- Create new proposals (text, parameter change, software upgrade)
- Vote on active proposals
- Deposit to proposals
- Governance parameter tracking
- Voting history

**Quick Start:**
```bash
cd governance
# Open index.html in a web browser
```

### 4. DEX Telemetry Dashboard
**Location:** `./dashboards/dex/`

Grafana-ready panels for monitoring Aura DEX swaps, market prices, and query validation errors.

**Features:**
- Per-pool swap effective price trends (`dex_swap_effective_price`)
- Market price sample sizes per coin (`dex_market_price_sample_size`)
- Spot-price & user-order validation error counters (`dex_query_validation_failed_total`)

**Quick Start:**
```bash
cd dashboards/dex
# Import dex-dashboard.json into Grafana (Prometheus datasource required)
```

### 5. AI Assistant Ops Dashboard
**Location:** `./dashboards/aiassistant/assistant-dashboard.json`

High-level visibility for the new AI assistant module and voucher lifecycle.

**Features:**
- Heartbeat success counters + average heartbeat age per assistant.
- Voucher issuance vs redemption (fed by the `aura-voucher` CLI via Pushgateway).
- Misbehavior report table so SOC teams can escalate jailed/tombstoned assistants quickly.

**Quick Start:**
```bash
cd dashboards/aiassistant
# Import assistant-dashboard.json into Grafana (Prometheus datasource required)
```

## Configuration

All dashboards share a common configuration file that centralizes network settings and API endpoints.

**Configuration File:** `config.js`

### Key Configuration Options

```javascript
// Network endpoints
AuraConfig.endpoints.rest = 'http://localhost:1317';  // REST API
AuraConfig.endpoints.rpc = 'http://localhost:26657';   // RPC endpoint

// Enable/disable mock mode for development
AuraConfig.dashboard.mockMode = false;

// Network parameters
AuraConfig.network.chainId = 'aura-1';
AuraConfig.network.denom = 'aura';
```

### Environment Variables

You can override configuration using environment variables:

```bash
export AURA_REST_ENDPOINT=https://api.aura.network:1317
export AURA_RPC_ENDPOINT=https://rpc.aura.network:26657
export AURA_MOCK_MODE=false
```

## API Endpoints

The dashboards use standard Cosmos SDK REST API endpoints plus AURA-specific custom module endpoints:

### Standard Cosmos SDK Endpoints
- **Staking:** `/cosmos/staking/v1beta1/*`
- **Governance:** `/cosmos/gov/v1beta1/*`
- **Distribution:** `/cosmos/distribution/v1beta1/*`
- **Slashing:** `/cosmos/slashing/v1beta1/*`
- **Bank:** `/cosmos/bank/v1beta1/*`

### AURA Custom Module Endpoints
- **Validator Security:** `/aura/validatorsecurity/v1beta1/*`
- **DEX:** `/aura/dex/v1beta1/*`
- **Bridge:** `/aura/bridge/v1beta1/*`
- **Network Security:** `/aura/networksecurity/v1beta1/*`
- **Governance Extensions:** `/aura/governance/v1beta1/*`

See `config.js` for the complete list of available endpoints.

## Development

### Mock Mode

All dashboards support mock mode for development and testing without a live blockchain connection:

```javascript
// In your dashboard's API service
AuraConfig.dashboard.mockMode = true;
```

When enabled, dashboards will use realistic mock data instead of making actual API calls.

### Testing

Each dashboard includes its own test suite:

- **Validator:** `./validator/tests/`
- **Staking:** `./staking/tests/`
- **Governance:** `./governance/tests/`

Run tests by opening the test runner HTML files in your browser or using the provided test scripts.

## Deployment

### Option 1: Static Hosting

All dashboards are static HTML/JS applications and can be hosted on any web server:

```bash
# Using Python's built-in server
cd dashboards
python -m http.server 8080

# Using Node.js http-server
npm install -g http-server
http-server -p 8080
```

Then navigate to:
- http://localhost:8080/validator
- http://localhost:8080/staking
- http://localhost:8080/governance

### Option 2: Docker

Each dashboard includes Docker configuration (see individual dashboard directories).

### Option 3: Integration with AURA Node

For production deployment, configure your AURA node to serve the REST API and point the dashboards to your node's endpoint:

1. Enable REST API in your AURA node configuration
2. Update `config.js` with your node's endpoints
3. Deploy dashboards to your web server
4. Configure CORS if necessary

## Chain-Specific Updates

The following AURA-specific parameters have been configured:

### Network Identity
- **Chain ID:** `aura-1`
- **Denomination:** `aura` (micro-aura for precision)
- **Decimals:** 6
- **Bech32 Prefixes:**
  - Account: `aura`
  - Validator: `auravaloper`
  - Consensus: `auravalcons`

### Governance
- **Minimum Deposit:** 10,000 AURA
- **Voting Period:** 14 days
- **Deposit Period:** 14 days
- **Quorum:** 33.4%
- **Threshold:** 50%
- **Veto Threshold:** 33.4%

### Staking
- **Unbonding Period:** 21 days
- **Max Validators:** 100
- **Bond Denomination:** aura

### Slashing
- **Signed Blocks Window:** 10,000 blocks
- **Min Signed Per Window:** 50%
- **Downtime Jail Duration:** 600 seconds
- **Slash Fraction (Double Sign):** 5%
- **Slash Fraction (Downtime):** 0.01%

## Browser Compatibility

All dashboards are tested and compatible with:
- Chrome/Chromium (recommended)
- Firefox
- Safari
- Edge

## Security Considerations

### Production Deployment
1. Always use HTTPS in production
2. Configure proper CORS headers on your REST API
3. Never expose private keys through the web interface
4. Consider implementing rate limiting on API calls
5. Use environment variables for sensitive configuration

### Transaction Signing
The current dashboards use mock mode for transaction signing. For production use, you should integrate with:
- Keplr Wallet
- CosmJS for programmatic signing
- Ledger hardware wallet support

## Migration Notes from PAW

The following changes were made during migration:

1. **Branding:** All PAW references replaced with AURA
2. **Denomination:** Changed from 'paw' to 'aura'
3. **Address Prefixes:** Updated to use 'aura' prefix family
4. **API Endpoints:** Configured for AURA's module structure
5. **Parameters:** Updated to match AURA's chain configuration

## Troubleshooting

### Dashboard won't connect to blockchain

1. Check that your AURA node is running
2. Verify REST API is enabled in node config
3. Confirm endpoint URLs in `config.js`
4. Check browser console for CORS errors
5. Try enabling mock mode for testing

### Data not refreshing

1. Check browser console for API errors
2. Verify cache settings in `config.js`
3. Clear browser cache and reload
4. Check network tab for failed requests

### Proposals/Validators not showing

1. Ensure your node is fully synced
2. Check that governance/staking modules are enabled
3. Verify API endpoints are correct
4. Try mock mode to test UI functionality

## Support

For issues specific to the dashboards:
- Check individual dashboard README files
- Review browser console errors
- Enable mock mode to isolate API issues

For AURA blockchain issues:
- Consult AURA documentation
- Check node logs
- Verify module configurations

## Contributing

When contributing to the dashboards:

1. Test in mock mode first
2. Test with live AURA node
3. Maintain compatibility with standard Cosmos SDK endpoints
4. Update configuration for new AURA modules
5. Add tests for new features
6. Update relevant documentation

## License

See LICENSE file in the root directory.

## Directory Structure

```
dashboards/
├── config.js                 # Shared configuration
├── README.md                 # This file
├── validator/                # Validator monitoring dashboard
│   ├── index.html
│   ├── app.js
│   ├── components/
│   ├── services/
│   │   └── validatorAPI.js
│   ├── assets/
│   └── tests/
├── staking/                  # Delegator staking dashboard
│   ├── index.html
│   ├── app.js
│   ├── components/
│   ├── services/
│   │   └── stakingAPI.js
│   ├── styles/
│   └── tests/
└── governance/               # Governance portal
    ├── index.html
    ├── app.js
    ├── components/
    ├── services/
    │   └── governanceAPI.js
    ├── assets/
    └── tests/
```

## Next Steps

1. **Production Configuration:**
   - Set up production REST/RPC endpoints
   - Configure HTTPS and CORS
   - Set up monitoring and logging

2. **Wallet Integration:**
   - Integrate Keplr wallet
   - Implement transaction signing
   - Add wallet connection UI

3. **Enhanced Features:**
   - Add AURA-specific module dashboards (DEX, Bridge, etc.)
   - Implement real-time WebSocket updates
   - Add advanced analytics and reporting

4. **Testing:**
   - Run comprehensive test suites
   - Perform security audit
   - Load testing with production data

5. **Documentation:**
   - Create user guides
   - Add video tutorials
   - Document API integration
