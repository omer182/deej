package deej

import (
	"os"
	"testing"

	"go.uber.org/zap"
)

// mockNotifier implements the Notifier interface for testing
type mockNotifier struct{}

func (m *mockNotifier) Notify(title, message string) {}

// TestSerialConfigAutoDetection tests that the config loads with auto-detection
func TestSerialConfigAutoDetection(t *testing.T) {
	// Create a temporary config file with auto-detection
	configContent := `
slider_mapping:
  0: master
serial_connection_info:
  com_port: "auto"
  baud_rate: 115200
`
	if err := os.WriteFile("config.yaml", []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
	defer os.Remove("config.yaml")

	logger := zap.NewNop().Sugar()
	notifier := &mockNotifier{}

	config, err := NewConfig(logger, notifier)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.SerialConnectionInfo.COMPort != "auto" {
		t.Errorf("Expected com_port to be 'auto', got '%s'", config.SerialConnectionInfo.COMPort)
	}

	if config.SerialConnectionInfo.BaudRate != 115200 {
		t.Errorf("Expected baud_rate to be 115200, got %d", config.SerialConnectionInfo.BaudRate)
	}
}

// TestSerialConfigExplicitPort tests that the config loads with an explicit COM port
func TestSerialConfigExplicitPort(t *testing.T) {
	// Create a temporary config file with explicit COM port
	configContent := `
slider_mapping:
  0: master
serial_connection_info:
  com_port: "COM4"
  baud_rate: 9600
`
	if err := os.WriteFile("config.yaml", []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
	defer os.Remove("config.yaml")

	logger := zap.NewNop().Sugar()
	notifier := &mockNotifier{}

	config, err := NewConfig(logger, notifier)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.SerialConnectionInfo.COMPort != "COM4" {
		t.Errorf("Expected com_port to be 'COM4', got '%s'", config.SerialConnectionInfo.COMPort)
	}

	if config.SerialConnectionInfo.BaudRate != 9600 {
		t.Errorf("Expected baud_rate to be 9600, got %d", config.SerialConnectionInfo.BaudRate)
	}
}

// TestSerialConfigDefaults tests that default values are used when not specified
func TestSerialConfigDefaults(t *testing.T) {
	// Create a temporary config file without serial settings
	configContent := `
slider_mapping:
  0: master
`
	if err := os.WriteFile("config.yaml", []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
	defer os.Remove("config.yaml")

	logger := zap.NewNop().Sugar()
	notifier := &mockNotifier{}

	config, err := NewConfig(logger, notifier)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.SerialConnectionInfo.COMPort != "auto" {
		t.Errorf("Expected default com_port to be 'auto', got '%s'", config.SerialConnectionInfo.COMPort)
	}

	if config.SerialConnectionInfo.BaudRate != 115200 {
		t.Errorf("Expected default baud_rate to be 115200, got %d", config.SerialConnectionInfo.BaudRate)
	}
}

// TestSerialConfigBaudRateDefault tests that baud_rate defaults to 115200
func TestSerialConfigBaudRateDefault(t *testing.T) {
	// Create a temporary config file with only com_port specified
	configContent := `
slider_mapping:
  0: master
serial_connection_info:
  com_port: "COM3"
`
	if err := os.WriteFile("config.yaml", []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
	defer os.Remove("config.yaml")

	logger := zap.NewNop().Sugar()
	notifier := &mockNotifier{}

	config, err := NewConfig(logger, notifier)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.SerialConnectionInfo.COMPort != "COM3" {
		t.Errorf("Expected com_port to be 'COM3', got '%s'", config.SerialConnectionInfo.COMPort)
	}

	if config.SerialConnectionInfo.BaudRate != 115200 {
		t.Errorf("Expected default baud_rate to be 115200 when not specified, got %d", config.SerialConnectionInfo.BaudRate)
	}
}
