.PHONY: install test update-golden lint lint-all lint-all-fix fmt-all

# Default fallback directory if not specified on the command line
DIR ?= ./...

install:
	go install ./cmd/openapi2go

test:
	go test ./generator -run TestGenerate

update-golden:
	UPDATE_GOLDEN=1 go test ./generator -run TestGenerate

lint:
	golangci-lint run $(DIR)

lint-new:
	golangci-lint run --new $(DIR)

lint-fix:
	golangci-lint run --fix $(DIR)

fmt:
	golangci-lint fmt $(DIR)
