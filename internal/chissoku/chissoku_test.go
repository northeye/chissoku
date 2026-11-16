package chissoku

import (
	"encoding/json"
	"testing"
)

func TestParseData(t *testing.T) {
	t.Parallel()
	tests := []struct {
		line string
		ok   bool
		co2  int64
		hum  float64
		tmp  float64
	}{
		{"CO2=1242,HUM=31.3,TMP=29.4", true, 1242, 31.3, 29.4},
		{"CO2=400,HUM=50.0,TMP=-3.5", true, 400, 50.0, -3.5},
		{"", false, 0, 0, 0},
		{"OK", false, 0, 0, 0},
		{"CO2=1,HUM=1.2.3,TMP=0", false, 0, 0, 0},
	}
	for _, tt := range tests {
		d, ok := parseData(tt.line)
		if ok != tt.ok {
			t.Errorf("parseData(%q) ok = %v, want %v", tt.line, ok, tt.ok)
			continue
		}
		if !ok {
			continue
		}
		if d.CO2 != tt.co2 || d.Humidity != tt.hum || d.Temperature != tt.tmp {
			t.Errorf("parseData(%q) = %+v", tt.line, d)
		}
	}
}

func TestEncode(t *testing.T) {
	t.Parallel()
	c := &Chissoku{Options: Options{
		Tags: []string{"Living"},
		Maps: map[string]string{"room": "1"},
	}}
	d, _ := parseData("CO2=1242,HUM=31.3,TMP=29.4")
	b, err := c.encode(d)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got["co2"] != float64(1242) || got["room"] != "1" || got["timestamp"] == nil {
		t.Errorf("unexpected output: %s", b)
	}
	if tags, ok := got["tags"].([]any); !ok || len(tags) != 1 || tags[0] != "Living" {
		t.Errorf("unexpected tags: %s", b)
	}
}

func TestRegisterOutputters(t *testing.T) {
	t.Parallel()
	var c Chissoku
	names := c.RegisterOutputters()
	if len(names) != 2 || names[0] != "stdout" || names[1] != "mqtt" {
		t.Errorf("RegisterOutputters() = %v", names)
	}
}
