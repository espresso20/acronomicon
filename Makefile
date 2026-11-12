.PHONY: all test build dist windows macos macos-app macos-pkg macos-zip windows-zip package clean

VERSION ?= $(shell git describe --tags --exact-match 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)
PKGVER := $(if $(filter dev,$(VERSION)),0.0.0,$(patsubst v%,%,$(VERSION)))
STAGE := dist/stage
# Substitutes the packaging placeholder version (1.0.0) in a staged copy.
SETVER = sed -i '' -e 's/1\.0\.0/$(PKGVER)/g'

all: test build

test:
	go test -v ./...

build:
	go build -ldflags="$(LDFLAGS)" -o acronomicon .

dist: windows macos package

windows:
	mkdir -p dist cmd/windows-installer/payload
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/acronomicon-windows-amd64.exe .
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o dist/acronomicon-windows-arm64.exe .
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

package: macos-app macos-pkg macos-zip windows-zip
	rm -rf $(STAGE)

# Acronomicon.app: launcher in MacOS/, binary in Resources/ (per AcronomiconLauncher).
macos-app: macos
	rm -rf dist/Acronomicon.app
	mkdir -p dist/Acronomicon.app/Contents/MacOS dist/Acronomicon.app/Contents/Resources
	cp packaging/macos/AcronomiconLauncher dist/Acronomicon.app/Contents/MacOS/AcronomiconLauncher
	cp dist/acronomicon-macos-universal dist/Acronomicon.app/Contents/Resources/acronomicon
	cp packaging/macos/Info.plist dist/Acronomicon.app/Contents/Info.plist
	$(SETVER) dist/Acronomicon.app/Contents/Info.plist
	chmod 755 dist/Acronomicon.app/Contents/MacOS/AcronomiconLauncher dist/Acronomicon.app/Contents/Resources/acronomicon
	codesign --force --deep -s - dist/Acronomicon.app

# Unsigned flat pkg: /usr/local/bin/acronomicon + /Applications/Acronomicon.app.
macos-pkg: macos-app
	rm -rf $(STAGE)/pkg dist/acronomicon-macos-universal.pkg
	mkdir -p $(STAGE)/pkg/root/usr/local/bin $(STAGE)/pkg/root/Applications $(STAGE)/pkg/resources $(STAGE)/pkg/scripts
	cp dist/acronomicon-macos-universal $(STAGE)/pkg/root/usr/local/bin/acronomicon
	chmod 755 $(STAGE)/pkg/root/usr/local/bin/acronomicon
	ditto dist/Acronomicon.app $(STAGE)/pkg/root/Applications/Acronomicon.app
	cp packaging/macos/scripts/postinstall $(STAGE)/pkg/scripts/postinstall
	chmod 755 $(STAGE)/pkg/scripts/postinstall
	cp packaging/macos/welcome.html packaging/macos/conclusion.html packaging/macos/distribution.xml $(STAGE)/pkg/resources/
	$(SETVER) $(STAGE)/pkg/resources/welcome.html $(STAGE)/pkg/resources/conclusion.html $(STAGE)/pkg/resources/distribution.xml
	pkgbuild --analyze --root $(STAGE)/pkg/root $(STAGE)/pkg/components.plist
	plutil -replace 0.BundleIsRelocatable -bool NO $(STAGE)/pkg/components.plist
	pkgbuild --root $(STAGE)/pkg/root --component-plist $(STAGE)/pkg/components.plist \
		--identifier io.acronomicon.cli --version $(PKGVER) --install-location / \
		--scripts $(STAGE)/pkg/scripts $(STAGE)/pkg/acronomicon-component.pkg
	productbuild --distribution $(STAGE)/pkg/resources/distribution.xml \
		--resources $(STAGE)/pkg/resources --package-path $(STAGE)/pkg \
		dist/acronomicon-macos-universal.pkg
	rm -rf $(STAGE)/pkg

macos-zip: macos-app
	rm -rf $(STAGE)/macos dist/acronomicon-macos-universal-package.zip
	mkdir -p $(STAGE)/macos/Acronomicon
	cp dist/acronomicon-macos-universal $(STAGE)/macos/Acronomicon/acronomicon
	ditto dist/Acronomicon.app $(STAGE)/macos/Acronomicon/Acronomicon.app
	cp "packaging/macos/Install Acronomicon.command" "packaging/macos/Uninstall Acronomicon.command" packaging/macos/README-MACOS.txt $(STAGE)/macos/Acronomicon/
	$(SETVER) $(STAGE)/macos/Acronomicon/README-MACOS.txt
	chmod 755 "$(STAGE)/macos/Acronomicon/Install Acronomicon.command" "$(STAGE)/macos/Acronomicon/Uninstall Acronomicon.command" $(STAGE)/macos/Acronomicon/acronomicon
	ditto -c -k --norsrc --keepParent $(STAGE)/macos/Acronomicon dist/acronomicon-macos-universal-package.zip
	rm -rf $(STAGE)/macos

# Scripts expect acronomicon.exe / acronomicon-setup.exe; rename inside the staging dir only.
windows-zip: windows
	rm -rf $(STAGE)/windows dist/acronomicon-windows-*-package.zip
	for arch in amd64 arm64; do \
		d=$(STAGE)/windows/$$arch/Acronomicon; \
		mkdir -p $$d; \
		cp dist/acronomicon-windows-$$arch.exe $$d/acronomicon.exe; \
		cp dist/acronomicon-windows-$$arch-setup.exe $$d/acronomicon-setup.exe; \
		cp "packaging/windows/Launch Acronomicon.bat" packaging/windows/README-WINDOWS.txt packaging/windows/install.ps1 packaging/windows/uninstall.ps1 $$d/; \
		$(SETVER) $$d/README-WINDOWS.txt; \
		(cd $(STAGE)/windows/$$arch && zip -X -r -q ../../../acronomicon-windows-$$arch-package.zip Acronomicon -x '*.DS_Store'); \
	done
	rm -rf $(STAGE)/windows

clean:
	rm -rf acronomicon dist cmd/windows-installer/payload/*.exe
