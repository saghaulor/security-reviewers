# Top-level Makefile for security_reviewer.
# Provides integration-level build orchestration for all project binaries.
# Component build logic lives in each component's own Makefile.

.PHONY: help verify-opengrep-mcp build-opengrep-mcp

OPENGREP_BIN := .claude/hooks/bin/opengrep-mcp

## help: List available targets (default)
help:
	@echo "security_reviewer Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  verify-opengrep-mcp   Verify the opengrep-mcp binary is installed and statically linked"
	@echo "  build-opengrep-mcp    Build and install opengrep-mcp binary (see Plan 09-02)"

## verify-opengrep-mcp: Verify the opengrep-mcp binary exists, is executable, and is statically linked.
## This target is the TDD RED-phase gate: it must fail until build-opengrep-mcp has run.
verify-opengrep-mcp:
	@test -f $(OPENGREP_BIN) || (echo "FAIL: $(OPENGREP_BIN) does not exist. Run: make build-opengrep-mcp" && exit 1)
	@test -x $(OPENGREP_BIN) || (echo "FAIL: $(OPENGREP_BIN) is not executable" && exit 1)
	@readelf -d $(OPENGREP_BIN) 2>&1 | grep -q "no dynamic section" || ldd $(OPENGREP_BIN) 2>&1 | grep -q "not a dynamic executable" || (echo "FAIL: $(OPENGREP_BIN) is not statically linked" && exit 1)
	@echo "OK: opengrep-mcp binary verified"

## build-opengrep-mcp: Build and install the opengrep-mcp binary from the standalone repo.
build-opengrep-mcp:
	@mkdir -p .claude/hooks/bin
	$(MAKE) -C /home/saghaulor/code/opengrep-mcp build
	@cp /home/saghaulor/code/opengrep-mcp/bin/opengrep-mcp .claude/hooks/bin/opengrep-mcp
