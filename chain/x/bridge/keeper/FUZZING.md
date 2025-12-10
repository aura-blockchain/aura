# Bridge Module Fuzz Testing

This directory contains Go native fuzz tests (Go 1.18+) for the Bridge module.

## Fuzz Tests

### FuzzSignatureVerification

Tests signature verification logic with random and malformed signatures.

**Security properties tested:**
- Rejects invalid signatures
- Rejects signatures from unauthorized validators
- Requires minimum threshold of valid signatures
- Handles malformed signature data gracefully
- Prevents signature forgery

**Run:**
```bash
go test -v ./x/bridge/keeper -fuzz=FuzzSignatureVerification -fuzztime=30s
```

### FuzzTransferAmountValidation

Tests transfer amount validation with edge cases.

**Security properties tested:**
- Rejects zero/negative amounts
- Enforces max transfer limits
- Handles integer overflow gracefully
- Validates amount consistency across operations

**Edge cases:**
- Zero amounts
- Negative amounts (via overflow)
- Maximum int64 values
- Amounts exceeding configured limits

**Run:**
```bash
go test -v ./x/bridge/keeper -fuzz=FuzzTransferAmountValidation -fuzztime=30s
```

### FuzzCrossChainMessageParsing

Tests parsing of cross-chain messages with malformed data.

**Security properties tested:**
- Handles malformed transaction hashes
- Validates chain identifiers
- Rejects invalid address formats
- Handles malformed denom strings
- Prevents injection attacks via message fields

**Attack vectors tested:**
- SQL injection attempts
- XSS attempts
- Path traversal attempts
- Control character injection
- Excessively long strings

**Run:**
```bash
go test -v ./x/bridge/keeper -fuzz=FuzzCrossChainMessageParsing -fuzztime=30s
```

### FuzzAddressValidation

Tests address validation across different chain formats (Aura/PAW/XAI).

**Security properties tested:**
- Rejects invalid Bech32 addresses
- Handles different address formats per chain
- Prevents address format confusion attacks
- Validates address checksums
- Handles malformed address data

**Run:**
```bash
go test -v ./x/bridge/keeper -fuzz=FuzzAddressValidation -fuzztime=30s
```

## Running All Fuzz Tests

Run all fuzz tests for a short duration:
```bash
go test -v ./x/bridge/keeper -fuzz=. -fuzztime=10s
```

Run all fuzz tests for extended duration (continuous fuzzing):
```bash
go test -v ./x/bridge/keeper -fuzz=. -fuzztime=5m
```

## Corpus Management

Go fuzzing automatically maintains a corpus of interesting test cases in:
```
testdata/fuzz/FuzzSignatureVerification/
testdata/fuzz/FuzzTransferAmountValidation/
testdata/fuzz/FuzzCrossChainMessageParsing/
testdata/fuzz/FuzzAddressValidation/
```

These corpus entries are valuable regression tests. Commit them to the repository.

## CI Integration

For CI/CD, run fuzz tests in regression mode (no new fuzzing, just replay corpus):
```bash
go test -v ./x/bridge/keeper
```

For extended fuzzing in CI (if time permits):
```bash
go test -v ./x/bridge/keeper -fuzz=. -fuzztime=1m
```

## Security Invariants

All fuzz tests enforce critical security invariants:

1. **No panics**: Invalid input must never cause panics
2. **Validation**: All inputs must be validated before processing
3. **Threshold enforcement**: Signature thresholds must never be bypassed
4. **Amount validation**: Zero/negative amounts must be rejected
5. **Injection prevention**: No SQL/XSS/path traversal injection possible
6. **Address validation**: Invalid addresses must be rejected

## Adding New Fuzz Tests

When adding new fuzz tests:

1. Use the `func FuzzXxx(f *testing.F)` pattern
2. Add seed corpus with `f.Add(...)` calls
3. Document security properties being tested
4. Ensure tests enforce security invariants
5. Add entries to this README

## References

- [Go Fuzzing Tutorial](https://go.dev/doc/tutorial/fuzz)
- [Go Fuzzing Documentation](https://go.dev/security/fuzz/)
- [Cosmos SDK Security Best Practices](https://docs.cosmos.network/main/build/building-modules/security)
