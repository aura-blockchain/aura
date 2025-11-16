# CI/CD Setup Verification Script for aura Project
# Run this to verify your CI/CD pipeline is properly configured

Write-Host "🔍 Verifying CI/CD Setup for aura Project..." -ForegroundColor Cyan
Write-Host ""

$errors = 0
$warnings = 0

# Check if workflow file exists
Write-Host "✓ Checking workflow file..." -ForegroundColor Yellow
if (Test-Path ".github/workflows/comprehensive-ci.yml") {
    Write-Host "  ✅ comprehensive-ci.yml exists" -ForegroundColor Green
} else {
    Write-Host "  ❌ comprehensive-ci.yml not found!" -ForegroundColor Red
    $errors++
}

# Check if this is a git repository
Write-Host "`n✓ Checking git repository..." -ForegroundColor Yellow
if (Test-Path ".git") {
    Write-Host "  ✅ Git repository detected" -ForegroundColor Green

    # Check if remote is configured
    $remote = git remote get-url origin 2>$null
    if ($remote) {
        Write-Host "  ✅ Remote origin: $remote" -ForegroundColor Green
    } else {
        Write-Host "  ⚠️  No remote origin configured" -ForegroundColor Yellow
        $warnings++
    }
} else {
    Write-Host "  ❌ Not a git repository!" -ForegroundColor Red
    $errors++
}

# Check Go installation
Write-Host "`n✓ Checking Go..." -ForegroundColor Yellow
try {
    $goVersion = go version 2>&1
    if ($goVersion -match "go1\.(2[0-9]|[3-9][0-9])") {
        Write-Host "  ✅ Go installed: $goVersion" -ForegroundColor Green
    } else {
        Write-Host "  ⚠️  Go version may not be optimal: $goVersion" -ForegroundColor Yellow
        Write-Host "     Recommended: Go 1.25+" -ForegroundColor Yellow
        $warnings++
    }
} catch {
    Write-Host "  ❌ Go not found!" -ForegroundColor Red
    $errors++
}

# Check PHP installation
Write-Host "`n✓ Checking PHP..." -ForegroundColor Yellow
try {
    $phpVersion = php -v 2>&1
    if ($phpVersion -match "PHP 8\.[0-9]") {
        Write-Host "  ✅ PHP installed: " + ($phpVersion -split "`n")[0] -ForegroundColor Green
    } else {
        Write-Host "  ⚠️  PHP version may not be optimal" -ForegroundColor Yellow
        Write-Host "     Recommended: PHP 8.4+" -ForegroundColor Yellow
        $warnings++
    }
} catch {
    Write-Host "  ⚠️  PHP not found (optional for Go-only builds)" -ForegroundColor Yellow
    $warnings++
}

# Check Go modules
Write-Host "`n✓ Checking Go modules..." -ForegroundColor Yellow
if (Test-Path "chain/go.mod") {
    Write-Host "  ✅ chain/go.mod exists" -ForegroundColor Green
} else {
    Write-Host "  ❌ chain/go.mod not found!" -ForegroundColor Red
    $errors++
}

if (Test-Path "proto/go.mod") {
    Write-Host "  ✅ proto/go.mod exists" -ForegroundColor Green
} else {
    Write-Host "  ⚠️  proto/go.mod not found" -ForegroundColor Yellow
    $warnings++
}

# Check Composer
Write-Host "`n✓ Checking Composer..." -ForegroundColor Yellow
if (Test-Path "composer.json") {
    Write-Host "  ✅ composer.json exists" -ForegroundColor Green
} else {
    Write-Host "  ⚠️  composer.json not found" -ForegroundColor Yellow
    $warnings++
}

if (Test-Path "tools/bin/composer.phar") {
    Write-Host "  ✅ Composer binary found" -ForegroundColor Green
} else {
    Write-Host "  ⚠️  Composer binary not found at tools/bin/composer.phar" -ForegroundColor Yellow
    $warnings++
}

# Check test structure
Write-Host "`n✓ Checking test structure..." -ForegroundColor Yellow
if (Test-Path "chain/tests" -or Test-Path "tests") {
    Write-Host "  ✅ Test directory exists" -ForegroundColor Green
} else {
    Write-Host "  ⚠️  No test directory found" -ForegroundColor Yellow
    $warnings++
}

# Check for golangci-lint config
Write-Host "`n✓ Checking linter configuration..." -ForegroundColor Yellow
if (Test-Path ".golangci.yml" -or Test-Path ".golangci.yaml") {
    Write-Host "  ✅ golangci-lint config found" -ForegroundColor Green
} else {
    Write-Host "  ℹ️  No golangci-lint config (will use defaults)" -ForegroundColor Cyan
}

# Check GitHub Actions status
Write-Host "`n✓ Checking GitHub Actions status..." -ForegroundColor Yellow
if ($remote -and $remote -match "github.com[:/](.+?)(?:\.git)?$") {
    $repo = $matches[1]
    Write-Host "  ℹ️  Repository: $repo" -ForegroundColor Cyan
    Write-Host "  🔗 GitHub Actions: https://github.com/$repo/actions" -ForegroundColor Cyan
    Write-Host "  ℹ️  Visit the URL above to check pipeline status" -ForegroundColor Cyan
} else {
    Write-Host "  ⚠️  Unable to determine GitHub repository" -ForegroundColor Yellow
    $warnings++
}

# Summary
Write-Host "`n" + ("=" * 70) -ForegroundColor Cyan
Write-Host "📊 VERIFICATION SUMMARY" -ForegroundColor Cyan
Write-Host ("=" * 70) -ForegroundColor Cyan

if ($errors -eq 0 -and $warnings -eq 0) {
    Write-Host "✅ Perfect! Your CI/CD setup looks great!" -ForegroundColor Green
} elseif ($errors -eq 0) {
    Write-Host "⚠️  Setup is functional but has $warnings warning(s)" -ForegroundColor Yellow
} else {
    Write-Host "❌ Found $errors error(s) and $warnings warning(s)" -ForegroundColor Red
}

Write-Host "`n📚 Next Steps:" -ForegroundColor Cyan
Write-Host "  1. Check GitHub Actions: https://github.com/$repo/actions" -ForegroundColor White
Write-Host "  2. Add Codecov token (optional): See CI_CD_SETUP_GUIDE.md" -ForegroundColor White
Write-Host "  3. Add SonarQube token (optional): See CI_CD_SETUP_GUIDE.md" -ForegroundColor White
Write-Host "  4. Review the comprehensive guide: CI_CD_SETUP_GUIDE.md" -ForegroundColor White

Write-Host "`n✨ Done!" -ForegroundColor Green
