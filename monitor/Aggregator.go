package monitor

import (
	"fmt"
	"math"
	"sync"
	"time"
)

const SHORT_INTERVAL = 2 * time.Minute
const MEDIUM_INTERVAL = 10 * time.Minute
const LONG_INTERAVL = time.Hour

type Aggregator struct {
	duration    time.Duration
	website     string
	last        *QueueElement
	first       *QueueElement
	heap        MaxHeap
	count       int
	errorCount  int
	sumResTime  int64 // TODO - change this (less than 64bits is needed)
	statusCount map[int]int
	statusAgg   map[int]int
	AlertStatus bool
	mutex       sync.Mutex
}

type Aggregators struct {
	Short  *Aggregator // 2 min
	Medium *Aggregator // 10 min
	Long   *Aggregator // 1 hour
}

type QueueElement struct {
	Value     *PingLog
	Timestamp time.Time
	next      *QueueElement
	heapIndex int
}

func NewAggregators(website string) Aggregators {
	short := Aggregator{
		duration:    SHORT_INTERVAL,
		website:     website,
		last:        nil,
		first:       nil,
		statusCount: make(map[int]int),
		statusAgg:   make(map[int]int),
		AlertStatus: false,
		heap:        NewMaxHeap(),
	}
	medium := Aggregator{
		duration:    MEDIUM_INTERVAL,
		website:     website,
		last:        nil,
		first:       nil,
		statusCount: make(map[int]int),
		statusAgg:   make(map[int]int),
		AlertStatus: false,
		heap:        NewMaxHeap(),
	}
	long := Aggregator{
		duration:    LONG_INTERAVL,
		website:     website,
		last:        nil,
		first:       nil,
		statusCount: make(map[int]int),
		statusAgg:   make(map[int]int),
		AlertStatus: false,
		heap:        NewMaxHeap(),
	}
	return Aggregators{
		Short:  &short,
		Medium: &medium,
		Long:   &long,
	}
}

// Aggregate PingLog and update metrics
func (a *Aggregator) Add(e QueueElement) (error, Alert) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	// Eventual alert
	var alert Alert

	// Add e to queue
	if a.first != nil {
		a.first.next = &e
		a.first = &e
	} else {
		a.first = &e
	}

	// Add &e to MaxHeap
	a.heap.insert(&e)

	// Update metrics
	if a.first.Value.Error != nil {
		a.errorCount++
	} else {
		a.count++
		a.sumResTime += int64(math.Floor(a.first.Value.ResponseTime.Seconds() * 1000))
		a.statusCount[a.first.Value.Status]++
		a.statusAgg[a.first.Value.Status/100]++
	}

	if a.last == nil {
		a.last = &e
	}

	// Dequeue and delete from heap outdated PingLogs, and update metrics
	for a.last.Timestamp.Add(a.duration).Before(a.first.Timestamp) {
		last := *a.last
		if last.Value.Error != nil {
			a.errorCount--
		} else {
			a.count--
			a.sumResTime -= int64(math.Floor(last.Value.ResponseTime.Seconds() * 1000))
			a.statusCount[last.Value.Status]--
			a.statusAgg[last.Value.Status/100]--
		}
		a.heap.delete(a.last)
		a.last = a.last.next

	}

	// Emit alerts on availability crash / resuming
	availability := float32(a.statusCount[200]) / float32(a.count+a.errorCount)
	if availability <= 0.80 && a.AlertStatus == false {
		a.AlertStatus = true
		timestamp := a.first.Timestamp
		formattedTimestamp := formatTime(timestamp)
		alert = Alert{
			Init:      true,
			Url:       a.website,
			Timestamp: timestamp,
			Value:     availability,
			Message:   fmt.Sprint("[Website ", a.website, " is down !\n Availability: ", availability, "%; time: ", formattedTimestamp, "](fg-red)"),
		}
	} else if availability > 0.8 && a.AlertStatus == true {
		a.AlertStatus = false
		timestamp := a.first.Timestamp
		formattedTimestamp := formatTime(timestamp)
		alert = Alert{
			Init:      true,
			Url:       a.website,
			Timestamp: timestamp,
			Value:     availability,
			Message:   fmt.Sprint("[Website ", a.website, " is up again !\n Availability: ", availability, "%; time: ", formattedTimestamp, "](fg-green)"),
		}
	}

	return nil, alert
}

func (a *Aggregator) GetAvgResTime() (float32, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return float32(a.sumResTime) / float32(a.count), nil
}

func (a *Aggregator) GetStatusCount(status int) (int, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.statusCount[status], nil
}

func (a *Aggregator) GetAvailability() (float32, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	out := float32(a.statusCount[200]) / float32(a.count+a.errorCount)

	return out, nil
}

func (a *Aggregator) GetStatusShare(status int) (float32, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	out := float32(a.statusCount[status]) / float32(a.count)

	return out, nil
}

func (a *Aggregator) GetStatusAgg(firstDigit int) (float32, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	out := float32(a.statusAgg[firstDigit]) / float32(a.count)

	return out, nil
}

func (a *Aggregator) GetMaxResTime() (float32, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	out := float32(math.Floor(a.heap.getMax().Value.ResponseTime.Seconds() * 1000))
	return out, nil
}

func (a *Aggregator) GetErrorCount() (int, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.errorCount, nil
}

func (a *Aggregator) GetCount() (int, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.count, nil
}

func (a *Aggregator) GetErrorShare() (float32, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	out := float32(a.errorCount) / float32(a.count)

	return out, nil
}

func (a *Aggregator) GetDuration() time.Duration {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.duration
}

func formatTime(time time.Time) string {
	return time.String()[:19]
}
