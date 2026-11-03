# Acronomicon Windows PowerShell Uninstaller
$ErrorActionPreference = 'SilentlyContinue'

$InstallDir = Join-Path $env:LOCALAPPDATA 'Programs\Acronomicon'
Write-Host "Uninstalling Acronomicon from $InstallDir..." -ForegroundColor Yellow

$UserPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if ($UserPath) {
    $Parts = $UserPath -split ';' | Where-Object { $_ -and ($_.TrimEnd('\') -ine $InstallDir.TrimEnd('\')) }
    [Environment]::SetEnvironmentVariable('Path', ($Parts -join ';'), 'User')
}

Remove-Item -Path (Join-Path ([Environment]::GetFolderPath('Desktop')) 'Acronomicon.lnk') -Force
Remove-Item -Path (Join-Path ([Environment]::GetFolderPath('Programs')) 'Acronomicon.lnk') -Force
Remove-Item -Path 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\Acronomicon' -Recurse -Force
Remove-Item -Path $InstallDir -Recurse -Force

Write-Host "[OK] Acronomicon has been removed." -ForegroundColor Green
