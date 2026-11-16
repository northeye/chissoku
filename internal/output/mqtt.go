package output

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	mqttPublishTimeout = 10 * time.Second
	// quiesce time (msec) for disconnecting
	mqttDisconnectQuiesce = 250
)

// URL schemes which paho treats as TLS connections
var mqttTLSSchemes = []string{"ssl://", "tls://", "mqtts://", "tcps://"}

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
	Qos int `help:"MQTT QoS (0, 1 or 2) to publish. default: '0'"`
	// MQTT Retain
	Retain bool `help:"MQTT Retain flag. default: 'false'" default:"false"`
	// SSL CA File
	CAFile string `name:"ssl-cafile" help:"Root CA Cert file for SSL" type:"existingfile"`
	// SSL Cert
	Cert string `name:"ssl-cert" help:"Cert file for SSL client authentication" type:"existingfile"`
	// SSL Key
	Key string `name:"ssl-key" help:"Private key file for SSL client authentication" type:"existingfile"`
	// Username
	Username string `help:"username for MQTT v3.1/3.1.1 authentication"`
	// Password
	Password string `help:"password for MQTT v3.1/3.1.1 authentication"`

	// mqtt Client interface
	client mqtt.Client
}

// Name outputter interface method
func (*Mqtt) Name() string {
	return "mqtt"
}

// Initialize initialize outputter
func (m *Mqtt) Initialize(ctx context.Context) error {
	if m.Address == "" {
		return errors.New("--mqtt.address is required")
	}
	if m.Topic == "" {
		return errors.New("--mqtt.topic is required")
	}
	if m.Qos < 0 || m.Qos > 2 {
		return fmt.Errorf("invalid QoS: %d", m.Qos)
	}

	o := mqtt.NewClientOptions().AddBroker(m.Address)
	if m.ClientID != "" {
		o.SetClientID(m.ClientID)
	}
	if m.Username != "" {
		o.SetUsername(m.Username)
		o.SetPassword(m.Password)
	}
	if m.isTLS() {
		cfg, err := m.newTLSConfig()
		if err != nil {
			return fmt.Errorf("prepare SSL/TLS configuration: %w", err)
		}
		o.SetTLSConfig(cfg)
	}

	m.client = mqtt.NewClient(o)
	if t := m.client.Connect(); t.Wait() && t.Error() != nil {
		return fmt.Errorf("connect to %s: %w", m.Address, t.Error())
	}

	m.start(ctx, m.publish)
	return nil
}

// Close outputter interface method
// close the MQTT connection
func (m *Mqtt) Close() {
	m.Base.Close()
	if m.client != nil && m.client.IsConnected() {
		m.client.Disconnect(mqttDisconnectQuiesce)
	}
}

func (m *Mqtt) publish(b []byte) {
	t := m.client.Publish(m.Topic, byte(m.Qos), m.Retain, b)
	if !t.WaitTimeout(mqttPublishTimeout) {
		slog.Warn("Publish timed out", "outputter", m.Name(), "topic", m.Topic)
		return
	}
	if err := t.Error(); err != nil {
		slog.Error("Publish failed", "outputter", m.Name(), "topic", m.Topic, "error", err)
		return
	}
	slog.Debug("Published data", "outputter", m.Name(), "topic", m.Topic)
}

func (m *Mqtt) isTLS() bool {
	for _, s := range mqttTLSSchemes {
		if strings.HasPrefix(m.Address, s) {
			return true
		}
	}
	return false
}

func (m *Mqtt) newTLSConfig() (*tls.Config, error) {
	cfg := &tls.Config{}

	if m.CAFile != "" {
		ca, err := os.ReadFile(m.CAFile)
		if err != nil {
			return nil, fmt.Errorf("read root CA: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(ca) {
			return nil, fmt.Errorf("no valid certificate found in %s", m.CAFile)
		}
		cfg.RootCAs = pool
	}

	switch {
	case m.Cert != "" && m.Key != "":
		cert, err := tls.LoadX509KeyPair(m.Cert, m.Key)
		if err != nil {
			return nil, fmt.Errorf("load client certificate: %w", err)
		}
		cfg.Certificates = []tls.Certificate{cert}
	case m.Cert != "" || m.Key != "":
		return nil, errors.New("both --mqtt.ssl-cert and --mqtt.ssl-key must be specified")
	}

	slog.Debug("Prepared SSL/TLS configuration", "outputter", m.Name(),
		"rootCA", m.CAFile != "", "clientCert", len(cfg.Certificates) > 0)
	return cfg, nil
}
