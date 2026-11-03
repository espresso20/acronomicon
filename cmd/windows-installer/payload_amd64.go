//go:build windows && amd64

package main

import _ "embed"

//go:embed payload/acronomicon-windows-amd64.exe
var embeddedBinary []byte

const archLabel = "x64 (amd64)"
