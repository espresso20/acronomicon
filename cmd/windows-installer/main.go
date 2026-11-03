//go:build windows

package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

const (
	appVersion = "1.0.0"
	appName    = "Acronomicon"
	exeName    = "acronomicon.exe"
	uninstName = "acronomicon-uninstall.exe"

	reset   = "\x1b[0m"
	bold    = "\x1b[1m"
	dim     = "\x1b[2m"
	cyan    = "\x1b[96m"
	yellow  = "\x1b[93m"
	green   = "\x1b[92m"
	magenta = "\x1b[95m"
	red     = "\x1b[91m"
	white   = "\x1b[97m"
)

func initConsole() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getStdHandle := kernel32.NewProc("GetStdHandle")
	getConsoleMode := kernel32.NewProc("GetConsoleMode")
	setConsoleMode := kernel32.NewProc("SetConsoleMode")
	setConsoleOutputCP := kernel32.NewProc("SetConsoleOutputCP")
	setConsoleCP := kernel32.NewProc("SetConsoleCP")
	setConsoleTitleW := kernel32.NewProc("SetConsoleTitleW")

	_, _, _ = setConsoleOutputCP.Call(uintptr(65001))
	_, _, _ = setConsoleCP.Call(uintptr(65001))

	if titlePtr, err := syscall.UTF16PtrFromString("Acronomicon Setup Wizard (" + archLabel + ")"); err == nil {
		_, _, _ = setConsoleTitleW.Call(uintptr(unsafe.Pointer(titlePtr)))
	}

	const enableVirtualTerminalProcessing = 0x0004
	stdOut := uintptr(^uint32(11) + 1)
	stdErr := uintptr(^uint32(12) + 1)

	for _, h := range []uintptr{stdOut, stdErr} {
		handle, _, _ := getStdHandle.Call(h)
		if handle == 0 || handle == ^uintptr(0) {
			continue
		}
		var mode uint32
		r, _, _ := getConsoleMode.Call(handle, uintptr(unsafe.Pointer(&mode)))
		if r != 0 {
			_, _, _ = setConsoleMode.Call(handle, uintptr(mode|enableVirtualTerminalProcessing))
		}
	}
}

func printBanner(subtitle string) {
	fmt.Println(cyan + bold + "╔═══════════════════════════════════════════════════════════════════════════════╗" + reset)
	fmt.Println(cyan + bold + "║    _____                                             .__                      ║" + reset)
	fmt.Println(cyan + bold + "║   /  _  \\   ____ _______  ____   ____   ____   _____ |__| ____  ____   ____   ║" + reset)
	fmt.Println(cyan + bold + "║  /  /_\\  \\_/ ___\\\\_  __ \\/  _ \\ /    \\ /  _ \\ /     \\|  |/ ___\\/  _ \\ /    \\  ║" + reset)
	fmt.Println(cyan + bold + "║ /    |    \\  \\___ |  | \\(  <_> )   |  (  <_> )  Y Y  \\  \\  \\__(  <_> )   |  \\ ║" + reset)
	fmt.Println(cyan + bold + "║ \\____|__  /\\___  >|__|   \\____/|___|  /\\____/|__|_|  /__|\\___  >____/|___|  / ║" + reset)
	fmt.Println(cyan + bold + "║         \\/     \\/                   \\/             \\/        \\/           \\/  ║" + reset)
	fmt.Printf("%s║         %s%-62s%s        ║%s\n", cyan+bold, yellow+bold, subtitle, cyan+bold, reset)
	fmt.Println(cyan + bold + "╚═══════════════════════════════════════════════════════════════════════════════╝" + reset)
	fmt.Println()
}

func defaultInstallDir() string {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		home, _ := os.UserHomeDir()
		localAppData = filepath.Join(home, "AppData", "Local")
	}
	return filepath.Join(localAppData, "Programs", appName)
}

func main() {
	initConsole()

	silent := false
	uninstall := false
	installDir := defaultInstallDir()

	// If invoked as acronomicon-uninstall.exe, default to uninstall mode
	exeBase := strings.ToLower(filepath.Base(os.Args[0]))
	if strings.Contains(exeBase, "uninstall") {
		uninstall = true
	}

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		arg := strings.ToLower(args[i])
		switch arg {
		case "--silent", "-s", "/s", "/quiet", "--quiet":
			silent = true
		case "--uninstall", "-u", "/uninstall", "uninstall":
			uninstall = true
		case "--dir", "-d":
			if i+1 < len(args) {
				installDir = args[i+1]
				i++
			}
		}
	}

	if uninstall {
		runUninstall(installDir, silent)
		return
	}

	runInstall(installDir, silent)
}

func runInstall(installDir string, silent bool) {
	printBanner(fmt.Sprintf("✦ WINDOWS INSTALLER v%s • %s ✦", appVersion, archLabel))

	fmt.Printf("  %sTarget Directory:%s %s\n\n", white+bold, reset, cyan+installDir+reset)

	// Step 1: Create directory
	fmt.Printf("  %s[1/4]%s Creating installation directory... ", cyan+bold, reset)
	if err := os.MkdirAll(installDir, 0755); err != nil {
		fmt.Printf("%sFAILED%s\n  Error: %v\n", red+bold, reset, err)
		pauseIfInteractive(silent)
		os.Exit(1)
	}
	fmt.Printf("%sOK%s\n", green+bold, reset)

	// Step 2: Write acronomicon.exe and uninstaller
	targetExe := filepath.Join(installDir, exeName)
	fmt.Printf("  %s[2/4]%s Installing %s (%d KB)... ", cyan+bold, reset, exeName, len(embeddedBinary)/1024)
	if err := os.WriteFile(targetExe, embeddedBinary, 0755); err != nil {
		fmt.Printf("%sFAILED%s\n  Error: %v\n", red+bold, reset, err)
		pauseIfInteractive(silent)
		os.Exit(1)
	}

	// Copy current installer binary as acronomicon-uninstall.exe
	if selfPath, err := os.Executable(); err == nil {
		if selfBytes, err := os.ReadFile(selfPath); err == nil {
			_ = os.WriteFile(filepath.Join(installDir, uninstName), selfBytes, 0755)
		}
	}
	fmt.Printf("%sOK%s\n", green+bold, reset)

	// Step 3: Add to Windows User PATH & register in Installed Apps
	fmt.Printf("  %s[3/4]%s Adding %s to Windows User PATH & Registry... ", cyan+bold, reset, appName)
	if err := configureWindowsEnvironment(installDir, targetExe); err != nil {
		fmt.Printf("%sWARNING%s (%v)\n", yellow+bold, reset, err)
	} else {
		fmt.Printf("%sOK%s\n", green+bold, reset)
	}

	// Step 4: Create Start Menu & Desktop Shortcuts
	fmt.Printf("  %s[4/4]%s Creating Start Menu & Desktop shortcuts... ", cyan+bold, reset)
	if err := createShortcuts(installDir, targetExe); err != nil {
		fmt.Printf("%sWARNING%s (%v)\n", yellow+bold, reset, err)
	} else {
		fmt.Printf("%sOK%s\n", green+bold, reset)
	}

	// Also add current installDir to current process PATH so immediate launch works seamlessly
	_ = os.Setenv("PATH", installDir+";"+os.Getenv("PATH"))

	fmt.Println()
	fmt.Println(green + bold + "  ╭──────────────────────────────────────────────────────────────────────────╮" + reset)
	fmt.Println(green + bold + "  │  ✔ Acronomicon has been installed on your Windows machine!              │" + reset)
	fmt.Println(green + bold + "  ╰──────────────────────────────────────────────────────────────────────────╯" + reset)
	fmt.Println()
	fmt.Printf("  %sHow to use Acronomicon in CMD, PowerShell, or Windows Terminal:%s\n", yellow+bold, reset)
	fmt.Printf("    • %sacronomicon%s              %s(Launch interactive full-color REPL)%s\n", cyan+bold, reset, dim, reset)
	fmt.Printf("    • %sacronomicon k8s%s          %s(Look up any cloud acronym card)%s\n", cyan+bold, reset, dim, reset)
	fmt.Printf("    • %sacronomicon search zero%s  %s(Search across 400+ cloud acronyms)%s\n", cyan+bold, reset, dim, reset)
	fmt.Printf("    • %sacronomicon quiz%s         %s(Play the interactive acronym quiz)%s\n", cyan+bold, reset, dim, reset)
	fmt.Println()

	if silent {
		return
	}

	fmt.Printf("  %sPress [Enter] to launch Acronomicon now, or type 'q' to exit setup:%s ", magenta+bold, reset)
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	trimmed := strings.TrimSpace(strings.ToLower(line))
	if trimmed != "q" && trimmed != "quit" && trimmed != "exit" && trimmed != "n" {
		cmd := exec.Command(targetExe)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		_ = cmd.Run()
	}
}

func runUninstall(installDir string, silent bool) {
	printBanner("✦ WINDOWS UNINSTALLER ✦")
	fmt.Printf("  Removing %s from %s...\n\n", appName, cyan+installDir+reset)

	psScript := fmt.Sprintf(`
$ErrorActionPreference = 'SilentlyContinue'
$installDir = '%s'

# 1. Remove from User PATH
$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if ($userPath) {
    $parts = $userPath -split ';' | Where-Object { $_ -and ($_.TrimEnd('\') -ine $installDir.TrimEnd('\')) }
    [Environment]::SetEnvironmentVariable('Path', ($parts -join ';'), 'User')
}

# 2. Remove shortcuts
$desktopLnk = Join-Path ([Environment]::GetFolderPath('Desktop')) 'Acronomicon.lnk'
$startMenuLnk = Join-Path ([Environment]::GetFolderPath('Programs')) 'Acronomicon.lnk'
Remove-Item -Path $desktopLnk -Force -ErrorAction SilentlyContinue
Remove-Item -Path $startMenuLnk -Force -ErrorAction SilentlyContinue

# 3. Remove Uninstall registry key
Remove-Item -Path 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\Acronomicon' -Recurse -Force -ErrorAction SilentlyContinue
`, strings.ReplaceAll(installDir, "'", "''"))

	_ = runPowerShell(psScript)
	_ = os.Remove(filepath.Join(installDir, exeName))

	fmt.Printf("  %s✔ Acronomicon has been uninstalled.%s\n", green+bold, reset)
	pauseIfInteractive(silent)
}

func configureWindowsEnvironment(installDir, targetExe string) error {
	uninstPath := filepath.Join(installDir, uninstName)
	psScript := fmt.Sprintf(`
$ErrorActionPreference = 'Stop'
$installDir = '%s'
$targetExe = '%s'
$uninstExe = '%s'

# Add to User PATH if not already present
$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if (-not $userPath) { $userPath = '' }
$parts = $userPath -split ';' | Where-Object { $_ -ne '' }
$exists = $false
foreach ($p in $parts) {
    if ($p.TrimEnd('\') -ieq $installDir.TrimEnd('\')) {
        $exists = $true
        break
    }
}
if (-not $exists) {
    $newPath = ($parts + $installDir) -join ';'
    [Environment]::SetEnvironmentVariable('Path', $newPath, 'User')
}

# Register in Windows Installed Apps (HKCU Uninstall)
$regPath = 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\Acronomicon'
New-Item -Path $regPath -Force | Out-Null
Set-ItemProperty -Path $regPath -Name 'DisplayName' -Value 'Acronomicon — Bible of Cloud Tech Acronyms'
Set-ItemProperty -Path $regPath -Name 'DisplayVersion' -Value '%s'
Set-ItemProperty -Path $regPath -Name 'Publisher' -Value 'Acronomicon Cloud CLI'
Set-ItemProperty -Path $regPath -Name 'DisplayIcon' -Value $targetExe
Set-ItemProperty -Path $regPath -Name 'InstallLocation' -Value $installDir
Set-ItemProperty -Path $regPath -Name 'UninstallString' -Value ('"' + $uninstExe + '" --uninstall')
Set-ItemProperty -Path $regPath -Name 'NoModify' -Value 1 -Type DWord
Set-ItemProperty -Path $regPath -Name 'NoRepair' -Value 1 -Type DWord
`,
		strings.ReplaceAll(installDir, "'", "''"),
		strings.ReplaceAll(targetExe, "'", "''"),
		strings.ReplaceAll(uninstPath, "'", "''"),
		appVersion,
	)
	return runPowerShell(psScript)
}

func createShortcuts(installDir, targetExe string) error {
	psScript := fmt.Sprintf(`
$ErrorActionPreference = 'Stop'
$targetExe = '%s'
$installDir = '%s'
$wsh = New-Object -ComObject WScript.Shell

$startMenuLnk = Join-Path ([Environment]::GetFolderPath('Programs')) 'Acronomicon.lnk'
$s1 = $wsh.CreateShortcut($startMenuLnk)
$s1.TargetPath = $targetExe
$s1.WorkingDirectory = $installDir
$s1.Description = 'The Acronomicon — Bible of Cloud Tech Acronyms (400+ Entries)'
$s1.Save()

$desktopLnk = Join-Path ([Environment]::GetFolderPath('Desktop')) 'Acronomicon.lnk'
$s2 = $wsh.CreateShortcut($desktopLnk)
$s2.TargetPath = $targetExe
$s2.WorkingDirectory = $installDir
$s2.Description = 'The Acronomicon — Bible of Cloud Tech Acronyms (400+ Entries)'
$s2.Save()
`,
		strings.ReplaceAll(targetExe, "'", "''"),
		strings.ReplaceAll(installDir, "'", "''"),
	)
	return runPowerShell(psScript)
}

func runPowerShell(script string) error {
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", script)
	return cmd.Run()
}

func pauseIfInteractive(silent bool) {
	if silent {
		return
	}
	fmt.Printf("\n  Press [Enter] to close...")
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
}
