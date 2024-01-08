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

	// Intialize the outputter.
	// When it returns non-nil error the outputter will be disabled.
	Initialize(context.Context) error

	// Output the data.
	// This method must be non-blocking and light-weight.
	Output(*types.Data)

	// Close cleanup the outputter.
	Close()
}

// contextKeyDeactivateOutputterChannel context value key for DeactivateOutputterChannel
type contextKeyDeactivateOutputterChannel struct{}

// ContextWithDeactivateChannel new context with deactivate channel
func ContextWithDeactivateChannel(ctx context.Context, c chan string) context.Context {
	return context.WithValue(ctx, contextKeyDeactivateOutputterChannel{}, c)
}

// deactivate deactivate an outputter
func deactivate(ctx context.Context, o Outputter) {
	if c, ok := ctx.Value(contextKeyDeactivateOutputterChannel{}).(chan string); ok {
		slog.Debug("Deactivate", "outputter", o.Name())
		c <- o.Name()
	}
}
