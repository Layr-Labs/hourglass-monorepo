package metrics

import "fmt"

type StdOutMetricsContext struct {
}

func NewStdOutMetricsContext() *StdOutMetricsContext {
	return &StdOutMetricsContext{}
}

func (m StdOutMetricsContext) Emit(name string, value int) {
	fmt.Printf("Metric: %s Value: %d\n", name, value)
}
