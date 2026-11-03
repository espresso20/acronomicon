.PHONY: all test build dist windows macos clean

VERSION ?= $(shell git describe --tags --exact-match 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

all: test build

test:
	go test -v ./...

build:
	go build -ldflags="$(LDFLAGS)" -o acronomicon .

dist: windows macos

windows:
	mkdir -p dist cmd/windows-installer/payload
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/acronomicon-windows-amd64.exe .
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -ldflags="-s -w" -o dist/acronomicon-windows-arm64.exe .
	cp dist/acronomicon-windows-amd64.exe cmd/windows-installer/payload/
	cp dist/acronomicon-windows-arm64.exe cmd/windows-installer/payload/
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/acronomicon-windows-amd64-setup.exe ./cmd/windows-installer
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -ldflags="-s -w" -o dist/acronomicon-windows-arm64-setup.exe ./cmd/windows-installer

macos:
	mkdir -p dist
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o dist/acronomicon-darwin-arm64 .
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/acronomicon-darwin-amd64 .
	lipo -create -output dist/acronomicon-macos-universal dist/acronomicon-darwin-arm64 dist/acronomicon-darwin-amd64
	codesign --force -s - dist/acronomicon-macos-universal

clean:
	rm -rf acronomicon dist cmd/windows-installer/payload/*.exe
