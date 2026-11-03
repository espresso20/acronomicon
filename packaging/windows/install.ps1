# Acronomicon Windows PowerShell Installer
$ErrorActionPreference = 'Stop'

$InstallDir = Join-Path $env:LOCALAPPDATA 'Programs\Acronomicon'
$SourceExe  = Join-Path $PSScriptRoot 'acronomicon.exe'
$TargetExe  = Join-Path $InstallDir 'acronomicon.exe'

Write-Host ""
Write-Host "=== Installing Acronomicon to $InstallDir ===" -ForegroundColor Cyan

New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
Copy-Item -Path $SourceExe -Destination $TargetExe -Force

# Add to User PATH if not already present
$UserPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if (-not $UserPath) { $UserPath = '' }
$Parts = $UserPath -split ';' | Where-Object { $_ -ne '' }
if (-not ($Parts | Where-Object { $_.TrimEnd('\') -ieq $InstallDir.TrimEnd('\') })) {
    $NewPath = ($Parts + $InstallDir) -join ';'
    [Environment]::SetEnvironmentVariable('Path', $NewPath, 'User')
    $env:Path = "$InstallDir;$env:Path"
    Write-Host "[OK] Added $InstallDir to your Windows User PATH." -ForegroundColor Green
}

# Create Start Menu & Desktop Shortcuts
$WshShell = New-Object -ComObject WScript.Shell
foreach ($Folder in @([Environment]::GetFolderPath('Programs'), [Environment]::GetFolderPath('Desktop'))) {
    $ShortcutPath = Join-Path $Folder 'Acronomicon.lnk'
    $Shortcut = $WshShell.CreateShortcut($ShortcutPath)
    $Shortcut.TargetPath = $TargetExe
    $Shortcut.WorkingDirectory = $InstallDir
    $Shortcut.Description = 'The Acronomicon — Bible of Cloud Tech Acronyms'
    $Shortcut.Save()
}

Write-Host "[OK] Created Start Menu and Desktop shortcuts." -ForegroundColor Green
Write-Host ""
Write-Host "Installation complete! Open any PowerShell, CMD, or Windows Terminal and run:" -ForegroundColor Yellow
Write-Host "  acronomicon" -ForegroundColor Cyan
Write-Host "  acronomicon k8s" -ForegroundColor Cyan
Write-Host ""
