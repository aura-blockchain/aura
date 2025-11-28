# PowerShell script to update all PAW references to AURA in dashboards
# This script updates all three dashboards: validator, staking, and governance

Write-Host "Starting AURA Dashboard Integration..." -ForegroundColor Green
Write-Host "=======================================" -ForegroundColor Green

# Define the base paths
$basePath = "C:\Users\decri\GitClones\aura\dashboards\dashboards"
$dashboards = @("validator", "staking", "governance")

# Statistics
$totalFiles = 0
$totalChanges = 0

foreach ($dashboard in $dashboards) {
    Write-Host "`nProcessing $dashboard dashboard..." -ForegroundColor Cyan
    $dashboardPath = Join-Path $basePath $dashboard

    if (!(Test-Path $dashboardPath)) {
        Write-Host "Warning: $dashboardPath does not exist!" -ForegroundColor Yellow
        continue
    }

    # Get all relevant files
    $files = Get-ChildItem -Path $dashboardPath -Recurse -Include *.js,*.html,*.md,*.css,*.json -Exclude node_modules,coverage,dist,build

    foreach ($file in $files) {
        $content = Get-Content $file.FullName -Raw -ErrorAction SilentlyContinue
        if ($null -eq $content) { continue }

        $originalContent = $content
        $fileChanges = 0

        # Replace PAW with AURA (case variations)
        $content = $content -replace 'PAW Network', 'AURA Network'
        $content = $content -replace 'PAW Blockchain', 'AURA Blockchain'
        $content = $content -replace 'PAW Chain', 'AURA Chain'
        $content = $content -replace 'paw-1', 'aura-1'
        $content = $content -replace 'paw-testnet-1', 'aura-testnet-1'
        $content = $content -replace 'pawvaloper', 'auravaloper'
        $content = $content -replace 'pawvalcons', 'auravalcons'
        $content = $content -replace 'pawpub', 'aurapub'
        $content = $content -replace 'upaw', 'uaura'
        $content = $content -replace '"paw"', '"aura"'
        $content = $content -replace "'paw'", "'aura'"

        # Update package names
        $content = $content -replace 'paw-validator-dashboard', '@aura/validator-dashboard'
        $content = $content -replace 'paw-staking-dashboard', '@aura/staking-dashboard'
        $content = $content -replace 'paw-governance-dashboard', '@aura/governance-dashboard'

        # Update comments and documentation
        $content = $content -replace '// PAW', '// AURA'
        $content = $content -replace '/\* PAW', '/* AURA'
        $content = $content -replace '# PAW', '# AURA'
        $content = $content -replace 'for PAW', 'for AURA'
        $content = $content -replace 'on PAW', 'on AURA'
        $content = $content -replace 'with PAW', 'with AURA'

        # Count changes
        if ($content -ne $originalContent) {
            $fileChanges = ($originalContent.ToCharArray() | Where-Object { $_ } | Measure-Object).Count - ($content.ToCharArray() | Where-Object { $_ } | Measure-Object).Count
            $fileChanges = [Math]::Abs($fileChanges)

            # Write updated content
            Set-Content -Path $file.FullName -Value $content -NoNewline
            $totalFiles++
            $totalChanges++
            Write-Host "  Updated: $($file.Name)" -ForegroundColor Green
        }
    }
}

Write-Host "`n=======================================" -ForegroundColor Green
Write-Host "Migration Complete!" -ForegroundColor Green
Write-Host "Total files updated: $totalFiles" -ForegroundColor Cyan
Write-Host "Total changes made: $totalChanges" -ForegroundColor Cyan
Write-Host "`nNext steps:" -ForegroundColor Yellow
Write-Host "1. Review changes in each dashboard" -ForegroundColor Yellow
Write-Host "2. Run tests: npm test in each dashboard directory" -ForegroundColor Yellow
Write-Host "3. Build dashboards: npm run build" -ForegroundColor Yellow
Write-Host "4. Start dashboards: npm start" -ForegroundColor Yellow
