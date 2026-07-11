[CmdletBinding()]
param(
    [Alias('h')]
    [switch] $Help
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version 2.0

if ($Help) {
    @'
Usage: install.ps1 [-Help]

Install the latest stable Papercuts release.

Environment:
  PAPERCUTS_VERSION      Release tag to install (default: latest)
  PAPERCUTS_INSTALL_DIR  Installation directory (default: ~/bin)
'@
    exit 0
}

[Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12

$repositoryUrl = 'https://github.com/Whamp/papercuts'
$installDirectory = $env:PAPERCUTS_INSTALL_DIR
if ([string]::IsNullOrWhiteSpace($installDirectory)) {
    $installDirectory = Join-Path $HOME 'bin'
}

$version = $env:PAPERCUTS_VERSION
if ([string]::IsNullOrWhiteSpace($version)) {
    $version = 'latest'
}
if ($version -eq 'latest') {
    $releaseUrl = "$repositoryUrl/releases/latest/download"
}
elseif ($version -match '^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-rc\.[1-9][0-9]*)?$') {
    $releaseUrl = "$repositoryUrl/releases/download/$version"
}
else {
    throw 'papercuts: PAPERCUTS_VERSION must be latest or a valid release tag'
}

$architecture = [Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString()
switch ($architecture) {
    'X64' { $arch = 'amd64' }
    'Arm64' { $arch = 'arm64' }
    default { throw "papercuts: unsupported architecture: $architecture" }
}

$temporary = Join-Path ([IO.Path]::GetTempPath()) ("papercuts-install-" + [guid]::NewGuid())
New-Item -ItemType Directory -Path $temporary | Out-Null
try {
    $checksums = Join-Path $temporary 'checksums.txt'
    Invoke-WebRequest -UseBasicParsing -Uri "$releaseUrl/checksums.txt" -OutFile $checksums

    $suffix = "_windows_$arch.zip"
    $archiveCandidates = @()
    foreach ($line in Get-Content $checksums) {
        if ($line -match '^([0-9a-fA-F]{64})\s{2}(papercuts_[^/\\\s]+_windows_(amd64|arm64)\.zip)$' -and $Matches[2].EndsWith($suffix)) {
            $archiveCandidates += [PSCustomObject]@{
                Hash = $Matches[1].ToLowerInvariant()
                Name = $Matches[2]
            }
        }
    }
    if ($archiveCandidates.Count -ne 1) {
        throw "papercuts: release does not contain exactly one archive for windows/$arch"
    }

    $expectedHash = $archiveCandidates[0].Hash
    $archiveName = $archiveCandidates[0].Name
    $archive = Join-Path $temporary $archiveName
    Invoke-WebRequest -UseBasicParsing -Uri "$releaseUrl/$archiveName" -OutFile $archive
    $actualHash = (Get-FileHash -Algorithm SHA256 $archive).Hash.ToLowerInvariant()
    if ($actualHash -ne $expectedHash) {
        throw "papercuts: checksum verification failed for $archiveName"
    }

    $extracted = Join-Path $temporary 'extracted'
    Expand-Archive -Path $archive -DestinationPath $extracted
    $source = Join-Path $extracted 'papercuts.exe'
    if (-not (Test-Path -LiteralPath $source -PathType Leaf)) {
        throw 'papercuts: release archive does not contain papercuts.exe'
    }

    New-Item -ItemType Directory -Force -Path $installDirectory | Out-Null
    $destination = Join-Path $installDirectory 'papercuts.exe'
    $staged = Join-Path $installDirectory ('.papercuts-install-' + [guid]::NewGuid() + '.exe')
    Copy-Item -LiteralPath $source -Destination $staged
    try {
        Move-Item -LiteralPath $staged -Destination $destination -Force
    }
    finally {
        Remove-Item -LiteralPath $staged -Force -ErrorAction SilentlyContinue
    }

    Write-Output "Installed papercuts to $destination"
    $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    $pathContainsInstallDirectory = $false
    foreach ($entry in ($userPath -split ';')) {
        if ($entry.TrimEnd('\') -eq $installDirectory.TrimEnd('\')) {
            $pathContainsInstallDirectory = $true
            break
        }
    }
    if (-not $pathContainsInstallDirectory) {
        $quotedInstallDirectory = $installDirectory.Replace("'", "''")
        Write-Output 'Add papercuts to your user PATH, then open a new terminal:'
        Write-Output "  [Environment]::SetEnvironmentVariable('Path', '$quotedInstallDirectory;' + [Environment]::GetEnvironmentVariable('Path', 'User'), 'User')"
    }
}
finally {
    Remove-Item -LiteralPath $temporary -Recurse -Force -ErrorAction SilentlyContinue
}
