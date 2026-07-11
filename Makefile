.PHONY: up stop force-up frontend wails wails-linux wails-windows wails-darwin install-linux package-linux k8s-start k8s-stop

up:
	docker compose up -d

stop:
	docker compose down

force-up:
	docker compose up -d --build --force-recreate

VERSION := $(shell git describe --tags --abbrev=0 --match 'v*' 2>/dev/null || echo v0.0.0)
BIN := jobtracker-$(VERSION)_dev
GOARCH := $(shell go env GOARCH)
VERSION_LDFLAG := -X github.com/tecnologer/jobtracker/handler.buildVersion=$(VERSION)

frontend:
	cd web && npm run build

wails: wails-linux wails-windows wails-darwin

# CGO against webkit2gtk: builds for the host arch only (release CI builds amd64 + arm64)
wails-linux: frontend
ifeq ($(shell uname),Linux)
	go build -tags desktop,production,webkit2_41 -ldflags "-s $(VERSION_LDFLAG)" -o $(BIN)_linux_$(GOARCH) ./cmd/desktop
else
	@echo "wails-linux: skipped — needs a Linux host (CGO against webkit2gtk)"
endif

# Wayland has no per-window icon protocol: the taskbar/dock icon comes from a
# .desktop entry whose name matches the app_id (JobTracker), so the binary must
# be installed alongside a desktop file + hicolor icon for the icon to show.
install-linux: wails-linux
	install -Dm755 $(BIN)_linux_$(GOARCH) $(HOME)/.local/bin/jobtracker-desktop
	install -Dm644 resources/images/icons/icon-512.png $(HOME)/.local/share/icons/hicolor/512x512/apps/jobtracker.png
	mkdir -p $(HOME)/.local/share/applications
	sed 's|^Exec=.*|Exec=$(HOME)/.local/bin/jobtracker-desktop|' resources/linux/JobTracker.desktop > $(HOME)/.local/share/applications/JobTracker.desktop
	-update-desktop-database $(HOME)/.local/share/applications 2>/dev/null
	-gtk-update-icon-cache -q -t $(HOME)/.local/share/icons/hicolor 2>/dev/null

# builds the release tar.gz (binary, icon, .desktop, install/uninstall scripts, README)
package-linux: wails-linux
ifeq ($(shell uname),Linux)
	bash scripts/package-linux.sh $(BIN)_linux_$(GOARCH) $(GOARCH)
else
	@echo "package-linux: skipped — needs a Linux host"
endif

# wails v2 on Windows is pure Go (webview2), so both arches cross-compile from Linux
wails-windows: frontend
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -tags desktop,production -ldflags "-s -H windowsgui $(VERSION_LDFLAG)" -o $(BIN)_windows_amd64.exe ./cmd/desktop
	GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -tags desktop,production -ldflags "-s -H windowsgui $(VERSION_LDFLAG)" -o $(BIN)_windows_arm64.exe ./cmd/desktop

# needs a Mac host: CGO against Cocoa; cannot cross-compile from Linux
wails-darwin: frontend
ifeq ($(shell uname),Darwin)
	CGO_LDFLAGS="-framework UniformTypeIdentifiers" go build -tags desktop,production -ldflags "-s $(VERSION_LDFLAG)" -o $(BIN)_darwin_$(GOARCH) ./cmd/desktop
	bash scripts/package-darwin.sh $(BIN)_darwin_$(GOARCH) $(VERSION)
else
	@echo "wails-darwin: skipped — needs a Mac host (Cocoa via CGO; release CI builds it)"
endif

k8s-start:
	docker build -t jobtracker:latest .
	minikube image load jobtracker:latest
	kubectl apply -f k8s/jobtracker.yaml

k8s-stop:
	kubectl delete -f k8s/jobtracker.yaml
