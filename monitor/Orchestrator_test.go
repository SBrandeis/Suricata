package monitor

import (
	"testing"
	"time"
)

var pipeline_test chan PingLog
var alerts_test chan Alert
var orchestrator_test *Orchestrator

func setup() {
	pipeline_test = make(chan PingLog)
	alerts_test = make(chan Alert)
	orchestrator_test = &Orchestrator{
		pipeline:    pipeline_test,
		alerts:      alerts_test,
		pingers:     make(map[string]*Pinger),
		aggregators: make(map[string]*Aggregators),
		reports:     make(map[string]*Report),
	}
}

func TestGetOrchestrator(t *testing.T) {
	setup()
	orch1 := GetOrchestrator(pipeline_test, alerts_test)
	orch2 := GetOrchestrator(pipeline_test, alerts_test)
	if orch1 != orch2 {
		t.Error("Orchestrator is not singleton:", orch1, orch2)
	}
}

func TestOrchestrator_Register(t *testing.T) {
	setup()
	time.Sleep(10 * time.Millisecond)
	website := Website{Url: "example", CheckInterval: 100}

	// Handle alert emitting
	messages := make([]Alert, 0)
	loop := true
	defer func() { loop = false }()
	go func() {
		for loop {
			messages = append(messages, <-alerts_test)
		}
	}()
	err := orchestrator_test.Register(website)
	if err != nil {
		t.Error("Error while registering website:", err)
	}
	// Wait for alert to be emitted and caught
	time.Sleep(10 * time.Millisecond)
	if len(messages) != 1 {
		t.Error("Alert was not raised (first register)")
	}
	_, ok := orchestrator_test.aggregators[website.Url]
	if !ok {
		t.Error("Aggregators were not added")
	}
	_, ok = orchestrator_test.reports[website.Url]
	if !ok {
		t.Error("Report was not added")
	}
	_, ok = orchestrator_test.pingers[website.Url]
	if !ok {
		t.Error("Pinger was not added")
	}
	err = orchestrator_test.Register(website)
	if err == nil {
		t.Error("Registering a website twice should return a non-nil error")
	}
	// Wait for alert to be emitted and caught
	time.Sleep(10 * time.Millisecond)
	if len(messages) != 2 {
		t.Error("Alert was not raised (second register)")
	}

}

func TestOrchestrator_Unregister(t *testing.T) {
	setup()
	website := Website{Url: "example", CheckInterval: 100}

	// Handle alert emitting
	messages := make([]Alert, 0)
	loop := true
	defer func() { loop = false }()
	go func() {
		for loop {
			messages = append(messages, <-alerts_test)
		}
	}()

	err := orchestrator_test.Register(website)
	if err != nil {
		t.Error("Error while registering website:", err)
	}
	_, err = orchestrator_test.Start(website.Url)
	if err != nil {
		t.Error("Error while starting website:", err)
	}
	// Waiting for pinger to start
	time.Sleep(10 * time.Millisecond)
	err = orchestrator_test.Unregister(website.Url)
	if err == nil {
		t.Error("Un-registering a website while monitoring should return a non-nil error")
	}
	_, err = orchestrator_test.Pause(website.Url)
	if err != nil {
		t.Error("Error while pausing website:", err)
	}

	err = orchestrator_test.Unregister(website.Url)
	if err != nil {
		t.Error("Error while un-registering website:", err)
	}
	_, ok := orchestrator_test.aggregators[website.Url]
	if ok {
		t.Error("Aggregators were not deleted")
	}
	_, ok = orchestrator_test.reports[website.Url]
	if ok {
		t.Error("Report was not deleted")
	}
	_, ok = orchestrator_test.pingers[website.Url]
	if ok {
		t.Error("Pinger was not deleted")
	}
	// Wait for alert to be emitted and caught
	time.Sleep(10 * time.Millisecond)
	if len(messages) != 5 {
		t.Error("Alert was not raised (first unregister)")
	}

	err = orchestrator_test.Unregister(website.Url)
	if err == nil {
		t.Error("Registering a website twice should return a non-nil error")
	}
	// Wait for alert to be emitted and caught
	time.Sleep(10 * time.Millisecond)
	if len(messages) != 6 {
		t.Error("Alert was not raised (second unregister)")
	}

}

func TestOrchestrator_Start(t *testing.T) {
	setup()

	website := Website{Url: "example", CheckInterval: 100}

	// Handle alert emitting
	messages := make([]Alert, 0)
	loop := true
	defer func() { loop = false }()
	go func() {
		for loop {
			messages = append(messages, <-alerts_test)
		}
	}()

	err := orchestrator_test.Register(website)
	if err != nil {
		t.Error("Error while registering website:", err)
	}

	_, err = orchestrator_test.Start(website.Url)
	if err != nil {
		t.Error("Error while starting website:", err)
	}
	// Wait for alert to be emitted and caught
	time.Sleep(10 * time.Millisecond)
	if len(messages) != 2 {
		t.Error("Alert was not raised")
	}
	if !orchestrator_test.pingers[website.Url].IsRunning {
		t.Error("Pinger was not started")
	}
	if !orchestrator_test.reports[website.Url].Active {
		t.Error("Report is not active")
	}

	// Should not crash
	orchestrator_test.Start(website.Url)

	if !orchestrator_test.pingers[website.Url].IsRunning {
		t.Error("Pinger was not started")
	}
	if !orchestrator_test.reports[website.Url].Active {
		t.Error("Report is not active")
	}

}

func TestOrchestrator_Pause(t *testing.T) {
	setup()

	website := Website{Url: "example", CheckInterval: 100}

	// Handle alert emitting
	messages := make([]Alert, 0)
	loop := true
	defer func() { loop = false }()
	go func() {
		for loop {
			messages = append(messages, <-alerts_test)
		}
	}()

	err := orchestrator_test.Register(website)
	if err != nil {
		t.Error("Error while registering website:", err)
	}
	_, err = orchestrator_test.Start(website.Url)
	if err != nil {
		t.Error("Error while starting website:", err)
	}
	// Waiting for pinger to start
	time.Sleep(10 * time.Millisecond)

	_, err = orchestrator_test.Pause(website.Url)
	if err != nil {
		t.Error("Error while pausing website:", err)
	}
	// Waiting for pinger to stop
	time.Sleep(10 * time.Millisecond)

	if orchestrator_test.pingers[website.Url].IsRunning {
		t.Error("Pinger was not paused")
	}
	if orchestrator_test.reports[website.Url].Active {
		t.Error("Report is active")
	}

}
