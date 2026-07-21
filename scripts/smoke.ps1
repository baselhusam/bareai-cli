$ErrorActionPreference = "Stop"

$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$Bin = Join-Path $Root "bareai.exe"

if (-not (Test-Path $Bin)) {
    Write-Host "building bareai..."
    Push-Location $Root
    go build -o bareai.exe ./cmd/bareai
    Pop-Location
}

function Run-Bareai {
    param([string[]]$Args)
    Write-Host "+ bareai $($Args -join ' ')"
    $env:NO_COLOR = "1"
    & $Bin @Args
    if ($LASTEXITCODE -ne 0) {
        throw "bareai $($Args -join ' ') failed with exit $LASTEXITCODE"
    }
}

Run-Bareai @("status", "--json") | Out-Null
Run-Bareai @("gpu", "--json") | Out-Null
Run-Bareai @("docker", "--json") | Out-Null
Run-Bareai @("llm", "--json") | Out-Null
Run-Bareai @("inspect", "--json") | Out-Null
Run-Bareai @("probe", "--endpoint", "http://127.0.0.1:59999", "--runtime", "ollama", "--json") | Out-Null

Write-Host "smoke: all commands exited 0"
