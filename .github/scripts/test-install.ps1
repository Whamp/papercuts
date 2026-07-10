$ErrorActionPreference = 'Stop'
$PSNativeCommandUseErrorActionPreference = $true

$repositoryRoot = (Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path
$installer = Join-Path $repositoryRoot 'install.ps1'
$temporary = Join-Path ([IO.Path]::GetTempPath()) ("papercuts-install-test-" + [guid]::NewGuid())
$releaseDirectory = Join-Path $temporary 'release'
$fixtureDirectory = Join-Path $temporary 'fixture'
$installDirectory = Join-Path $temporary 'bin'
New-Item -ItemType Directory -Force -Path $releaseDirectory, $fixtureDirectory | Out-Null

try {
    $fixtureBinary = Join-Path $fixtureDirectory 'papercuts.exe'
    if ([string]::IsNullOrWhiteSpace($env:PAPERCUTS_TEST_FIXTURE_BINARY)) {
        $ldflags = '-s -w -X github.com/Whamp/papercuts/internal/buildinfo.version=v1.2.3 -X github.com/Whamp/papercuts/internal/buildinfo.commit=fixture -X github.com/Whamp/papercuts/internal/buildinfo.buildDate=1970-01-01T00:00:00Z'
        & go build -trimpath -ldflags $ldflags -o $fixtureBinary ./cmd/papercuts
        if ($LASTEXITCODE -ne 0) { throw 'failed to build fixture binary' }
    }
    else {
        Copy-Item -LiteralPath $env:PAPERCUTS_TEST_FIXTURE_BINARY -Destination $fixtureBinary
    }

    $fixtureHash = (Get-FileHash -Algorithm SHA256 $fixtureBinary).Hash
    $archiveName = 'papercuts_1.2.3_windows_amd64.zip'
    $archive = Join-Path $releaseDirectory $archiveName
    Compress-Archive -Path $fixtureBinary -DestinationPath $archive
    $archiveHash = (Get-FileHash -Algorithm SHA256 $archive).Hash.ToLowerInvariant()
    Set-Content -NoNewline -Path (Join-Path $releaseDirectory 'checksums.txt') -Value "$archiveHash  $archiveName`n"

    $requestLog = Join-Path $temporary 'requests.log'
    $env:PAPERCUTS_TEST_REQUEST_LOG = $requestLog
    function Invoke-PapercutsTestWebRequest {
        param(
            [Parameter(Mandatory = $true)] [string] $Uri,
            [Parameter(Mandatory = $true)] [string] $OutFile,
            [switch] $UseBasicParsing
        )
        $null = $UseBasicParsing
        Add-Content -Path $env:PAPERCUTS_TEST_REQUEST_LOG -Value $Uri
        Copy-Item -Force (Join-Path $releaseDirectory (Split-Path -Leaf $Uri)) $OutFile
    }
    Set-Alias -Name Invoke-WebRequest -Value Invoke-PapercutsTestWebRequest

    $env:PAPERCUTS_INSTALL_DIR = $installDirectory
    $env:PAPERCUTS_VERSION = $null
    $output = & $installer | Out-String
    $installed = Join-Path $installDirectory 'papercuts.exe'
    if ((Get-FileHash -Algorithm SHA256 $installed).Hash -ne $fixtureHash) {
        throw 'installer did not install the fixture binary'
    }
    if ($output -notlike '*SetEnvironmentVariable*') { throw 'installer did not print the PATH command' }

    Set-Content -Path $installed -Value 'old'
    Set-Content -Path $requestLog -Value $null
    $env:PAPERCUTS_VERSION = 'v1.2.3'
    & $installer | Out-Null
    if ((Get-Content $requestLog) -notcontains 'https://github.com/Whamp/papercuts/releases/download/v1.2.3/checksums.txt') {
        throw 'pinned install did not use the pinned release URL'
    }
    if ((Get-FileHash -Algorithm SHA256 $installed).Hash -ne $fixtureHash) {
        throw 'pinned install did not replace the old binary'
    }

    Copy-Item -Force $fixtureBinary $installed
    $preservedHash = (Get-FileHash -Algorithm SHA256 $installed).Hash
    Set-Content -NoNewline -Path (Join-Path $releaseDirectory 'checksums.txt') -Value "$('0' * 64)  $archiveName`n"
    $failed = $false
    try {
        $env:PAPERCUTS_VERSION = $null
        & $installer | Out-Null
    }
    catch {
        $failed = $true
    }
    if (-not $failed) { throw 'installer accepted an invalid archive checksum' }
    if ((Get-FileHash -Algorithm SHA256 $installed).Hash -ne $preservedHash) {
        throw 'checksum failure replaced the existing binary'
    }

    $failed = $false
    try {
        $env:PAPERCUTS_VERSION = 'v1/../../malicious'
        & $installer | Out-Null
    }
    catch {
        $failed = $true
    }
    if (-not $failed) { throw 'installer accepted a malformed pinned version' }
}
finally {
    Remove-Item Env:PAPERCUTS_INSTALL_DIR -ErrorAction SilentlyContinue
    Remove-Item Env:PAPERCUTS_VERSION -ErrorAction SilentlyContinue
    Remove-Item Env:PAPERCUTS_TEST_REQUEST_LOG -ErrorAction SilentlyContinue
    Remove-Item -Recurse -Force $temporary -ErrorAction SilentlyContinue
}
