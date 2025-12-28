package deej

import (
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

// mockSerialConnection simulates a serial connection for testing
type mockSerialConnection struct {
	writeBuffer []string
	readData    []string
	readIndex   int
	closed      bool
}

func (m *mockSerialConnection) Read(p []byte) (n int, err error) {
	if m.readIndex >= len(m.readData) {
		time.Sleep(100 * time.Millisecond)
		return 0, io.EOF
	}

	data := m.readData[m.readIndex]
	m.readIndex++

	copy(p, []byte(data+"\n"))
	return len(data) + 1, nil
}

func (m *mockSerialConnection) Write(p []byte) (n int, err error) {
	if m.closed {
		return 0, io.ErrClosedPipe
	}
	m.writeBuffer = append(m.writeBuffer, string(p))
	return len(p), nil
}

func (m *mockSerialConnection) Close() error {
	m.closed = true
	return nil
}

// Helper to create test config file
func createTestConfig(t *testing.T, content string) func() {
	if err := os.WriteFile("config.yaml", []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
	return func() {
		os.Remove("config.yaml")
	}
}

// Helper function for absolute value
func absFloat(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}

// TestSliderDataParsing tests slider data parsing and normalization
func TestSliderDataParsing(t *testing.T) {
	logger := zap.NewNop().Sugar()
	notifier := &mockNotifier{}

	configContent := `
slider_mapping:
  0: master
  1: chrome
  2: discord
  3: spotify
  4: system
serial_connection_info:
  com_port: "COM4"
  baud_rate: 115200
`
	cleanup := createTestConfig(t, configContent)
	defer cleanup()

	config, err := NewConfig(logger, notifier)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	deej := &Deej{
		config:      config,
		logger:      logger,
		notifier:    notifier,
		stopChannel: make(chan bool),
	}

	sio, err := NewSerialIO(deej, logger)
	if err != nil {
		t.Fatalf("Failed to create SerialIO: %v", err)
	}

	// Subscribe to slider events
	eventChan := sio.SubscribeToSliderMoveEvents()

	// Test slider data parsing
	sliderData := []string{"4095", "2048", "1024", "512", "0"}
	expectedValues := []float32{1.0, 0.5, 0.25, 0.125, 0.0}

	// Handle sliders directly
	sio.handleSliders(sliderData)

	// Collect events with timeout
	receivedEvents := make(map[int]float32)
	timeout := time.After(500 * time.Millisecond)

eventLoop:
	for i := 0; i < len(sliderData); i++ {
		select {
		case event := <-eventChan:
			receivedEvents[event.SliderID] = event.PercentValue
		case <-timeout:
			break eventLoop
		}
	}

	// Verify events
	for i, expected := range expectedValues {
		received, ok := receivedEvents[i]
		if !ok {
			t.Errorf("No event received for slider %d", i)
			continue
		}
		if diff := absFloat(received - expected); diff > 0.01 {
			t.Errorf("Slider %d: expected value ~%.2f, got %.2f (diff: %.3f)", i, expected, received, diff)
		}
	}
}

// TestMuteButtonHandling tests mute button request/response handling
func TestMuteButtonHandling(t *testing.T) {
	logger := zap.NewNop().Sugar()
	notifier := &mockNotifier{}

	configContent := `
slider_mapping:
  0: master
mute_button_mapping:
  0: chrome
  1: discord
serial_connection_info:
  com_port: "COM4"
  baud_rate: 115200
`
	cleanup := createTestConfig(t, configContent)
	defer cleanup()

	config, err := NewConfig(logger, notifier)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	deej := &Deej{
		config:      config,
		logger:      logger,
		notifier:    notifier,
		stopChannel: make(chan bool),
	}

	sio, err := NewSerialIO(deej, logger)
	if err != nil {
		t.Fatalf("Failed to create SerialIO: %v", err)
	}

	// Create mock connection
	mockConn := &mockSerialConnection{
		writeBuffer: []string{},
	}
	sio.conn = mockConn
	sio.connected = true

	// Set up a mock consumer
	consumerCalled := false
	sio.setMuteButtonClickEventConsumer(func(events []MuteButtonClickEvent) (MuteButtonsState, error) {
		consumerCalled = true
		state := MuteButtonsState{
			MuteButtons: make([]bool, len(events)),
		}
		for i, event := range events {
			state.MuteButtons[i] = event.mute
		}
		return state, nil
	})

	// Handle mute button data
	muteData := []string{"true", "false"}
	sio.handleMuteButtons(muteData)

	if !consumerCalled {
		t.Error("Consumer was not called")
	}

	// Check response
	if len(mockConn.writeBuffer) == 0 {
		t.Fatal("No response was written")
	}

	response := strings.TrimSpace(mockConn.writeBuffer[0])
	expectedResponse := "MuteState|true|false"
	if response != expectedResponse {
		t.Errorf("Expected response '%s', got '%s'", expectedResponse, response)
	}
}

// TestDeviceSwitchHandling tests device switch request/response handling
func TestDeviceSwitchHandling(t *testing.T) {
	logger := zap.NewNop().Sugar()
	notifier := &mockNotifier{}

	configContent := `
slider_mapping:
  0: master
available_output_device:
  0: ["Speakers"]
  1: ["Headphones"]
serial_connection_info:
  com_port: "COM4"
  baud_rate: 115200
`
	cleanup := createTestConfig(t, configContent)
	defer cleanup()

	config, err := NewConfig(logger, notifier)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	deej := &Deej{
		config:      config,
		logger:      logger,
		notifier:    notifier,
		stopChannel: make(chan bool),
	}

	sio, err := NewSerialIO(deej, logger)
	if err != nil {
		t.Fatalf("Failed to create SerialIO: %v", err)
	}

	// Create mock connection
	mockConn := &mockSerialConnection{
		writeBuffer: []string{},
	}
	sio.conn = mockConn
	sio.connected = true

	// Set up a mock consumer
	consumerCalled := false
	sio.setToggleOutputDeviceEventConsumer(func(event ToggleOutoutDeviceClickEvent) (OutputDeviceState, error) {
		consumerCalled = true
		return OutputDeviceState{selectedOutputDevice: event.selectedOutputDevice}, nil
	})

	// Handle device switch
	deviceData := []string{"1"}
	sio.handleSwitchOutput(deviceData)

	if !consumerCalled {
		t.Error("Consumer was not called")
	}

	// Check response
	if len(mockConn.writeBuffer) == 0 {
		t.Fatal("No response was written")
	}

	response := strings.TrimSpace(mockConn.writeBuffer[0])
	expectedResponse := "OutputDevice|1"
	if response != expectedResponse {
		t.Errorf("Expected response '%s', got '%s'", expectedResponse, response)
	}
}

// TestNoiseReduction tests noise reduction with config thresholds
func TestNoiseReduction(t *testing.T) {
	logger := zap.NewNop().Sugar()
	notifier := &mockNotifier{}

	configContent := `
slider_mapping:
  0: master
serial_connection_info:
  com_port: "COM4"
  baud_rate: 115200
noise_reduction: "default"
`
	cleanup := createTestConfig(t, configContent)
	defer cleanup()

	config, err := NewConfig(logger, notifier)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	deej := &Deej{
		config:      config,
		logger:      logger,
		notifier:    notifier,
		stopChannel: make(chan bool),
	}

	sio, err := NewSerialIO(deej, logger)
	if err != nil {
		t.Fatalf("Failed to create SerialIO: %v", err)
	}

	// Subscribe to slider events
	eventChan := sio.SubscribeToSliderMoveEvents()

	// Set initial value
	sio.handleSliders([]string{"2048"})

	// Drain initial event
	select {
	case <-eventChan:
	case <-time.After(100 * time.Millisecond):
	}

	// Send a value very close to the current one (should be filtered by noise reduction)
	sio.handleSliders([]string{"2060"})

	// Should NOT receive an event
	select {
	case event := <-eventChan:
		t.Errorf("Received event when noise should have been filtered: %+v", event)
	case <-time.After(100 * time.Millisecond):
		// Expected: no event
	}

	// Send a value significantly different (should pass through)
	sio.handleSliders([]string{"3000"})

	// Should receive an event
	select {
	case event := <-eventChan:
		if event.SliderID != 0 {
			t.Errorf("Expected slider ID 0, got %d", event.SliderID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Did not receive event for significant change")
	}
}

// TestInvertSliders tests the invert_sliders config option
func TestInvertSliders(t *testing.T) {
	logger := zap.NewNop().Sugar()
	notifier := &mockNotifier{}

	configContent := `
slider_mapping:
  0: master
serial_connection_info:
  com_port: "COM4"
  baud_rate: 115200
invert_sliders: true
`
	cleanup := createTestConfig(t, configContent)
	defer cleanup()

	config, err := NewConfig(logger, notifier)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	deej := &Deej{
		config:      config,
		logger:      logger,
		notifier:    notifier,
		stopChannel: make(chan bool),
	}

	sio, err := NewSerialIO(deej, logger)
	if err != nil {
		t.Fatalf("Failed to create SerialIO: %v", err)
	}

	// Subscribe to slider events
	eventChan := sio.SubscribeToSliderMoveEvents()

	// Send max value (4095)
	sio.handleSliders([]string{"4095"})

	// With invert, 4095 should become 0.0
	select {
	case event := <-eventChan:
		if event.PercentValue != 0.0 {
			t.Errorf("With invert, 4095 should become 0.0, got %.2f", event.PercentValue)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Did not receive event")
	}

	// Send min value (0)
	sio.handleSliders([]string{"0"})

	// With invert, 0 should become 1.0
	select {
	case event := <-eventChan:
		if event.PercentValue != 1.0 {
			t.Errorf("With invert, 0 should become 1.0, got %.2f", event.PercentValue)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Did not receive event")
	}
}

// TestMalformedPacketValidation tests that values > 4095 are normalized
func TestMalformedPacketValidation(t *testing.T) {
	logger := zap.NewNop().Sugar()
	notifier := &mockNotifier{}

	configContent := `
slider_mapping:
  0: master
serial_connection_info:
  com_port: "COM4"
  baud_rate: 115200
`
	cleanup := createTestConfig(t, configContent)
	defer cleanup()

	config, err := NewConfig(logger, notifier)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	deej := &Deej{
		config:      config,
		logger:      logger,
		notifier:    notifier,
		stopChannel: make(chan bool),
	}

	sio, err := NewSerialIO(deej, logger)
	if err != nil {
		t.Fatalf("Failed to create SerialIO: %v", err)
	}

	// Create mock connection
	mockConn := &mockSerialConnection{
		writeBuffer: []string{},
	}
	sio.conn = mockConn
	sio.connected = true

	// Subscribe to slider events
	eventChan := sio.SubscribeToSliderMoveEvents()

	// Send malformed data (value > 4095)
	sio.handleSliders([]string{"5000"})

	// Should normalize to 1.0
	select {
	case event := <-eventChan:
		if event.PercentValue != 1.0 {
			t.Errorf("Expected normalized value 1.0 for >4095, got %.2f", event.PercentValue)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Did not receive event")
	}

	// Check that OK was sent
	if len(mockConn.writeBuffer) == 0 {
		t.Fatal("No response was written")
	}
}

// TestGetCurrentOutputDevice tests the query pattern for current device
func TestGetCurrentOutputDevice(t *testing.T) {
	logger := zap.NewNop().Sugar()
	notifier := &mockNotifier{}

	configContent := `
slider_mapping:
  0: master
available_output_device:
  0: ["Speakers"]
  1: ["Headphones"]
serial_connection_info:
  com_port: "COM4"
  baud_rate: 115200
`
	cleanup := createTestConfig(t, configContent)
	defer cleanup()

	config, err := NewConfig(logger, notifier)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	deej := &Deej{
		config:      config,
		logger:      logger,
		notifier:    notifier,
		stopChannel: make(chan bool),
	}

	sio, err := NewSerialIO(deej, logger)
	if err != nil {
		t.Fatalf("Failed to create SerialIO: %v", err)
	}

	// Create mock connection
	mockConn := &mockSerialConnection{
		writeBuffer: []string{},
	}
	sio.conn = mockConn
	sio.connected = true

	// Set up a mock consumer that returns device 1 as current
	sio.setToggleOutputDeviceEventConsumer(func(event ToggleOutoutDeviceClickEvent) (OutputDeviceState, error) {
		if event.selectedOutputDevice == -1 {
			// Query: return current device
			return OutputDeviceState{selectedOutputDevice: 1}, nil
		}
		// Switch: return requested device
		return OutputDeviceState{selectedOutputDevice: event.selectedOutputDevice}, nil
	})

	// Query current device
	sio.handleGetCurrentOutputDevice()

	// Check response
	if len(mockConn.writeBuffer) == 0 {
		t.Fatal("No response was written")
	}

	response := strings.TrimSpace(mockConn.writeBuffer[0])
	expectedResponse := "OutputDevice|1"
	if response != expectedResponse {
		t.Errorf("Expected response '%s', got '%s'", expectedResponse, response)
	}
}
