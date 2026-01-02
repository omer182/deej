# Tech Stack

## Framework & Runtime
- **Backend Language:** Go 1.21+
- **Firmware Language:** C++17 (Arduino framework)
- **Package Manager:** Go modules (backend), PlatformIO (firmware)

**Rationale:** Go provides excellent Windows API bindings through CGO and strong concurrency primitives for handling serial I/O alongside audio session management. Arduino framework on ESP32 offers mature serial communication libraries and widespread community support.

## Backend (Windows PC Application)

### Core Libraries
- **go-wca** (moutend/go-wca v0.3.0): Windows Core Audio API bindings for volume control and session management
- **go-ole** (go-ole/go-ole v1.2.6): COM/OLE automation required for Windows Core Audio interactions
- **go-serial** (jacobsa/go-serial): USB serial communication library for ESP32 connection at 115200 baud
- **viper** (spf13/viper v1.7.1): YAML configuration file parsing and management
- **zap** (go.uber.org/zap v1.15.0): Structured logging library for debugging serial protocol and audio operations

**Rationale:** go-wca provides direct access to Windows Core Audio Session API (WASAPI), enabling true mute functionality and per-application volume control. go-serial handles cross-platform USB serial with proper timeout and reconnection support.

### Supporting Libraries
- **systray** (getlantern/systray): System tray icon and menu integration for background operation
- **beeep** (gen2brain/beeep): Desktop notifications for connection status and errors
- **go-ps** (mitchellh/go-ps v1.0.0): Process enumeration for mapping application names to audio sessions
- **fsnotify** (fsnotify/fsnotify v1.4.9): Config file hot-reloading without application restart
- **go-funk** (thoas/go-funk v0.7.0): Functional programming utilities for collection operations

**Rationale:** Background operation via system tray allows deej to run continuously without cluttering the taskbar. Config hot-reload enables users to adjust mappings without disconnecting hardware.

## Firmware (ESP32 Controller)

### Platform
- **Microcontroller:** ESP32 (ESP32-WROOM-32 recommended)
- **Framework:** Arduino framework (arduino-esp32 core)
- **Build System:** PlatformIO

**Rationale:** ESP32 provides 12-bit ADC resolution (0-4095) for smooth volume control, built-in USB serial (CDC), sufficient GPIO pins for 5 sliders + 3 buttons + LEDs, and costs under $10. PlatformIO offers better dependency management and build reproducibility than Arduino IDE.

### Communication
- **Protocol:** USB Serial (CDC)
- **Baud Rate:** 115200
- **Data Format:** Pipe-delimited text protocol (human-readable for debugging)

**Rationale:** USB serial is the simplest reliable communication method - hardware already connected via USB for power, eliminates WiFi configuration complexity, provides sub-50ms latency, and works across all Windows versions without driver installation.

### Hardware Interface
- **Analog Input:** 12-bit ADC pins (GPIO 33, 35, 32, 36, 34) for 5 potentiometer sliders
- **Digital Input:** GPIO pins (4, 14, 5) for 3 buttons with internal pull-up resistors and debouncing
- **Digital Output:** GPIO pins (12, 21, 18, 19) for LED indicators

**Rationale:** 12-bit ADC provides 4096 discrete steps for smooth volume adjustment. Dedicated GPIO pins avoid analog multiplexing complexity. Internal pull-ups simplify button wiring.

## Configuration

### Format
- **Config File:** YAML (config.yaml)
- **Location:** Application directory (same folder as deej.exe)

### Configuration Schema
The config file uses YAML format with the following structure:
- serial_port: auto-detect or explicit COM port
- baud_rate: 115200
- slider_mapping: maps slider indices (0-4) to device names or app executables
- mute_button_mapping: maps button indices to master or mic
- available_output_device: lists device names for toggle button

**Rationale:** YAML is human-readable and supports comments for user guidance. Auto-detection of COM port reduces setup friction. Named device/app mappings are more intuitive than process IDs.

## Platform Support

### Primary Platform
- **Operating System:** Windows 10/11 (64-bit)
- **Audio API:** Windows Core Audio (WASAPI)
- **System APIs:** COM/OLE for device enumeration and control

**Rationale:** Windows Core Audio provides the most reliable per-application volume control and true mute functionality. Focus on single platform ensures deep integration quality.

### Future Platforms
- **Linux:** Potential support via PulseAudio (jfreymuth/pulse library available)

## Communication Protocol

### Commands (ESP32 to PC)
- **Sliders:** Pipe-delimited values with 12-bit ADC readings (0-4095), sent every 50ms if changed
- **MuteButtons:** Boolean states as 1 or 0, sent when button state changes
- **SwitchOutput:** Device index 0 or 1, sent when toggle button pressed

### Responses (PC to ESP32)
- **OK:** Acknowledgment of slider update
- **MuteState:** Actual mute states from Windows Core Audio, sent after MuteButtons command
- **OutputDevice:** Active device index, sent after SwitchOutput command

**Rationale:** Text-based protocol is debuggable via serial monitor. Pipe delimiters are simple to parse. Newline terminators enable line-based reading. Bidirectional responses ensure LEDs always reflect actual Windows state.

## Build & Development

### Backend Build
- **Build Tool:** Go build toolchain (go build)
- **Dependencies:** Managed via go.mod
- **Platform-Specific:** Windows CGO enabled for WCA bindings

**Rationale:** Go's single-binary output simplifies distribution. CGO required for C-based Windows APIs but produces statically-linked executable.

### Firmware Build
- **IDE:** PlatformIO (VSCode extension or CLI)
- **Board Config:** ESP32 Dev Module (arduino-esp32 platform)
- **Upload:** USB serial connection (same cable used for operation)

**Rationale:** PlatformIO provides reproducible builds with locked dependency versions. Single USB cable for both programming and operation.

## Testing & Quality

### Backend Testing
- **Manual Testing:** Physical hardware interaction testing with actual ESP32 device
- **Windows Integration Testing:** Audio session and device control validation across different applications

**Rationale:** Hardware-dependent functionality requires manual testing with real devices. Windows audio API behavior varies by application, requiring broad compatibility testing.

### Firmware Testing
- **Serial Monitor Validation:** Verify outgoing message format and timing
- **LED State Verification:** Confirm LEDs match actual Windows state after backend responses
- **Reconnection Testing:** Unplug/replug USB during operation

### Code Quality
- **Coding Style:** Go standard formatting (gofmt), C++ follows Google style guide
- **Naming Conventions:** Descriptive names, avoid abbreviations except domain terms (ADC, GPIO, LED)
- **Error Handling:** Explicit error handling with structured logging, graceful degradation on serial timeout
- **Logging:** Structured logging via zap for backend, no serial logging from ESP32 (protocol-only)

**Rationale:** Consistent formatting aids collaboration. Explicit error handling prevents silent failures. Protocol-only serial communication avoids parsing ambiguity.
