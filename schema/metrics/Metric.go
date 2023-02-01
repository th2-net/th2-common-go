package metrics

type Metric interface {
	isEnabled() bool
	enable()
	disable()
}
