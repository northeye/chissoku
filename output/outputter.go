// Package output implements data outputter interfaces.
package output

import (
	"github.com/northeye/chissoku/options"
	"github.com/northeye/chissoku/types"
)

// Outputter the outputter interface
type Outputter interface {
	// Name returns unique name of the outputter.
	// Lower cased struct name is recomended.
	Name() string

	// Intialize the outputter.
	// When it returns non-nil error the outputter will be disabled.
	Initialize(*options.Options) error

	// Output the data.
	// This method must be non-blocking and light-weight.
	Output(*types.Data)

	// Close cleanup the outputter.
	Close()
}
