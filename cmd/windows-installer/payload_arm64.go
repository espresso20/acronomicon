//go:build windows && arm64

package main

import _ "embed"

//go:embed payload/acronomicon-windows-arm64.exe
var embeddedBinary []byte

const archLabel = "ARM64"
