package main

import (
	"fmt"
	"os"

	"github.com/saghaulor/claude-security-hooks/internal/hooks"
	"github.com/saghaulor/claude-security-hooks/internal/uuidgen"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: claude-security-hooks <preflight|validate|inject-context|uuid>")
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "preflight":
		err = hooks.Preflight(os.Stdin, os.Stdout, os.Stderr)
	case "validate":
		err = hooks.Validate(os.Stdin, os.Stdout, os.Stderr)
	case "inject-context":
		err = hooks.InjectContext(os.Stdin, os.Stdout, os.Stderr)
	case "uuid":
		fmt.Println(uuidgen.New())
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n", os.Args[1])
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
