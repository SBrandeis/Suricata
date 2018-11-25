package monitor

import (
	"fmt"
	"testing"
	"time"
)

// Testing on a scenario (pingLogs)
func TestAggregator_Alerts(t *testing.T) {
	var shouldBeAlert = []bool{false, false, false, true, false, false, true, true}

	agg := NewAggregators("http://www.example.com")

	queueEls := make([]QueueElement, 0)

	// Generate queue elements holding the logs
	for _, log := range pingLogs {
		queueEls = append(queueEls, QueueElement{
			Timestamp: log.Time,
			Value:     log,
			next:      nil,
		})
	}

	// Aggregate each element
	for idx, qEl := range queueEls {
		err, alert := agg.Short.Add(qEl)
		if err != nil {
			t.Error(err)
		}
		if alert.Init {
			if !shouldBeAlert[idx] {
				t.Error("Log #", idx, "should not raise an alert, but got:", alert)
			}
		} else {
			if shouldBeAlert[idx] {
				t.Error("Log #", idx, "should have raised an alert, but got:", alert)
			}
		}
		availability, err := agg.Short.GetAvailability()
		if err != nil {
			t.Error("An error occurred while retrieving Availability:", err)
		}
		fmt.Println("Alert raised:", alert.Init, "Availability:", availability)
	}
}

func TestAggregator_Count(t *testing.T) {
	agg := NewAggregators("http://www.example.com")

	expectedCount := []int{1, 2, 3, 3, 3, 3, 2, 2}

	queueEls := make([]QueueElement, 0)

	// Generate queue elements holding the logs
	for _, log := range pingLogs {
		queueEls = append(queueEls, QueueElement{
			Timestamp: log.Time,
			Value:     log,
			next:      nil,
		})
	}

	for idx, qEl := range queueEls {
		err, _ := agg.Short.Add(qEl)
		if err != nil {
			t.Error("An error occurred while aggregating log:", err)
		}
		count, err := agg.Short.GetCount()
		if err != nil {
			t.Error("An error occurred while retrieving count:", err)
		}
		t.Log("Actual count:", count, "Expected count:", expectedCount[idx])
		if count != expectedCount[idx] {
			t.Error("Count does not match expected output")
		}

	}
}

func TestAggregator_GetAvgResTime(t *testing.T) {
	agg := NewAggregators("http://www.example.com")

	expectedAvgRT := []float32{50., 50., 50., 60., 60., 60., 50., 50.}

	queueEls := make([]QueueElement, 0)

	// Generate queue elements holding the logs
	for _, log := range pingLogs {
		queueEls = append(queueEls, QueueElement{
			Timestamp: log.Time,
			Value:     log,
			next:      nil,
		})
	}

	for idx, qEl := range queueEls {
		err, _ := agg.Short.Add(qEl)
		if err != nil {
			t.Error("An error occurred while aggregating log:", err)
		}
		avgRT, err := agg.Short.GetAvgResTime()
		if err != nil {
			t.Error("An error occurred while retrieving count:", err)
		}
		t.Log("Actual avg res time:", avgRT, "Expected avg res time:", expectedAvgRT[idx])
		if avgRT != expectedAvgRT[idx] {
			t.Error("AvgRT does not match expected avgRT")
		}

	}
}

func TestAggregator_GetMaxResTime(t *testing.T) {
	agg := NewAggregators("http://www.example.com")

	// TODO - Implement this
	expectedAvgRT := []float32{50., 50., 50., 80., 80., 80., 50., 50.}

	queueEls := make([]QueueElement, 0)

	// Generate queue elements holding the logs
	for _, log := range pingLogs {
		queueEls = append(queueEls, QueueElement{
			Timestamp: log.Time,
			Value:     log,
			next:      nil,
		})
	}

	for idx, qEl := range queueEls {
		err, _ := agg.Short.Add(qEl)
		if err != nil {
			t.Error("An error occurred while aggregating log:", err)
		}
		t.Log("heap:")
		c, err := agg.Short.GetCount()
		for i, e := range agg.Short.heap.values {
			t.Log("index:", i, "value:", e.Value.ResponseTime, "storedIndex:", e.heapIndex, "count:", c)
		}
		maxRT, err := agg.Short.GetMaxResTime()
		if err != nil {
			t.Error("An error occurred while retrieving max res time:", err)
		}
		t.Log("Actual max res time:", maxRT, "Expected max res time:", expectedAvgRT[idx])
		if maxRT != expectedAvgRT[idx] {
			t.Error("maxRT does not match expected maxRT")
		}

	}

}

func TestAggregator_GetStatusCount(t *testing.T) {
	agg := NewAggregators("http://www.example.com")

	expected200Counts := []int{1, 2, 3, 2, 2, 2, 2, 1}
	expected500Counts := []int{0, 0, 0, 1, 1, 1, 0, 0}
	expected401Counts := []int{0, 0, 0, 0, 0, 0, 0, 1}
	expectedCount := []int{1, 2, 3, 3, 3, 3, 2, 2}

	queueEls := make([]QueueElement, 0)

	// Generate queue elements holding the logs
	for _, log := range pingLogs {
		queueEls = append(queueEls, QueueElement{
			Timestamp: log.Time,
			Value:     log,
			next:      nil,
		})
	}

	for idx, qEl := range queueEls {
		err, _ := agg.Short.Add(qEl)
		if err != nil {
			t.Error("An error occurred while aggregating log:", err)
		}
		count200, err := agg.Short.GetStatusCount(200)
		if err != nil {
			t.Error("An error occurred while retrieving count of status 200:", err)
		}
		count500, err := agg.Short.GetStatusCount(500)
		if err != nil {
			t.Error("An error occurred while retrieving count of status 500:", err)
		}

		count401, err := agg.Short.GetStatusCount(401)
		if err != nil {
			t.Error("An error occurred while retrieving count of status 401:", err)
		}

		count, err := agg.Short.GetCount()
		if err != nil {
			t.Error("An error occurred while retrieving count:", err)
		}
		t.Log("Actual 200 count:", count200, "Expected 200 count:", expected200Counts[idx])
		t.Log("Actual 500 count:", count500, "Expected 500 count:", expected500Counts[idx])
		t.Log("Actual 401 count:", count401, "Expected 401 count:", expected401Counts[idx])
		t.Log("Actual count:", count, "Expected count:", expectedCount[idx])

		if count200 != expected200Counts[idx] {
			t.Error("Count of status 200 does not match expectation")
		}
		if count500 != expected500Counts[idx] {
			t.Error("Count of status 500 does not match expectation")
		}
		if count401 != expected401Counts[idx] {
			t.Error("Count of status 401 does not match expectation")
		}
		if count != expectedCount[idx] {
			t.Error("Count of requests does not match expectation")
		}

	}

}

var pingLogs = []*PingLog{
	{
		Status:       200,
		Error:        nil,
		ResponseTime: 50 * time.Millisecond,
		Time:         time.Date(2018, 11, 11, 11, 10, 0, 0, time.Local),
	},
	{
		Status:       200,
		Error:        nil,
		ResponseTime: 50 * time.Millisecond,
		Time:         time.Date(2018, 11, 11, 11, 11, 0, 0, time.Local),
	},
	{
		Status:       200,
		Error:        nil,
		ResponseTime: 50 * time.Millisecond,
		Time:         time.Date(2018, 11, 11, 11, 12, 0, 0, time.Local),
	},
	// This ping should raise an alert: availability under 80 %
	{
		Status:       500,
		Error:        nil,
		ResponseTime: 80 * time.Millisecond,
		Time:         time.Date(2018, 11, 11, 11, 13, 0, 0, time.Local),
	},
	{
		Status:       200,
		Error:        nil,
		ResponseTime: 50 * time.Millisecond,
		Time:         time.Date(2018, 11, 11, 11, 14, 0, 0, time.Local),
	},
	// This ping should raise an alert: availability resumes
	{
		Status:       200,
		Error:        nil,
		ResponseTime: 50 * time.Millisecond,
		Time:         time.Date(2018, 11, 11, 11, 15, 0, 0, time.Local),
	},
	{
		Status:       200,
		Error:        nil,
		ResponseTime: 50 * time.Millisecond,
		Time:         time.Date(2018, 11, 11, 11, 16, 30, 0, time.Local),
	},
	{
		Status:       401,
		Error:        nil,
		ResponseTime: 50 * time.Millisecond,
		Time:         time.Date(2018, 11, 11, 11, 17, 30, 0, time.Local),
	},
}
