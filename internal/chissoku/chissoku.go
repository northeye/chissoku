// Package chissoku is core part of chissoku
package chissoku

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	"github.com/northeye/chissoku/internal/output"
	"github.com/northeye/chissoku/internal/types"
)

// how long to wait for the device to respond to the STP command on shutdown
const stopTimeout = time.Second

// Chissoku main program
type Chissoku struct {
	// Options
	Options Options `embed:""`

	// Stdout output
	output.Stdout `prefix:"stdout." group:"Stdout Output:"`
	// MQTT output
	output.Mqtt `prefix:"mqtt." group:"MQTT Output:"`

	// available outputters
	outputters map[string]output.Outputter
}

// AfterApply kong hook
func (c *Chissoku) AfterApply(opts *Options) error {
	var writer io.Writer = os.Stderr
	level := slog.LevelInfo
	if opts.Debug {
		level = slog.LevelDebug
	}
	if opts.Quiet {
		writer = io.Discard
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: level})))
	return nil
}

// Run run the program
func (c *Chissoku) Run() error {
	slog.Debug("Start chissoku", "name", programName, "version", version)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithCancel(ctx)

	// outputters deactivating themselves are notified through this channel
	deactivated := make(chan string)
	octx := output.WithDeactivator(ctx, func(o output.Outputter) {
		select {
		case deactivated <- o.Name():
		case <-ctx.Done():
		}
	})

	active := c.initOutputters(octx)
	defer func() {
		// cancel first so that no outputter blocks on deactivation while being closed
		cancel()
		for _, o := range active {
			o.Close()
		}
	}()
	if len(active) == 0 {
		return errors.New("no active outputters are available")
	}

	dev, err := openDevice(c.Options.Device)
	if err != nil {
		return err
	}
	defer dev.Close()

	if err := dev.prepare(); err != nil {
		return err
	}

	records := make(chan *types.Data)
	readErr := make(chan error, 1)
	go func() { readErr <- dev.read(ctx, records) }()

	c.dispatch(ctx, active, records, deactivated)

	// shutdown: stop receiving signals so that a second interrupt kills the process immediately
	stop()
	cancel()
	if err := dev.send(commandSTP); err != nil {
		slog.Debug("Failed to stop the device", "error", err)
	}
	select {
	case err := <-readErr:
		return err
	case <-time.After(stopTimeout):
		return errors.New("no response from device")
	}
}

// dispatch delivers records to the active outputters until ctx is canceled,
// records is closed or all outputters are deactivated.
func (c *Chissoku) dispatch(ctx context.Context, active map[string]output.Outputter, records <-chan *types.Data, deactivated <-chan string) {
	for {
		select {
		case <-ctx.Done():
			slog.Debug("Context canceled", "cause", context.Cause(ctx))
			return
		case name := <-deactivated:
			if o, ok := active[name]; ok {
				delete(active, name)
				o.Close()
			}
			if len(active) == 0 {
				slog.Debug("No outputters are alive")
				return
			}
		case d, ok := <-records:
			if !ok {
				slog.Debug("Reader channel has been closed")
				return
			}
			b, err := c.encode(d)
			if err != nil {
				slog.Error("Failed to encode data", "error", err)
				continue
			}
			for _, o := range active {
				o.Output(b)
			}
		}
	}
}

// encode encodes the data as JSON with tags and additional maps
func (c *Chissoku) encode(d *types.Data) ([]byte, error) {
	d.Tags = c.Options.Tags
	b, err := json.Marshal(d)
	if err != nil || len(c.Options.Maps) == 0 {
		return b, err
	}

	var raw types.RawData
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, err
	}
	for k, v := range c.Options.Maps {
		raw[k] = v
	}
	return json.Marshal(raw)
}

// initOutputters initializes the outputters specified by the options and returns the active ones
func (c *Chissoku) initOutputters(ctx context.Context) map[string]output.Outputter {
	active := make(map[string]output.Outputter, len(c.Options.Output))
	for _, name := range c.Options.Output {
		o, ok := c.outputters[name]
		if !ok {
			continue
		}
		if _, dup := active[name]; dup {
			continue
		}
		if err := o.Initialize(ctx); err != nil {
			slog.Error("Initialize outputter", "outputter", name, "error", err)
			continue
		}
		active[name] = o
	}
	return active
}

// RegisterOutputters registers the outputters embedded in Chissoku and returns their names
func (c *Chissoku) RegisterOutputters() []string {
	c.outputters = make(map[string]output.Outputter)
	var names []string
	rv := reflect.ValueOf(c).Elem()
	rt := rv.Type()
	for i := range rt.NumField() {
		if !rt.Field(i).IsExported() {
			continue
		}
		if o, ok := rv.Field(i).Addr().Interface().(output.Outputter); ok {
			names = append(names, o.Name())
			c.outputters[o.Name()] = o
		}
	}
	return names
}
