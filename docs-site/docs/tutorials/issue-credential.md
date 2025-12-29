---
sidebar_position: 3
---

# Issue a Verifiable Credential

**Difficulty:** Intermediate | **Time:** 10 minutes

Learn to issue W3C-compliant Verifiable Credentials on Aura.

## Prerequisites

- [Created a DID](./create-did)
- Registered as an issuer (requires governance approval for production)

## Credential Types

Aura supports standard credential types:
- **IdentityCredential**: Basic identity verification
- **AgeCredential**: Age verification (18+, 21+)
- **ResidencyCredential**: Location verification
- **EmploymentCredential**: Employment status

## Steps

### 1. Register as Issuer (Testnet)

```bash
aurad tx vcregistry register-issuer \
  --name "My Organization" \
  --credential-types "IdentityCredential,AgeCredential" \
  --from my-wallet \
  --chain-id aura-testnet-1 \
  -y
```

### 2. Issue a Credential

```bash
aurad tx vcregistry issue-credential \
  --subject-did "did:aura:1subject..." \
  --credential-type "IdentityCredential" \
  --claims '{"verified": true, "level": "basic"}' \
  --expiration "2026-01-01T00:00:00Z" \
  --from my-wallet \
  --chain-id aura-testnet-1 \
  -y
```

### 3. Query the Credential

```bash
aurad query vcregistry credential <credential-id>
```

## Credential Lifecycle

1. **Issuance**: Issuer creates and signs credential
2. **Active**: Credential can be verified
3. **Suspended**: Temporarily invalid (can be reactivated)
4. **Revoked**: Permanently invalid

### Revoking a Credential

```bash
aurad tx vcregistry revoke-credential <credential-id> \
  --reason "Information outdated" \
  --from my-wallet \
  -y
```

## Verification

Anyone can verify a credential:

```bash
aurad query vcregistry verify-credential <credential-id>
```

Returns verification status, issuer info, and expiration.

## Next Steps

- [Module Guide: VC Registry](/docs/modules/vcregistry) - Deep dive into credentials
- [Governance](./governance) - Become a verified issuer on mainnet
