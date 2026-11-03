//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

func initTerminal() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getStdHandle := kernel32.NewProc("GetStdHandle")
	getConsoleMode := kernel32.NewProc("GetConsoleMode")
	setConsoleMode := kernel32.NewProc("SetConsoleMode")
	setConsoleOutputCP := kernel32.NewProc("SetConsoleOutputCP")
	setConsoleCP := kernel32.NewProc("SetConsoleCP")
	setConsoleTitleW := kernel32.NewProc("SetConsoleTitleW")

	// Set UTF-8 codepage (65001) so box-drawing characters render cleanly in Windows CMD & PowerShell
	_, _, _ = setConsoleOutputCP.Call(uintptr(65001))
	_, _, _ = setConsoleCP.Call(uintptr(65001))

	if titlePtr, err := syscall.UTF16PtrFromString("The Acronomicon — Bible of Cloud Tech Acronyms"); err == nil {
		_, _, _ = setConsoleTitleW.Call(uintptr(unsafe.Pointer(titlePtr)))
	}

	const enableVirtualTerminalProcessing = 0x0004
	stdOut := uintptr(^uint32(11) + 1) // STD_OUTPUT_HANDLE (-11)
	stdErr := uintptr(^uint32(12) + 1) // STD_ERROR_HANDLE (-12)

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
