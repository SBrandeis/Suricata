package monitor

import (
	"errors"
	"sync"
	"time"
)

type Orchestrator struct {
	pipeline    chan PingLog
	alerts      chan Alert
	pingers     map[string]*Pinger
	aggregators map[string]*Aggregators
	reports     map[string]*Report
}

var (
	orchestrator *Orchestrator
	once         sync.Once
)

// Singleton pattern
func GetOrchestrator(pipeline chan PingLog, alerts chan Alert) *Orchestrator {
	once.Do(func() {
		orchestrator = &Orchestrator{
			pipeline:    pipeline,
			alerts:      alerts,
			pingers:     make(map[string]*Pinger),
			aggregators: make(map[string]*Aggregators),
			reports:     make(map[string]*Report),
		}
	})

	go func() {
		for {
			orchestrator.AggLog(<-orchestrator.pipeline)
		}
	}()

	return orchestrator
}

// Forward incoming PingLog to Aggregators
func (o *Orchestrator) AggLog(log PingLog) error {
	wrapper := QueueElement{
		Value:     &log,
		Timestamp: log.Time,
		next:      nil,
	}
	agg, exists := o.aggregators[log.Website]
	if !exists {
		return errors.New("NO AGGREGATOR FOR WEBSITE " + log.Website)
	}
	err, alert := agg.Short.Add(wrapper)
	if err != nil {
		return err
	}
	err, _ = agg.Medium.Add(wrapper)
	if err != nil {
		return err
	}
	err, _ = agg.Long.Add(wrapper)
	if err != nil {
		return err
	}
	if alert.Init {
		o.alerts <- alert
	}
	return nil
}

// Register a new website
func (o *Orchestrator) Register(website Website) error {
	_, registered := o.pingers[website.Url]
	if registered {
		alert := Alert{
			Url:       website.Url,
			Init:      true,
			Timestamp: time.Now(),
			Message:   "Website " + website.Url + " is already registered for monitoring, aborting",
			Value:     0.,
		}
		o.alerts <- alert
		return errors.New("WEBSITE " + website.Url + " ALREADY REGISTERED")
	}
	newPinger := NewPinger(o.pipeline, website)
	o.pingers[website.Url] = &newPinger
	err := o.addAggregators(website.Url)
	if err != nil {
		return err
	}
	o.reports[website.Url], err = NewReport(website)
	if err != nil {
		return err
	}
	alert := Alert{
		Url:       website.Url,
		Init:      true,
		Timestamp: time.Now(),
		Message:   "Website " + website.Url + " is registered for monitoring",
		Value:     0.,
	}
	o.alerts <- alert
	return nil
}

// Unregister (delete) a websitye
func (o *Orchestrator) Unregister(url string) error {
	pinger, registered := o.pingers[url]
	if !registered {
		alert := Alert{
			Url:       url,
			Init:      true,
			Timestamp: time.Now(),
			Message:   "Website " + url + " is not registered, aborting",
			Value:     0.,
		}
		o.alerts <- alert
		return errors.New("WEBSITE " + url + " IS NOT REGISTERED")
	}
	if pinger.IsRunning {
		alert := Alert{
			Url:       url,
			Init:      true,
			Timestamp: time.Now(),
			Message:   "Website " + url + " is still running, aborting",
			Value:     0.,
		}
		o.alerts <- alert
		return errors.New("PINGER FOR WEBSITE " + url + " IS STILL RUNNING")
	}
	delete(o.pingers, url)
	delete(o.reports, url)
	err := o.deleteAggregators(url)

	o.alerts <- Alert{
		Url:       url,
		Init:      true,
		Timestamp: time.Now(),
		Message:   "Website " + url + " is unregistered",
		Value:     0.,
	}
	return err
}

// Start/resume monitoring a website
func (o *Orchestrator) Start(website string) (bool, error) {
	pinger, registered := o.pingers[website]
	if !registered {
		return false, errors.New("WEBSITE " + website + " IS NOT REGISTERED")
	}
	if !pinger.IsRunning {
		r := o.reports[website]
		r.Active = true
		o.reports[website] = r
		go pinger.Start()
		o.alerts <- Alert{
			Url:       website,
			Init:      true,
			Timestamp: time.Now(),
			Message:   "Website " + website + " monitoring has started",
			Value:     0.,
		}
		return true, nil
	}
	return false, nil
}

// Pause monitoring a website
func (o *Orchestrator) Pause(url string) (bool, error) {
	pinger, registered := o.pingers[url]
	if !registered {
		return false, errors.New("WEBSITE " + url + " IS NOT REGISTERED")
	}
	if pinger.IsRunning {
		r := o.reports[url]
		r.Active = false
		o.reports[url] = r
		pinger.Pause()
		o.alerts <- Alert{
			Url:       url,
			Init:      true,
			Timestamp: time.Now(),
			Message:   "Website " + url + " monitoring is paused",
			Value:     0.,
		}
		return true, nil
	}
	return false, nil
}

// Start / Pause
func (o *Orchestrator) Toggle(url string) (bool, error) {
	pinger, registered := o.pingers[url]
	if !registered {
		return false, errors.New("WEBSITE " + url + " IS NOT REGISTERED")
	}
	if pinger.IsRunning {
		pinger.Pause()
		return false, nil
	} else {
		go pinger.Start()
		return true, nil
	}

}

// Checks whether monitoring of url is active
func (o *Orchestrator) IsActive(url string) (bool, error) {
	pinger, registered := o.pingers[url]
	if !registered {
		return false, errors.New("WEBSITE " + url + " IS NOT REGISTERED")
	}
	return pinger.IsRunning, nil
}

// Get all registered Pingers
func (o *Orchestrator) GetPingers() ([]*Pinger, error) {
	var pingers []*Pinger
	for _, pinger := range o.pingers {
		pingers = append(pingers, pinger)
	}
	return pingers, nil
}

// Add aggregators for url
func (o *Orchestrator) addAggregators(url string) error {
	_, exists := o.aggregators[url]
	if exists {
		return errors.New("WEBSITE " + url + " ALREADY HAVE AGGREGATORS")
	}
	agg := NewAggregators(url)
	o.aggregators[url] = &agg
	return nil
}

// Delete aggregators for url
func (o *Orchestrator) deleteAggregators(url string) error {
	_, exists := o.aggregators[url]
	if !exists {
		return errors.New("WEBSITE " + url + " HAS NO AGGREGATORS")
	}
	delete(o.aggregators, url)
	return nil
}

// Get all registered aggregators
func (o *Orchestrator) GetAggregators() ([]*Aggregators, error) {
	var aggregators []*Aggregators
	for _, agg := range o.aggregators {
		aggregators = append(aggregators, agg)
	}
	return aggregators, nil
}

// Get aggregators for url
func (o *Orchestrator) GetAggregator(url string) (*Aggregators, error) {
	agg, exists := o.aggregators[url]
	if !exists {
		return nil, errors.New("NO AGGREGATOR FOR WEBSITE " + url)
	}
	return agg, nil
}

// Get Report for url
func (o *Orchestrator) GetReport(url string) *Report {
	return o.reports[url]
}

// Get Summary for url
func (o *Orchestrator) GetSummary(url string) ([][]string, error) {
	report, exists := o.reports[url]
	if !exists {
		return nil, errors.New("NO REPORT FOR WEBSITE " + url)
	}
	return report.Summary(), nil
}

// Update Short Term report for url
func (o *Orchestrator) UpdateShortReport(url string) error {
	err := o.reports[url].ShortTerm.Update(o.aggregators[url].Short)
	return err
}

// Update Medium Term report for url
func (o *Orchestrator) UpdateMediumReport(url string) error {
	err := o.reports[url].MediumTerm.Update(o.aggregators[url].Medium)
	return err
}

// Update Long Term report for url
func (o *Orchestrator) UpdateLongReport(url string) error {
	err := o.reports[url].LongTerm.Update(o.aggregators[url].Long)
	return err
}

// Start monitoring for all registered websites
func (o *Orchestrator) StartAll() error {
	pingers, err := o.GetPingers()
	if err != nil {
		return err
	}
	for _, pinger := range pingers {
		_, err = o.Start(pinger.Url)
		if err != nil {
			return err
		}
	}
	return err
}

// Pause monitoring for all registered websites
func (o *Orchestrator) PauseAll() error {
	pingers, err := o.GetPingers()
	if err != nil {
		return err
	}
	for _, pinger := range pingers {
		_, err = o.Pause(pinger.Url)
		if err != nil {
			return err
		}
	}
	return err
}

// Unregister all websites
func (o *Orchestrator) UnregisterAll() error {
	pingers, err := o.GetPingers()
	if err != nil {
		return err
	}
	for _, pinger := range pingers {
		err = o.Unregister(pinger.Url)
		if err != nil {
			return err
		}
	}
	return err
}
