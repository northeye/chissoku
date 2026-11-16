package chissoku

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.bug.st/serial"

	"github.com/northeye/chissoku/internal/types"
)

// UD-CO2S commands and responses
const (
	commandSTP = `STP`
	commandID  = `ID?`
	commandSTA = `STA`
	responseOK = `OK`
	responseNG = `NG`
)

const (
	readTimeout  = 10 * time.Second
	commandDelay = 100 * time.Millisecond
)

var dataPattern = regexp.MustCompile(`CO2=(\d+),HUM=([0-9.]+),TMP=([0-9.-]+)`)

// device UD-CO2S serial device
type device struct {
	port    serial.Port
	scanner *bufio.Scanner
}

// openDevice opens the serial device
func openDevice(name string) (*device, error) {
	port, err := serial.Open(name, &serial.Mode{
		BaudRate: 115200,
		DataBits: 8,
		StopBits: serial.OneStopBit,
		Parity:   serial.NoParity,
	})
	if err != nil {
		return nil, fmt.Errorf("open serial device %q: %w", name, err)
	}
	if err := port.SetReadTimeout(readTimeout); err != nil {
		port.Close()
		return nil, fmt.Errorf("set read timeout: %w", err)
	}
	return &device{port: port, scanner: bufio.NewScanner(port)}, nil
}

// Close closes the serial port
func (d *device) Close() error {
	slog.Debug("Close serial port")
	return d.port.Close()
}

// send sends a command to the device
func (d *device) send(cmd string) error {
	slog.Debug("Sending command", "command", cmd)
	if _, err := io.WriteString(d.port, cmd+"\r\n"); err != nil {
		return fmt.Errorf("send command %q: %w", cmd, err)
	}
	return nil
}

// prepare initializes the device and starts measurement
func (d *device) prepare() error {
	for _, cmd := range []string{commandSTP, commandID, commandSTA} {
		if err := d.send(cmd); err != nil {
			return err
		}
		time.Sleep(commandDelay)
		res, err := d.waitResponse(cmd)
		if err != nil {
			return err
		}
		slog.Debug("Command succeeded", "command", cmd, "response", res)
	}
	slog.Info("Prepared UD-CO2S")
	return nil
}

// waitResponse skips lines until OK or NG response is received
func (d *device) waitResponse(cmd string) (string, error) {
	for d.scanner.Scan() {
		line := d.scanner.Text()
		switch {
		case strings.HasPrefix(line, responseOK):
			return line, nil
		case strings.HasPrefix(line, responseNG):
			return "", fmt.Errorf("command %q failed: %s", cmd, line)
		}
	}
	if err := d.scanner.Err(); err != nil {
		return "", fmt.Errorf("wait response for %q: %w", cmd, err)
	}
	return "", fmt.Errorf("no response for command %q", cmd)
}

// read reads measurement data from the device and sends them to out until `OK STP` is received.
// After ctx is canceled, the data are discarded. out is closed on return.
func (d *device) read(ctx context.Context, out chan<- *types.Data) error {
	defer close(out)
	for d.scanner.Scan() {
		line := d.scanner.Text()
		if strings.HasPrefix(line, responseOK+" "+commandSTP) {
			return nil
		}
		data, ok := parseData(line)
		if !ok {
			slog.Warn("Read unmatched string", "str", line)
			continue
		}
		select {
		case out <- data:
		case <-ctx.Done():
		}
	}
	if err := d.scanner.Err(); err != nil {
		return fmt.Errorf("read serial device: %w", err)
	}
	return nil
}

// parseData parses a measurement line like `CO2=1242,HUM=31.3,TMP=29.4`
func parseData(line string) (*types.Data, bool) {
	m := dataPattern.FindStringSubmatch(line)
	if m == nil {
		return nil, false
	}
	co2, err := strconv.ParseInt(m[1], 10, 64)
	if err != nil {
		return nil, false
	}
	hum, err := strconv.ParseFloat(m[2], 64)
	if err != nil {
		return nil, false
	}
	tmp, err := strconv.ParseFloat(m[3], 64)
	if err != nil {
		return nil, false
	}
	return &types.Data{
		CO2:         co2,
		Humidity:    hum,
		Temperature: tmp,
		Timestamp:   types.ISO8601Time(time.Now()),
	}, true
}
