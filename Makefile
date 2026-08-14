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

lint-all:
	golangci-lint run -c .golangci.all.yaml $(DIR)

lint-all-fix:
	golangci-lint run -c .golangci.all.yaml --fix $(DIR)

fmt-all:
	golangci-lint fmt -c .golangci.all.yaml $(DIR)
