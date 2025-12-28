package deej

import (
	"os"
	"testing"
	"time"

	"go.uber.org/zap"
)

// TestSerialIOCreation tests that SerialIO is successfully created
func TestSerialIOCreation(t *testing.T) {
	configContent := `
slider_mapping:
  0: master
  1: chrome
serial_connection_info:
  com_port: "COM4"
  baud_rate: 115200
`
	if err := os.WriteFile("config.yaml", []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
	defer os.Remove("config.yaml")

	logger := zap.NewNop().Sugar()
	notifier := &mockNotifier{}

	deej, err := NewDeej(logger, false)
	if err != nil {
		t.Fatalf("Failed to create Deej: %v", err)
	}

	// Load config first
	if err := deej.config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create SerialIO
	serialIO, err := NewSerialIO(deej, deej.logger)
	if err != nil {
		t.Fatalf("Failed to create SerialIO: %v", err)
	}

	if serialIO == nil {
		t.Error("SerialIO instance is nil")
	}
}

// TestSerialIOAssignedToBothControllers tests that SerialIO is assigned to both controllers
func TestSerialIOAssignedToBothControllers(t *testing.T) {
	configContent := `
slider_mapping:
  0: master
  1: chrome
mute_button_mapping:
  0: chrome
  1: discord
available_output_device:
  0: ["Speakers"]
  1: ["Headphones"]
serial_connection_info:
  com_port: "COM4"
  baud_rate: 115200
`
	if err := os.WriteFile("config.yaml", []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
	defer os.Remove("config.yaml")

	logger := zap.NewNop().Sugar()

	deej, err := NewDeej(logger, false)
	if err != nil {
		t.Fatalf("Failed to create Deej: %v", err)
	}

	// Load config
	if err := deej.config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create SerialIO and assign to both controllers
	serialIO, err := NewSerialIO(deej, deej.logger)
	if err != nil {
		t.Fatalf("Failed to create SerialIO: %v", err)
	}

	deej.deejSlidersController = serialIO
	deej.deejButtonsController = serialIO

	// Verify both controllers are set to the same instance
	if deej.deejSlidersController == nil {
		t.Error("deejSlidersController is nil")
	}

	if deej.deejButtonsController == nil {
		t.Error("deejButtonsController is nil")
	}

	// Verify they're the same instance
	if deej.deejSlidersController != deej.deejButtonsController {
		t.Error("Expected both controllers to point to the same SerialIO instance")
	}
}

// TestConsumerRegistration tests that button event consumers can be registered
func TestConsumerRegistration(t *testing.T) {
	configContent := `
slider_mapping:
  0: master
mute_button_mapping:
  0: chrome
  1: discord
available_output_device:
  0: ["Speakers"]
  1: ["Headphones"]
serial_connection_info:
  com_port: "COM4"
  baud_rate: 115200
`
	if err := os.WriteFile("config.yaml", []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
	defer os.Remove("config.yaml")

	logger := zap.NewNop().Sugar()

	deej, err := NewDeej(logger, false)
	if err != nil {
		t.Fatalf("Failed to create Deej: %v", err)
	}

	// Load config
	if err := deej.config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create SerialIO
	serialIO, err := NewSerialIO(deej, deej.logger)
	if err != nil {
		t.Fatalf("Failed to create SerialIO: %v", err)
	}

	deej.deejButtonsController = serialIO

	// Register mock consumers
	muteConsumerCalled := false
	deviceConsumerCalled := false

	deej.deejButtonsController.setMuteButtonClickEventConsumer(func(events []MuteButtonClickEvent) (MuteButtonsState, error) {
		muteConsumerCalled = true
		return MuteButtonsState{MuteButtons: make([]bool, len(events))}, nil
	})

	deej.deejButtonsController.setToggleOutputDeviceEventConsumer(func(event ToggleOutoutDeviceClickEvent) (OutputDeviceState, error) {
		deviceConsumerCalled = true
		return OutputDeviceState{selectedOutputDevice: event.selectedOutputDevice}, nil
	})

	// Verify consumers are registered by calling them directly
	serialIO.handleMuteButtons([]string{"true", "false"})
	serialIO.handleSwitchOutput([]string{"1"})

	if !muteConsumerCalled {
		t.Error("Mute consumer was not called")
	}

	if !deviceConsumerCalled {
		t.Error("Device consumer was not called")
	}
}

// TestSessionMapSubscription tests that session map can subscribe to slider events
func TestSessionMapSubscription(t *testing.T) {
	configContent := `
slider_mapping:
  0: master
  1: chrome
serial_connection_info:
  com_port: "COM4"
  baud_rate: 115200
`
	if err := os.WriteFile("config.yaml", []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
	defer os.Remove("config.yaml")

	logger := zap.NewNop().Sugar()

	deej, err := NewDeej(logger, false)
	if err != nil {
		t.Fatalf("Failed to create Deej: %v", err)
	}

	// Load config
	if err := deej.config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create SerialIO
	serialIO, err := NewSerialIO(deej, deej.logger)
	if err != nil {
		t.Fatalf("Failed to create SerialIO: %v", err)
	}

	deej.deejSlidersController = serialIO

	// Subscribe to slider events (simulating what session map does)
	eventChannel := deej.deejSlidersController.SubscribeToSliderMoveEvents()

	if eventChannel == nil {
		t.Fatal("Event channel is nil")
	}

	// Send a slider event
	go serialIO.handleSliders([]string{"2048", "4095"})

	// Wait for event
	select {
	case event := <-eventChannel:
		if event.SliderID < 0 || event.SliderID > 1 {
			t.Errorf("Unexpected slider ID: %d", event.SliderID)
		}
		if event.PercentValue < 0.0 || event.PercentValue > 1.0 {
			t.Errorf("Unexpected percent value: %f", event.PercentValue)
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("Did not receive slider event")
	}
}
