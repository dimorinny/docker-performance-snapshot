package main

type (
	Metric struct {
		Timestamp int64
		Value     float64
	}

	Reporter interface {
		Report(name string, description string, metrics []*Metric) error
	}
)
