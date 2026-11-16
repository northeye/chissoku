package output

import (
	"context"
	"fmt"
	"os"
)

// Stdout outputter for Stdout
type Stdout struct {
	Base
	Iterations int64 `help:"Inactive on maximum iterations INT"`

	// iteration counter (accessed only from the emitter loop)
	count int64
}

// Name outputter interface method
func (*Stdout) Name() string {
	return "stdout"
}

// Initialize initialize outputter
func (s *Stdout) Initialize(ctx context.Context) error {
	s.count = 0
	s.start(ctx, func(b []byte) { s.write(ctx, b) })
	return nil
}

func (s *Stdout) write(ctx context.Context, b []byte) {
	if s.Iterations > 0 && s.count >= s.Iterations {
		return
	}
	fmt.Fprintf(os.Stdout, "%s\n", b)
	s.count++
	if s.Iterations > 0 && s.count == s.Iterations {
		deactivate(ctx, s)
	}
}
