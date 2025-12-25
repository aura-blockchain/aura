# PHP Tooling Explanation

## Why PHP Files Exist

This is a **Go/Cosmos SDK blockchain project**, but includes PHP development tooling for potential future PHP-based wallet integrations or web components.

## PHP File Locations

### 1. Development Dependencies (vendor/)

The `vendor/` directory contains PHP development tools installed via Composer:

- **PHPUnit** (`phpunit/phpunit`) - Testing framework
- **PHPStan** (`phpstan/phpstan`) - Static analysis
- **PHP CodeSniffer** (`squizlabs/php_codesniffer`) - Code style checker
- **WordPress Coding Standards** (`wp-coding-standards/wpcs`) - WordPress style rules
- **Other dependencies** - Support libraries for the above tools

**Size:** ~51MB (excluded from git via `.gitignore`)

### 2. Placeholder Test

- **File:** `/home/hudson/blockchain-projects/aura/tests/PlaceholderTest.php`
- **Purpose:** Minimal PHPUnit test to ensure tooling works
- **Content:** Single passing test (`assertTrue(true)`)

### 3. Node.js Dependencies (node_modules/)

Contains one PHP file from the `flatted` npm package:
- `sdk/javascript/node_modules/flatted/php/flatted.php` - PHP port of JSON serialization library

**Size:** Part of node_modules (excluded from git)

## Configuration Files

- **composer.json** - Defines PHP dev dependencies and autoload paths
  - Autoload namespace: `Aequitas\Wallet\` → `wallet/php/` (directory doesn't exist yet)
  - Scripts: `phpunit`, `phpstan`, `phpcs`, `phpcbf`

- **package.json** - Node.js/Husky setup for pre-commit hooks
  - Script: `php-checks` runs `composer test && composer phpcs`

- **phpunit.xml.dist** - PHPUnit configuration (exists in root)

## Current Usage

**NONE.** The PHP tooling is currently unused because:

1. No PHP wallet implementation exists yet (`wallet/php/` directory missing)
2. Only one placeholder test exists
3. Main project is Go-based (Cosmos SDK blockchain)

## Future Plans

The PHP tooling is prepared for:

- **PHP Wallet SDK** - Potential PHP implementation for wallet operations
- **Web Integration** - PHP-based web components or admin panels
- **WordPress Plugin** - If WordPress integration is needed (WPCS installed)

## Should This Be Removed?

**Recommendation:** Keep it. The tooling is:
- Properly gitignored (vendor/, node_modules/)
- Minimal footprint in repo (only config files + 1 test)
- Ready for future use if PHP components are added
- Common pattern for blockchain projects (multi-language SDK support)

## Current SDK Status

The project has **complete SDK coverage** without PHP:

- **Go SDK** - Full implementation (`sdk/go/`)
- **JavaScript/TypeScript SDK** - Full implementation (`sdk/javascript/`)
- **Python SDK** - Full implementation (`sdk/python/`)

If PHP SDK development begins, the tooling and structure are already in place.

## Maintenance

To verify PHP tooling works:

```bash
# Install dependencies
composer install

# Run placeholder test
composer test

# Check code style (will fail if no PHP files found)
composer phpcs

# Static analysis (will pass with no files)
composer phpstan
```

**Last verified:** 2025-12-25 (tooling functional but unused)
