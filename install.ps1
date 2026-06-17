# Premium Windows Installer for Nora Compiler & Runtime
# Pure 7-bit ASCII for maximum safety and compatibility across all PowerShell versions.

$ErrorActionPreference = "Stop"

# Clear host and print banner
Clear-Host

Write-Host " _  _  ___  ____   __   " -ForegroundColor Magenta
Write-Host "( \( )/ _ \(  _ \ / _\  " -ForegroundColor Magenta
Write-Host " )  (( (_) ))   //    \ " -ForegroundColor Magenta
Write-Host "(_)\_)\___/(__\_)\_/\_/ " -ForegroundColor Magenta
Write-Host ""
Write-Host "=== NORA LANGUAGE SYSTEM INSTALLER ===" -ForegroundColor Cyan
Write-Host ""

# Step 1: Detect Platform
Write-Host "[1/4] Detecting OS and Architecture..." -ForegroundColor Cyan
$OSName = "Windows"
$Arch = $env:PROCESSOR_ARCHITECTURE
Write-Host "  - Target OS: " -NoNewline
Write-Host $OSName -ForegroundColor Green
Write-Host "  - Architecture: " -NoNewline
Write-Host $Arch -ForegroundColor Green
Write-Host ""

# Step 2: Establish Install Paths
$NoraDir = Join-Path $HOME ".nora"
$BinDir = Join-Path $NoraDir "bin"
$StdDir = Join-Path $NoraDir "std"

Write-Host "[2/4] Setting up installation directories..." -ForegroundColor Cyan
Write-Host "  - Destination: " -NoNewline
Write-Host $NoraDir -ForegroundColor Yellow

if (-not (Test-Path $BinDir)) {
    New-Item -ItemType Directory -Force -Path $BinDir | Out-Null
}
if (-not (Test-Path $StdDir)) {
    New-Item -ItemType Directory -Force -Path $StdDir | Out-Null
}
Write-Host "  [OK] Directories successfully prepared." -ForegroundColor Green
Write-Host ""

# Step 3: Compile / Install compiler and runtime
Write-Host "[3/4] Compiling Nora Compiler and Runtime..." -ForegroundColor Cyan

# Verify Go installation
$goCheck = Get-Command go -ErrorAction SilentlyContinue
if (-not $goCheck) {
    Write-Host "  [ERROR] Go (golang) is not installed or not in PATH!" -ForegroundColor Red
    Write-Host "  Please install Go to compile Nora from source." -ForegroundColor Yellow
    exit 1
}

Write-Host "  - Compiling binary..."
$exePath = Join-Path $BinDir "nora.exe"
go build -o $exePath pkg/cmd/nora/main.go
if ($LASTEXITCODE -eq 0) {
    Write-Host "  [OK] Nora compiler compiled successfully." -ForegroundColor Green
} else {
    Write-Host "  [ERROR] Failed to compile Nora compiler!" -ForegroundColor Red
    exit 1
}

Write-Host "  - Copying standard library to $StdDir..." -ForegroundColor Yellow

if (Test-Path $StdDir) {
    Remove-Item -Recurse -Force (Join-Path $StdDir "*") -ErrorAction SilentlyContinue | Out-Null
}
Copy-Item -Path "std\*" -Destination $StdDir -Recurse -Force
Write-Host "  [OK] Standard library successfully copied." -ForegroundColor Green
Write-Host ""

# Step 4: Configure environment PATH
Write-Host "[4/4] Integrating with Environment PATH..." -ForegroundColor Cyan

# Fetch current User Path from registry
$userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
$noraPathExists = $userPath -split ";" | Where-Object { $_ -like '*\.nora\bin' }

if ($noraPathExists) {
    Write-Host "  - PATH is already configured for Nora." -ForegroundColor Yellow
} else {
    Write-Host "  - Appending Nora to Windows User PATH registry..." -ForegroundColor Yellow
    $newUserPath = $userPath
    if (-not $newUserPath.EndsWith(";")) {
        $newUserPath += ";"
    }
    $newUserPath += $BinDir
    [Environment]::SetEnvironmentVariable("PATH", $newUserPath, "User")
    Write-Host "  [OK] PATH successfully updated in registry." -ForegroundColor Green
}

# Update current active session path as well so they can run it immediately
if ($env:PATH -notlike '*\.nora\bin*') {
    $env:PATH += ";$BinDir"
}

Write-Host "================================================================" -ForegroundColor Green
Write-Host "         NORA LANGUAGE INSTALLED SUCCESSFULLY!" -ForegroundColor Green
Write-Host "================================================================" -ForegroundColor Green
Write-Host "Nora is now installed at: $NoraDir" -ForegroundColor Yellow
Write-Host ""
Write-Host "To start using Nora immediately in this terminal, you can run:"
Write-Host "  nora help" -ForegroundColor Cyan
Write-Host "  nora targets" -ForegroundColor Cyan
Write-Host ""
Write-Host "The PATH variable has been permanently configured for all new terminals."
Write-Host "================================================================" -ForegroundColor Green
