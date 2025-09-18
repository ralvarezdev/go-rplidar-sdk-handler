package go_rplidar_sdk_handler

type (
	// CardinalDirection is an enum to represent the different cardinal directions that the RPLiDAR can face.
	CardinalDirection uint8
)

const (
	CardinalDirectionNil CardinalDirection = iota
	CardinalDirectionNorth
	CardinalDirectionWest
	CardinalDirectionEast
	CardinalDirectionSouth
	CardinalDirectionNorthwest
	CardinalDirectionNortheast
	CardinalDirectionSouthwest
	CardinalDirectionSoutheast
	CardinalDirectionWestNorthwest
	CardinalDirectionNorthNorthwest
	CardinalDirectionEastNortheast
	CardinalDirectionNorthNortheast
	CardinalDirectionWestSouthwest
	CardinalDirectionEastSoutheast
	CardinalDirectionSouthSouthwest
	CardinalDirectionSouthSoutheast
)

var (
	// CardinalDirectionNames maps a given CardinalDirection to its string name
	CardinalDirectionNames = map[CardinalDirection]string{
		CardinalDirectionNorth:          "north",
		CardinalDirectionWest:           "west",
		CardinalDirectionEast:           "east",
		CardinalDirectionSouth:          "south",
		CardinalDirectionNorthwest:      "northwest",
		CardinalDirectionNortheast:      "northeast",
		CardinalDirectionSouthwest:      "southwest",
		CardinalDirectionSoutheast:      "southeast",
		CardinalDirectionWestNorthwest:  "west-northwest",
		CardinalDirectionNorthNorthwest: "north-northwest",
		CardinalDirectionEastNortheast:  "east-northeast",
		CardinalDirectionNorthNortheast: "north-northeast",
		CardinalDirectionWestSouthwest:  "west-southwest",
		CardinalDirectionEastSoutheast:  "east-southeast",
		CardinalDirectionSouthSouthwest: "south-southwest",
		CardinalDirectionSouthSoutheast: "south-southeast",
	}

	// CardinalDirectionAngles maps a given CardinalDirection to its angle in degrees
	CardinalDirectionAngles = map[CardinalDirection]float64{
		CardinalDirectionNorth:          0.0,
		CardinalDirectionNorthNortheast: 22.5,
		CardinalDirectionNortheast:      45.0,
		CardinalDirectionEastNortheast:  67.5,
		CardinalDirectionEast:           90.0,
		CardinalDirectionEastSoutheast:  112.5,
		CardinalDirectionSoutheast:      135.0,
		CardinalDirectionSouthSoutheast: 157.5,
		CardinalDirectionSouth:          180.0,
		CardinalDirectionSouthSouthwest: 202.5,
		CardinalDirectionSouthwest:      225.0,
		CardinalDirectionWestSouthwest:  247.5,
		CardinalDirectionWest:           270.0,
		CardinalDirectionWestNorthwest:  292.5,
		CardinalDirectionNorthwest:      315.0,
		CardinalDirectionNorthNorthwest: 337.5,
	}

	// CardinalDirections is a slice of all valid CardinalDirection values
	CardinalDirections = []CardinalDirection{
		CardinalDirectionNorth,
		CardinalDirectionWest,
		CardinalDirectionEast,
		CardinalDirectionSouth,
		CardinalDirectionNorthwest,
		CardinalDirectionNortheast,
		CardinalDirectionSouthwest,
		CardinalDirectionSoutheast,
		CardinalDirectionWestNorthwest,
		CardinalDirectionNorthNorthwest,
		CardinalDirectionEastNortheast,
		CardinalDirectionNorthNortheast,
		CardinalDirectionWestSouthwest,
		CardinalDirectionEastSoutheast,
		CardinalDirectionSouthSouthwest,
		CardinalDirectionSouthSoutheast,
	}
)

// String returns the string representation of the CardinalDirection
//
// Returns:
//
// The string representation of the CardinalDirection enum
func (r CardinalDirection) String() string {
	return CardinalDirectionNames[r]
}

// Angle returns the angle in degrees of the CardinalDirection
//
// Returns:
//
// The angle in degrees of the CardinalDirection enum
func (r CardinalDirection) Angle() float64 {
	return CardinalDirectionAngles[r]
}
