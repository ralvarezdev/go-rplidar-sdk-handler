package go_rplidar_sdk_handler

import (
	"math"
)

// GetAverageDistanceFromAngle calculates the average distance for a given list of angles.
//
// Parameters:
//
// measures: A pointer to an array of 360 Measure pointers indexed by angle.
// middleAngle: The middle angle to start the averaging from.
// width: The sum of the angles to consider with both sides and the middle angle.
//
// Returns:
//
// The average distance for the specified angles, or an error if the width is not valid.
func GetAverageDistanceFromAngle(
	measures *[360]*Measure,
	middleAngle int,
	width int,
) (float64, error) {
	var totalDistance float64
	var count int

	// Calculate the range of angles to consider
	if width%2 == 0 {
		return 0, ErrAngleWidthMustBeOdd
	}
	if width < 1 {
		return 0, ErrAngleWidthTooSmall
	}
	if width >= 360 {
		return 0, ErrAngleWidthTooLarge
	}

	// Check if the width is 1, in which case we only consider the middle angle
	if width == 1 {
		measure := measures[middleAngle]
		if measure == nil {
			return 0.0, nil
		}
		return measure.GetDistance(), nil
	}

	// Calculate the angles to consider
	var angles []int
	widthPerSide := (width - 1) / 2
	leftAngle := middleAngle - widthPerSide
	rightAngle := middleAngle + widthPerSide
	if leftAngle < 0 {
		for angle := 360 + leftAngle; angle < 360; angle++ {
			angles = append(angles, angle)
		}
	}
	if rightAngle >= 360 {
		for angle := 0; angle <= rightAngle-360; angle++ {
			angles = append(angles, angle)
		}
	}
	for angle := max(leftAngle, 0); angle <= min(360, rightAngle); angle++ {
		angles = append(angles, angle)
	}

	// Calculate the average distance
	for _, angle := range angles {
		measure := measures[angle]
		if measure == nil {
			continue
		}

		// Check the distance and quality
		if measure.GetDistance() == 0.0 || measure.GetQuality() == 0 {
			continue
		}

		totalDistance += measure.GetDistance()
		count++
	}
	return totalDistance / float64(count), nil
}

// GetAverageDistanceFromDirection calculates the average distance for a given direction.
//
// Parameters:
//
// measures: A pointer to an array of 360 Measure pointers indexed by angle.
// width: The sum of the angles to consider with both sides and the middle angle.
// direction: The direction to calculate the average distance for.
//
// Returns:
//
// The average distance for the specified direction, or an error if the direction is not valid.
func GetAverageDistanceFromDirection(
	measures *[360]*Measure,
	width int,
	direction CardinalDirection,
) (float64, error) {
	directionAngle := direction.Angle()

	// Round the angle
	if directionAngle >= 180 {
		directionAngle = math.Ceil(directionAngle)
	} else {
		directionAngle = math.Floor(directionAngle)
	}

	return GetAverageDistanceFromAngle(
		measures,
		int(directionAngle),
		width,
	)
}

// GetAverageDistancesFromDirections calculates the average distances for the specified directions.
//
// Parameters:
//
// measures: A pointer to an array of 360 Measure pointers indexed by angle.
// width: The sum of the angles to consider with both sides and the middle angle.
// directions: The directions to calculate the average distances for.
//
// Returns:
//
// A map with directions as keys and their average distances as values, or an error if any direction is not valid.
func GetAverageDistancesFromDirections(
	measures *[360]*Measure,
	width int,
	directions ...CardinalDirection,
) (map[CardinalDirection]float64, error) {
	avgDistances := make(map[CardinalDirection]float64)
	for _, direction := range directions {
		avgDistance, err := GetAverageDistanceFromDirection(
			measures, width, direction,
		)
		if err != nil {
			return nil, err
		}
		avgDistances[direction] = avgDistance
	}
	return avgDistances, nil
}