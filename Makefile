.PHONY: install test update-golden coverage coverage-html

include make/lint.mk

install:
	go install ./cmd/openapi2go

test:
	go test $(DIR)

update-golden:
	UPDATE_GOLDEN=1 go test $(DIR)

coverage:
	go test $(DIR) -coverprofile=coverage.out
	go tool cover -func=coverage.out

coverage-html:
	go test $(DIR) -coverprofile=coverage.out
	go tool cover -html=coverage.out
