// Command majucau-worker runs as the registered Windows service by default.
// The explicit --console mode is reserved for local smoke tests.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	console := flag.Bool("console", false, "run in foreground for local smoke tests; do not use as the installed service")
	flag.Parse()
	if !*console {
		if handled, err := tryRunAsService("MajucauWorker", runWorker); handled {
			if err != nil {
				fmt.Fprintln(os.Stderr, "worker service failed")
			}
			return
		}
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := runWorker(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "worker console failed:", err)
		os.Exit(1)
	}
}
