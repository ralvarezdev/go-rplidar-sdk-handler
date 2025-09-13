package go_rplidar_sdk_handler

import (
	"context"
)

type (
	// Handler is the interface to handle the RPLiDAR devices
	Handler interface {
		Run(ctx context.Context, stopFn func()) error
		IsRunning() bool
		GetMeasures() *[360]*Measure
		GetRotationCompletedChannel() <-chan RotationCompleted
	}
)
