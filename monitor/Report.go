package monitor

import (
	"fmt"
	"math"
)

type Measures struct {
	Period           string
	AvgRes           float32
	MaxRes           float32
	Share2XX         float32
	Share5XX         float32
	Share3XX         float32
	Share4XX         float32
	Availability     float32
	UnsuccessfulRate float32
}

type Report struct {
	Url           string
	Active        bool
	CheckInterval int
	ShortTerm     Measures
	MediumTerm    Measures
	LongTerm      Measures
}

func NewReport(website Website) (*Report, error) {
	report := Report{
		Active:        false,
		Url:           website.Url,
		CheckInterval: website.CheckInterval,
		ShortTerm: Measures{
			Period:           "Past 2 min",
			AvgRes:           -1.,
			MaxRes:           -1.,
			Share2XX:         -1.,
			Share5XX:         -1.,
			Share3XX:         -1.,
			Share4XX:         -1.,
			Availability:     -1.,
			UnsuccessfulRate: -1.,
		},
		MediumTerm: Measures{
			Period:           "Past 10 min",
			AvgRes:           -1.,
			MaxRes:           -1.,
			Share2XX:         -1.,
			Share5XX:         -1.,
			Share3XX:         -1.,
			Share4XX:         -1.,
			Availability:     -1.,
			UnsuccessfulRate: -1.,
		},
		LongTerm: Measures{
			Period:           "Past 1 hour",
			AvgRes:           -1.,
			MaxRes:           -1.,
			Share2XX:         -1.,
			Share5XX:         -1.,
			Share3XX:         -1.,
			Share4XX:         -1.,
			Availability:     -1.,
			UnsuccessfulRate: -1.,
		},
	}
	return &report, nil
}

func (r *Report) Summary() [][]string {

	header := fmt.Sprint("[", r.Url, "](fg-bold)")
	if !r.Active {
		header = fmt.Sprint(header, " [(sleeping)](fg-yellow)")
	}
	summary := [][]string{
		{
			header,
			"[" + r.MediumTerm.Period + "](fg-bold)",
			formatMs(r.MediumTerm.AvgRes, 100.),
			formatMs(r.MediumTerm.MaxRes, 800.),
			formatShare(r.MediumTerm.Availability, 0.8, 1.),
			formatShare(r.MediumTerm.Share2XX, 0.8, 1.),
			formatShare(r.MediumTerm.Share5XX, 0., 0.05),
			formatShare(r.MediumTerm.Share4XX, 0., 0.05),
			formatShare(r.MediumTerm.UnsuccessfulRate, 0, 0.05),
		},
		{
			fmt.Sprint("check interval: [", r.CheckInterval, " ms](fg-bold)"),
			"[" + r.LongTerm.Period + "](fg-bold)",
			formatMs(r.LongTerm.AvgRes, 100.),
			formatMs(r.LongTerm.MaxRes, 800.),
			formatShare(r.LongTerm.Availability, 0.8, 1.),
			formatShare(r.LongTerm.Share2XX, 0.8, 1.),
			formatShare(r.LongTerm.Share5XX, 0., 0.05),
			formatShare(r.LongTerm.Share4XX, 0., 0.05),
			formatShare(r.LongTerm.UnsuccessfulRate, 0, 0.05),
		},
	}

	return summary
}

func (m *Measures) Update(aggregator *Aggregator) error {
	availability, err := aggregator.GetAvailability()
	if err != nil {
		return err
	}
	m.Availability = availability

	avgRes, err := aggregator.GetAvgResTime()
	if err != nil {
		return err
	}
	m.AvgRes = avgRes

	maxRes, err := aggregator.GetMaxResTime()
	if err != nil {
		return err
	}
	m.MaxRes = maxRes

	share5XX, err := aggregator.GetStatusAgg(5)
	if err != nil {
		return err
	}
	m.Share5XX = share5XX

	share2XX, err := aggregator.GetStatusAgg(2)
	if err != nil {
		return err
	}
	m.Share2XX = share2XX

	share4XX, err := aggregator.GetStatusAgg(4)
	if err != nil {
		return err
	}
	m.Share4XX = share4XX

	share3XX, err := aggregator.GetStatusAgg(3)
	if err != nil {
		return err
	}
	m.Share3XX = share3XX

	errCount, err := aggregator.GetErrorCount()
	if err != nil {
		return err
	}
	count, err := aggregator.GetCount()
	if err != nil {
		return err
	}

	m.UnsuccessfulRate = float32(errCount) / float32(errCount+count)
	return nil
}

func formatShare(value float32, low float32, high float32) string {
	if value < 0. {
		return "collecting..."
	}
	str := fmt.Sprint(math.Floor(float64(value)*100), " %")
	if value < low {
		str = fmt.Sprint("[", str, "](fg-red)")
		return str
	}
	if value > high {
		str = fmt.Sprint("[", str, "](fg-red)")
		return str
	}
	return str

}

func formatMs(value float32, high float32) string {
	if value < 0. {
		return "collecting..."
	}
	str := fmt.Sprint(math.Floor(float64(value)*100)/100, " ms")
	if value > high {
		str = fmt.Sprint("[", str, "](fg-red)")
	}
	return str
}
