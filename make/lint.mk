# This module contains code quality targets

# Default fallback directory if not specified on the command line
DIR ?= ./...

.PHONY: lint lint-new lint-fix fmt fmt-dry

lint:
	golangci-lint run $(DIR)

lint-new:
	golangci-lint run --new $(DIR)

lint-fix:
	golangci-lint run --fix $(DIR)

fmt:
	golangci-lint fmt $(DIR)

fmt-dry:
	golangci-lint fmt -d $(DIR)