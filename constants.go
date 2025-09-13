package go_rplidar_sdk_handler

import (
	"time"
)

const (
	// SlamtecC1BaudRate is the RPLiDAR C1 baud rate
	SlamtecC1BaudRate = 460800

	// HandlerStartedMessage is the message logged when the handler starts
	HandlerStartedMessage = "RPLiDAR handler started"

	// CloseTimeout is the timeout for closing the handler
	CloseTimeout = 5 * time.Second

	// UltraSimpleChannelArgument is the argument for the channel in the ultra_simple executable
	UltraSimpleChannelArgument = "--channel"

	// UltraSimpleSerialArgument is the argument for the serial port in the ultra_simple executable
	UltraSimpleSerialArgument = "--serial"

	// SyncBitCharacter is the sync bit character
	SyncBitCharacter = "S"

	// AngleIndex is the index of the angle in the measure string
	AngleIndex = 0

	// DistanceIndex is the index of the distance in the measure string
	DistanceIndex = 1

	// QualityIndex is the index of the quality in the measure string
	QualityIndex = 2
)

var (
	// LinuxSlamtecC1Port is the RPLiDAR C1 default port in Linux systems
	LinuxSlamtecC1Port = "/dev/ttyUSB0"

	// InitialSizeBuffer is the initial size of the buffer for reading lines
	InitialSizeBuffer = 1024 * 1024 // 1 MB

	// MaxSizeBuffer is the maximum size of the buffer for reading lines
	MaxSizeBuffer = 1024 * 1024 * 10 // 10 MB

	// StdoutTag is the tag for standard output logs
	StdoutTag = "STDOUT"

	// StderrTag is the tag for standard error logs
	StderrTag = "STDERR"

	// IgnoreFirstStdoutMessages is the number of initial stdout messages to ignore
	IgnoreFirstStdoutMessages = 6

	// HandlerLoggerProducerTag is the logger producer tag for RPLiDAR
	HandlerLoggerProducerTag = "RPLiDAR_HANDLER"

	// AttributesSeparator is the attributes separator
	AttributesSeparator = ","
)
