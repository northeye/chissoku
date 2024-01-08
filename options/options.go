// Package options implements global options
package options

import (
	"github.com/alecthomas/kong"
)

// Options command line options
type Options struct {
	// The serial device
	Device string `arg:"" help:"specify the serial device like '/dev/ttyACM0'"`
	// Output
	Output []string `short:"o" enum:"${outputters}" help:"at least one output must be specified (available: ${enum})" default:"stdout"`
	// Quiet
	Quiet bool `short:"q" help:"don't output any process logs to STDERR"`
	// Tags tagging data
	Tags []string `short:"t" help:"Add tags field to json output, comma-separated strings ex: 'one,two,three'"`
	// Version
	Version kong.VersionFlag `short:"v" help:"show program version"`
	// Debug
	Debug bool `short:"d" help:"print debug log"`
}

// ContextKeyOptions context value key for global Options
type ContextKeyOptions struct{}
