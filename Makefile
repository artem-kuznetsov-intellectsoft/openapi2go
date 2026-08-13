.PHONY: install test

install:
	go install ./cmd/openapi2go

test:
	UPDATE_GOLDEN=1 go test ./generator -run TestGenerate
