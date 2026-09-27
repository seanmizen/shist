# ---------- config ----------
APP     ?= shist
PKG     ?= .
BIN     ?= bin
DIST    ?= dist

# version comes from the git tag (v1.0.0 -> 1.0.0), falling back to a commit sha
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
VERSION := $(patsubst v%,%,$(VERSION))
LDFLAGS := -s -w -X main.version=$(VERSION)

OS      ?= $(shell go env GOOS)
ARCH    ?= $(shell go env GOARCH)
EXT     := $(if $(filter windows,$(OS)),.exe,)   # add .exe on Windows
OUT     := $(BIN)/$(APP)-$(OS)-$(ARCH)$(EXT)

# used by "make release"
MATRIX_OS   := linux darwin windows
MATRIX_ARCH := amd64 arm64

# ---------- targets ----------
.PHONY: all build clean install release version test bench fuzz

all: build

version:
	@echo $(VERSION)

build:
	@echo "→ building $(APP) $(VERSION) for $(OS)/$(ARCH)"
	@mkdir -p $(BIN)
	GOOS=$(OS) GOARCH=$(ARCH) go build -trimpath -ldflags '$(LDFLAGS)' -o $(OUT) $(PKG)

test:
	go vet ./...
	go test ./...

bench:
	go test -run '^$$' -bench . -benchmem ./...

# -fuzz only takes one target per run
FUZZTIME ?= 30s
fuzz:
	go test -run '^$$' -fuzz FuzzParse   -fuzztime $(FUZZTIME) ./internal/history
	go test -run '^$$' -fuzz FuzzCompile -fuzztime $(FUZZTIME) ./internal/format
	go test -run '^$$' -fuzz FuzzSGR     -fuzztime $(FUZZTIME) ./internal/format

clean:
	@rm -rf $(BIN) $(DIST)

install: build
ifeq ($(OS),windows)
	@echo "Copy $(OUT) somewhere on your PATH (e.g. %USERPROFILE%\\bin) or install via Scoop/Chocolatey."
else
	@sudo cp $(OUT) /usr/local/bin/$(APP)
	@echo "installed $(APP) -> /usr/local/bin/$(APP)"
endif

# cross-compile every target into $(DIST) as .tar.gz (unix) / .zip (windows),
# then write checksums.txt over the lot
release: clean
	@echo "→ packaging $(APP) $(VERSION)"
	@mkdir -p $(DIST)
	@for o in $(MATRIX_OS); do \
	  for a in $(MATRIX_ARCH); do \
	    ext=$$( [ "$$o" = "windows" ] && echo ".exe" ); \
	    name="$(APP)_$(VERSION)_$${o}_$${a}"; \
	    stage="$(DIST)/stage/$$name"; \
	    echo "  → $$o/$$a"; \
	    mkdir -p "$$stage" || exit 1; \
	    GOOS=$$o GOARCH=$$a go build -trimpath -ldflags '$(LDFLAGS)' \
	      -o "$$stage/$(APP)$$ext" $(PKG) || exit 1; \
	    cp readme.md LICENSE "$$stage/" || exit 1; \
	    if [ "$$o" = "windows" ]; then \
	      (cd "$$stage" && zip -q "$(CURDIR)/$(DIST)/$$name.zip" "$(APP)$$ext" readme.md LICENSE) || exit 1; \
	    else \
	      tar -czf "$(DIST)/$$name.tar.gz" -C "$$stage" "$(APP)$$ext" readme.md LICENSE || exit 1; \
	    fi; \
	  done; \
	done
	@rm -rf $(DIST)/stage
	@cd $(DIST) && if command -v sha256sum >/dev/null 2>&1; then \
	  sha256sum *.tar.gz *.zip > checksums.txt; \
	else \
	  shasum -a 256 *.tar.gz *.zip > checksums.txt; \
	fi
	@echo "→ $(DIST)/"
	@ls -1sh $(DIST)
