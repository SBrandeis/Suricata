package monitor

import "time"

type Alert struct {
	Url       string
	Timestamp time.Time
	Value     float32
	Message   string
	Init      bool
}
