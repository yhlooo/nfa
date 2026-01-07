package main

import (
	"context"
	"log"
	"syscall"

	"github.com/yhlooo/nfa/pkg/commands"
	"github.com/yhlooo/nfa/pkg/ctxutil"
)

func main() {
	ctx, cancel := ctxutil.Notify(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cmd := commands.NewCommand("nfa")
	if err := cmd.ExecuteContext(ctx); err != nil {
		log.Fatal(err)
	}
}
