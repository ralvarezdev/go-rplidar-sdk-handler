package go_rplidar_sdk_handler

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"

	gostringsconvert "github.com/ralvarezdev/go-strings/convert"
	goconcurrentlogger "github.com/ralvarezdev/go-concurrent-logger"
)

type (
	// RotationCompleted is a signal sent when a full rotation is completed
	RotationCompleted struct{}

	// Measure is a struct that represents a single measurement from the RPLiDAR.
	Measure struct {
		angle      float64
		distance   float64
		quality    int
		hasSyncBit bool
	}

	// DefaultHandler is the handler for the Slamtec RPLiDAR devices
	DefaultHandler struct {
		handlerMutex          sync.Mutex
		measuresMutex         sync.RWMutex
		isRunning             atomic.Bool
		logger                goconcurrentlogger.Logger
		handlerLoggerProducer goconcurrentlogger.LoggerProducer
		baudRate              int
		isUpsideDown          bool
		angleAdjustment       float64
		measures              [360]*Measure
		stdoutLinesRead       int
		ultraSimplePath      string
		maxDistanceLimit    float64
		port string
	}
)

// validateAngle validates the angle value.
//
// Parameters:
//
// angle: Angle value to validate.
// hasSyncBit: Indicates if the measurement has a sync bit.
//
// Returns:
//
// An error if the angle is invalid.
func validateAngle(angle float64, hasSyncBit bool) error {
	// Check if the angle corresponds to a measure with sync bit
	if !hasSyncBit {
		if angle < 0 || angle >= 360 {
			return fmt.Errorf(
				"angle without sync bit must be in [0, 360), got %f",
				angle,
			)
		}
	} else if angle < 0 {
		return fmt.Errorf(
			"angle with sync bit must be non-negative, got %f",
			angle,
		)
	}
	return nil
}

// NewMeasure creates a new Measure instance.
//
// Parameters:
//
// angle: Angle of the measurement in degrees.
// distance: Distance of the measurement in millimeters.
// quality: Quality of the measurement.
// hasSyncBit: Indicates if the measurement has a sync bit.
// isUpsideDown: Indicates if the LIDAR is upside down.
// angleAdjustment: Angle adjustment to apply to the angle.
//
// Returns:
//
// A Measure instance, or an error if any parameter is invalid.
func NewMeasure(
	angle, distance float64,
	quality int,
	hasSyncBit bool,
	isUpsideDown bool,
	angleAdjustment float64,
) (*Measure, error) {
	// Validate angle
	if err := validateAngle(angle, hasSyncBit); err != nil {
		return nil, err
	}

	// Ensure the angle is between 0 and 360 if it has a sync bit
	if hasSyncBit {
		angle = angle - 360.0
	}

	// Adjust angle if the LIDAR is upside down
	if isUpsideDown {
		angle = 360.0 - angle
	}

	// Apply angle adjustment
	if angleAdjustment != 0 {
		angle = angle + angleAdjustment
	}

	// Normalize angle to be within [0, 360)
	if angle < 0 {
		angle = angle + 360.0
	} else if angle >= 360.0 {
		angle = angle - 360.0
	}

	return &Measure{
		angle:      angle,
		distance:   distance,
		quality:    quality,
		hasSyncBit: hasSyncBit,
	}, nil
}

// NewMeasureFromString creates a new Measure instance from a string representation of the measurement.
//
// Parameters:
//
// measureStr: String representation of the measurement.
// isUpsideDown: Indicates if the RPLiDAR is upside down.
// angleAdjustment: Angle adjustment to apply to the angle.
//
// Returns:
//
// A Measure instance, or an error if the string is invalid.
func NewMeasureFromString(
	measureStr string,
	isUpsideDown bool,
	angleAdjustment float64,
) (*Measure, error) {
	// Trim and split
	fields := strings.Fields(measureStr)

	// Check if it has sync bit
	hasSyncBit := false
	if len(fields) == 4 && fields[0] == SyncBitCharacter {
		hasSyncBit = true
		fields = fields[1:]
	}

	// Check number of fields
	if len(fields) != 3 {
		return nil, fmt.Errorf("expected 3 fields, got %d", len(fields))
	}

	// Parse fields
	var angle float64
	if err := gostringsconvert.ToFloat64(
		fields[AngleIndex],
		&angle,
	); err != nil {
		return nil, fmt.Errorf("failed to parse angle: %w", err)
	}

	var distance float64
	if err := gostringsconvert.ToFloat64(
		fields[DistanceIndex],
		&distance,
	); err != nil {
		return nil, fmt.Errorf("failed to parse distance: %w", err)
	}

	var quality int
	if err := gostringsconvert.ToInt(
		fields[QualityIndex],
		&quality,
	); err != nil {
		return nil, fmt.Errorf("failed to parse quality: %w", err)
	}

	// Create the Measure instance
	return NewMeasure(
		angle,
		distance,
		quality,
		hasSyncBit,
		isUpsideDown,
		angleAdjustment,
	)
}

// GetAngle returns the angle of the measurement.
//
// Returns:
//
// The angle of the measurement in degrees.
func (m *Measure) GetAngle() float64 {
	return m.angle
}

// GetDistance returns the distance of the measurement.
//
// Returns:
//
// The distance of the measurement in millimeters.
func (m *Measure) GetDistance() float64 {
	return m.distance
}

// GetQuality returns the quality of the measurement.
//
// Returns:
//
// The quality of the measurement.
func (m *Measure) GetQuality() int {
	return m.quality
}

// String returns the string representation of the Measure.
//
// Returns:
//
// The string representation of the Measure.
func (m *Measure) String() string {
	return fmt.Sprintf(
		"%f%s%f%s%d",
		m.angle,
		AttributesSeparator,
		m.distance,
		AttributesSeparator,
		m.quality,
	)
}

// IsRotationCompleted determines if a full rotation has been completed
//
// Returns:
//
// True if a full rotation has been completed, false otherwise.
func (m *Measure) IsRotationCompleted() bool {
	return m.hasSyncBit
}

// NewDefaultHandler creates a new DefaultHandler instance.
//
// Parameters:
//
// baudRate: Baud rate for the serial communication.
// port: SerialCommunication port for the RPLiDAR.
// isUpsideDown: If true, the RPLiDAR is upside down, and angles will be adjusted accordingly.
// angleAdjustment: Optional angle adjustment to apply to the angles.
// logger: Logger instance for logging messages.
// ultraSimplePath: Path to the ultra_simple executable.
// maxDistanceLimit: Maximum distance limit for valid measurements.
//
// Returns:
//
// A pointer to a DefaultHandler instance or an error if any parameter is invalid.
func NewDefaultHandler(
	baudRate int,
	port string,
	isUpsideDown bool,
	angleAdjustment float64,
	logger goconcurrentlogger.Logger,
	ultraSimplePath      string,
	maxDistanceLimit    float64,
) (*DefaultHandler, error) {
	// Check if the logger is nil
	if logger == nil {
		return nil, goconcurrentlogger.ErrNilLogger
	}

	// Check if the ultra simple path is empty
	if strings.TrimSpace(ultraSimplePath) == "" {
		return nil, ErrEmptyUltraSimplePath
	}

	// Check if the max distance limit is valid
	if maxDistanceLimit <= 0 {
		return nil, ErrInvalidMaxDistanceLimit
	}

	// Create a new DefaultHandler instance
	handler := &DefaultHandler{
		logger:          logger,
		baudRate:        baudRate,
		port:            port,
		isUpsideDown:    isUpsideDown,
		angleAdjustment: angleAdjustment,
		ultraSimplePath:      ultraSimplePath,
		maxDistanceLimit:    maxDistanceLimit,
	}

	return handler, nil
}

// NewSlamtecC1Handler creates a new DefaultHandler instance configured for the Slamtec RPLiDAR C1 model.
//
// Parameters:
//
// port: SerialCommunication port for the RPLiDAR C1.
// isUpsideDown: If true, the RPLiDAR is upside down, and angles will be adjusted accordingly.
// angleAdjustment: Optional angle adjustment to apply to the angles.
// logger: Logger instance for logging messages.
// ultraSimplePath: Path to the ultra_simple executable.
// maxDistanceLimit: Maximum distance limit for valid measurements.
//
// Returns:
//
// A pointer to a DefaultHandler instance or an error if any parameter is invalid.
func NewSlamtecC1Handler(
	port string,
	isUpsideDown bool,
	angleAdjustment float64,
	logger goconcurrentlogger.Logger,
	ultraSimplePath      string,
	maxDistanceLimit    float64,
) (*DefaultHandler, error) {
	return NewDefaultHandler(
		SlamtecC1BaudRate,
		port,
		isUpsideDown,
		angleAdjustment,
		logger,
		ultraSimplePath,
		maxDistanceLimit,
	)
}

// IsRunning checks if the handler is currently running.
//
// Returns:
//
// True if the handler is running, false otherwise.
func (h *DefaultHandler) IsRunning() bool {
	return h.isRunning.Load()
}

// runToWrap is the internal function to read incoming measures from the RPLiDAR and process them.
//
// Parameters:
//
// ctx: Context for managing cancellation and timeouts.
// stopFn: Function to stop the context in case of an error.
//
// Returns:
//
// An error if any issue occurs during reading or processing measures.
func (h *DefaultHandler) runToWrap(ctx context.Context, stopFn func()) error {
	// Initialize the measures slice
	h.measures = [360]*Measure{}

	// Reset the stdout lines read counter
	h.stdoutLinesRead = 0

	// Log the start of reading measures
	h.handlerLoggerProducer.Info(HandlerStartedMessage)

	// Check if the ultra simple executable exists
	if _, err := os.Stat(h.ultraSimplePath); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf(
			"ultra simple executable not found at path: %s",
			h.ultraSimplePath,
		)
	}

	// Arguments (do not include the executable itself)
	args := []string{
		UltraSimpleChannelArgument,
		UltraSimpleSerialArgument,
		h.port,
		strconv.Itoa(h.baudRate),
	}

	// Execute the command with a context
	cmd := exec.CommandContext(ctx, h.ultraSimplePath, args...)

	// Stream output line by line
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe error: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe error: %w", err)
	}

	// Start the command
	if err = cmd.Start(); err != nil {
		return fmt.Errorf("start command error: %w", err)
	}

	// Create an error group to wait for all goroutines to finish
	g := &errgroup.Group{}

	// Stream stdout
	g.Go(
		goconcurrentlogger.StopContextAndLogOnError(
			ctx,
			stopFn, 
			func(ctx context.Context) error {
				return h.scanLines(
					ctx,
					StdoutTag,
					stdout,
					h.handleStdoutLine,
				)
			},
			h.handlerLoggerProducer,
		),
	)

	// Stream stderr
	g.Go(
		goconcurrentlogger.StopContextAndLogOnError(
			ctx,
			stopFn, 
			func(ctx context.Context) error {
				return h.scanLines(
					ctx,
					StderrTag,
					stderr,
					h.handleStderrLine,
				)
			},
			h.handlerLoggerProducer,
		),
	)

	// Wait for completion or context cancel
	if err = g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		h.handlerLoggerProducer.Warning(fmt.Sprintf("Error reading lines: %v", err))
		return err
	}

	// Close the stdout and stderr pipes
	if err = stdout.Close(); err != nil {
		return fmt.Errorf("stdout close error: %w", err)
	}
	if err = stderr.Close(); err != nil {
		return fmt.Errorf("stderr close error: %w", err)
	}

	// Signal the process to stop
	_ = cmd.Process.Signal(os.Interrupt) // Unix

	// Sleep for a grace period to allow clean shutdown
	time.Sleep(CloseTimeout)

	// Hard kill fallback if still running after grace period
	if cmd.ProcessState == nil || !cmd.ProcessState.Exited() {
		_ = cmd.Process.Kill()
	}
	return nil

}

// Run reads incoming measures from the RPLiDAR and processes them.
//
// Parameters:
//
// ctx: Context for managing cancellation and timeouts.
// stopFn: Function to stop the context in case of an error.
//
// Returns:
//
// An error if any issue occurs during reading or processing measures.
func (h *DefaultHandler) Run(ctx context.Context, stopFn func()) error {
	h.handlerMutex.Lock()

	// Check if it's already running
	if h.IsRunning() {
		h.handlerMutex.Unlock()
		return ErrHandlerAlreadyRunning
	}
	defer func() {
		h.handlerMutex.Lock()

		// Set running to false
		h.isRunning.Store(false)

		h.handlerMutex.Unlock()
	}()

	// Set running to true
	h.isRunning.Store(true)

	h.handlerMutex.Unlock()

	// Create a logger producer
	handlerLoggerProducer, err := h.logger.NewProducer(
		HandlerLoggerProducerTag,
	)
	if err != nil {
		return fmt.Errorf("failed to create handler logger producer: %w", err)
	}
	h.handlerLoggerProducer = handlerLoggerProducer
	defer h.handlerLoggerProducer.Close()

	return goconcurrentlogger.LogOnError(
		func() error {
			return h.runToWrap(ctx, stopFn)
		},
		h.handlerLoggerProducer,
	)
}

// scanLines reads lines from the provided reader and processes them using the given lineHandler.
//
// Parameters:
//
// ctx: Context for managing cancellation and timeouts.
// tag: Tag to identify the source of the lines (e.g., "stdout" or "stderr").
// r: Reader to read lines from.
// lineHandler: Function to process each line.
//
// Returns:
//
// An error if any issue occurs during reading or processing lines.
func (h *DefaultHandler) scanLines(
	ctx context.Context,
	tag string,
	r interface{ Read([]byte) (int, error) },
	lineHandler func(string) error,
) error {
	// Check if the lineHandler is nil
	if lineHandler == nil {
		return ErrNilLineHandler
	}

	// Create a new scanner
	sc := bufio.NewScanner(r)

	// Set the buffer size
	buf := make([]byte, 0, InitialSizeBuffer)
	sc.Buffer(buf, MaxSizeBuffer)

	for sc.Scan() {
		select {
		case <-ctx.Done():
			h.handlerLoggerProducer.Info(
				fmt.Sprintf(
					"Context done while reading lines from %s: %v",
					tag,
					ctx.Err(),
				),
			)
			// Return context error
			return ctx.Err()
		default:
			// Read the line
			line := strings.TrimSpace(sc.Text())

			// Process the line
			if h.handlerLoggerProducer.IsDebug() {
				h.handlerLoggerProducer.Debug(
					fmt.Sprintf(
						"Received line from %s: %s",
						tag,
						line,
					),
				)
			}

			// Handle the line
			if err := lineHandler(line); err != nil {
				return err
			}
		}
	}

	// Check for scanning errors
	if err := sc.Err(); err != nil {
		return fmt.Errorf("scan error: %w", err)
	}
	return nil
}

// handleStdoutLine processes a single line from stdout.
//
// Parameters:
//
// line: The line to process.
//
// Returns:
//
// An error if any issue occurs during processing the line.
func (h *DefaultHandler) handleStdoutLine(line string) error {
	// Increment the stdout lines read counter
	h.stdoutLinesRead++

	// Check if the message should be ignored
	if h.stdoutLinesRead <= IgnoreFirstStdoutMessages {
		return nil
	}

	// Create a measure from the given string
	measure, err := NewMeasureFromString(
		line,
		h.isUpsideDown,
		h.angleAdjustment,
	)
	if err != nil {
		h.handlerLoggerProducer.Warning(
			fmt.Sprintf(
				"Failed to parse measure: %v",
				err,
			),
		)
		return nil // Ignore parsing errors
	}

	// Check if the RPLiDAR has completed a full rotation
	if measure.IsRotationCompleted() {
		h.handlerLoggerProducer.Info("Full rotation completed.")
	}

	// Check if the distance is within the maximum limit
	if measure.GetDistance() < 0 || measure.GetDistance() > h.maxDistanceLimit {
		return nil // Ignore out-of-range distances
	}

	// Lock the measures for writing
	h.measuresMutex.Lock()

	// Store the measure in the measures
	angle := int(measure.GetAngle()) % 360
	h.measures[angle] = measure

	// Unlock the measures
	h.measuresMutex.Unlock()
	return nil
}

// GetMeasures returns a copy of the current measures.
//
// Returns:
//
// A copy of the current measures.
func (h *DefaultHandler) GetMeasures() *[360]*Measure {
	// Lock the measures for reading
	h.measuresMutex.RLock()
	defer h.measuresMutex.RUnlock()

	// Create a copy of the measures
	measuresCopy := [360]*Measure{}
	copy(measuresCopy[:], h.measures[:])
	return &measuresCopy
}

// GetAverageDistanceFromAngle calculates the average distance for a given angle.
//
// Parameters:
//
// midleAngle: The middle angle to calculate the average distance for.
// width: The sum of the angles to consider with both sides and the middle angle.
//
// Returns:
//
// The average distance for the specified angle, or an error if the angle is not valid.
func (h *DefaultHandler) GetAverageDistanceFromAngle(
	middleAngle int,
	width int,
) (float64, error) {
	// Get the current measures
	measures := h.GetMeasures()

	return GetAverageDistanceFromAngle(
		measures,
		middleAngle,
		width,
	)
}

// GetAverageDistanceFromDirection calculates the average distance for a given direction.
//
// Parameters:
//
// width: The sum of the angles to consider with both sides and the middle angle.
// direction: The direction to calculate the average distance for.
//
// Returns:
//
// The average distance for the specified direction, or an error if the direction is not valid.
func (h *DefaultHandler) GetAverageDistanceFromDirection(
	width int,
	direction CardinalDirection,
) (float64, error) {
	// Get the current measures
	measures := h.GetMeasures()

	return GetAverageDistanceFromDirection(
		measures,
		width,
		direction,
	)
}

// GetAverageDistancesFromDirections calculates the average distances for the specified directions.
//
// Parameters:
//
// width: The sum of the angles to consider with both sides and the middle angle.
// directions: The directions to calculate the average distances for.
//
// Returns:
//
// A map with directions as keys and their average distances as values, or an error if any direction is not valid.
func (h *DefaultHandler) GetAverageDistancesFromDirections(
	width int,
	directions ...CardinalDirection,
) (map[CardinalDirection]float64, error) {	
	// Get the current measures
	measures := h.GetMeasures()
	
	return GetAverageDistancesFromDirections(
		measures,
		width,
		directions...,
	)
}

// GetAverageDistancesFromAllDirections calculates the average distances for all cardinal directions.
//
// Parameters:
//
// width: The sum of the angles to consider with both sides and the middle angle.
//
// Returns:
//
// A map with all cardinal directions as keys and their average distances as values, or an error if any direction is not valid.
func (h *DefaultHandler) GetAverageDistancesFromAllDirections(
	width int,
) (map[CardinalDirection]float64, error) {
	// Get the current measures
	measures := h.GetMeasures()

	return GetAverageDistanceFromAllDirections(
		measures,
		width,
	)
}

// handleStderrLine processes a single line from stderr.
//
// Parameters:
//
// line: The line to process.
//
// Returns:
//
// An error if any issue occurs during processing the line.
func (h *DefaultHandler) handleStderrLine(line string) error {
	// Log the stderr line as a warning
	h.handlerLoggerProducer.Warning(fmt.Sprintf("stderr: %s", line))
	return nil
}