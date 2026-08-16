.PHONY: install test update-golden coverage coverage-html

include make/lint.mk

install:
	go install ./cmd/openapi2go

test:
	go test $(DIR)

# Regeneration runs on its own, before anything else is compiled. `go test
# ./...` builds every test binary up front, so a single combined invocation
# would compile the fixture packages' hand-written exec tests against the
# pre-regeneration goldens that TestGenerate is concurrently rewriting — and
# silently validate stale code. The second and third steps then run against
# what was actually written.
update-golden:
	UPDATE_GOLDEN=1 go test ./generator -run TestGenerate
	go build ./...
	go test $(DIR)

coverage:
	go test $(DIR) -coverprofile=coverage.out
	go tool cover -func=coverage.out

coverage-html:
	go test $(DIR) -coverprofile=coverage.out
	go tool cover -html=coverage.out
