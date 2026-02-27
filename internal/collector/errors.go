package collector

import "fmt"

// CollectorError wraps an error with the collector that produced it.
type CollectorError struct {
	Collector string
	Err       error
}

func (e *CollectorError) Error() string {
	return fmt.Sprintf("%s: %v", e.Collector, e.Err)
}

func (e *CollectorError) Unwrap() error {
	return e.Err
}
