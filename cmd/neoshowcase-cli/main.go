package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/traP-jp/neoshowcase-cli/internal/runner"
)

var version = "dev"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	os.Exit(runner.Execute(ctx, os.Args[1:], os.Stdout, os.Stderr, version))
}
