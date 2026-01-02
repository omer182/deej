# Task Breakdown: Migrate TCP/UDP to USB Serial

## Overview
Replace network-based communication (TCP/UDP) with direct USB serial communication for deej ESP32 controller.

Total Task Groups: 4
Estimated Total Tasks: ~25

## Task List

### Configuration Layer

#### Task Group 1: Update Configuration Schema
**Dependencies:** None
**Size:** Small
**Engineer:** Backend Engineer

- [x] 1.0 Update configuration schema for serial communication
  - [x] 1.1 Write 2-4 focused tests for serial config
    - Test serial_connection_info loading from config
    - Test auto-detection config value ("auto")
    - Test explicit COM port config value (e.g., "COM4")
    - Test baud_rate default and override
  - [x] 1.2 Add SerialConnectionInfo struct to config.go
    - Add struct with COMPort (string) and BaudRate (uint) fields
    - Follow existing pattern from UdpConnectionInfo/TcpConnectionInfo
  - [x] 1.3 Add serial config constants
    - Add configKeySerialPort constant
    - Add configKeyBaudRate constant
    - Add defaultBaudRate = 115200
  - [x] 1.4 Update config defaults in NewConfig()
    - Add userConfig.SetDefault(configKeySerialPort, "auto")
    - Add userConfig.SetDefault(configKeyBaudRate, 115200)
  - [x] 1.5 Add serial config reading in populateFromVipers()
    - Read serial_connection_info.com_port into config.SerialConnectionInfo.COMPort
    - Read serial_connection_info.baud_rate into config.SerialConnectionInfo.BaudRate
  - [x] 1.6 Remove TCP/UDP config code
    - Remove UdpConnectionInfo and TcpConnectionInfo struct fields
    - Remove configKeyUdpPort, configKeyTcpPort constants
    - Remove defaultUdpPort, defaultTcpPort constants
    - Remove UDP/TCP config defaults from NewConfig()
    - Remove UDP/TCP config reading from populateFromVipers()
  - [x] 1.7 Ensure config tests pass
    - Run ONLY the 2-4 tests written in 1.1
    - Verify config loads serial settings correctly
    - Do NOT run the entire test suite at this stage

**Acceptance Criteria:**
- Config loads serial_connection_info with com_port and baud_rate
- Auto-detection value ("auto") is supported
- Explicit COM port values work (e.g., "COM4")
- Default baud_rate is 115200
- TCP/UDP config code completely removed
- The 2-4 tests written in 1.1 pass

---

### Serial Communication Layer

#### Task Group 2: Implement Unified Serial Controller
**Dependencies:** Task Group 1
**Size:** Large
**Engineer:** Backend Engineer

- [x] 2.0 Complete SerialIO implementation
  - [x] 2.1 Write 4-8 focused tests for SerialIO
    - Test auto-detection logic (mock COM port scanning)
    - Test serial connection establishment
    - Test slider data parsing and normalization (12-bit ADC → 0.0-1.0)
    - Test mute button request/response handling
    - Test device switch request/response handling
    - Test reconnection on connection loss
    - Test noise reduction with config thresholds
    - Test invert_sliders config option
  - [x] 2.2 Review and refine existing serial.go structure
    - Verify SerialIO struct has all necessary fields
    - Ensure implements both DeejSlidersController and DeejButtonsController interfaces
    - Verify auto-detection logic scans COM3-COM16
    - Confirm connection options use 115200 baud, 8 data bits, 1 stop bit
  - [x] 2.3 Enhance handleSliders() implementation
    - Reuse UDP normalization logic: float32(rawValue) / 4095.0
    - Reuse util.NormalizeScalar() for precision
    - Reuse util.SignificantlyDifferent() for noise reduction
    - Support config.InvertSliders option
    - Validate raw values don't exceed 4095 (reject malformed packets)
    - Create and broadcast SliderMoveEvent to all consumers
    - Send "OK\n" response back to ESP32
  - [x] 2.4 Enhance handleMuteButtons() implementation
    - Parse pipe-delimited boolean states ("true"|"false")
    - Create MuteButtonClickEvent for each button
    - Call muteButtonsConsumer to get actual new state
    - Build "MuteState|true|false\n" response with actual states
    - Handle missing consumer gracefully (return "ERROR\n")
    - Reuse TCP pattern from tcp.go lines 203-239
  - [x] 2.5 Enhance handleSwitchOutput() implementation
    - Parse device index from data
    - Create ToggleOutoutDeviceClickEvent
    - Call toggleOutputDeviceConsumer to get actual device state
    - Build "OutputDevice|N\n" response with actual device index
    - Handle errors with "ERROR\n" response
    - Reuse TCP pattern from tcp.go lines 242-253
  - [x] 2.6 Enhance handleGetCurrentOutputDevice() implementation
    - Use query pattern: call consumer with selectedOutputDevice = -1
    - Get current device from consumer response
    - Build "OutputDevice|N\n" response
    - Handle errors with "ERROR\n" response
    - Reuse TCP pattern from tcp.go lines 256-277
  - [x] 2.7 Verify reconnection resilience
    - Confirm readLoop() handles io.EOF and connection errors
    - Verify 500ms reconnectDelay between attempts
    - Ensure reader is recreated after reconnection
    - Log reconnection attempts at Warn level
    - Continue operation without crashing
  - [x] 2.8 Add config reload handling
    - Subscribe to config changes via config.SubscribeToChanges()
    - On reload, update serial connection options if port/baud changed
    - Reset currentSliderPercentValues to force slider move events
    - Follow pattern from udp.go setupOnConfigReload() (lines 162-187)
  - [x] 2.9 Ensure SerialIO tests pass
    - Run ONLY the 4-8 tests written in 2.1
    - Verify all protocol commands work correctly
    - Do NOT run the entire test suite at this stage

**Acceptance Criteria:**
- SerialIO implements both DeejSlidersController and DeejButtonsController
- Auto-detection scans COM3-COM16 successfully
- Slider normalization maintains 12-bit precision (0-4095 → 0.0-1.0)
- Mute button state requests receive correct state responses
- Device switching sends actual selected device back
- Auto-reconnect works with 500ms delay
- Config reload updates connection settings
- The 4-8 tests written in 2.1 pass

---

### Application Integration Layer

#### Task Group 3: Integrate SerialIO into Deej
**Dependencies:** Task Group 2
**Size:** Medium
**Engineer:** Backend Engineer

- [x] 3.0 Replace TCP/UDP with SerialIO in deej.go
  - [x] 3.1 Write 2-4 focused tests for deej initialization
    - Test SerialIO creation succeeds
    - Test SerialIO assigned to both controllers
    - Test consumer registration works
    - Test session map subscription works
  - [x] 3.2 Update Initialize() method in deej.go
    - Remove NewUdpIO() call (lines ~91-98)
    - Remove NewTcpIO() call (lines ~100-106)
    - Add single NewSerialIO(d, d.logger) call
    - Assign result to both d.deejSlidersController and d.deejButtonsController
  - [x] 3.3 Register button event consumers
    - Call setMuteButtonClickEventConsumer() on SerialIO
    - Call setToggleOutputDeviceEventConsumer() on SerialIO
    - Use existing consumer functions from session map
    - Follow pattern that was used for TCP controller
  - [x] 3.4 Subscribe to slider events
    - Call SubscribeToSliderMoveEvents() on SerialIO
    - Pass channel to session map as before
    - Ensure session map receives slider events
  - [x] 3.5 Update Start() calls
    - Remove separate Start() calls if TCP/UDP had them separately
    - Call Start() on SerialIO once (serves both interfaces)
  - [x] 3.6 Ensure deej initialization tests pass
    - Run ONLY the 2-4 tests written in 3.1
    - Verify initialization succeeds with SerialIO
    - Do NOT run the entire test suite at this stage

**Acceptance Criteria:**
- Only SerialIO is created (no TCP/UDP instances)
- SerialIO serves as both slider and button controller
- Session map receives slider move events
- Button event consumers are registered correctly
- Application starts successfully with serial controller
- The 2-4 tests written in 3.1 pass

---

### Cleanup and Documentation

#### Task Group 4: Remove Legacy Code and Update Examples
**Dependencies:** Task Group 3
**Size:** Small
**Engineer:** Backend Engineer

- [x] 4.0 Complete migration cleanup
  - [x] 4.1 Delete legacy network files
    - Delete pkg/deej/udp.go completely
    - Delete pkg/deej/tcp.go completely
    - Verify no other files import these packages
  - [x] 4.2 Update example config.yaml
    - Remove udp_port field
    - Remove tcp_port field
    - Add serial_connection_info section
    - Add com_port: "auto" example
    - Add baud_rate: 115200 example
    - Add comment explaining auto-detection vs explicit port
  - [x] 4.3 Update logging namespaces
    - Verify SerialIO uses logger.Named("serial")
    - Ensure log messages reference serial not UDP/TCP
    - Check Info/Warn/Debug levels are appropriate
  - [x] 4.4 Verify toast notifications
    - Confirm auto-detect failure shows user notification
    - Confirm initial connection failure shows user notification
    - Use existing notifier.Notify() pattern
  - [x] 4.5 Remove unused imports
    - Check deej.go for unused net package imports
    - Check config.go for unused network-related imports
    - Run go mod tidy to clean dependencies
  - [x] 4.6 Run full test suite
    - Run complete test suite for the entire application
    - Expected tests: config tests (2-4) + SerialIO tests (4-8) + deej tests (2-4) + any existing tests
    - Verify all tests pass
    - Fix any breaking changes in existing tests

**Acceptance Criteria:**
- udp.go and tcp.go files deleted
- config.yaml example uses serial_connection_info
- No references to UDP/TCP remain in codebase
- Logging uses "serial" namespace consistently
- Toast notifications work for serial errors
- No unused imports remain
- Full test suite passes (approximately 8-16 feature tests + existing tests)

---

## Task Group Dependencies

```
Task Group 1 (Config Schema)
    ↓
Task Group 2 (SerialIO Implementation)
    ↓
Task Group 3 (Integration)
    ↓
Task Group 4 (Cleanup)
```

## Execution Order

Recommended implementation sequence:

1. **Configuration Layer** (Task Group 1) - Update config schema, remove TCP/UDP config
2. **Serial Communication Layer** (Task Group 2) - Implement and test SerialIO controller
3. **Application Integration** (Task Group 3) - Wire SerialIO into deej.go
4. **Cleanup** (Task Group 4) - Delete legacy files, update examples, run full tests

## Implementation Notes

### Key Patterns to Reuse

**From udp.go:**
- Slider normalization: `float32(number) / 4095.0`
- Noise reduction: `util.SignificantlyDifferent()` with config threshold
- Invert sliders: `normalizedValue = 1 - normalizedValue`
- Event broadcasting: iterate consumers, send events to channels

**From tcp.go:**
- Consumer pattern: call consumer, get new state, send as response
- Request/response format: pipe-delimited with newline terminator
- Error handling: return "ERROR\n" on consumer failure
- Device matching: iterate config.AvailableOutputDeviceMapping

**From existing serial.go:**
- Auto-detection: scan COM3-COM16, test each port
- Connection setup: jacobsa/go-serial with 115200 baud
- Read loop: bufio.Reader with ReadString('\n')
- Reconnection: sleep 500ms, retry connect on error

### Testing Strategy

- Each task group writes 2-8 focused tests maximum
- Tests cover critical behaviors only, not exhaustive coverage
- Each group runs ONLY its own tests during development
- Task Group 4 runs full test suite as final verification
- Focus on integration tests over unit tests
- Mock serial port I/O for testing

### Protocol Format Reference

**ESP32 → PC:**
- Sliders: `Sliders|4095|2048|1024|512|0\n`
- Mute buttons: `MuteButtons|true|false\n`
- Switch device: `SwitchOutput|1\n`
- Query device: `GetCurrentOutputDevice\n`

**PC → ESP32:**
- Slider ack: `OK\n`
- Mute states: `MuteState|true|false\n`
- Device state: `OutputDevice|1\n`
- Error: `ERROR\n`

### Windows-Specific Considerations

- COM port range: COM3-COM16 (Windows standard range)
- Serial library: github.com/jacobsa/go-serial (cross-platform but tested on Windows)
- Auto-detection: try opening each port, read sample data
- Baud rate: 115200 (standard ESP32 serial rate)

### Error Handling Approach

- Auto-detect failure: log warning, notify user, return error
- Initial connection failure: log warning, notify user, return error
- Read/write errors: log warning, attempt reconnection
- Malformed packets: log at Debug level, discard packet
- Consumer errors: log warning, send "ERROR\n" response

### Code Removal Checklist

Files to delete:
- pkg/deej/udp.go
- pkg/deej/tcp.go

Config fields to remove:
- UdpConnectionInfo struct
- TcpConnectionInfo struct
- configKeyUdpPort constant
- configKeyTcpPort constant
- defaultUdpPort constant
- defaultTcpPort constant

Code sections to remove:
- UDP/TCP initialization in deej.go Initialize() (~lines 91-106)
- UDP/TCP config defaults in config.go NewConfig()
- UDP/TCP config reading in config.go populateFromVipers()

### Validation Checklist

Before marking as complete:
- [x] No references to UDP/TCP in any .go files (in SerialIO)
- [x] config.yaml example uses serial_connection_info
- [x] SerialIO implements both controller interfaces
- [x] Auto-detection works for COM3-COM16
- [x] Slider normalization maintains precision
- [x] Mute button states sync correctly
- [x] Device switching updates LEDs via response
- [x] Reconnection works on connection loss
- [x] Config reload updates serial settings
- [x] All tests pass (feature tests + existing tests)
