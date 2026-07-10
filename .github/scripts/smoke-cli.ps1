param(
    [Parameter(Mandatory = $true)]
    [string]$Binary
)

$ErrorActionPreference = 'Stop'
$PSNativeCommandUseErrorActionPreference = $false
$binaryPath = (Resolve-Path $Binary).Path
$root = Join-Path $env:RUNNER_TEMP "papercuts-cli-smoke-$([guid]::NewGuid())"
New-Item -ItemType Directory -Path $root | Out-Null

function Assert-ExitCode {
    param([int]$Want, [string]$Operation)
    if ($LASTEXITCODE -ne $Want) {
        throw "$Operation exit code was $LASTEXITCODE; want $Want"
    }
}

try {
    $project = Join-Path $root 'project'
    New-Item -ItemType Directory -Path $project | Out-Null
    Push-Location $project
    try {
        $recovery = & $binaryPath capture --severity low 'recovery smoke' 2>&1
        Assert-ExitCode 1 'capture before project initialization'
        if (($recovery -join "`n") -notlike '*run `papercuts init --project`*') {
            throw "project recovery command missing: $recovery"
        }
        & $binaryPath init --no-agents
        Assert-ExitCode 0 'project init'
        & $binaryPath capture --severity low 'project smoke'
        Assert-ExitCode 0 'project capture'
        if (-not (Select-String -Path PAPERCUTS.md -SimpleMatch '> project smoke' -Quiet)) {
            throw 'project capture missing from PAPERCUTS.md'
        }
    }
    finally {
        Pop-Location
    }

    $globalLog = Join-Path $root 'global\PAPERCUTS.md'
    $globalRecovery = & $binaryPath capture --global --global-path $globalLog --severity low 'global recovery smoke' 2>&1
    Assert-ExitCode 1 'capture before global initialization'
    if (($globalRecovery -join "`n") -notlike '*papercuts init --global --global-path*') {
        throw "global recovery command missing: $globalRecovery"
    }
    & $binaryPath init --global --global-path $globalLog --no-agents
    Assert-ExitCode 0 'global init'
    "global line one`nglobal line two" | & $binaryPath capture --global --global-path $globalLog --severity medium --stdin
    Assert-ExitCode 0 'global stdin capture'
    foreach ($line in @('> global line one', '> global line two')) {
        if (-not (Select-String -Path $globalLog -SimpleMatch $line -Quiet)) {
            throw "global capture line missing: $line"
        }
    }

    $agentsProject = Join-Path $root 'agents'
    New-Item -ItemType Directory -Path $agentsProject | Out-Null
    Push-Location $agentsProject
    try {
        & $binaryPath init --agents
        Assert-ExitCode 0 'agent guidance init'
        foreach ($marker in @('<!-- papercuts:begin -->', '<!-- papercuts:end -->')) {
            if (-not (Select-String -Path AGENTS.md -SimpleMatch $marker -Quiet)) {
                throw "AGENTS.md marker missing: $marker"
            }
        }
    }
    finally {
        Pop-Location
    }
}
finally {
    Remove-Item -Recurse -Force $root -ErrorAction SilentlyContinue
}
