package output

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"time"

	"github.com/northeye/chissoku/options"
	"github.com/northeye/chissoku/types"
)

// Stdout outputter for Stdout
type Stdout struct {
	Base
}

// Initialize initialize outputter
func (s *Stdout) Initialize(_ *options.Options) error {
	s.Base.Initialize(nil)
	go func() {
		var cur *types.Data
		s.write(<-s.r) // ouput first data
		tick := time.NewTicker(time.Second * time.Duration(s.Interval))
		for {
			select {
			case <-tick.C:
				if cur == nil {
					continue
				}
				s.write(cur)
				cur = nil // dismiss
			case d, more := <-s.r:
				if !more {
					slog.Debug("Output cannel has been closed", "outputter", s.Name())
					return
				}
				cur = d
			}
		}
	}()
	return nil
}

// Name outputter interface method
func (s *Stdout) Name() string {
	return strings.ToLower(reflect.Indirect(reflect.ValueOf(s)).Type().Name())
}

func (s *Stdout) write(d *types.Data) {
	b, err := json.Marshal(d)
	if err != nil {
		slog.Error("json.Marshal", "error", err, "outputter", s.Name())
		return
	}
	fmt.Println(string(b))
}
