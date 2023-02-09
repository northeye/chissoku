package main

// Options command line options
type Options struct {
	Device string `short:"d" long:"device" description:"specify the serial device" required:"true"`
	// Don't print to STDOUT
	NoStdout bool `short:"n" long:"no-stdout" description:"don't output to stdout"`
	// Output interval (sec)
	Interval int `short:"i" long:"interval" description:"interval (second) for output. default: 60"`
	// Quiet
	Quiet bool `long:"quiet" description:"don't output any process logs to STDERR"`

	// MQTT Address
	MqttAddress string `short:"m" long:"mqtt-address" description:"MQTT address like tcp://mosquitto:1883"`
	// MQTT ClientID
	ClientID string `short:"c" long:"client-id" description:"Client ID for MQTT. default: \"chissoku\""`
	// MQTT Topic
	Topic string `short:"t" long:"topic" description:"MQTT Topic name to publish"`
	// MQTT QoS
	QoS int `short:"q" long:"qos" description:"MQTT Topic name to publish. default: 1"`
	// MQTT retained
	Retained bool `short:"r" long:"retained" description:"MQTT Publish retained flag. default: false"`
	// SSL CA File
	CAFile string `long:"ca-cert" description:"Root CA Cert file for SSL"`
	// SSL Cert
	CertFile string `long:"client-cert" description:"Cert file for SSL Client Auth"`
	// SSL Key
	KeyFile string `long:"private-key" description:"Pivate key file for SSL Client Auth"`
	// Username
	Username string `short:"u" long:"username" description:"username for MQTT authentication"`
	// Password
	Password string `short:"p" long:"password" description:"password for MQTT authentication"`
	// Tags tagging data
	Tags []string `long:"tags" description:"Add tags field to json output. ex: \"one,two,three\""`

	// Version
	Version bool `short:"v" long:"version" description:"print version"`
}
