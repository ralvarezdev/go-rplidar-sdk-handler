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
		GetAverageDistanceFromAngle(
			middleAngle int,
			width int,
		) (float64, error)
		GetAverageDistanceFromDirection(
			width int,
			direction CardinalDirection,
		) (float64, error)
		GetAverageDistancesFromDirections(
			width int,
			directions ...CardinalDirection,
		) (map[CardinalDirection]float64, error)
	}
)
