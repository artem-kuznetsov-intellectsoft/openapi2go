.PHONY: install test update-golden lint lint-all lint-all-fix fmt-all

install:
	go install ./cmd/openapi2go

test:
	go test ./generator -run TestGenerate

update-golden:
	UPDATE_GOLDEN=1 go test ./generator -run TestGenerate

lint:
	golangci-lint run ./...

lint-all:
	golangci-lint run -c .golangci.all.yaml ./...

lint-all-fix:
	golangci-lint run -c .golangci.all.yaml --fix ./...

fmt-all:
	golangci-lint fmt -c .golangci.all.yaml ./...
