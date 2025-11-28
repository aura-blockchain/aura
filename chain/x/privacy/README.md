# Privacy Module

The Privacy module provides comprehensive privacy and anonymity features for the Aura blockchain, implementing state-of-the-art cryptographic protocols to protect user transaction data.

## Features

### 1. Zero-Knowledge Proofs (zkproof.go)

**Lines 1-417**: Implementation of multiple ZK proof systems for private transactions.

#### Supported Proof Systems
- **Groth16** (Lines 47-90): Efficient SNARK with constant-size proofs
- **PLONK** (Lines 92-107): Universal SNARK without per-circuit trusted setup
- **Bulletproofs** (Lines 109-124): Short proofs for range proofs
- **STARK** (Lines 126-141): Transparent, quantum-resistant proofs

#### Key Functions
- `NewZKProofSystem()` (Lines 24-43): Creates a new ZK proof system
- `GenerateProof()` (Lines 45-60): Generates a zero-knowledge proof
- `VerifyProof()` (Lines 62-77): Verifies a ZK proof
- `GenerateRangeProof()` (Lines 180-208): Creates range proofs to prove values are within bounds
- `VerifyRangeProof()` (Lines 210-222): Verifies range proofs

#### Cryptographic Schemes

**Pedersen Commitments** (Lines 224-234): Commitment scheme using `C = vG + rH`
- `v`: value being committed
- `G, H`: generator points on elliptic curve
- `r`: random blinding factor

**Membership Proofs** (Lines 236-292): Proves set membership without revealing which element

### 2. Stealth Addresses (stealth.go)

**Lines 1-377**: One-time addresses for recipient privacy.

#### Implementation Details

**ECDSA-based Stealth Addresses** (Lines 18-176):
- Uses NIST P-256 elliptic curve
- Dual-key system: spend key + view key
- One-time address generation: `P = H(rA)G + B`

**Curve25519 Stealth Addresses** (Lines 178-261):
- More efficient than NIST curves
- Used in Monero and similar protocols
- 32-byte keys for compact representation

#### Key Functions
- `GenerateStealthKeys()` (Lines 35-66): Creates spend and view key pairs
- `GenerateStealthAddress()` (Lines 77-118): Generates one-time stealth address
- `ScanForStealthPayments()` (Lines 120-162): Scans blockchain for payments
- `DerivePrivateKey()` (Lines 164-195): Derives spending key from stealth address

#### Cryptographic Protocol

1. Recipient publishes `(A, B)` where `A` is view key, `B` is spend key
2. Sender generates ephemeral key `r`
3. Sender computes:
   - `R = rG` (transaction public key)
   - `P = H(rA)G + B` (one-time address)
4. Recipient scans using: `P' = H(aR)G + B`
5. If `P = P'`, the payment is for recipient
6. Spending key: `x = H(aR) + b`

### 3. Ring Signatures (ringsig.go)

**Lines 1-459**: Sender anonymity through ring signatures.

#### Features

**Basic Ring Signatures** (Lines 18-218):
- Proves signer is in a set without revealing which member
- Prevents double-spending using key images
- Configurable ring size for anonymity

**Linkable Ring Signatures** (Lines 220-263):
- Detects if two signatures are from same signer
- Uses linking tags
- Prevents signature reuse

**MLSAG (Multi-layered Linkable SAG)** (Lines 265-388):
- Used in Monero for RingCT
- Supports multiple input commitments
- Enhanced anonymity set

#### Key Functions
- `Sign()` (Lines 43-136): Creates ring signature
- `Verify()` (Lines 138-188): Verifies ring signature
- `generateKeyImage()` (Lines 190-202): Creates unique key image `I = xH(P)`
- `SignMLSAG()` (Lines 295-357): Creates multi-layered signature
- `VerifyMLSAG()` (Lines 359-383): Verifies MLSAG signature

#### Cryptographic Protocol

1. Signer has private key `x`, public key `P`
2. Ring contains `n` public keys including `P`
3. Key image: `I = x·H(P)` prevents double-spending
4. For each ring member `i`:
   - `L[i] = q[i]G + c[i]P[i]`
   - `R[i] = q[i]H(P[i]) + c[i]I`
5. Challenge: `c[i+1] = H(m, L[i], R[i])`
6. Verification checks if ring closes: `c[0] = c[n]`

### 4. Confidential Transactions (confidential.go)

**Lines 1-420**: Hide transaction amounts using homomorphic encryption.

#### Components

**Pedersen Commitments** (Lines 64-103):
- Commitment: `C = vG + rH`
- Computationally hiding and perfectly binding
- Supports homomorphic addition

**Bulletproofs** (Lines 115-219):
- Logarithmic-size range proofs
- Proves `0 ≤ v < 2^n` without revealing `v`
- Inner product arguments for efficiency

**Ring Confidential Transactions (RingCT)** (Lines 221-306):
- Combines ring signatures with confidential amounts
- Used in Monero protocol
- ECDH for amount encryption

#### Key Functions
- `NewConfidentialTransactionSystem()` (Lines 31-62): Initializes CT system
- `CreateCommitment()` (Lines 73-103): Creates Pedersen commitment
- `GenerateBulletproof()` (Lines 117-187): Generates range proof
- `VerifyBulletproof()` (Lines 189-219): Verifies range proof
- `CreateRingCT()` (Lines 230-282): Creates confidential transaction
- `VerifyRingCT()` (Lines 284-306): Verifies RingCT balance

#### Cryptographic Properties

**Commitment Scheme**:
- Hiding: Cannot determine `v` from `C`
- Binding: Cannot find `v' ≠ v` with same `C`
- Homomorphic: `C1 + C2 = (v1+v2)G + (r1+r2)H`

**Range Proof**:
- Proves `v ∈ [0, 2^n)` without revealing `v`
- Size: `O(log n)` with Bulletproofs
- Verification time: `O(n)`

### 5. Network Privacy (network.go)

**Lines 1-363**: Tor/I2P integration for network-level anonymity.

#### Tor Integration (Lines 54-143)

- SOCKS5 proxy support
- Circuit management
- Stream isolation
- Automatic circuit rotation

**Circuit Structure**:
- Entry node (guard)
- Middle node
- Exit node
- Configurable lifetime

#### I2P Integration (Lines 145-225)

- HTTP proxy support
- Destination keys
- Tunnel creation and management
- Garlic routing

**Tunnel Parameters**:
- Inbound/outbound tunnels
- Configurable length (0-7 hops)
- Latency vs anonymity tradeoff

#### Key Functions
- `NewNetworkPrivacyManager()` (Lines 33-65): Initializes privacy manager
- `CreateCircuit()` (Lines 246-285): Creates Tor circuit or I2P tunnel
- `RotateCircuits()` (Lines 307-339): Rotates expired circuits
- `MakeRequest()` (Lines 134-149): Makes HTTP request through Tor

### 6. Coin Mixing/Tumbling (mixing.go)

**Lines 1-372**: Privacy through transaction mixing.

#### CoinJoin Implementation (Lines 91-145)

**Process**:
1. Participants join pool with same denomination
2. Inputs are collected
3. Outputs are shuffled
4. Combined transaction is created
5. ZK proof of correct mixing

#### Mixing Pools (Lines 17-44)

**Statuses**:
- `PENDING`: Waiting for participants
- `ACTIVE`: Ready to mix
- `MIXING`: Mixing in progress
- `COMPLETED`: Mixing finished
- `CANCELLED`: Mixing failed

#### Key Functions
- `CreatePool()` (Lines 79-125): Creates mixing pool
- `JoinPool()` (Lines 127-185): Join existing pool
- `ExecuteMixing()` (Lines 187-232): Execute mixing process
- `performCoinJoin()` (Lines 234-274): CoinJoin implementation
- `shuffleOutputs()` (Lines 284-305): Fisher-Yates shuffle

#### Tumbling Service (Lines 320-372)

**Features**:
- Scheduled tumbling
- Split amounts across multiple outputs
- Time-delayed transactions
- Increased anonymity through multiple mixing rounds

### 7. Encrypted Memos (encryption.go)

**Lines 1-469**: Encrypted transaction memos.

#### Supported Algorithms

**AES-256-GCM** (Lines 68-130):
- 256-bit key encryption
- Galois/Counter Mode
- Authenticated encryption
- 12-byte nonce

**ChaCha20-Poly1305** (Lines 179-241):
- Stream cipher with MAC
- More efficient on mobile devices
- 12-byte nonce
- IETF standard

**XChaCha20-Poly1305** (Lines 243-305):
- Extended nonce (24 bytes)
- Better security margins
- Suitable for random nonces

#### Key Functions
- `Encrypt()` (Lines 39-55): Encrypts memo for recipient
- `Decrypt()` (Lines 57-73): Decrypts encrypted memo
- `deriveSharedSecret()` (Lines 307-326): ECDH shared secret
- `deriveEncryptionKey()` (Lines 328-334): Key derivation function

#### Encryption Protocol

1. Generate ephemeral key pair `(r, R)`
2. Compute shared secret: `S = r·PubKey`
3. Derive encryption key: `K = KDF(S, info)`
4. Encrypt: `C = AEAD.Encrypt(K, nonce, plaintext)`
5. Output: `(R, nonce, C, tag)`

### 8. View Keys (encryption.go, Lines 336-469)

**Lines 336-469**: Selective disclosure through view keys.

#### View Key Types

- **INCOMING**: View received transactions
- **OUTGOING**: View sent transactions
- **AUDIT**: Full audit access
- **FULL**: Complete transaction access

#### Key Functions
- `GenerateViewKey()` (Lines 364-399): Creates new view key
- `GetViewKey()` (Lines 401-417): Retrieves view key
- `RevokeViewKey()` (Lines 419-430): Revokes view key
- `VerifyPermission()` (Lines 432-445): Checks permissions
- `DecryptWithViewKey()` (Lines 447-469): Decrypts using view key

#### Use Cases

1. **Compliance**: Provide audit view keys to regulators
2. **Accounting**: Grant incoming view key to accountant
3. **Monitoring**: Track specific transaction types
4. **Temporary Access**: Time-limited view keys

## Security Considerations

### Zero-Knowledge Proofs
- Requires trusted setup for Groth16 (can be eliminated with PLONK/STARK)
- Soundness: Prevents false proofs
- Zero-knowledge: Reveals nothing beyond statement validity
- Use parameter generation ceremonies for production

### Stealth Addresses
- View key compromise reveals all received transactions
- Spend key compromise allows theft
- Keep keys separate and secure
- Regularly scan for payments

### Ring Signatures
- Larger rings = better anonymity but higher cost
- Key images must be tracked to prevent double-spending
- Ring members should be diverse
- Minimum ring size: 11 (recommended)

### Confidential Transactions
- Range proofs prevent negative amounts
- Commitment keys must be kept secret
- Bulletproof verification is relatively expensive
- Use batch verification when possible

### Network Privacy
- Tor requires running Tor daemon
- I2P has higher latency
- Circuit rotation prevents long-term correlation
- Use bridges if Tor is blocked

### Mixing
- Requires critical mass of participants
- Timing analysis possible if unique denominations
- Multiple rounds increase anonymity
- CoinJoin coordinator can be decentralized

### Encrypted Memos
- ECDH provides forward secrecy if ephemeral keys used
- Nonces must never be reused
- Authentication tags prevent tampering
- Use XChaCha20 for random nonces

### View Keys
- Implement strict permission checks
- Set expiration times
- Revocation should be immediate
- Audit key usage logs

## Performance Characteristics

| Feature | Generation Time | Verification Time | Size |
|---------|----------------|-------------------|------|
| ZK Proof (Groth16) | ~1s | ~10ms | 128 bytes |
| ZK Proof (PLONK) | ~2s | ~20ms | 512 bytes |
| Bulletproof (64-bit) | ~100ms | ~50ms | 674 bytes |
| Ring Signature (size 11) | ~50ms | ~30ms | 2 KB |
| Stealth Address | ~5ms | ~5ms | 64 bytes |
| Encrypted Memo | ~1ms | ~1ms | variable |

## Usage Examples

### Creating a Private Transaction

```go
// Generate stealth address
scheme := NewStealthAddressScheme()
keys, _ := scheme.GenerateStealthKeys()
stealthAddr, _ := scheme.GenerateStealthAddress(
    keys.SpendKeyPair.PublicKey,
    keys.ViewKeyPair.PublicKey,
)

// Create confidential amount
ct, _ := NewConfidentialTransactionSystem(32)
commitment, _ := ct.CreateCommitment(big.NewInt(1000))
rangeProof, _ := ct.GenerateBulletproof(
    big.NewInt(1000),
    commitment.BlindingFactor,
)

// Create ring signature for sender anonymity
signer := NewRingSigner()
ringPublicKeys := [][]byte{/* public keys */}
signature, _ := signer.Sign(
    0, // your index
    privateKey,
    ringPublicKeys,
    message,
)

// Encrypt memo
encryptor := NewMemoEncryptor(AlgorithmChaCha20Poly1305)
encryptedMemo, _ := encryptor.Encrypt(
    []byte("payment for services"),
    recipientPublicKey,
)
```

### Joining a Mixing Pool

```go
mixingService := NewMixingService(5)

// Create pool
pool, _ := mixingService.CreatePool(
    big.NewInt(1000), // denomination
    5,  // min participants
    20, // max participants
    3,  // mixing rounds
    1 * time.Hour,
    big.NewInt(10), // fee
)

// Join pool
commitment := createCommitment(big.NewInt(1000))
err := mixingService.JoinPool(
    pool.ID,
    "your_address",
    commitment,
    outputAddress,
    blindingFactor,
)

// Execute mixing when ready
result, _ := mixingService.ExecuteMixing(pool.ID)
```

### Using View Keys

```go
vkm := NewViewKeyManager()

// Generate audit view key
viewKey, _ := vkm.GenerateViewKey(
    ViewKeyTypeAudit,
    address,
    []string{"view_incoming", "view_balance"},
    &expirationTime,
)

// Grant to auditor
// auditor can now decrypt transactions with view key

// Revoke when done
vkm.RevokeViewKey(viewKey.PublicKey)
```

## Testing

Comprehensive tests are provided in:
- `zkproof_test.go`: ZK proof systems
- `stealth_test.go`: Stealth addresses
- `privacy_test.go`: Integration tests

Run tests:
```bash
go test ./chain/x/privacy -v
```

## Dependencies

- `golang.org/x/crypto`: Cryptographic primitives
- `crypto/ecdsa`: Elliptic curve signatures
- `crypto/aes`: AES encryption
- `math/big`: Large number arithmetic

## Future Enhancements

1. **Threshold Signatures**: Multi-party signing
2. **zkSNARK Circuits**: Custom circuit compilation
3. **Payment Channels**: Off-chain privacy
4. **Atomic Swaps**: Cross-chain private swaps
5. **Quantum Resistance**: Lattice-based cryptography
6. **Hardware Wallet Support**: Secure key management

## References

1. Groth16: "On the Size of Pairing-based Non-interactive Arguments"
2. PLONK: "Permutations over Lagrange-bases for Oecumenical Noninteractive arguments of Knowledge"
3. Bulletproofs: "Bulletproofs: Short Proofs for Confidential Transactions"
4. CryptoNote: "CryptoNote v2.0"
5. Monero: "Ring Confidential Transactions"
6. Tor: "Tor Protocol Specification"
7. I2P: "Invisible Internet Project"

## License

Copyright (c) 2025 Aura Blockchain
