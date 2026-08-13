.PHONY: install update-golden

install:
	go install ./cmd/openapi2go

update-golden:
	UPDATE_GOLDEN=1 go test ./generator -run TestGenerate
