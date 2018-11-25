package monitor

import (
	"net/http"
	"time"
)

type Pinger struct {
	out        chan<- PingLog
	Url        string
	Interval   int
	IsRunning  bool
	httpClient *http.Client
}

type PingLog struct {
	Time         time.Time
	Website      string
	Error        error
	Status       int
	ResponseTime time.Duration
}

func NewPinger(outChan chan<- PingLog, website Website) Pinger {
	tr := &http.Transport{
		DisableKeepAlives:     true,
		ResponseHeaderTimeout: time.Duration(website.CheckInterval) * time.Millisecond,
	}
	client := &http.Client{Transport: tr}
	return Pinger{
		out:        outChan,
		Url:        website.Url,
		Interval:   website.CheckInterval,
		IsRunning:  false,
		httpClient: client,
	}
}

func (p *Pinger) Start() {
	p.IsRunning = true
	tick := time.NewTicker(time.Duration(p.Interval) * time.Millisecond)

	for p.IsRunning {
		<-tick.C
		startTime := time.Now()
		res, err := p.httpClient.Get(p.Url)
		if err != nil {
			p.out <- PingLog{
				startTime,
				p.Url,
				err,
				0,
				time.Now().Sub(startTime),
			}
		} else {
			p.out <- PingLog{
				startTime,
				p.Url,
				nil,
				res.StatusCode,
				time.Now().Sub(startTime),
			}
			res.Body.Close()
		}
	}
}

func (p *Pinger) Pause() {
	p.IsRunning = false
}
