package contexts

import (
	"context"
	"log"
	"os"
	"os/signal"
)

// WithSignals returns a context that is canceled when the process receives one of the given signals.
// When ctx is nil, a default Background context is used.
// When signals is empty, the context will be canceled by the default os.Interrupt signal.
func WithSignals(ctx context.Context, signals ...os.Signal) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}

	if len(signals) == 0 {
		signals = []os.Signal{os.Interrupt}
	}

	ctx, cancel := context.WithCancel(ctx)
	c := make(chan os.Signal, 1)
	signal.Notify(c, signals...)
	go func() {
		sig := <-c
		log.Printf("Received signal: %q. Terminating...", sig.String())
		cancel()
	}()

	return ctx, cancel
}
