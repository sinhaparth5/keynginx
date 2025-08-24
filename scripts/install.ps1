param(
    [string]$InstallPath = "$env:USERPROFILE\AppData\Local\Microsoft\WindowsApps"
)

$ErrorActionPreference = "Stop"

# Colors
function Write-ColorOutput($ForegroundColor) {
    $fc = $host.UI.RawUI.ForegroundColor
    $host.UI.RawUI.ForegroundColor = $ForegroundColor
    if ($args) {
        Write-Output $args
    }
    $host.UI.RawUI.ForegroundColor = $fc
}

function Write-Blue { Write-ColorOutput Blue $args }
function Write-Green { Write-ColorOutput Green $args }
function Write-Yellow { Write-ColorOutput Yellow $args }
function Write-Red { Write-ColorOutput Red $args }

Write-Blue "KeyNginx Windows Installer"
Write-Blue "========================="

# Detect architecture
$arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
Write-Yellow "Detected architecture: windows-$arch"

# Get latest release
Write-Blue "Fetching latest release..."
try {
    $latestRelease = Invoke-RestMethod -Uri "https://api.github.com/repos/sinhaparth5/keynginx/releases/latest"
    $version = $latestRelease.tag_name
    Write-Green "Latest version: $version"
}
catch {
    Write-Red "Failed to fetch latest release"
    exit 1
}

# Download URL
$downloadUrl = "https://github.com/sinhaparth5/keynginx/releases/download/$version/keynginx-$version-windows-$arch.zip"

# Create temp directory
$tempDir = [System.IO.Path]::GetTempPath() + [System.Guid]::NewGuid().ToString()
New-Item -ItemType Directory -Path $tempDir | Out-Null

try {
    Write-Blue "Downloading keynginx..."
    $zipPath = "$tempDir\keynginx.zip"
    Invoke-WebRequest -Uri $downloadUrl -OutFile $zipPath

    Write-Blue "Extracting..."
    Expand-Archive -Path $zipPath -DestinationPath $tempDir

    # Find the binary
    $binaryPath = Get-ChildItem -Path $tempDir -Name "keynginx.exe" -Recurse | Select-Object -First 1
    if (!$binaryPath) {
        Write-Red "Binary not found in archive"
        exit 1
    }

    $fullBinaryPath = "$tempDir\$binaryPath"

    # Create install directory if it doesn't exist
    if (!(Test-Path $InstallPath)) {
        New-Item -ItemType Directory -Path $InstallPath -Force | Out-Null
    }

    # Install
    Write-Blue "Installing to $InstallPath..."
    Copy-Item $fullBinaryPath "$InstallPath\keynginx.exe"

    # Verify installation
    try {
        $version = & "$InstallPath\keynginx.exe" version
        Write-Green "✅ KeyNginx installed successfully!"
        Write-Blue "Version: $version"
        Write-Green ""
        Write-Green "Get started:"
        Write-Green "  keynginx init --domain myapp.local"
        Write-Green "  keynginx --help"
    }
    catch {
        Write-Red "Installation may have failed - cannot run keynginx"
        Write-Yellow "Try adding $InstallPath to your PATH or run keynginx.exe directly"
    }
}
finally {
    # Cleanup
    Remove-Item -Path $tempDir -Recurse -Force
}