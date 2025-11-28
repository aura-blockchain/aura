# AURA Validator Dashboard - Quick Start Guide

## Instant Setup (3 Steps)

### Step 1: Navigate to Directory
```bash
cd C:\Users\decri\GitClones\PAW\dashboards\validator
```

### Step 2: Start with Docker (Recommended)
```bash
docker-compose up -d
```

### Step 3: Open Browser
```
http://localhost:8080
```

That's it! The dashboard is now running.

---

## Alternative: Local Development

### Step 1: Install Dependencies
```bash
npm install
```

### Step 2: Start Development Server
```bash
npm run dev
```

### Step 3: Open Browser
```
http://localhost:8080
```

---

## First Use

### 1. Add Your First Validator

1. Click **"Add Validator"** button (top right)
2. Enter validator address: `auravaloper1...`
3. (Optional) Add a display name
4. Click **"Add"**

### 2. Explore the Dashboard

- **Overview**: See validator status and statistics
- **Delegations**: View all delegators
- **Rewards**: Track rewards over time
- **Performance**: Monitor performance metrics
- **Uptime**: Check block signing history
- **Signing**: View signing statistics
- **Slashing**: See slash events (if any)
- **Settings**: Configure alerts and preferences

### 3. Configure Blockchain Endpoints

If connecting to a real blockchain:

**Edit `services/validatorAPI.js`:**
```javascript
static baseURL = 'https://your-lcd-endpoint.com';
```

**Edit `services/websocket.js`:**
```javascript
this.wsURL = 'wss://your-websocket-endpoint.com/websocket';
```

---

## Running Tests

### All Tests
```bash
npm test
```

### Unit Tests Only
```bash
npm run test:unit
```

### Integration Tests Only
```bash
npm run test:integration
```

### E2E Tests Only
```bash
npm run test:e2e
```

---

## Docker Commands

### Start Dashboard
```bash
docker-compose up -d
```

### Stop Dashboard
```bash
docker-compose down
```

### View Logs
```bash
docker-compose logs -f
```

### Restart Dashboard
```bash
docker-compose restart
```

### Run Tests in Docker
```bash
docker-compose --profile test up test-runner
```

---

## Troubleshooting

### Issue: "Port 8080 already in use"

**Solution:**
```bash
# Change port in docker-compose.yml
ports:
  - "8081:80"  # Use 8081 instead
```

### Issue: "Cannot connect to blockchain"

**Solution:**
1. Verify blockchain is running
2. Check endpoint URLs in configuration
3. Enable mock data mode (automatic fallback)

### Issue: "WebSocket connection failed"

**Solution:**
1. Check WebSocket endpoint is accessible
2. Verify firewall rules
3. Dashboard will auto-retry connection

---

## Configuration Files

### package.json
Dependencies and npm scripts

### docker-compose.yml
Docker container configuration

### nginx.conf
Web server configuration

### jest.config.js
Unit/integration test configuration

### playwright.config.js
E2E test configuration

---

## File Structure Overview

```
dashboards/validator/
├── index.html              # Main page
├── app.js                  # Main app logic
├── assets/css/styles.css   # Styling
├── components/             # UI components
├── services/               # API & WebSocket
└── tests/                  # Test suite
```

---

## Key Features

✅ Real-time monitoring
✅ Multi-validator support
✅ Responsive design
✅ Comprehensive statistics
✅ Interactive charts
✅ Delegation management
✅ Performance tracking
✅ Uptime monitoring
✅ Alert configuration
✅ 85+ tests included

---

## Next Steps

1. **Configure Endpoints** - Connect to your blockchain
2. **Add Validators** - Monitor your validators
3. **Set Alerts** - Configure uptime alerts
4. **Explore Features** - Try all dashboard sections
5. **Run Tests** - Verify everything works

---

## Support

- 📖 Full Documentation: `README.md`
- 🧪 Test Results: `TEST_RESULTS.md`
- 📊 Implementation Details: `IMPLEMENTATION_SUMMARY.md`

---

## Production Deployment

For production use:
1. Use HTTPS endpoints
2. Configure proper CORS
3. Enable security headers
4. Set up monitoring
5. Configure backups

See `README.md` for detailed production deployment guide.

---

**Ready to use! 🚀**
