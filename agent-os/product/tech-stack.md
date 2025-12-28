# Tech Stack

## Framework & Runtime
- **Backend Language:** Go 1.14+
- **Firmware Language:** C++17 (Arduino framework)
- **Package Manager:** Go modules (backend), PlatformIO/Arduino (firmware)

## Backend (Windows PC Application)

### Core Libraries
- **go-wca** (moutend/go-wca v0.3.0): Windows Core Audio API bindings for volume control and session management
- **go-ole** (go-ole/go-ole v1.2.6): COM/OLE automation for Windows API interactions
- **go-serial** (jacobsa/go-serial): USB serial communication library for ESP32 connection at 115200 baud
- **viper** (spf13/viper v1.7.1): YAML configuration file parsing and management
- **zap** (go.uber.org/zap v1.15.0): Structured logging library

### Supporting Libraries
- **systray** (getlantern/systray): System tray icon and menu integration
- **beeep** (gen2brain/beeep): Desktop notifications for user feedback
- **go-ps** (mitchellh/go-ps v1.0.0): Process enumeration for app-specific volume targeting
- **fsnotify** (fsnotify/fsnotify v1.4.9): Config file hot-reloading
- **go-funk** (thoas/go-funk v0.7.0): Functional programming utilities

## Firmware (ESP32 Controller)

### Platform
- **Microcontroller:** ESP32 or ESP8266
- **Framework:** Arduino framework
- **Build System:** PlatformIO (or Arduino IDE)

### Communication
- **Protocol:** USB Serial (migrating from UDP/TCP)
- **Baud Rate:** 115200
- **Data Format:** Pipe-delimited text protocol (e.g., `Sliders|0|512|512|0|0`)

### Hardware Interface
- **Analog Input:** ADC pins for 5 potentiometer sliders (10-bit resolution: 0-1023)
- **Digital Input:** GPIO pins for 2 mute buttons (with debouncing)
- **Digital Input:** GPIO pin for output device toggle button
- **Digital Output:** GPIO pins for LED indicators (mute states, device selection)

## Configuration

### Format
- **Config File:** YAML (`config.yaml`)
- **Location:** Application directory

### Configuration Schema
```yaml
# Serial communication settings (replacing udp_port)
serial_port: "auto"  # or specific COM port like "COM3"
baud_rate: 115200

# Slider to audio session mappings
slider_mapping:
  0: "device name or app.exe"
  1: "another target"
  # ... up to index 4

# Mute button mappings
mute_button_mapping:
  0: master
  1: mic

# Output device toggle options
available_output_device:
  0: "Speakers (Realtek(R) Audio)"
  1: "Headphones (HyperX Cloud III Wireless)"
```

## Platform Support

### Primary Platform
- **Operating System:** Windows 10/11
- **Audio API:** Windows Core Audio (WCA)
- **System APIs:** COM/OLE for device enumeration and control

### Future Platforms
- **Linux:** Potential support via PulseAudio (jfreymuth/pulse library already included)

## Communication Protocol

### Commands (ESP32 → PC)
- **Sliders:** `Sliders|<val0>|<val1>|<val2>|<val3>|<val4>` (values 0-1023)
- **MuteButtons:** `MuteButtons|<bool0>|<bool1>` (true/false states)
- **SwitchOutput:** `SwitchOutput|<index>` (device index from config)

### Bidirectional Enhancement (PC → ESP32)
- **MuteState:** State updates for LED feedback on mute buttons
- **DeviceState:** Active output device index for LED indication
- **Acknowledgment:** Connection handshake and status messages

## Build & Development

### Backend Build
- **Build Tool:** Go build toolchain
- **Dependencies:** Go modules (`go.mod`)
- **Platform-Specific:** Windows CGO for WCA bindings

### Firmware Build
- **IDE:** PlatformIO (recommended) or Arduino IDE
- **Board Config:** ESP32 Dev Module or ESP8266 board definitions
- **Upload:** USB serial connection for firmware flashing

## Testing & Quality

### Backend Testing
- **Manual Testing:** Physical hardware interaction testing
- **Windows Integration Testing:** Audio session and device control validation

### Code Quality
- **Coding Style:** Follow Go standard formatting (gofmt)
- **Naming Conventions:** Descriptive names, minimal abbreviations
- **Error Handling:** Explicit error handling with clear user messages
- **Logging:** Structured logging via zap for debugging and monitoring
