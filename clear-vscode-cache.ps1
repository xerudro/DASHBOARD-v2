# Clear VS Code Copilot and general cache
Write-Host "Clearing VS Code cache..." -ForegroundColor Cyan

$paths = @(
    "$env:APPDATA\Code\Cache",
    "$env:APPDATA\Code\CachedData",
    "$env:APPDATA\Code\CachedExtensionVSIXs",
    "$env:APPDATA\Code\User\workspaceStorage",
    "$env:LOCALAPPDATA\Programs\Microsoft VS Code\resources\app\extensions\github.copilot"
)

foreach ($path in $paths) {
    if (Test-Path $path) {
        Write-Host "Clearing: $path" -ForegroundColor Yellow
        Remove-Item -Path $path -Recurse -Force -ErrorAction SilentlyContinue
    }
}

Write-Host "Cache cleared! Please restart VS Code." -ForegroundColor Green
