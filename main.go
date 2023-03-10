// the chissoku program
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"time"

	"github.com/alecthomas/kong"
	"go.bug.st/serial"
)

// ISO8601Time utility
type ISO8601Time time.Time

// ISO8601 date time format
const ISO8601 = `2006-01-02T15:04:05.000Z07:00`

// MarshalJSON interface function
func (t ISO8601Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(t).Format(ISO8601))
}

// Data - the data
type Data struct {
	CO2         int64       `json:"co2"`
	Humidity    float64     `json:"humidity"`
	Temperature float64     `json:"temperature"`
	Tags        []string    `json:"tags,omitempty"`
	Timestamp   ISO8601Time `json:"timestamp"`
}

// initialize and prepare the device
func prepareDevice(p serial.Port, s *bufio.Scanner) error {
	logInfo("Prepare device...:")
	defer logPrintln("")
	for _, c := range []string{"STP", "ID?", "STA"} {
		logPrintf(" %v", c)
		if _, err := p.Write([]byte(c + "\r\n")); err != nil {
			return err
		}
		time.Sleep(time.Millisecond * 100) // wait
		for s.Scan() {
			t := s.Text()
			if t[:2] == `OK` {
				break
			} else if t[:2] == `NG` {
				return fmt.Errorf(" command `%v` failed", c)
			}
		}
	}
	logPrint(" OK.")
	return nil
}

func main() {
	var opts Options

	kong.Parse(&opts,
		kong.Name(ProgramName),
		kong.Vars{"version": "v" + Version},
		kong.Description(`A CO2 sensor reader`))

	if opts.Quiet {
		logWriter = io.Discard
	}

	port, err := serial.Open(opts.Device, &serial.Mode{
		BaudRate: 115200,
		DataBits: 8,
		StopBits: serial.OneStopBit,
		Parity:   serial.NoParity,
	})
	if err != nil {
		logErrorf("Opening serial: %+v: %v\n", err, opts.Device)
		os.Exit(1)
	}
	defer func() { port.Write([]byte("STP\r\n")); time.Sleep(time.Millisecond * 100); port.Close() }()

	// serial reader
	port.SetReadTimeout(time.Second * 10)
	s := bufio.NewScanner(port)
	s.Split(bufio.ScanLines)

	if err := prepareDevice(port, s); err != nil {
		logErrorln(err.Error())
		port.Close()
		os.Exit(1)
	}

	// mqtt
	client := newMqttClient(&opts)
	if client != nil {
		logInfoln("Connecting to MQTT broker...")
		if t := client.Connect(); t.Wait() && t.Error() != nil {
			logErrorf("%v, disable MQTT output at the time.\n", t.Error())
			client = nil
		} else {
			defer client.Disconnect(1000)
		}
	}

	// trap SIGINT
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)
	// signal handler
	go func() {
		<-sigch
		port.Write([]byte("STP\r\n"))
	}()

	// serial reader channel
	r := make(chan *Data)
	// publisher channel
	p := make(chan *Data)

	// publisher
	go func() {
		for d := range p {
			d.Tags = opts.Tags
			b, err := json.Marshal(d)
			if err != nil {
				logError(err.Error())
				continue
			}
			if !opts.NoStdout {
				fmt.Println(string(b))
			}
			if client != nil {
				client.Publish(opts.Topic, byte(opts.Qos), false, b)
			}
		}
	}()

	// periodical dispatcher
	go func() {
		var cur *Data // current data
		p <- <-r      // send the first data
		tick := time.Tick(time.Second * time.Duration(opts.Interval))
		for {
			select {
			case <-tick:
				if cur == nil {
					continue
				}
				p <- cur  // publish current
				cur = nil // dismiss
			case cur = <-r:
			}
		}
	}()

	// reader (main)
	re := regexp.MustCompile(`CO2=(\d+),HUM=([0-9\.]+),TMP=([0-9\.-]+)`)
	for s.Scan() {
		d := &Data{Timestamp: ISO8601Time(time.Now())}
		text := s.Text()
		m := re.FindAllStringSubmatch(text, -1)
		if len(m) > 0 {
			d.CO2, _ = strconv.ParseInt(m[0][1], 10, 64)
			d.Humidity, _ = strconv.ParseFloat(m[0][2], 64)
			d.Temperature, _ = strconv.ParseFloat(m[0][3], 64)
			d.Timestamp = ISO8601Time(time.Now())
			r <- d
		} else if text[:6] == `OK STP` {
			return // exit 0
		} else {
			logWarningf("Read unmatched string: %v", text)
		}
	}
	if s.Err() != nil {
		logError(s.Err().Error())
	}
}
