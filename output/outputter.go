// Package output implements data outputter interfaces.
package output

import (
	"context"
	"log/slog"

	"github.com/northeye/chissoku/types"
)

// Outputter the outputter interface
type Outputter interface {
	// Name returns unique name of the outputter.
	// Lower cased struct name is recommended.
	Name() string

	// Intialize the outputter and run the receiver loop as own goroutine.
	// When it returns non-nil error the outputter will be disabled.
	// The context can be receive event of parent's termination through the `ctx.Done()`.
	Initialize(context.Context) error

	// Output the data.
	// This method must be non-blocking and light-weight.
	Output(*types.Data)

	// Close cleanup the outputter.
	// It's automatically invoked upon the parent's termination.
	// If there is a need to call it from the child side, it should be implemented as atomically and maintain idempotence.
	// So, `sync.OnceFunc` is useful.
	Close()
}

// contextKeyDeactivateOutputterChannel context value key for DeactivateOutputterChannel
type contextKeyDeactivateOutputterChannel struct{}

// ContextWithDeactivateChannel new context with deactivate channel
func ContextWithDeactivateChannel(ctx context.Context, c chan string) context.Context {
	return context.WithValue(ctx, contextKeyDeactivateOutputterChannel{}, c)
}

// notify deactivation of an outputter to parent's main loop.
func deactivate(ctx context.Context, o Outputter) {
	if c, ok := ctx.Value(contextKeyDeactivateOutputterChannel{}).(chan string); ok {
		slog.Debug("Deactivate", "outputter", o.Name())
		c <- o.Name()
	}
}
