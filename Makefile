.DEFAULT_GOAL := build

BINARY_NAME ?= fetchtracker
CMD_DIR ?= cmd/fetchtracker
TESTDATA ?= /tmp/testdata

MD := mkdir
DD := dd


.PHONY: build testdata

build:
	@echo "Building application..."
	@go build -o $(BINARY_NAME) $(CMD_DIR)/*.go

testdata:
	@echo "Building testdata..."
	$(MD) -p $(TESTDATA)
	$(MD) -p $(TESTDATA)/one
	$(DD) if=/dev/urandom of=$(TESTDATA)/one/file1.img bs=1M count=5 > /dev/null 2>&1
	$(DD) if=/dev/urandom of=$(TESTDATA)/one/file2.img bs=1M count=7 > /dev/null 2>&1
	$(DD) if=/dev/urandom of=$(TESTDATA)/one/file3.img bs=1M count=10 > /dev/null 2>&1

