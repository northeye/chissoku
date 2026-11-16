// Package output implements data outputter interfaces.
package output

import (
	"context"
	"log/slog"
)

// Outputter the outputter interface
type Outputter interface {
	// Name returns unique name of the outputter.
	// Lower cased struct name is recommended.
	Name() string

	// Initialize the outputter and run the receiver loop as own goroutine.
	// When it returns non-nil error the outputter will be disabled.
	// The context is canceled upon the parent's termination.
	Initialize(context.Context) error

	// Output the data.
	// This method must be non-blocking and light-weight.
	Output([]byte)

	// Close cleanup the outputter.
	// It's invoked by the parent when the outputter has been deactivated or upon the parent's termination.
	// It must be idempotent.
	Close()
}

// contextKeyDeactivator context value key for the deactivator function
type contextKeyDeactivator struct{}

// WithDeactivator returns a new context carrying f, which is called when an outputter deactivates itself.
// f must not block indefinitely.
func WithDeactivator(ctx context.Context, f func(Outputter)) context.Context {
	return context.WithValue(ctx, contextKeyDeactivator{}, f)
}

// deactivate notifies the parent that the outputter should be removed from the active outputters.
// The parent will call o.Close() afterwards.
func deactivate(ctx context.Context, o Outputter) {
	if f, ok := ctx.Value(contextKeyDeactivator{}).(func(Outputter)); ok {
		slog.Debug("Deactivate", "outputter", o.Name())
		f(o)
	}
}
