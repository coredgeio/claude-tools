$ErrorActionPreference = "Stop"

$Repo = "coredgeio/claude-tools"
$Binary = "claudeline"
$InstallDir = Join-Path $env:USERPROFILE ".claude"
$Settings = Join-Path $InstallDir "settings.json"

# detect arch
$Arch = if ([System.Environment]::Is64BitOperatingSystem) {
    if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }
} else {
    Write-Error "Unsupported architecture"; exit 1
}

$Asset = "${Binary}-windows-${Arch}.exe"
$Url = "https://github.com/${Repo}/releases/download/statusline-latest/${Asset}"
$Dest = Join-Path $InstallDir "${Binary}.exe"

Write-Host "Downloading $Asset..."
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
Invoke-WebRequest -Uri $Url -OutFile $Dest

# patch settings.json
$Command = $Dest -replace "\\", "/"
if (Test-Path $Settings) {
    $json = Get-Content $Settings -Raw | ConvertFrom-Json
    if (-not $json.statusLine) {
        $json | Add-Member -MemberType NoteProperty -Name "statusLine" -Value @{}
    }
    $json.statusLine.command = $Command
    $json | ConvertTo-Json -Depth 10 | Set-Content $Settings
} else {
    @{ statusLine = @{ command = $Command } } | ConvertTo-Json -Depth 10 | Set-Content $Settings
}

Write-Host "Installed to $Dest"
Write-Host "Restart Claude Code to apply."
