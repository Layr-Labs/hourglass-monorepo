package metrics

type MetricsContext interface {
	Emit(name string, value int)
}
