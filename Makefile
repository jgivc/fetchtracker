.DEFAULT_GOAL := build

BINARY_NAME ?= fetchtracker
CMD_DIR ?= cmd/fetchtracker

.PHONY: build

build:
	@echo "Building application..."
	@go build -o $(BINARY_NAME) $(CMD_DIR)/*.go
