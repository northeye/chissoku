package main

import (
	"crypto/tls"
	"crypto/x509"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

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

	if opts.Cert != "" && opts.Key != "" {
		if cert, err := tls.LoadX509KeyPair(opts.Cert, opts.Key); err == nil {
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
