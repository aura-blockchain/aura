# Security Policy

## Supported Versions

We take security seriously and actively maintain the following versions of the AURA blockchain project:

| Version | Supported | End of Support |
| ------- | --------- | -------------- |
| Latest  | ✓         | N/A            |
| 0.1.x   | ✓         | 2026-12-31     |
| < 0.1   | ✗         | Unsupported    |

We recommend always using the latest version to ensure you have the most recent security patches and improvements.

## Reporting a Vulnerability

We appreciate the security research community's efforts in helping us maintain the security of the AURA blockchain. If you believe you have found a security vulnerability, please report it responsibly.

### How to Report

**Please DO NOT report security vulnerabilities through public GitHub issues.**

Instead, please report security vulnerabilities by:

1. **GitHub Security Advisory**: Use [GitHub's private vulnerability reporting feature](https://github.com/aequitas/aura/security/advisories/new)
2. **Email**: Send details to **security@aequitas-labs.io** with subject line starting with `[SECURITY]`

### What to Include

Please include as much of the following information as possible:

- Type of vulnerability (e.g., buffer overflow, SQL injection, authentication bypass, etc.)
- Full paths of source file(s) related to the manifestation of the vulnerability
- The location of the affected source code (tag/branch/commit or direct URL)
- Any special configuration required to reproduce the issue
- Step-by-step instructions to reproduce the vulnerability
- Proof-of-concept or exploit code (if possible)
- Impact of the vulnerability, including how an attacker might exploit it
- Your name and contact information (optional, for credit if desired)

### Response Timeline

- **Initial Response**: We aim to acknowledge receipt of your report within 48 hours
- **Triage & Assessment**: We will assess the vulnerability and communicate the severity within 5 business days
- **Fix Development**: Critical vulnerabilities will be prioritized for patching within 10 business days
- **Disclosure Timeline**: We follow a 90-day responsible disclosure timeline before public announcement

## Security Best Practices for Contributors

### Code Review

- All security-sensitive code must undergo rigorous peer review
- At least two approvals required for changes to cryptographic or authentication modules
- Security-focused code reviews should consider threat models and attack vectors

### Dependency Management

- Keep all dependencies up-to-date with security patches
- Use `composer audit` for PHP dependencies and review regularly
- Pin versions in production to prevent unexpected updates
- Only use verified and well-maintained packages

### Secrets & Credentials

- Never commit passwords, API keys, private keys, or sensitive credentials
- Use `.env.example` files to document required environment variables (without values)
- Rotate credentials immediately if accidentally committed
- Use GitHub's secret scanning and pre-commit hooks to prevent accidental commits

### Cryptography

- Use industry-standard cryptographic libraries (no custom implementations)
- Apply cryptographic functions correctly; consult security-critical documentation
- Use strong key derivation functions and proper hashing for sensitive data
- Document all cryptographic decisions and reasoning

### Testing

- Write security-focused unit tests covering edge cases and invalid inputs
- Include integration tests for authentication and authorization flows
- Perform penetration testing on security-critical features
- Maintain at least 80% code coverage for security modules

### Documentation

- Document security assumptions and threat models
- Clearly mark security-sensitive code with comments
- Maintain records of security-related changes in SECURITY_AUDIT_REPORT.md

## Security Incident Response

In the event of a confirmed security vulnerability:

1. **Immediate Actions**: Disable affected features if necessary to prevent exploitation
2. **Fix Development**: Develop and test patches with urgent priority
3. **Security Advisory**: Prepare a detailed security advisory with mitigation steps
4. **Notification**: Notify affected parties and provide upgrade paths
5. **Post-Incident**: Conduct a post-mortem to prevent similar issues

## Security Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [CWE/SANS Top 25](https://cwe.mitre.org/top25/)
- [NIST Cybersecurity Framework](https://www.nist.gov/cyberframework)
- [PHP Security](https://www.php.net/manual/en/security.php)
- [Go Security](https://golang.org/doc/effective_go#security)

## License

This security policy is part of the AURA blockchain project and is licensed under the same license as the project.
