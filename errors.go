package go_rplidar_sdk_handler

import (
	"errors"
)

var (
	ErrNilHandler            = errors.New("handler cannot be nil")
	ErrNilLineHandler        = errors.New("line handler cannot be nil")
	ErrHandlerAlreadyRunning = errors.New("handler is already running")
	ErrEmptyUltraSimplePath   = errors.New("ultra_simple path cannot be empty")
	ErrInvalidMaxDistanceLimit = errors.New("max distance limit must be greater than zero")
	ErrAngleWidthMustBeOdd          = errors.New("angle width must be odd")
	ErrAngleWidthTooSmall           = errors.New("angle width must be greater than 0")
	ErrAngleWidthTooLarge           = errors.New("angle width must be less than 360 degrees")
)
