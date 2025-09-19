.DEFAULT_GOAL := build

BINARY_NAME ?= fetchtracker
CMD_DIR ?= cmd/fetchtracker
TESTDATA ?= /tmp/testdata
TESTDATA_MARKER := $(TESTDATA)/.built
BUILD_DIR=releases
VERSION?=$(shell git describe --tags 2>/dev/null || echo "v0.0")
PLATFORMS=linux-amd64 linux-arm64 linux-386 darwin-amd64 darwin-arm64

MD := mkdir
DD := dd

GREEN=\033[0;32m
BLUE=\033[0;34m
NC=\033[0m

.PHONY: build
build:
	@echo "Building application..."
	@go build -o $(BINARY_NAME) $(CMD_DIR)/*.go

$(TESTDATA_MARKER):
	@echo "Building testdata..."
	$(MD) -p $(TESTDATA)
	$(MD) -p $(TESTDATA)/one
	$(DD) if=/dev/urandom of=$(TESTDATA)/one/file1.img bs=1M count=5 > /dev/null 2>&1
	$(DD) if=/dev/urandom of=$(TESTDATA)/one/file2.img bs=1M count=7 > /dev/null 2>&1
	$(DD) if=/dev/urandom of=$(TESTDATA)/one/file3.img bs=1M count=10 > /dev/null 2>&1
	$(MD) -p $(TESTDATA)/two
	$(DD) if=/dev/urandom of=$(TESTDATA)/two/file1.img bs=1M count=2 > /dev/null 2>&1
	$(DD) if=/dev/urandom of=$(TESTDATA)/two/file2.img bs=1M count=8 > /dev/null 2>&1
	@touch $@

testdata: $(TESTDATA_MARKER)

.PHONY: run
run: testdata
	@echo "Run docker..."
	@SHARE_PATH=$(TESTDATA) docker compose -f deploy/docker-compose.yml up --build

.PHONY: build-all
build-all: $(PLATFORMS)
	@echo "$(GREEN)All builds completed!$(NC)"

$(PLATFORMS):
	@echo "$(BLUE)Building for $@...$(NC)"
	@mkdir -p $(BUILD_DIR)

	$(eval OS := $(word 1,$(subst -, ,$@)))
	$(eval ARCH := $(word 2,$(subst -, ,$@)))

	GOOS=$(OS) GOARCH=$(ARCH) go build -o $(BUILD_DIR)/$(BINARY_NAME)-$(VERSION).$(OS)-$(ARCH) $(CMD_DIR)/*.go


.PHONY: linux
linux: linux-amd64 linux-arm64 linux-386

.PHONY: darwin
darwin: darwin-amd64 darwin-arm64

.PHONY: release
release: build-all
	@echo "$(GREEN)Creating release archives...$(NC)"
	@for file in $(BUILD_DIR)/*; do \
		tar -czf $$file.tar.gz $$file deploy/config.example.yml; \
	done

.PHONY: clean
clean:
	@echo "$(GREEN)Cleaning...$(NC)"
	rm -rf $(BUILD_DIR) $(BINARY_NAME)
