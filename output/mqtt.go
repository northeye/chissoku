// Package output implements data outputter interfaces.
package output

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"log/slog"
	"os"
	"reflect"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/northeye/chissoku/options"
	"github.com/northeye/chissoku/types"
)

// Mqtt MQTT Outputter
type Mqtt struct {
	Base

	// MQTT Address
	Address string `help:"MQTT address like 'tcp://mosquitto:1883'"`
	// MQTT ClientID
	ClientID string `help:"Client ID for MQTT. default: 'chissoku'" default:"chissoku"`
	// MQTT Topic
	Topic string `help:"MQTT Topic name to publish"`
	// MQTT Qos
	Qos int `help:"MQTT Topic name to publish. default: '0'"`
	// SSL CA File
	CAFile string `name:"ssl-cafile" help:"Root CA Cert file for SSL" type:"existingfile"`
	// SSL Cert
	Cert string `name:"ssl-cert" help:"Cert file for SSL client authentication" type:"existingfile"`
	// SSL Key
	Key string `name:"ssl-key" help:"Pivate key file for SSL client authentication" type:"existingfile"`
	// Username
	Username string `help:"username for MQTT v3.1/3.1.1 authentication"`
	// Password
	Password string `help:"password for MQTT v3.1/3.1.1 authentication"`

	// mqtt mqtt Client interface
	client mqtt.Client
}

// Initialize initialize outputter
func (m *Mqtt) Initialize(_ *options.Options) error {
	m.Base.Initialize(nil)
	o := mqtt.NewClientOptions()
	o.AddBroker(m.Address)
	if m.ClientID != "" {
		o.SetClientID(m.ClientID)
	}
	if strings.HasPrefix(m.Address, `ssl://`) {
		if t, err := m.newTLSConfig(); err == nil {
			o.SetTLSConfig(t)
		} else {
			slog.Warn("could not enfoce SSL/TLS option, disabled at this time.", "err", err)
		}
	}
	m.client = mqtt.NewClient(o)
	if t := m.client.Connect(); t.Wait() && t.Error() != nil {
		return t.Error()
	}

	go func() {
		var cur *types.Data
		m.publish(<-m.r) // publish first data
		tick := time.NewTicker(time.Second * time.Duration(m.Interval))
		for {
			select {
			case <-tick.C:
				if cur == nil {
					continue
				}
				m.publish(cur)
				cur = nil // dismiss
			case d, more := <-m.r:
				if !more {
					slog.Debug("reader channel has been closed", "outputter", m.Name())
					tick.Stop()
					return
				}
				cur = d
			}
		}
	}()

	return nil
}

// Name outputter interface method
func (m *Mqtt) Name() string {
	return strings.ToLower(reflect.Indirect(reflect.ValueOf(m)).Type().Name())
}

// Close outputter interface method
// clsoe the MQTT connection
func (m *Mqtt) Close() {
	if m.client.IsConnected() {
		m.client.Disconnect(1000)
	}
}

func (m *Mqtt) publish(d *types.Data) {
	b, err := json.Marshal(d)
	if err != nil {
		slog.Error("json.Marshal", "error", err, "outputter", m.Name())
		return
	}
	token := m.client.Publish(m.Topic, byte(m.Qos), false, b)
	slog.Debug("Publish data", "outputter", m.Name(), "token", token)
}

func (m *Mqtt) newTLSConfig() (*tls.Config, error) {
	state := map[string]any{}
	defer func() {
		msg := "Preparing SSL/TLS Configuration"
		level := slog.LevelInfo
		for _, err := range state {
			if e, ok := err.(error); ok && e != nil {
				level = slog.LevelError
				break
			}
		}
		slog.Log(context.Background(), level, msg, "state", state)
	}()

	cfg := &tls.Config{
		InsecureSkipVerify: false,
	}

	if m.CAFile != "" {
		if ca, err := os.ReadFile(m.CAFile); err == nil {
			certpool := x509.NewCertPool()
			certpool.AppendCertsFromPEM(ca)
			cfg.RootCAs = certpool
			state["RootCA"] = "OK"
		} else {
			state["RootCA"] = err
			return nil, err
		}
	}

	if m.Cert != "" && m.Key != "" {
		if cert, err := tls.LoadX509KeyPair(m.Cert, m.Key); err == nil {
			cfg.Certificates = []tls.Certificate{cert}
			state["ClientCert"] = "OK"
		} else {
			state["ClientCert"] = err
			return nil, err
		}
	}
	return cfg, nil
}
