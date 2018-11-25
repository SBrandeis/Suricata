package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	ui "github.com/gizak/termui"
	"log"
	"os"
	"strconv"
	"strings"
	"suricata/cui"
	"suricata/monitor"
	"time"
)

//
const DEFAULT_CHECKING_INTERVAL int = 1000
const REFRESH_INTERVAL_MEDIUM = time.Second * 10
const REFRESH_INTERVAL_LONG = time.Minute

// Info to display to the user
var info = []string{
	"press q to quit",
	"press s to resume monitoring",
	"press p to pause monitoring",
	"availability: % of status code 200",
	"Unsuccessful requests can have timed out",
}

// Loads ./config.sample by default
var configFile = flag.String("cfg", "./config.sample", "Config file containing the websites to monitor and the check inbtervals")

var websites []monitor.Website
var urls []string

func main() {
	flag.Parse()

	w, u, err := parseConfig(*configFile)
	if err != nil {
		panic("Failed to read config file !")
	}
	websites = w
	urls = u

	err = ui.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer ui.Close()

	pipeline := make(chan monitor.PingLog)
	alerts := make(chan monitor.Alert)
	messages := make([]string, 0)

	orchestrator := monitor.GetOrchestrator(pipeline, alerts)
	display := cui.GetDisplay()

	go func(display *cui.Display) {
		stopTick := time.NewTimer(30 * time.Minute)
		mediumTick := time.NewTicker(REFRESH_INTERVAL_MEDIUM)
		longTick := time.NewTicker(REFRESH_INTERVAL_LONG)
		loop := true

		for loop {
			select {
			// Every 10s, update medium term data
			case <-mediumTick.C:
				updateMedium(orchestrator)
				display.UpdateMeasures(urls, orchestrator)
				render(display)

				// Every 1mn, update long term data
			case <-longTick.C:
				updateLong(orchestrator)
				display.UpdateMeasures(urls, orchestrator)
				render(display)

			case alert := <-alerts:
				// TODO - write message in log file
				messages = append(messages, alert.Message)
				if len(messages) > 8 {
					messages = messages[len(messages)-8:]
				}
				display.UpdateMessages(messages)
				render(display)

			case <-stopTick.C:
				ui.StopLoop()
				loop = false
			}
		}
	}(display)

	launch(orchestrator, websites)
	defer stop(orchestrator)

	orchestrator.StartAll()

	display.UpdateInfo(info)
	display.UpdateMessages(messages)
	display.UpdateMeasures(urls, orchestrator)
	render(display)

	// Launch orchestrator and register websites
	ui.Handle("q", func(ui.Event) {
		ui.StopLoop()
	})

	ui.Handle("s", func(ui.Event) {
		orchestrator.StartAll()
		display.UpdateMeasures(urls, orchestrator)
		render(display)
	})

	ui.Handle("p", func(ui.Event) {
		orchestrator.PauseAll()
		display.UpdateMeasures(urls, orchestrator)
		render(display)
	})

	ui.Handle("<Resize>", func(e ui.Event) {
		payload := e.Payload.(ui.Resize)
		ui.Body.Width = payload.Width
		ui.Body.Align()
		ui.Clear()
		ui.Render(ui.Body)
	})

	ui.Loop()
}

func launch(orchestrator *monitor.Orchestrator, websites []monitor.Website) {
	for _, website := range websites {
		err := orchestrator.Register(website)
		if err != nil {
			fmt.Println("ERROR - ", err)
		}
	}
}

func stop(orchestrator *monitor.Orchestrator) {
	err := orchestrator.PauseAll()
	if err != nil {
		panic(err)
	}
	err = orchestrator.UnregisterAll()
	if err != nil {
		panic(err)
	}
}

func updateMedium(orchestrator *monitor.Orchestrator) error {
	for _, website := range websites {
		err := orchestrator.UpdateMediumReport(website.Url)
		if err != nil {
			return err
		}
	}
	return nil
}

func updateLong(orchestrator *monitor.Orchestrator) error {
	for _, website := range websites {
		err := orchestrator.UpdateLongReport(website.Url)
		if err != nil {
			return err
		}
	}
	return nil
}

func render(display *cui.Display) error {
	ui.Body.Rows = make([]*ui.Row, 0)
	rows, err := display.Render()
	ui.Body.AddRows(rows...)
	ui.Body.Align()
	ui.Render(ui.Body)
	return err
}

func parseConfig(fileLocation string) ([]monitor.Website, []string, error) {
	file, err := os.Open(fileLocation)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	websites := make([]monitor.Website, 0)
	urls := make([]string, 0)
	foundUrls := make(map[string]bool)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var url string
		var interval int

		line := scanner.Text()
		params := strings.Split(line, ",")
		if len(params) > 2 {
			return nil, nil, errors.New("INVALID CONFIG FILE: 1 WEBSITE PER lINE")
		}

		if !(strings.Index(params[0], "http://") == 0 || strings.Index(params[0], "https://") == 0) {
			url = "http://" + params[0]
		} else {
			url = params[0]
		}

		if len(params) == 2 {
			interval, err = strconv.Atoi(params[1])
			if err != nil {
				return nil, nil, errors.New("INVALID CONFIG FILE: INTERVAL MUST BE INTEGER")
			}
		} else {
			interval = DEFAULT_CHECKING_INTERVAL
		}

		website := monitor.Website{Url: url, CheckInterval: interval}
		if _, ok := foundUrls[url]; !ok {
			websites = append(websites, website)
			urls = append(urls, url)
			foundUrls[url] = true
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}
	return websites, urls, nil
}
