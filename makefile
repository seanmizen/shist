# ---------- config ----------
APP     ?= shist
PKG     ?= ./src/main
BIN     ?= bin

OS      ?= $(shell go env GOOS)
ARCH    ?= $(shell go env GOARCH)
EXT     := $(if $(filter windows,$(OS)),.exe,)   # add .exe on Windows
OUT     := $(BIN)/$(APP)-$(OS)-$(ARCH)$(EXT)

# used by “make release”
MATRIX_OS   := linux darwin windows
MATRIX_ARCH := amd64 arm64

# ---------- targets ----------
.PHONY: all build clean install release

all: build

build:
	@echo "→ building $(OS)/$(ARCH)"
	@mkdir -p $(BIN)
	GOOS=$(OS) GOARCH=$(ARCH) go build -o $(OUT) $(PKG)

clean:
	@rm -rf $(BIN)

install: build
ifeq ($(OS),windows)
	@echo "Copy $(OUT) somewhere on your PATH (e.g. %USERPROFILE%\\bin) or install via Scoop/Chocolatey."
else
	@sudo cp $(OUT) /usr/local/bin/$(APP)
	@echo "installed $(APP) -> /usr/local/bin/$(APP)"
endif

release: clean
	@mkdir -p $(BIN)
	@for o in $(MATRIX_OS); do \
	  for a in $(MATRIX_ARCH); do \
	    ext=$$( [ "$$o" = "windows" ] && echo ".exe" ); \
	    echo "→ $$o/$$a"; \
	    GOOS=$$o GOARCH=$$a go build -o $(BIN)/$(APP)-$$o-$$a$$ext $(PKG); \
	  done; \
	done
