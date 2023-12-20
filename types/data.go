// Package types defines some data types
package types

import (
	"encoding/json"
	"time"
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
