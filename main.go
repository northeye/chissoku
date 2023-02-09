// the chissoku program
package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	flags "github.com/jessevdk/go-flags"
	"go.bug.st/serial"
)

// ISO8601Time utitlity
type ISO8601Time time.Time

// ISO8601 date time format
const ISO8601 = `2006-01-02T15:04:05.000+09:00`

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

func newTLSConfig(opts *Options) (*tls.Config, error) {
	logInfof("Preparing SSL/TLS Configuration...")
	cfg := &tls.Config{
		InsecureSkipVerify: false,
	}

	if opts.CAFile != "" {
		if c, err := os.ReadFile(opts.CAFile); err == nil {
			certpool := x509.NewCertPool()
			certpool.AppendCertsFromPEM(c)
			cfg.RootCAs = certpool
			logPrint(" RootCA")
		} else {
			logError(err.Error())
			return nil, err
		}
	}

	if opts.CertFile != "" && opts.KeyFile != "" {
		if cert, err := tls.LoadX509KeyPair(opts.CertFile, opts.KeyFile); err == nil {
			cfg.Certificates = []tls.Certificate{cert}
			logPrint(" ClientCert")
		} else {
			logError(err.Error())
			return nil, err
		}
	}
	logPrint(" OK.")
	return cfg, nil
}

func newMqttClient(opts *Options) mqtt.Client {
	if opts.MqttAddress == "" {
		return nil
	}
	o := mqtt.NewClientOptions()
	o.AddBroker(opts.MqttAddress)
	if opts.ClientID != "" {
		o.SetClientID(opts.ClientID)
	}
	if opts.MqttAddress[:6] == `ssl://` {
		if t, err := newTLSConfig(opts); err == nil {
			o.SetTLSConfig(t)
		} else {
			logWarningln("could not enfoce SSL/TLS option, disabled at this time.")
		}
	}
	return mqtt.NewClient(o)
}

// initialize and prepare the device
func prepareDevice(w io.Writer, s *bufio.Scanner) error {
	logInfo("Prepare device...:")
	defer logPrint("\n")
	for _, c := range []string{"STP", "ID?", "STA"} {
		logPrintf(" %v", c)
		if _, err := w.Write([]byte(c + "\r\n")); err != nil {
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
	opts.QoS = 1
	opts.ClientID = `chissoku`
	opts.Interval = 60
	if _, err := flags.Parse(&opts); err != nil {
		os.Exit(1)
	}
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
	port.SetReadTimeout(time.Duration(time.Second * 10))
	s := bufio.NewScanner(port)
	s.Split(bufio.ScanLines)

	// mqtt
	client := newMqttClient(&opts)
	if client != nil {
		logInfoln("Connecting to MQTT broker...")
		if t := client.Connect(); t.Wait() && t.Error() != nil {
			logErrorf("%v, disable MQTT output at the time.\n", t.Error())
			client = nil
		}
	}

	if err := prepareDevice(port.(io.Writer), s); err != nil {
		logError(" " + err.Error())
		port.Close()
		os.Exit(1)
	}

	// trap SIGINT
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)
	// signal handler
	go func() {
		<-sigch
		port.Write([]byte("STP\r\n"))
		time.Sleep(time.Millisecond * 100)
		port.Close()
		if client != nil {
			client.Disconnect(1000)
		}
		os.Exit(0)
	}()

	// serial reader channel
	r := make(chan *Data)
	// publisher channel
	p := make(chan *Data)

	// publisher
	go func() {
		for {
			select {
			case d := <-p:
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
					client.Publish(opts.Topic, byte(opts.QoS), opts.Retained, b)
				}
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
				p <- cur
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
		} else {
			logWarningf("Read unmatched string: %v", text)
		}
	}
	if s.Err() != nil {
		logError(s.Err().Error())
	}
}
