// Package chissoku implements main chissoku program
package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.bug.st/serial"

	"github.com/alecthomas/kong"
	"github.com/northeye/chissoku/options"
	"github.com/northeye/chissoku/output"
	"github.com/northeye/chissoku/types"
)

func main() {
	var c Chissoku
	ctx := kong.Parse(&c,
		kong.Name(ProgramName),
		kong.Vars{"version": "v" + Version, "outputters": c.OutputterNames()},
		kong.Description(`A CO2 sensor reader`),
		kong.Bind(&c.Options))
	if err := ctx.Run(); err != nil {
		slog.Error("chissoku.Run()", "error", err)
		os.Exit(1)
	}
}

// Chissoku main program
type Chissoku struct {
	// Options
	Options options.Options `embed:""`

	// Stdout output
	output.Stdout `prefix:"stdout." group:"Stdout Output:"`
	// MQTT output
	output.Mqtt `prefix:"mqtt." group:"MQTT Output:"`

	// available outputters
	outputters map[string]output.Outputter

	// reader channel
	rchan chan *types.Data

	// serial device
	port serial.Port
	// serial scanner
	scanner *bufio.Scanner
}

// AfterApply kong hook
func (c *Chissoku) AfterApply(opts *options.Options) error {
	var writer io.Writer = os.Stderr
	level := slog.LevelInfo
	if opts.Debug {
		level = slog.LevelDebug
	}
	if opts.Quiet {
		writer = io.Discard
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: level})))

	c.rchan = make(chan *types.Data)

	enabled := opts.Output[:0]
	for _, v := range opts.Output {
		if o, ok := c.outputters[v]; ok {
			if err := o.Initialize(opts); err != nil {
				slog.Error("Initialize outputter", "outputter", o.Name(), "error", err)
				continue
			}
			enabled = append(enabled, v)
		}
	}
	opts.Output = enabled
	if len(opts.Output) == 0 {
		return fmt.Errorf("no outputters are avaiable")
	}
	return nil
}

const (
	// CommandSTP the STP Command
	CommandSTP string = `STP`
	// CommandID the ID? Command
	CommandID string = `ID?`
	// CommandSTA the STA Command
	CommandSTA string = `STA`
	// ResponseOK the OK response
	ResponseOK string = `OK`
	// ResponseNG the NG response
	ResponseNG string = `NG`
)

func (c *Chissoku) cleanup() {
	if c.port != nil {
		slog.Debug("Closing Serial port")
		// nolint: errcheck
		c.port.Write([]byte(CommandSTP + "\r\n"))
		// nolint: errcheck
		c.port.Close()
	}
	time.Sleep(time.Millisecond * 100)
	for _, v := range c.Options.Output {
		c.outputters[v].Close()
	}
}

// Run run the program
func (c *Chissoku) Run() (err error) {
	slog.Debug("Start", "name", ProgramName, "version", Version)

	opts := &c.Options

	if c.port, err = serial.Open(opts.Device, &serial.Mode{
		BaudRate: 115200,
		DataBits: 8,
		StopBits: serial.OneStopBit,
		Parity:   serial.NoParity,
	}); err != nil {
		slog.Error("Opening serial", "error", err, "device", opts.Device)
		return err
	}
	c.port.SetReadTimeout(time.Second * 10)

	// initialize UD-CO2S
	if err := c.prepareDevice(); err != nil {
		return err
	}

	// signalHandler
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)
	// signal handler
	go func() {
		<-sigch
		c.cleanup()
		os.Exit(128)
	}()

	// main
	go c.dispatch()

	return c.readDevice()
}

// readDevice read data from serial device
func (c *Chissoku) readDevice() error {
	re := regexp.MustCompile(`CO2=(\d+),HUM=([0-9\.]+),TMP=([0-9\.-]+)`)
	// as main loop
	for c.scanner.Scan() {
		text := c.scanner.Text()
		m := re.FindAllStringSubmatch(text, -1)
		if len(m) > 0 {
			d := &types.Data{Timestamp: types.ISO8601Time(time.Now()), Tags: c.Options.Tags}
			d.CO2, _ = strconv.ParseInt(m[0][1], 10, 64)
			d.Humidity, _ = strconv.ParseFloat(m[0][2], 64)
			d.Temperature, _ = strconv.ParseFloat(m[0][3], 64)
			c.rchan <- d
		} else if text[:6] == `OK STP` {
			return nil // exit 0
		} else {
			slog.Warn("Read unmatched string", "str", text)
		}
	}
	if c.scanner.Err() != nil {
		slog.Error("Scanner read error", "error", c.scanner.Err())
		return c.scanner.Err()
	}
	return nil
}

func (c *Chissoku) dispatch() {
	for d := range c.rchan {
		for _, v := range c.Options.Output {
			c.outputters[v].Output(d)
		}
	}
	slog.Debug("Reader channel has ben closed")
}

// initialize and prepare the device
func (c *Chissoku) prepareDevice() (err error) {
	c.scanner = bufio.NewScanner(c.port)
	c.scanner.Split(bufio.ScanLines)

	commands := []string{CommandSTP, CommandID, CommandSTA}
	do := make([]string, 0, len(commands))
	defer func() {
		level := slog.LevelInfo
		if err != nil {
			level = slog.LevelError
		}
		slog.Log(context.Background(), level, "Prepare UD-CO2S", "commands", do, "error", err)
	}()
	for _, cmd := range commands {
		do = append(do, cmd)
		if _, err = c.port.Write([]byte(cmd + "\r\n")); err != nil {
			return
		}
		time.Sleep(time.Millisecond * 100) // wait
		for c.scanner.Scan() {
			t := c.scanner.Text()
			if strings.HasPrefix(t[:2], ResponseOK) {
				break
			} else if strings.HasPrefix(t[:2], ResponseNG) {
				return fmt.Errorf("command `%v` failed", cmd)
			}
		}
	}
	return
}

// OutputterNames returns names of impleneted outputter
func (c *Chissoku) OutputterNames() (names string) {
	enum := []string{}
	if c.outputters != nil {
		for k := range c.outputters {
			enum = append(enum, k)
		}
		return strings.Join(enum, ",")
	}
	c.outputters = make(map[string]output.Outputter)
	rv := reflect.Indirect(reflect.ValueOf(c))
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if !f.IsExported() {
			continue
		}
		if value, ok := rv.Field(i).Addr().Interface().(output.Outputter); ok {
			name := value.Name()
			enum = append(enum, name)
			c.outputters[name] = value
		}
	}
	return strings.Join(enum, ",")
}
