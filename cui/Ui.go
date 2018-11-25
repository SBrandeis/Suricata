package cui

import (
	ui "github.com/gizak/termui"
	"suricata/monitor"
	"sync"
)

type Display struct {
	messages *ui.List
	info     *ui.List
	measures *ui.Table
}

var (
	disp *Display
	once sync.Once
)

//Singleton
func GetDisplay() *Display {
	once.Do(func() {
		disp = &Display{
			messages: newMsgHolder(),
			info:     newInfoHolder(),
			measures: newMeasures(),
		}
	})
	return disp
}

// Update Messages component
func (u *Display) UpdateMessages(messages []string) error {
	msg := newMsgHolder()
	msg.Items = messages
	u.messages = msg
	return nil
}

// Update Info component
func (u *Display) UpdateInfo(controls []string) error {
	newControls := newInfoHolder()
	newControls.Items = controls
	u.info = newControls
	return nil
}

// Update Measures Table
func (u *Display) UpdateMeasures(websites []string, o *monitor.Orchestrator) error {
	measures := newMeasures()

	rows := [][]string{
		// Headers of the report table
		{"website", "period", "average response", "max response time", "availability", "[2XX](fg-green)", "[5XX](fg-red)", "[4XX](fg-yellow)", "[Unsuccessful %](fg-magenta)"},
	}

	// Populate table with data from Summary
	for _, website := range websites {
		summary, err := o.GetSummary(website)

		if err != nil {
			return err
		}
		rows = append(rows, summary...)
	}

	measures.Rows = rows

	measures.Analysis()
	measures.SetSize()
	u.measures = measures
	return nil
}

// Displays components
func (u Display) Render() ([]*ui.Row, error) {
	display := make([]*ui.Row, 0)
	display = append(display, []*ui.Row{
		ui.NewRow(
			ui.NewCol(8, 0, u.messages),
			ui.NewCol(4, 0, u.info),
		),
		ui.NewRow(ui.NewCol(12, 0, u.measures)),
	}...)
	return display, nil
}

func newMsgHolder() *ui.List {
	msg := ui.NewList()
	msg.Overflow = "hidden"
	msg.ItemFgColor = ui.ColorWhite
	msg.Border = true
	msg.BorderLabel = "Messages"
	msg.Height = 11
	return msg
}

func newInfoHolder() *ui.List {
	ctl := ui.NewList()
	ctl.Items = []string{"Loading..."}
	ctl.BorderLabel = "Controls"
	ctl.BorderFg = ui.ColorWhite
	ctl.ItemFgColor = ui.ColorYellow
	ctl.Height = 11
	return ctl
}

func newMeasures() *ui.Table {
	measures := ui.NewTable()
	measures.FgColor = ui.ColorWhite
	measures.TextAlign = ui.AlignLeft
	measures.Separator = true
	measures.Border = true
	return measures
}
