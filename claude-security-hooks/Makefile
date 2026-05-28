.PHONY: build test check-skill install clean lint verify-static

BIN := bin/claude-security-hooks
HOOK_DIR := ../.claude/hooks/bin

build:
	CGO_ENABLED=0 go build -ldflags='-s -w' -o $(BIN) ./cmd/claude-security-hooks

# check-skill validates that SKILL.md is a well-formed skill file with operational
# content (not just documentation). Runs as part of `make test` to prevent the
# "doc-as-skill" regression where an agent replaces instructions with a README.
check-skill:
	go test ./internal/skillcheck/... -v -count=1

test: check-skill
	go test ./...

install: build
	mkdir -p $(HOOK_DIR)
	cp $(BIN) $(HOOK_DIR)/claude-security-hooks

clean:
	rm -rf bin

lint:
	go vet ./...

verify-static: build
	@file $(BIN) | grep -q "statically linked" || (echo "H1 FAILED: binary is not statically linked" && exit 1)
