package deej

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jacobsa/go-serial/serial"
	"go.uber.org/zap"

	"github.com/tomerhh/deej/pkg/deej/util"
)

// SerialIO provides a deej-aware abstraction layer for managing serial I/O
type SerialIO struct {
	comPort  string
	baudRate uint

	deej   *Deej
	logger *zap.SugaredLogger

	sliderMoveConsumers []chan SliderMoveEvent

	muteButtonsConsumer        MuteButtonConsumer
	toggleOutputDeviceConsumer ToggleOutputDeviceConsumer

	currentSliderPercentValues []float32

	conn        io.ReadWriteCloser
	connOptions serial.OpenOptions

	stopChannel chan bool
	connected   bool
}

const (
	writeTimeout   = 50 * time.Millisecond
	readTimeout    = 2 * time.Second
	commandTimeout = 3 * time.Second

	reconnectDelay = 500 * time.Millisecond
)

var (
	expectedLinePattern = regexp.MustCompile(`^\w+(\|\w+)*$`)
)

// NewSerialIO creates a SerialIO instance that uses auto-detection to find the ESP32
func NewSerialIO(deej *Deej, logger *zap.SugaredLogger) (*SerialIO, error) {
	logger = logger.Named("serial")

	sio := &SerialIO{
		deej:                       deej,
		logger:                     logger,
		sliderMoveConsumers:        []chan SliderMoveEvent{},
		currentSliderPercentValues: make([]float32, deej.config.SliderMapping.NumSliders()),
		stopChannel:                make(chan bool),
		connected:                  false,
	}

	// Initialize current slider values to -1.0 to force initial events
	for idx := range sio.currentSliderPercentValues {
		sio.currentSliderPercentValues[idx] = -1.0
	}

	logger.Debug("Created serial i/o instance")

	// Use values from config
	sio.setupSerialConnection(deej.config.SerialConnectionInfo.COMPort, deej.config.SerialConnectionInfo.BaudRate)

	// Set up config reload handling
	sio.setupOnConfigReload()

	return sio, nil
}

func (sio *SerialIO) setupSerialConnection(comPort string, baudRate uint) {
	sio.comPort = comPort
	sio.baudRate = baudRate

	sio.connOptions = serial.OpenOptions{
		PortName:              sio.comPort,
		BaudRate:              sio.baudRate,
		DataBits:              8,
		StopBits:              1,
		MinimumReadSize:       0,
		InterCharacterTimeout: 100,
	}

	sio.logger.Debugw("Set up serial connection options", "comPort", comPort, "baudRate", baudRate)
}

// Start attempts to connect to the serial port and begin reading lines
func (sio *SerialIO) Start() error {

	// If no port specified, try auto-detection
	if sio.comPort == "" || sio.comPort == "auto" {
		sio.logger.Info("Auto-detecting serial port...")
		detectedPort, err := sio.autoDetectPort()
		if err != nil {
			sio.logger.Warnw("Failed to auto-detect serial port", "error", err)

			// Notify user of auto-detect failure
			sio.deej.notifier.Notify("deej - Serial Auto-Detect Failed",
				"Could not automatically detect the serial port. Please specify a port in config.yaml.")

			return fmt.Errorf("auto-detect serial port: %w", err)
		}
		sio.logger.Infow("Auto-detected serial port", "port", detectedPort)
		sio.setupSerialConnection(detectedPort, sio.baudRate)
	}

	// Attempt first connection
	if err := sio.connect(); err != nil {
		sio.logger.Warnw("Failed initial serial connection", "error", err)

		// Notify user of initial connection failure
		sio.deej.notifier.Notify("deej - Serial Connection Failed",
			fmt.Sprintf("Could not connect to serial port %s. Check the connection and config.", sio.comPort))

		return fmt.Errorf("initial serial connection: %w", err)
	}

	// Start reading lines in a goroutine
	go sio.readLoop()

	return nil
}

// Stop signals the serial connection to stop and closes the port
func (sio *SerialIO) Stop() {
	sio.logger.Debug("Stopping serial i/o")
	sio.stopChannel <- true

	if sio.conn != nil {
		sio.conn.Close()
	}

	sio.connected = false
}

// SubscribeToSliderMoveEvents returns an unbuffered channel that receives
// slider move events as they occur
func (sio *SerialIO) SubscribeToSliderMoveEvents() chan SliderMoveEvent {
	ch := make(chan SliderMoveEvent)
	sio.sliderMoveConsumers = append(sio.sliderMoveConsumers, ch)

	return ch
}

func (sio *SerialIO) setMuteButtonClickEventConsumer(consumer MuteButtonConsumer) {
	sio.muteButtonsConsumer = consumer
}

func (sio *SerialIO) setToggleOutputDeviceEventConsumer(consumer ToggleOutputDeviceConsumer) {
	sio.toggleOutputDeviceConsumer = consumer
}

// setupOnConfigReload subscribes to config changes and updates connection settings
func (sio *SerialIO) setupOnConfigReload() {
	configReloadedChannel := sio.deej.config.SubscribeToChanges()

	const stopDelay = 50 * time.Millisecond

	go func() {
		for {
			select {
			case <-configReloadedChannel:
				// Reset current slider values to force slider move events
				go func() {
					<-time.After(stopDelay)
					for idx := range sio.currentSliderPercentValues {
						sio.currentSliderPercentValues[idx] = -1.0
					}
				}()

				// If connection params have changed, update connection options
				newPort := sio.deej.config.SerialConnectionInfo.COMPort
				newBaud := sio.deej.config.SerialConnectionInfo.BaudRate

				if newPort != sio.comPort || newBaud != sio.baudRate {
					sio.logger.Infow("Serial config changed, updating connection",
						"oldPort", sio.comPort,
						"newPort", newPort,
						"oldBaud", sio.baudRate,
						"newBaud", newBaud)

					sio.setupSerialConnection(newPort, newBaud)

					// Reconnect with new settings
					if sio.connected && sio.conn != nil {
						sio.conn.Close()
						sio.connected = false

						time.Sleep(reconnectDelay)
						if err := sio.connect(); err != nil {
							sio.logger.Warnw("Failed to reconnect with new settings", "error", err)
						}
					}
				}
			}
		}
	}()
}

// autoDetectPort attempts to find the ESP32 by scanning available COM ports
func (sio *SerialIO) autoDetectPort() (string, error) {
	// On Windows, scan COM3-COM16
	possiblePorts := []string{
		"COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "COM10",
		"COM11", "COM12", "COM13", "COM14", "COM15", "COM16",
	}

	sio.logger.Debug("Scanning for available COM ports...")

	for _, port := range possiblePorts {
		sio.logger.Debugw("Trying port", "port", port)

		// Try to open the port
		testOptions := serial.OpenOptions{
			PortName:              port,
			BaudRate:              sio.baudRate,
			DataBits:              8,
			StopBits:              1,
			MinimumReadSize:       0,
			InterCharacterTimeout: 100,
		}

		conn, err := serial.Open(testOptions)
		if err != nil {
			// Port doesn't exist or is in use
			continue
		}

		// Port opened successfully, try to read some data
		sio.logger.Debugw("Port opened, testing for ESP32 data", "port", port)

		// Wait a moment for data
		time.Sleep(500 * time.Millisecond)

		reader := bufio.NewReader(conn)
		reader.ReadString('\n') // Discard first line (might be partial)

		// Try to read a valid line
		line, err := reader.ReadString('\n')
		conn.Close()

		if err == nil && sio.isValidLine(line) {
			sio.logger.Infow("Found ESP32", "port", port)
			return port, nil
		}
	}

	return "", errors.New("no valid serial port found")
}

// connect attempts to establish a serial connection
func (sio *SerialIO) connect() error {
	sio.logger.Debugw("Attempting serial connection", "port", sio.comPort, "baud", sio.baudRate)

	conn, err := serial.Open(sio.connOptions)
	if err != nil {
		return fmt.Errorf("open serial port: %w", err)
	}

	// Set the connection
	sio.conn = conn
	sio.connected = true

	sio.logger.Infow("Connected to serial port", "port", sio.comPort)

	return nil
}

// readLoop continuously reads lines from the serial port
func (sio *SerialIO) readLoop() {
	sio.logger.Debug("Started read loop")
	reader := bufio.NewReader(sio.conn)

	for {
		select {
		case <-sio.stopChannel:
			sio.logger.Debug("Stopped read loop")
			return
		default:
			// Read until newline
			line, err := reader.ReadString('\n')

			if err != nil {
				if err == io.EOF {
					sio.logger.Warn("Serial connection closed, attempting reconnect...")
				} else {
					sio.logger.Warnw("Error reading from serial", "error", err)
				}

				sio.connected = false

				// Try to reconnect
				time.Sleep(reconnectDelay)
				if reconnErr := sio.connect(); reconnErr != nil {
					sio.logger.Warnw("Reconnection failed", "error", reconnErr)
					continue
				}

				// Recreate reader after reconnection
				reader = bufio.NewReader(sio.conn)
				continue
			}

			// Trim and process the line
			line = strings.TrimSpace(line)

			if line != "" && sio.isValidLine(line) {
				sio.handleLine(line)
			} else if line != "" {
				sio.logger.Debugw("Invalid line format, discarding", "line", line)
			}
		}
	}
}

// isValidLine checks if a line matches the expected protocol format
func (sio *SerialIO) isValidLine(line string) bool {
	line = strings.TrimSpace(line)
	return expectedLinePattern.MatchString(line)
}

// handleLine processes a received line and dispatches to appropriate handlers
func (sio *SerialIO) handleLine(line string) {
	sio.logger.Debugw("Received line", "line", line)

	parts := strings.Split(line, "|")
	if len(parts) == 0 {
		return
	}

	command := parts[0]
	data := parts[1:]

	switch command {
	case "Sliders":
		sio.handleSliders(data)
	case "MuteButtons":
		sio.handleMuteButtons(data)
	case "SwitchOutput":
		sio.handleSwitchOutput(data)
	case "GetCurrentOutputDevice":
		sio.handleGetCurrentOutputDevice()
	default:
		sio.logger.Debugw("Unknown command", "command", command)
	}
}

// handleSliders processes slider data and sends move events
func (sio *SerialIO) handleSliders(data []string) {
	numSliders := len(data)

	if numSliders != sio.deej.config.SliderMapping.NumSliders() {
		sio.logger.Warnw("Received unexpected number of sliders",
			"expected", sio.deej.config.SliderMapping.NumSliders(),
			"received", numSliders)
		// Send OK anyway
		sio.sendResponse("OK")
		return
	}

	moveEvents := []SliderMoveEvent{}

	for sliderIdx, stringValue := range data {
		number, err := strconv.Atoi(stringValue)
		if err != nil {
			sio.logger.Warnw("Invalid slider value", "value", stringValue, "error", err)
			continue
		}

		// Validate raw values don't exceed 4095 - normalize if they do
		var dirtyFloat float32
		if number > 4095 {
			dirtyFloat = 1.0
			sio.logger.Debugw("Got value > 4095, normalizing to 1.0", "value", number, "slider", sliderIdx)
		} else if number < 0 {
			dirtyFloat = 0.0
		} else {
			// Normalize 12-bit ADC (0-4095) to 0.0-1.0
			dirtyFloat = float32(number) / 4095.0
		}

		// Normalize to 2 points of precision using util function
		normalizedScalar := util.NormalizeScalar(dirtyFloat)

		// Apply invert if configured
		if sio.deej.config.InvertSliders {
			normalizedScalar = 1.0 - normalizedScalar
		}

		// Check if significantly different (noise reduction)
		if util.SignificantlyDifferent(sio.currentSliderPercentValues[sliderIdx], normalizedScalar, sio.deej.config.NoiseReductionLevel) {

			// Update current value
			sio.currentSliderPercentValues[sliderIdx] = normalizedScalar

			// Create move event
			moveEvents = append(moveEvents, SliderMoveEvent{
				SliderID:     sliderIdx,
				PercentValue: normalizedScalar,
			})

			if sio.deej.Verbose() {
				sio.logger.Debugw("Slider moved", "event", moveEvents[len(moveEvents)-1])
			}
		}
	}

	// Broadcast events to all consumers
	if len(moveEvents) > 0 {
		for _, consumer := range sio.sliderMoveConsumers {
			for _, moveEvent := range moveEvents {
				consumer <- moveEvent
			}
		}
	}

	// Send OK response
	sio.sendResponse("OK")
}

// handleMuteButtons processes mute button data and sends state back
func (sio *SerialIO) handleMuteButtons(data []string) {
	if sio.muteButtonsConsumer == nil {
		sio.logger.Warn("No mute button consumer registered")
		sio.sendResponse("ERROR")
		return
	}

	events := []MuteButtonClickEvent{}

	for buttonIdx, stringValue := range data {
		muteState, err := strconv.ParseBool(stringValue)
		if err != nil {
			sio.logger.Warnw("Invalid mute button value", "value", stringValue, "error", err)
			continue
		}
		events = append(events, MuteButtonClickEvent{
			MuteButtonID: buttonIdx,
			mute:         muteState,
		})

		if sio.deej.Verbose() {
			sio.logger.Debugw("Mute button clicked", "event", events[len(events)-1])
		}
	}

	if len(events) == 0 {
		sio.sendResponse("ERROR")
		return
	}

	// Call consumer to get actual state
	newState, err := sio.muteButtonsConsumer(events)
	if err != nil {
		sio.logger.Warnw("Error handling mute buttons", "error", err)
		sio.sendResponse("ERROR")
		return
	}

	// Build response with actual mute states
	responseParts := []string{"MuteState"}
	for _, muted := range newState.MuteButtons {
		if muted {
			responseParts = append(responseParts, "true")
		} else {
			responseParts = append(responseParts, "false")
		}
	}

	response := strings.Join(responseParts, "|")
	sio.sendResponse(response)
}

// handleSwitchOutput processes output device switching
func (sio *SerialIO) handleSwitchOutput(data []string) {
	if sio.toggleOutputDeviceConsumer == nil {
		sio.logger.Warn("No toggle output device consumer registered")
		sio.sendResponse("ERROR")
		return
	}

	if len(data) == 0 {
		sio.logger.Warn("No device index provided for SwitchOutput")
		sio.sendResponse("ERROR")
		return
	}

	deviceIdx, err := strconv.Atoi(data[0])
	if err != nil {
		sio.logger.Warnw("Invalid device index", "value", data[0], "error", err)
		sio.sendResponse("ERROR")
		return
	}

	event := ToggleOutoutDeviceClickEvent{
		selectedOutputDevice: deviceIdx,
	}

	if sio.deej.Verbose() {
		sio.logger.Debugw("Output device switch requested", "deviceIdx", deviceIdx)
	}

	// Call consumer to get actual state
	newState, err := sio.toggleOutputDeviceConsumer(event)
	if err != nil {
		sio.logger.Warnw("Error handling output device switch", "error", err)
		sio.sendResponse("ERROR")
		return
	}

	// Send response with actual device index
	response := fmt.Sprintf("OutputDevice|%d", newState.selectedOutputDevice)
	sio.sendResponse(response)
}

// handleGetCurrentOutputDevice sends the current output device index
func (sio *SerialIO) handleGetCurrentOutputDevice() {
	if sio.toggleOutputDeviceConsumer == nil {
		sio.logger.Warn("No toggle output device consumer registered")
		sio.sendResponse("ERROR")
		return
	}

	// Call with a query event (negative index means query)
	event := ToggleOutoutDeviceClickEvent{
		selectedOutputDevice: -1,
	}

	newState, err := sio.toggleOutputDeviceConsumer(event)
	if err != nil {
		sio.logger.Warnw("Error getting current output device", "error", err)
		sio.sendResponse("ERROR")
		return
	}

	if sio.deej.Verbose() {
		sio.logger.Debugw("Current output device queried", "deviceIdx", newState.selectedOutputDevice)
	}

	response := fmt.Sprintf("OutputDevice|%d", newState.selectedOutputDevice)
	sio.sendResponse(response)
}

// sendResponse writes a response to the serial port
func (sio *SerialIO) sendResponse(response string) {
	if !sio.connected || sio.conn == nil {
		sio.logger.Warn("Cannot send response: not connected")
		return
	}

	responseWithNewline := response + "\n"

	_, err := sio.conn.Write([]byte(responseWithNewline))
	if err != nil {
		sio.logger.Warnw("Error writing response", "error", err, "response", response)
		return
	}

	sio.logger.Debugw("Sent response", "response", response)
}
