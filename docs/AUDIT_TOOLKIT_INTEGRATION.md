# Audit Toolkit Integration Guide for Aura Blockchain

**Based on:** blockchain_audit_toolkit.json v1.0.0
**Last Updated:** November 13, 2025

---

## Overview

This document integrates the Blockchain Security Audit Toolkit with Aura's implemented security features, providing a comprehensive guide for conducting security audits using specialized tools and agent workflows.

---

## 1. Tool Integration Mapping

### Smart Contract Static Analysis

**Aura Relevance:** While Aura is Cosmos SDK-based (Go), these tools are valuable for:
- Any EVM-compatible bridge contracts
- Cross-chain smart contracts
- Future Ethereum integration

#### Recommended Tools:

**1. Slither**
- **Purpose:** Static analysis for Solidity/Vyper
- **Install:** `pip install slither-analyzer`
- **Usage:** `slither . --exclude-dependencies`
- **Integration:** Add to CI/CD for bridge contract validation
- **Aura Application:** Scan bridge contracts before deployment

**2. Mythril**
- **Purpose:** Symbolic execution for EVM bytecode
- **Install:** `pip install mythril` or Docker
- **Usage:** `myth analyze contracts/Bridge.sol`
- **Aura Application:** Deep analysis of high-value bridge contracts

**3. Manticore**
- **Purpose:** Symbolic execution with custom analyses
- **Install:** Follow docs at https://manticore.readthedocs.io/
- **Aura Application:** Path exploration for complex bridge logic

### Smart Contract Fuzzing

**1. Echidna**
- **Purpose:** Property-based fuzzer
- **Homepage:** https://github.com/crytic/echidna
- **Aura Application:** Fuzz bridge invariants and state transitions

**2. Foundry/Forge**
- **Purpose:** Ethereum development toolkit with fuzzing
- **Install:** `curl -L https://foundry.paradigm.xyz | bash`
- **Usage:** `forge test --fuzz-runs 10000`
- **Aura Application:** Invariant testing for bridge contracts

### General Code Static Analysis

**1. CodeQL** ⭐ **HIGH PRIORITY**
- **Purpose:** Semantic code analysis for Go/Rust/JS
- **Setup:** GitHub Actions integration
- **Aura Application:** Scan ALL Aura Go modules
- **Integration:**

```yaml
# .github/workflows/codeql-analysis.yml
name: CodeQL Analysis
on: [push, pull_request]
jobs:
  analyze:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        language: ['go', 'javascript']
    steps:
      - uses: actions/checkout@v3
      - uses: github/codeql-action/init@v2
        with:
          languages: ${{ matrix.language }}
          queries: security-extended
      - uses: github/codeql-action/analyze@v2
```

**2. Semgrep** ⭐ **HIGH PRIORITY**
- **Purpose:** Fast pattern-based static analysis
- **Install:** `pip install semgrep`
- **Usage:** `semgrep --config=auto .`
- **Aura Application:** Enforce secure coding standards across all modules

**Configuration for Aura:**

```yaml
# .semgrep.yml
rules:
  - id: dangerous-panic
    pattern: panic($...)
    message: Avoid panic in production code
    languages: [go]
    severity: ERROR

  - id: unsafe-sql-query
    pattern: |
      db.Query($X, ...)
    message: Use parameterized queries
    languages: [go]
    severity: ERROR

  - id: hardcoded-secret
    pattern: |
      const $KEY = "..."
    message: Never hardcode secrets
    languages: [go]
    severity: CRITICAL
```

### Dependency & Supply Chain

**1. cargo-audit** (for Rust components)
- **Install:** `cargo install cargo-audit`
- **Usage:** `cargo audit`
- **Aura Application:** Currently not needed (no Rust), but useful for future

**2. gosec** ⭐ **IMPLEMENTED - ENHANCE**
- **Install:** `go install github.com/securego/gosec/v2/cmd/gosec@latest`
- **Usage:** `gosec ./...`
- **Current Status:** Already in CODE_QUALITY_FRAMEWORK.md
- **Enhancement:** Add custom rules for Cosmos SDK patterns

**Custom gosec config for Aura:**

```yaml
# .gosec.json
{
  "tests": true,
  "exclude-generated": true,
  "severity": "medium",
  "confidence": "medium",
  "excludes": [
    "G104"  // Exclude unhandled errors (use golangci-lint instead)
  ],
  "include": [
    "G101",  // Look for hard coded credentials
    "G102",  // Bind to all interfaces
    "G103",  // Audit the use of unsafe block
    "G104",  // Audit errors not checked
    "G401",  // Detect the usage of DES, RC4, MD4 or MD5
    "G501",  // Import blacklist: crypto/md5
    "G502",  // Import blacklist: crypto/des
    "G503",  // Import blacklist: crypto/rc4
    "G504"   // Import blacklist: net/http/cgi
  ]
}
```

**3. npm audit** (for any JS/TS frontends)
- **Usage:** `npm audit --audit-level=moderate`
- **Aura Application:** Scan wallet GUI dependencies

**4. Syft (SBOM)** ⭐ **NEW - HIGH VALUE**
- **Purpose:** Generate Software Bill of Materials
- **Install:** `curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh`
- **Usage:** `syft packages dir:. -o spdx-json > sbom.json`
- **Aura Application:** Document all dependencies for security tracking

**5. Trivy** ⭐ **NEW - HIGH VALUE**
- **Purpose:** Container and filesystem vulnerability scanner
- **Install:** Binary available at https://github.com/aquasecurity/trivy
- **Usage:**
  ```bash
  trivy fs .
  trivy image aura-node:latest
  ```
- **Aura Application:** Scan Docker images before deployment

**Add to CI/CD:**

```yaml
# .github/workflows/trivy-scan.yml
name: Trivy Security Scan
on: [push, pull_request]
jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          scan-ref: '.'
          format: 'sarif'
          output: 'trivy-results.sarif'
      - name: Upload results to GitHub Security
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: 'trivy-results.sarif'
```

### Fuzzing & Dynamic Analysis

**1. libFuzzer** (for C/C++/Rust)
- **Purpose:** Coverage-guided fuzzing
- **Aura Application:** Not currently applicable (no C/C++/Rust)

**2. AFL++** (for native code)
- **Purpose:** Advanced fuzzer
- **Aura Application:** Future use if native components added

**3. go-fuzz / Go built-in fuzzing** ⭐ **IMPLEMENTED - EXPAND**
- **Current Status:** Framework in CODE_QUALITY_FRAMEWORK.md
- **Enhancement:** Add fuzz tests for each module

**Example fuzz test for Bridge module:**

```go
// chain/x/bridge/keeper/fuzz_test.go
package keeper

import (
    "testing"
    "github.com/stretchr/testify/require"
)

func FuzzMerkleProofVerification(f *testing.F) {
    // Seed corpus
    f.Add([]byte("valid_proof"), []byte("valid_root"))

    f.Fuzz(func(t *testing.T, proof []byte, root []byte) {
        keeper, ctx := setupTest(t)

        // Should never panic
        defer func() {
            if r := recover(); r != nil {
                t.Errorf("Panic on input: %v", r)
            }
        }()

        // Verify proof (may be invalid but shouldn't crash)
        _, err := keeper.VerifyMerkleProof(ctx, proof, root)

        // Either succeeds or returns error (no panic)
        require.True(t, err == nil || err != nil)
    })
}

func FuzzTransferValidation(f *testing.F) {
    f.Add(uint64(1000), "cosmos1abc", "ethereum:0x123")

    f.Fuzz(func(t *testing.T, amount uint64, from string, to string) {
        keeper, ctx := setupTest(t)

        defer func() {
            if r := recover(); r != nil {
                t.Errorf("Panic on transfer validation: %v", r)
            }
        }()

        keeper.ValidateTransfer(ctx, amount, from, to)
    })
}
```

### Formatters & Linters

**1. clippy** (Rust)
- **Not applicable:** Aura is Go-based

**2. ESLint + security plugins** ⭐ **ADD FOR WALLET GUI**
- **Install:** `npm install --save-dev eslint eslint-plugin-security`
- **Config:**

```json
// .eslintrc.json
{
  "extends": [
    "eslint:recommended",
    "plugin:security/recommended"
  ],
  "plugins": ["security"],
  "rules": {
    "security/detect-object-injection": "error",
    "security/detect-non-literal-regexp": "warn",
    "security/detect-unsafe-regex": "error",
    "security/detect-buffer-noassert": "error",
    "security/detect-eval-with-expression": "error",
    "security/detect-no-csrf-before-method-override": "error"
  }
}
```

**3. ShellCheck** ⭐ **ADD TO CI**
- **Install:** Available in most package managers
- **Usage:** `shellcheck scripts/*.sh`
- **Add to Makefile:**

```makefile
.PHONY: lint-shell
lint-shell:
	@echo "Running ShellCheck on all scripts..."
	@find . -name "*.sh" -type f -exec shellcheck {} \;
```

---

## 2. MCP Server Integration

The toolkit recommends implementing custom MCP servers to expose audit tools to LLM agents. Here's the implementation plan:

### Priority 1: Critical MCP Servers

**1. CodeQL MCP Server**

**Implementation:**

```python
# scripts/mcp_servers/codeql_server.py
"""
MCP Server for CodeQL integration with Aura blockchain
"""
import subprocess
import json
from typing import Dict, Any, List

class CodeQLServer:
    def __init__(self, database_path: str = "./codeql-db"):
        self.database_path = database_path

    async def create_database(self, repo_path: str, language: str = "go") -> Dict[str, Any]:
        """Create CodeQL database for analysis"""
        cmd = [
            "codeql", "database", "create",
            self.database_path,
            f"--language={language}",
            f"--source-root={repo_path}"
        ]
        result = subprocess.run(cmd, capture_output=True, text=True)
        return {
            "status": "success" if result.returncode == 0 else "error",
            "output": result.stdout,
            "error": result.stderr
        }

    async def run_analysis(self, query_suite: str = "security-extended") -> Dict[str, Any]:
        """Run CodeQL analysis on database"""
        output_file = "codeql-results.sarif"
        cmd = [
            "codeql", "database", "analyze",
            self.database_path,
            f"{query_suite}",
            f"--format=sarif-latest",
            f"--output={output_file}"
        ]
        result = subprocess.run(cmd, capture_output=True, text=True)

        # Parse SARIF results
        with open(output_file, 'r') as f:
            results = json.load(f)

        return {
            "status": "success" if result.returncode == 0 else "error",
            "findings": self._parse_sarif(results),
            "output_file": output_file
        }

    def _parse_sarif(self, sarif: Dict) -> List[Dict]:
        """Parse SARIF format to simplified findings"""
        findings = []
        for run in sarif.get("runs", []):
            for result in run.get("results", []):
                findings.append({
                    "rule_id": result.get("ruleId"),
                    "level": result.get("level"),
                    "message": result["message"]["text"],
                    "locations": [
                        {
                            "file": loc["physicalLocation"]["artifactLocation"]["uri"],
                            "line": loc["physicalLocation"]["region"]["startLine"]
                        }
                        for loc in result.get("locations", [])
                    ]
                })
        return findings

# MCP server configuration
mcp_tools = {
    "codeql_create_database": CodeQLServer.create_database,
    "codeql_run_analysis": CodeQLServer.run_analysis,
    "codeql_list_alerts": lambda: CodeQLServer._parse_sarif(...)
}
```

**2. Semgrep MCP Server**

```python
# scripts/mcp_servers/semgrep_server.py
"""
MCP Server for Semgrep integration
"""
import subprocess
import json
from typing import Dict, List, Any

class SemgrepServer:
    async def scan(self, path: str = ".", ruleset: str = "auto") -> Dict[str, Any]:
        """Run Semgrep scan on specified path"""
        cmd = [
            "semgrep",
            "--config", ruleset,
            "--json",
            path
        ]
        result = subprocess.run(cmd, capture_output=True, text=True)

        if result.returncode == 0:
            findings = json.loads(result.stdout)
            return {
                "status": "success",
                "findings": self._parse_findings(findings),
                "summary": {
                    "total": len(findings.get("results", [])),
                    "by_severity": self._count_by_severity(findings)
                }
            }
        else:
            return {"status": "error", "error": result.stderr}

    async def scan_diff(self, path: str, base_ref: str, head_ref: str) -> Dict[str, Any]:
        """Scan only changed files between refs"""
        # Get changed files
        diff_cmd = ["git", "diff", "--name-only", f"{base_ref}..{head_ref}"]
        diff_result = subprocess.run(diff_cmd, capture_output=True, text=True)

        changed_files = diff_result.stdout.strip().split('\n')

        # Scan only changed files
        all_findings = []
        for file in changed_files:
            result = await self.scan(file)
            if result["status"] == "success":
                all_findings.extend(result["findings"])

        return {
            "status": "success",
            "findings": all_findings,
            "files_scanned": len(changed_files)
        }

    def _parse_findings(self, raw_findings: Dict) -> List[Dict]:
        """Parse Semgrep output to simplified format"""
        return [
            {
                "rule_id": finding["check_id"],
                "severity": finding["extra"]["severity"],
                "message": finding["extra"]["message"],
                "file": finding["path"],
                "line": finding["start"]["line"],
                "code": finding["extra"]["lines"]
            }
            for finding in raw_findings.get("results", [])
        ]

    def _count_by_severity(self, findings: Dict) -> Dict[str, int]:
        """Count findings by severity"""
        counts = {"ERROR": 0, "WARNING": 0, "INFO": 0}
        for finding in findings.get("results", []):
            severity = finding["extra"]["severity"]
            counts[severity] = counts.get(severity, 0) + 1
        return counts
```

**3. Repository MCP Server** (Already partially implemented via filesystem MCP)

**4. CI/Test Runner MCP Server**

```python
# scripts/mcp_servers/ci_server.py
"""
MCP Server for CI/CD integration
"""
import subprocess
from typing import Dict, Any

class CIServer:
    async def run_test_suite(self, repo_path: str, test_pattern: str = "./...") -> Dict[str, Any]:
        """Run Go test suite"""
        cmd = ["go", "test", "-v", "-race", "-coverprofile=coverage.out", test_pattern]
        result = subprocess.run(cmd, cwd=repo_path, capture_output=True, text=True)

        return {
            "status": "passed" if result.returncode == 0 else "failed",
            "output": result.stdout,
            "coverage": self._parse_coverage("coverage.out")
        }

    async def run_fuzzer(self, job_id: str, duration: str = "30s") -> Dict[str, Any]:
        """Run Go fuzzer for specified duration"""
        cmd = ["go", "test", "-fuzz=.", f"-fuzztime={duration}"]
        result = subprocess.run(cmd, capture_output=True, text=True)

        return {
            "status": "success" if result.returncode == 0 else "failed",
            "output": result.stdout,
            "crashes": self._parse_crashes(result.stdout)
        }

    def _parse_coverage(self, coverage_file: str) -> Dict[str, float]:
        """Parse coverage report"""
        cmd = ["go", "tool", "cover", "-func", coverage_file]
        result = subprocess.run(cmd, capture_output=True, text=True)

        # Extract total coverage percentage
        lines = result.stdout.strip().split('\n')
        if lines:
            last_line = lines[-1]
            if "total:" in last_line:
                coverage_str = last_line.split()[-1].rstrip('%')
                return {"total_coverage": float(coverage_str)}

        return {"total_coverage": 0.0}

    def _parse_crashes(self, output: str) -> int:
        """Count fuzzer crashes"""
        return output.count("--- FAIL:")
```

### MCP Server Deployment

**Install MCP SDK:**
```bash
pip install mcp
```

**Server Registration:**

```json
// .mcp/servers.json
{
  "servers": {
    "codeql": {
      "command": "python",
      "args": ["scripts/mcp_servers/codeql_server.py"],
      "env": {
        "CODEQL_HOME": "/usr/local/bin/codeql"
      }
    },
    "semgrep": {
      "command": "python",
      "args": ["scripts/mcp_servers/semgrep_server.py"]
    },
    "ci": {
      "command": "python",
      "args": ["scripts/mcp_servers/ci_server.py"]
    }
  }
}
```

---

## 3. Specialized Agent Workflows

Based on the toolkit's agent definitions, here are tailored workflows for Aura:

### Agent 1: Cosmos Module Security Auditor

**Role:** Audit Cosmos SDK modules for security vulnerabilities

**Targets:**
- All custom x/ modules (bridge, dex, governance, etc.)
- Module interactions and state transitions
- Keeper functions and message handlers

**Tools:**
- CodeQL (Go analysis)
- Semgrep (pattern matching)
- Go fuzzing (state transitions)
- gosec (security scanning)

**Workflow:**

1. **Inventory Phase:**
   ```bash
   # List all modules
   find chain/x -name "keeper.go" -type f
   ```

2. **Static Analysis:**
   ```bash
   # Run CodeQL
   codeql database create codeql-db --language=go
   codeql database analyze codeql-db security-extended --format=sarif-latest

   # Run Semgrep
   semgrep --config=p/golang chain/x/

   # Run gosec
   gosec -fmt=json -out=results.json ./chain/x/...
   ```

3. **Dynamic Testing:**
   ```bash
   # Run all tests with race detection
   go test -race -v ./chain/x/...

   # Run fuzz tests
   go test -fuzz=FuzzMerkleProof -fuzztime=10m ./chain/x/bridge/keeper
   ```

4. **Manual Review:**
   - Check for reentrancy vulnerabilities
   - Verify integer overflow protection
   - Review access control on all state changes
   - Validate input sanitization

5. **Report Generation:**
   - Aggregate findings by module
   - Assign severity (Critical/High/Medium/Low)
   - Provide code-level remediation
   - Generate test cases for fixes

### Agent 2: Node & Wallet Application Security

**Role:** Audit node software and wallet backends

**Targets:**
- RPC endpoints (gRPC, REST)
- P2P networking code
- Wallet key management
- API authentication

**Tools:**
- CodeQL (injection, auth issues)
- Semgrep (crypto patterns)
- go-fuzz (RPC fuzzing)
- Trivy (dependency scanning)

**Workflow:**

1. **RPC Security Audit:**
   ```bash
   # Find all RPC handlers
   grep -r "RegisterQueryServer\|RegisterMsgServer" chain/

   # Audit for injection vulnerabilities
   semgrep --config=p/sql-injection chain/
   ```

2. **Crypto Audit:**
   ```bash
   # Check for weak crypto
   semgrep --config=rules/crypto-weak.yml chain/

   # Verify key handling
   grep -r "PrivKey\|Mnemonic\|Seed" chain/ | semgrep --config=rules/secrets.yml
   ```

3. **Dependency Scan:**
   ```bash
   # Scan Go dependencies
   gosec -exclude=G104 ./...

   # Generate SBOM
   syft packages dir:./chain -o spdx-json > chain-sbom.json

   # Vulnerability scan
   trivy fs ./chain
   ```

4. **Fuzzing:**
   ```bash
   # Fuzz RPC endpoints
   go test -fuzz=FuzzQueryHandler -fuzztime=1h ./chain/x/*/client
   ```

### Agent 3: Frontend Wallet & GUI Security

**Role:** Review wallet and GUI security

**Targets:**
- Web wallet interfaces
- Browser extensions
- Mobile apps (if any)

**Tools:**
- ESLint + security plugins
- Semgrep (XSS, CSRF)
- npm audit

**Workflow:**

1. **Dependency Audit:**
   ```bash
   npm audit --audit-level=moderate
   npm audit fix
   ```

2. **Static Analysis:**
   ```bash
   eslint --config .eslintrc.json src/
   semgrep --config=p/xss src/
   ```

3. **Secrets Detection:**
   ```bash
   semgrep --config=p/secrets src/
   gitleaks detect
   ```

4. **UX Security Review:**
   - Verify clear transaction details
   - Check phishing protection
   - Review seed phrase handling
   - Validate signing flows

### Agent 4: DevOps & Infrastructure

**Role:** Harden CI/CD and infrastructure

**Targets:**
- GitHub Actions workflows
- Docker images
- Deployment scripts
- Kubernetes manifests

**Tools:**
- ShellCheck
- Trivy
- Syft
- actionlint

**Workflow:**

1. **CI/CD Audit:**
   ```bash
   # Lint GitHub Actions
   actionlint .github/workflows/*.yml

   # Check for secrets in workflows
   semgrep --config=p/ci .github/workflows/
   ```

2. **Container Security:**
   ```bash
   # Scan Docker images
   trivy image aura-node:latest

   # Generate SBOM
   syft packages docker:aura-node:latest
   ```

3. **Script Hardening:**
   ```bash
   # Lint all shell scripts
   find . -name "*.sh" -exec shellcheck {} \;
   ```

4. **Infrastructure as Code:**
   ```bash
   # Scan Terraform (if used)
   tfsec .

   # Scan Kubernetes manifests
   kubesec scan k8s/*.yaml
   ```

### Agent 5: Security Coordination

**Role:** Coordinate all agents and produce final report

**Workflow:**

1. **Component Inventory:**
   - List all modules, services, contracts
   - Map dependencies and interactions
   - Identify critical paths

2. **Task Delegation:**
   - Assign modules to Agent 1 (Cosmos Module Auditor)
   - Assign RPC/APIs to Agent 2 (AppSec)
   - Assign wallet to Agent 3 (Frontend)
   - Assign infra to Agent 4 (DevOps)

3. **Findings Aggregation:**
   - Collect reports from all agents
   - Deduplicate findings
   - Assign global severity

4. **Risk Assessment:**
   - Identify critical vulnerabilities
   - Map to OWASP Top 10
   - Assess exploitability

5. **Report Generation:**
   - Executive summary
   - Findings by component
   - Remediation recommendations
   - Timeline for fixes

---

## 4. Integration with Existing Aura Security

### Current Security Features → Audit Tools Mapping

| Security Feature | Audit Tool | Purpose |
|-----------------|------------|---------|
| Bridge Security | Slither, Mythril (for EVM contracts) | Verify bridge contract logic |
| DEX Security | go-fuzz | Fuzz AMM calculations |
| Cryptography Module | CodeQL, Semgrep | Detect weak crypto patterns |
| Network Security | go-fuzz | Fuzz gossip protocol |
| Monitoring | Prometheus queries | Validate metrics |
| Access Control | CodeQL | Find authz bypasses |
| Compliance Module | Semgrep | Verify PII handling |

### Recommended Audit Schedule

**Phase 1: Automated Scans (Weekly)**
- gosec on all PRs
- Semgrep on changed files
- npm audit on dependencies
- Trivy on Docker images

**Phase 2: Deep Analysis (Monthly)**
- Full CodeQL analysis
- Comprehensive fuzzing campaign
- SBOM generation
- Dependency vulnerability review

**Phase 3: Professional Audit (Pre-Mainnet)**
- Engage external auditors
- Full manual code review
- Penetration testing
- Economic audit

---

## 5. Quick Start Guide

### Day 1: Setup Tools

```bash
# Install Go tools
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Install Python tools
pip install semgrep slither-analyzer mythril

# Install system tools
# macOS
brew install shellcheck trivy syft

# Linux
apt-get install shellcheck
curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh
curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh
```

### Day 1: Run First Scans

```bash
# Static analysis
gosec ./chain/...
semgrep --config=auto ./chain

# Dependency scan
trivy fs ./chain
syft packages dir:./chain

# Container scan (if built)
docker build -t aura-node:audit .
trivy image aura-node:audit
```

### Week 1: Add to CI/CD

1. Copy workflow files from `.github/workflows/` examples above
2. Enable GitHub code scanning
3. Configure Dependabot
4. Set up scheduled scans

### Month 1: Deep Dive

1. Create CodeQL database for full analysis
2. Run 24-hour fuzz campaign on each module
3. Conduct manual code review of critical paths
4. Generate comprehensive audit report

---

## 6. Success Metrics

**Coverage:**
- ✅ 100% of Go code scanned with gosec
- ✅ 100% of Go code scanned with Semgrep
- ✅ 100% of dependencies scanned with Trivy
- ✅ 90%+ code coverage with fuzz testing

**Findings:**
- ✅ 0 Critical vulnerabilities
- ✅ <5 High vulnerabilities
- ✅ <20 Medium vulnerabilities

**Processes:**
- ✅ Automated scans on every PR
- ✅ Monthly deep analysis
- ✅ Quarterly external audit

---

## 7. Resources

**Official Documentation:**
- CodeQL: https://codeql.github.com/docs/
- Semgrep: https://semgrep.dev/docs/
- Trivy: https://aquasecurity.github.io/trivy/
- Syft: https://github.com/anchore/syft
- gosec: https://github.com/securego/gosec

**Aura-Specific:**
- Security Implementation: `SECURITY_IMPLEMENTATION_COMPLETE.md`
- Code Quality Framework: `docs/CODE_QUALITY_FRAMEWORK.md`
- Audit Framework: `docs/AUDIT_FRAMEWORK.md`

---

**Document Version:** 1.0
**Last Updated:** November 13, 2025
**Maintainer:** Aura Security Team
