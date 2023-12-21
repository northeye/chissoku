// Package output implements data outputter interfaces.
package output

import (
	"reflect"
	"strings"

	"github.com/northeye/chissoku/options"
	"github.com/northeye/chissoku/types"
)

// Base outputter base struct
type Base struct {
	// Output interval (sec)
	Interval int `long:"interval" help:"interval (second) for output. default: '60'" default:"60"`

	// receiver channel
	r chan *types.Data
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
func (b *Base) Initialize(_ *options.Options) (_ error) {
	b.r = make(chan *types.Data)
	return
}
