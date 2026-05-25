.PHONY: build test install clean lint verify-static

BIN := bin/claude-security-hooks
HOOK_DIR := ../.claude/hooks/bin

build:
	CGO_ENABLED=0 go build -ldflags='-s -w' -o $(BIN) ./cmd/claude-security-hooks

test:
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
