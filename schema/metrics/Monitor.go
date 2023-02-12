package metrics

type Monitor struct {
	Name        string
	FlagArbiter *FlagArbiter
}

func (mon *Monitor) IsEnabled() bool {
	return !mon.FlagArbiter.disabled.contains(mon.Name)
}

func (mon *Monitor) Enable() {
	mon.FlagArbiter.EnableMonitor(mon.Name)
}

func (mon *Monitor) Disable() {
	mon.FlagArbiter.DisableMonitor(mon.Name)
}
