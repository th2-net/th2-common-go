package metrics

type Metric interface {
	IsEnabled() bool
	Enable()
	Disable()
}
