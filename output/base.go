// Package output implements data outputter interfaces.
package output

import (
	"context"
	"reflect"
	"strings"

	"github.com/northeye/chissoku/types"
)

// Base outputter base struct
type Base struct {
	// Output interval (sec)
	Interval int `long:"interval" help:"interval (second) for output. default: '60'" default:"60"`

	// receiver channel
	r chan *types.Data
	// cancel
	cancel func()
}

// Close sample implementation
func (*Base) Close() {
}

// Name sample implementation
func (b *Base) Name() string {
	return strings.ToLower(reflect.TypeOf(b).Elem().Name())
}

// Output sample implementation
func (b *Base) Output(d *types.Data) {
	b.r <- d
}

// Initialize initialize outputter
func (b *Base) Initialize(_ context.Context) (_ error) {
	b.r = make(chan *types.Data, 1)
	return
}
