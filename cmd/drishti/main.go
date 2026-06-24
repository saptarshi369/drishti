// Command drishti is the Drishti daemon entrypoint.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/saptarshi369/drishti/internal/app"
)

// version is stamped at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := app.Run(ctx, version); err != nil {
		fmt.Fprintln(os.Stderr, "drishti:", err)
		os.Exit(1)
	}
}
