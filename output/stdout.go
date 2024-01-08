package output

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/northeye/chissoku/types"
)

// Stdout outputter for Stdout
type Stdout struct {
	Base
	Iterations int64 `help:"Inavtive on maximum iterations INT"`

	// iteration counter
	count atomic.Int64

	// close
	close func()
}

// Initialize initialize outputter
func (s *Stdout) Initialize(ctx context.Context) (_ error) {
	s.Base.Initialize(ctx)
	ctx, s.cancel = context.WithCancel(ctx)

	s.close = sync.OnceFunc(func() {
		deactivate(ctx, s)
		slog.Debug("Closing receiver channel", "outputter", s.Name())
		close(s.r)
	})

	go s.run(ctx)
	return
}

// Name outputter interface method
func (s *Stdout) Name() string {
	return strings.ToLower(reflect.TypeOf(s).Elem().Name())
}

// Close close channel
func (s *Stdout) Close() {
	s.close()
}

func (s *Stdout) run(ctx context.Context) {
	var cur *types.Data
	s.write(<-s.r) // output first data immediately
	tick := time.NewTicker(time.Second * time.Duration(s.Interval))
	for {
		select {
		case <-ctx.Done():
			cur = nil
			tick.Stop()
			s.Close()
		case <-tick.C:
			if cur == nil {
				continue
			}
			s.write(cur)
			cur = nil // dismiss
		case d, more := <-s.r:
			if !more {
				slog.Debug("Output channel has been closed", "outputter", s.Name())
				return
			}
			cur = d
		}
	}
}

func (s *Stdout) write(d *types.Data) {
	b, err := json.Marshal(d)
	if err != nil {
		slog.Error("json.Marshal", "error", err, "outputter", s.Name())
		return
	}
	if s.Iterations <= 0 {
		fmt.Println(string(b))
		return
	}
	if i := s.count.Add(1); i <= s.Iterations {
		fmt.Println(string(b))
		if i == s.Iterations {
			s.cancel()
		}
	}
}
