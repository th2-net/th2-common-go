package metrics

type Flag interface {
	IsEnabled() bool
	Enable()
	Disable()
}
