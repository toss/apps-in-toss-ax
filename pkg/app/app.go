package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/toss/apps-in-toss-ax/cmd"
)

func Run() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
		<-sigs
		os.Exit(137) // sigkill
	}()

	cfg := cmd.CommandConfig{
		Name:    "ax",
		Version: "0.1.0",
	}

	if err := cmd.NewCommand(cfg).ExecuteContext(ctx); err != nil {
		fmt.Printf("Err: %v\n", err)
		os.Exit(1)
	}
}
