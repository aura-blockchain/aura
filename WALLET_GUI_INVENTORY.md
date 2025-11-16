# AURA Wallet & GUI Code Inventory
## Existing Code from Crypto & PAW Projects

**Date:** November 13, 2025
**Source Projects:** /Crypto and /PAW (same license, reusable)

---

## 📦 What We Found

### 1. Browser Wallet Extensions (2 implementations)

#### From PAW Project
**Location:** `/c/Users/decri/GitClones/paw/external/crypto/browser-wallet-extension/`

**Features:**
- Chrome/Firefox compatible (Manifest V3)
- WalletConnect-style integration
- Mining controls (can adapt for AURA PoI)
- Wallet-to-wallet trading
- Order matching UI
- Automatic refresh and broadcasts

**Files:**
- `manifest.json` - Extension config
- `popup.html` - Main UI
- `popup.js` - UI logic
- `background.js` - Service worker
- `styles.css` - Styling

**API Integration:**
- REST API calls to blockchain node
- Storage permissions for wallet data
- Host permissions for localhost/network

#### From Crypto Project
**Location:** `/c/Users/decri/GitClones/Crypto/src/aixn/browser_wallet_extension/`

**Similar features, can merge best of both**

---

### 2. Electron Desktop Wallet

**Location:** `/c/Users/decri/GitClones/Crypto/src/aixn/electron/`

**Features:**
- Cross-platform desktop app (Windows, Mac, Linux)
- Auto-starts blockchain node
- Built-in explorer dashboard
- System tray integration
- QR code mobile linking
- Electron-builder for distributables

**Files:**
- `main.js` - Electron main process (3,266 lines)
- `preload.js` - Security bridge
- `package.json` - Dependencies and build config
- `assets/` - Icons, images
- README with build instructions

**Tech Stack:**
- Electron framework
- Node.js backend integration
- Auto-spawn Python node (adaptable for Go chain)
- Dashboard at `http://localhost:3000`

---

### 3. Mobile Wallet Components

**Location:** `/c/Users/decri/GitClones/Crypto/src/aixn/core/`

#### Mobile Wallet Bridge
**File:** `mobile_wallet_bridge.py`
- Bridge between mobile apps and blockchain
- QR code generation
- Session management
- Mobile-optimized API

#### Mobile Cache
**File:** `mobile_cache.py`
- Caching layer for mobile performance
- Offline data storage
- Sync protocols

#### Mobile Template
**File:** `templates/mobile.html`
- Mobile-responsive web interface
- Can be wrapped in React Native WebView
- QR scanning interface

---

### 4. Wallet Implementations

**Location:** `/c/Users/decri/GitClones/Crypto/src/aixn/core/`

#### Core Wallet
**File:** `wallet.py`
- Basic wallet functionality
- Key generation
- Transaction signing
- Balance management

#### Hardware Wallet Support
**Files:**
- `hardware_wallet.py` - Generic hardware wallet interface
- `hardware_wallet_ledger.py` - Ledger device integration
- Secure key storage
- Transaction approval flow

#### Exchange Wallet
**File:** `exchange_wallet.py`
- Trading wallet functionality
- Order management
- Liquidity pool integration

---

### 5. GUI Components

#### Simple Swap GUI
**File:** `/c/Users/decri/GitClones/Crypto/src/aixn/core/simple_swap_gui.py`

**Features:**
- One-click swap interface
- On-chain orderbook (P2P)
- Auto-matching buyers/sellers
- Progress tracking
- Price floor enforcement
- HTLC (Hash Time-Locked Contracts)

**Classes:**
- `SwapOrderType` - Buy/Sell orders
- `SwapOrderStatus` - Order lifecycle
- `SwapOrder` - Individual order
- GUI integration ready

---

### 6. Exchange Frontend

**Location:** `/c/Users/decri/GitClones/Crypto/exchange/frontend/`

**File:** `app.js`
- Trading interface
- Order book display
- Chart integration
- Real-time updates

---

## 🎯 Adaptation Plan for AURA

### Phase 1: Browser Wallet Extension (Priority 1)

**Adapt for AURA's new features:**

1. **QR Code Verification (Feature #1)**
   - Replace mining controls with QR generation
   - Add "Generate QR" button
   - Display QR code with expiration timer
   - Show attribute selection checkboxes

2. **Selective Disclosure (Feature #2)**
   - Add attribute selection UI
   - Voice command input field
   - Disclosure policy settings page
   - Real-time preview of what's shared

3. **Data Registry (Feature #3)**
   - Add "My Data" tab
   - Upload interface for documents/photos
   - IPFS integration
   - Geo-tagging for golf scores, photos
   - Verification status display

**New manifest.json:**
```json
{
  "manifest_version": 3,
  "name": "AURA Identity Wallet",
  "description": "Decentralized identity with QR verification, selective disclosure, and verified data storage",
  "version": "1.0.0",
  "action": {
    "default_popup": "popup.html"
  },
  "background": {
    "service_worker": "background.js"
  },
  "host_permissions": [
    "http://localhost:26657/*",
    "http://localhost:1317/*",
    "http://localhost:9090/*"
  ],
  "permissions": [
    "storage",
    "camera",
    "geolocation"
  ]
}
```

---

### Phase 2: Electron Desktop Wallet (Priority 2)

**Adapt `main.js` for AURA:**

```javascript
// Replace Python node spawn with Go chain
const { spawn } = require('child_process');
const auradProcess = spawn('aurad', ['start'], {
  env: {
    ...process.env,
    CHAIN_ID: 'aura-1'
  }
});

// Dashboard URL changes
const dashboardURL = 'http://localhost:1317/swagger/';
const explorerURL = 'http://localhost:3000';

// Add AURA-specific menu items
const menu = Menu.buildFromTemplate([
  {
    label: 'AURA',
    submenu: [
      { label: 'Generate QR Code', click: showQRGenerator },
      { label: 'My Data Registry', click: showDataRegistry },
      { label: 'Disclosure Settings', click: showDisclosureSettings },
      { type: 'separator' },
      { label: 'Quit', click: () => app.quit() }
    ]
  }
]);
```

**New Features:**
- Dashboard shows IR completion status
- QR code generator window
- Data registry browser
- Attribute management UI
- Hardware wallet integration (Ledger support)

---

### Phase 3: Mobile Wallet Bridge (Priority 3)

**Adapt for AURA (Go backend):**

```go
// mobile_bridge.go
package mobile

type MobileBridge struct {
    lcd     *LCDClient
    grpc    *GRPCClient
    cache   *MobileCache
    qrGen   *QRGenerator
}

func (m *MobileBridge) GenerateQR(vcIDs []string, context PresentationContext) (string, error) {
    // Call vcregistry CreatePresentation
    // Return QR code data as base64
}

func (m *MobileBridge) GetUserAttributes(address string) ([]AttributeVC, error) {
    // Fetch AttributeVCs for mobile display
}

func (m *MobileBridge) UploadToIPFS(data []byte) (string, error) {
    // Upload to IPFS, return CID
}

func (m *MobileBridge) StoreDataItem(item DataItem) error {
    // Store in dataregistry module
}
```

**Mobile App Integration:**
- React Native wrapper around bridge
- Native camera for QR scanning
- Native biometric (Face ID, Touch ID)
- GPS for geolocation
- Voice recognition for commands

---

### Phase 4: React Native Mobile App (Priority 4)

**Project Structure:**
```
aura-mobile/
├── src/
│   ├── screens/
│   │   ├── HomeScreen.tsx
│   │   ├── QRGenerateScreen.tsx
│   │   ├── QRScanScreen.tsx
│   │   ├── AttributesScreen.tsx
│   │   ├── DataRegistryScreen.tsx
│   │   └── SettingsScreen.tsx
│   ├── components/
│   │   ├── QRCodeDisplay.tsx
│   │   ├── AttributeSelector.tsx
│   │   ├── VoiceCommandInput.tsx
│   │   ├── DataItemCard.tsx
│   │   └── VerificationStatus.tsx
│   ├── services/
│   │   ├── AuraAPI.ts
│   │   ├── IPFS.ts
│   │   ├── Biometric.ts
│   │   └── Geolocation.ts
│   └── navigation/
│       └── AppNavigator.tsx
├── ios/
├── android/
└── package.json
```

**Key Libraries:**
- `react-native-camera` - QR scanning
- `react-native-biometrics` - Face ID/Touch ID
- `react-native-geolocation` - GPS
- `react-native-voice` - Voice commands
- `@react-native-community/async-storage` - Local cache
- `react-native-qrcode-svg` - QR generation

---

## 🔧 Technical Mapping

### API Endpoint Mapping

| Crypto/PAW API | AURA Equivalent | Module |
|----------------|-----------------|--------|
| `/wallet-trades/register` | `/aura/vcregistry/v1beta1/presentations` | vcregistry |
| `/wallet-trades/orders` | `/aura/dataregistry/v1beta1/items` | dataregistry |
| `/mining/start` | N/A (PoS chain) | - |
| `/mining/status` | `/cosmos/staking/v1beta1/delegations` | staking |
| `/wallet/balance` | `/cosmos/bank/v1beta1/balances` | bank |
| `/wallet/transactions` | `/cosmos/tx/v1beta1/txs` | tx |

### Data Structure Mapping

| Crypto/PAW | AURA | Notes |
|------------|------|-------|
| `WalletAddress` | `bech32 address` | Standard Cosmos format |
| `TradeOrder` | `DataItem` | Similar structure |
| `HTLCContract` | N/A | Not needed for AURA |
| `MiningReward` | `POI Reward` | Different mechanism |
| `SwapOrder` | `Presentation` | QR code presentation |

---

## 📱 Platform Coverage

After adaptation, AURA will have:

✅ **Browser Extension**
- Chrome, Firefox, Edge, Brave
- Manifest V3 compliant
- Full Feature #1, #2, #3 support

✅ **Desktop Wallet**
- Windows, macOS, Linux
- Electron-based
- Auto-start chain node
- Full-featured dashboard

✅ **Mobile Apps**
- iOS (React Native)
- Android (React Native)
- Camera QR scanning
- Biometric authentication
- Voice commands
- GPS geolocation

✅ **Web Interface**
- Mobile-responsive
- Progressive Web App (PWA)
- Can be accessed from any browser

---

## 🚀 Implementation Steps

### Step 1: Browser Extension (Week 1)

1. Copy `/paw/external/crypto/browser-wallet-extension/` to `/aura/wallet/browser-extension/`
2. Update `manifest.json` for AURA
3. Modify `popup.html` to add:
   - QR generation UI
   - Attribute selection checkboxes
   - Data registry tab
4. Update `popup.js` to call AURA APIs:
   - `/aura/vcregistry/v1beta1/presentations`
   - `/aura/vcregistry/v1beta1/attributes`
   - `/aura/dataregistry/v1beta1/items`
5. Update `background.js` for IPFS integration
6. Test with local AURA node

### Step 2: Desktop Wallet (Week 2)

1. Copy `/Crypto/src/aixn/electron/` to `/aura/wallet/desktop/`
2. Update `main.js` to spawn `aurad` instead of Python node
3. Create dashboard pages:
   - QR generator
   - Attribute manager
   - Data registry browser
4. Add system tray icons for AURA
5. Configure `electron-builder` for AURA branding
6. Build distributables for Windows/Mac/Linux

### Step 3: Mobile Bridge (Week 3)

1. Port `mobile_wallet_bridge.py` to Go
2. Implement `mobile_bridge.go` with AURA APIs
3. Add mobile-specific endpoints:
   - `/mobile/qr/generate`
   - `/mobile/attributes/list`
   - `/mobile/data/upload`
4. Implement caching layer
5. Test with mobile devices

### Step 4: React Native App (Week 4-6)

1. Initialize React Native project
2. Implement screens (Home, QR, Attributes, Data)
3. Integrate camera for QR scanning
4. Add biometric authentication
5. Implement voice command input
6. Connect to AURA mobile bridge
7. Test on iOS and Android
8. Submit to App Store and Google Play

---

## 💰 Estimated Effort

| Component | Effort | Priority |
|-----------|--------|----------|
| Browser Extension | 1 week | High |
| Desktop Wallet | 1 week | High |
| Mobile Bridge | 1 week | Medium |
| React Native App | 3 weeks | Medium |
| Testing & Polish | 2 weeks | High |
| **Total** | **8 weeks** | - |

---

## 🎯 Feature Integration Matrix

| Wallet Type | Feature #1 (QR) | Feature #2 (Disclosure) | Feature #3 (Data) |
|-------------|-----------------|-------------------------|-------------------|
| Browser Extension | ✅ Full | ✅ Full | ✅ Full |
| Desktop Wallet | ✅ Full | ✅ Full | ✅ Full |
| Mobile Bridge | ✅ Full | ✅ Full | ✅ Limited (no upload UI) |
| Mobile App | ✅ Full | ✅ Full | ✅ Full |
| Web Interface | ✅ Full | ✅ Full | ✅ View only |

---

## 📋 Next Actions

**Immediate (Today):**
1. Copy browser extension code to AURA project
2. Update manifest and APIs
3. Test with local node

**This Week:**
1. Complete browser extension
2. Start desktop wallet adaptation
3. Design mobile app screens

**Next Week:**
1. Implement mobile bridge in Go
2. Continue React Native app development
3. Begin app store submission prep

---

## 🎉 Conclusion

We found **extensive wallet and GUI code** from Crypto and PAW projects that can be adapted for AURA with minimal effort. Instead of building from scratch, we can:

- ✅ Adapt browser extension (1 week instead of 4 weeks)
- ✅ Adapt desktop wallet (1 week instead of 3 weeks)
- ✅ Use mobile bridge patterns (1 week instead of 2 weeks)
- ✅ Follow proven React Native patterns (3 weeks instead of 6 weeks)

**Total time saved: ~8 weeks!**

All code is from your own projects, so no licensing concerns. The adaption work is straightforward since the blockchain interaction patterns are similar (REST APIs, transactions, queries).

**Status:** Ready to begin adaptation! 🚀
