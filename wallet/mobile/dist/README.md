# Aura Mobile Install Artifacts (Testnet)

Artifacts generated from the Capacitor-wrapped web wallet (`wallet/web`) for quick Android installs. Build date: 2025-12-18T16:15:18Z

## Files
- `aura-wallet-testnet-debug.apk` — Android debug build (web bundle packaged via Capacitor)

SHA256:
```
3cd7d7295c5cfd9c6120f3f7235747758c6b166ca1ef259ae6641a4d7a5664f4  aura-wallet-testnet-debug.apk
```

## Build Recipe
```bash
cd wallet/mobile/capacitor
npm install           # already done
npx cap sync android
docker run --rm -v $PWD/../..:/workspace -w /workspace/wallet/mobile/capacitor/android \
  reactnativecommunity/react-native-android:latest \
  bash -lc "./gradlew assembleDebug"
cp android/app/build/outputs/apk/debug/app-debug.apk ../dist/aura-wallet-testnet-debug.apk
```

## Install (Android)
1) Enable “Install from unknown sources” on the device/emulator.  
2) Transfer `aura-wallet-testnet-debug.apk` and install.  
3) On first launch, allow network access; the app points at `https://rpc.aura-testnet.com`/`https://api.aura-testnet.com` via the embedded web bundle.  
4) Verify address prefix `aura` and set gas price to `0.025uaura` before sending funds.

## iOS / TestFlight
- The Capacitor project includes the Android build only. To create a TestFlight IPA, run `npx cap add ios` on macOS, open the generated Xcode project, and archive with a TestFlight profile. Use the same webDir (`../../web`) and chain-registry values from `docs/testnet/ONBOARDING_KIT.md`.

## Notes
- The APK is a packaged web wallet (no native signing/storage). Use hardware wallets or WalletConnect for high-value testing.
- Regenerate the APK after updating `wallet/web` assets or chain configuration.
