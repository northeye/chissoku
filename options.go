package main

import "github.com/alecthomas/kong"

// Options command line options
type Options struct {
	// The serial device
	Device string `arg:"" help:"specify the serial device like '/dev/ttyACM0'"`
	// Don't outout to STDOUT
	NoStdout bool `short:"n" long:"no-stdout" help:"don't output to stdout"`
	// Output interval (sec)
	Interval int `short:"i" long:"interval" help:"interval (second) for output. default: '60'" default:"60"`
	// Quiet
	Quiet bool `long:"quiet" help:"don't output any process logs to STDERR"`

	// MQTT Address
	MqttAddress string `short:"m" long:"mqtt-address" help:"MQTT address like 'tcp://mosquitto:1883'"`
	// MQTT ClientID
	ClientID string `short:"c" long:"client-id" help:"Client ID for MQTT. default: 'chissoku'" default:"chissoku"`
	// MQTT Topic
	Topic string `short:"t" long:"topic" help:"MQTT Topic name to publish"`
	// MQTT Qos
	Qos int `short:"q" long:"qos" help:"MQTT Topic name to publish. default: '0'"`
	// MQTT retained
	Retained bool `short:"r" long:"retained" help:"MQTT Publish retained flag. default: false"`
	// SSL CA File
	CAFile string `long:"cafile" help:"Root CA Cert file for SSL" type:"existingfile"`
	// SSL Cert
	Cert string `long:"cert" help:"Cert file for SSL client authentication" type:"existingfile"`
	// SSL Key
	Key string `long:"key" help:"Pivate key file for SSL client authentication" type:"existingfile"`
	// Username
	Username string `short:"u" long:"username" help:"username for MQTT v3.1/3.1.1 authentication"`
	// Password
	Password string `short:"p" long:"password" help:"password for MQTT v3.1/3.1.1 authentication"`
	// Tags tagging data
	Tags []string `long:"tags" help:"Add tags field to json output, comma-separated strings ex: 'one,two,three'"`

	// Version
	Version kong.VersionFlag `short:"v" long:"version" help:"show program version"`
}
