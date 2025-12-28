# Specification: Migrate TCP/UDP Network Communication to USB Serial

## Goal
Replace the current TCP/UDP network-based communication between the ESP32 firmware and the PC backend with direct USB serial communication, enabling bidirectional data exchange for slider values, mute button states, and audio device switching with LED feedback.

## User Stories
- As a user, I want to connect my ESP32 to my PC via USB cable instead of WiFi, so that I have a more reliable connection without network dependencies
- As a user, I want the ESP32 LEDs to reflect the actual mute states from the PC, so that I can see the real-time status of my audio sessions

## Specific Requirements

**Remove all TCP/UDP network communication code**
- Delete `pkg/deej/udp.go` entirely (handles slider data via UDP port 16990)
- Delete `pkg/deej/tcp.go` entirely (handles button states and device switching via TCP port 16991)
- Remove `UdpConnectionInfo` and `TcpConnectionInfo` from `pkg/deej/config.go`
- Remove `configKeyUdpPort`, `configKeyTcpPort`, `defaultUdpPort`, `defaultTcpPort` constants from config.go
- Remove UDP/TCP port reading logic from `populateFromVipers()` method in config.go
- Remove UDP and TCP controller initialization from `pkg/deej/deej.go` Initialize() method (lines 91-106)

**Implement unified serial controller**
- Use the existing draft `pkg/deej/serial.go` as the foundation
- Implement `DeejSlidersController` interface for slider events
- Implement `DeejButtonsController` interface for button events
- Create single `SerialIO` struct that handles both slider and button communication
- Support both manual COM port specification and auto-detection (scan COM3-COM16 on Windows)
- Use 115200 baud rate as specified by user requirements
- Integrate with existing `github.com/jacobsa/go-serial` library dependency

**Update configuration schema**
- Add new `serial_connection_info` section to replace UDP/TCP settings
- Include `com_port` field (string, accepts "auto" for auto-detection or specific port like "COM4")
- Include `baud_rate` field (uint, default 115200)
- Update `config.yaml` example to use new serial settings
- Maintain backwards compatibility for all other config fields (slider_mapping, mute_button_mapping, etc.)

**Define bidirectional serial protocol**
- ESP32 to PC: Send pipe-delimited messages terminated with newline `\n`
- Slider data: `Sliders|4095|2048|1024|512|0\n` (5 sliders, 12-bit ADC values 0-4095)
- Mute button requests: `MuteButtons|true|false\n` (boolean states for each button)
- Device switch requests: `SwitchOutput|1\n` (device index to switch to)
- Query current device: `GetCurrentOutputDevice\n`
- PC to ESP32: Send pipe-delimited responses terminated with newline `\n`
- Slider acknowledgment: `OK\n`
- Mute state response: `MuteState|true|false\n` (actual mute states after processing)
- Device state response: `OutputDevice|1\n` (actual selected device index)
- Error response: `ERROR\n` (when request cannot be processed)

**Maintain 12-bit ADC precision**
- ESP32 uses 12-bit ADC producing values 0-4095
- Normalize to 0.0-1.0 float range in backend: `normalizedValue = float32(rawValue) / 4095.0`
- Apply existing noise reduction logic from config (`low`, `default`, `high` thresholds)
- Apply existing `invert_sliders` config option if enabled
- Reuse existing `util.NormalizeScalar()` and `util.SignificantlyDifferent()` utility functions

**Implement connection resilience**
- Auto-reconnect on serial connection loss with 500ms retry delay
- Log warnings when serial read/write fails but continue attempting reconnection
- No timeout-based error LED blinking (simplified approach: connection either works or app shows error)
- Graceful handling of partial/corrupted lines (discard and wait for next valid message)

**Update deej.go initialization flow**
- Replace `NewUdpIO()` and `NewTcpIO()` calls with single `NewSerialIO()` call
- Assign SerialIO instance to both `d.deejSlidersController` and `d.deejButtonsController` (same instance implements both interfaces)
- Call both `setMuteButtonClickEventConsumer()` and `setToggleOutputDeviceEventConsumer()` on the SerialIO instance
- Subscribe session map to slider move events from SerialIO
- Keep all other initialization logic unchanged (session map, config loading, tray icon)

**Preserve existing event handling architecture**
- Maintain `SliderMoveEvent` struct (SliderID int, PercentValue float32)
- Maintain `MuteButtonClickEvent` struct (MuteButtonID int, mute bool)
- Maintain `ToggleOutoutDeviceClickEvent` struct (selectedOutputDevice int)
- Use existing channel-based consumer pattern for slider events
- Use existing function-based consumer pattern for button events
- Keep `MuteButtonsState` and `OutputDeviceState` response structs unchanged

**Update error handling and logging**
- Use existing zap.SugaredLogger with "serial" namespace
- Log connection status at Info level (connected, disconnected, reconnecting)
- Log protocol errors at Warn level (malformed packets, invalid values)
- Log detailed data at Debug level when verbose mode enabled
- Notify user via existing toast notifier on critical failures (auto-detect fails, initial connection fails)

**No firmware changes required initially**
- Existing draft `serial.go` appears designed to work with pipe-delimited protocol
- Firmware will need separate migration effort (not in this spec's scope)
- Firmware should send commands prefixed with operation type matching current UDP/TCP format
- Firmware should parse PC responses to update LED states

## Visual Design
No visual mockups provided for this feature.

## Existing Code to Leverage

**UdpIO slider handling logic (pkg/deej/udp.go)**
- Reuse `handleSliders()` logic for parsing pipe-delimited slider values
- Reuse ADC value normalization (lines 256-264): divide by 4095.0, apply invert, use util.NormalizeScalar()
- Reuse noise reduction comparison logic (line 267): util.SignificantlyDifferent() with config threshold
- Reuse slider move event creation and broadcasting pattern (lines 272-288)
- Reuse malformed packet validation (line 250): reject values > 4095

**TcpIO button handling logic (pkg/deej/tcp.go)**
- Reuse `handleMuteButtons()` consumer pattern (lines 203-239): parse bool states, call consumer, return new state as response
- Reuse `handleSwitchOutput()` consumer pattern (lines 242-253): parse device ID, call consumer, return actual device ID
- Reuse `handleGetCurrentOutputDevice()` logic (lines 256-277): query current device via util functions, match against config mapping
- Reuse request/response format with pipe delimiters

**Existing SerialIO draft implementation (pkg/deej/serial.go)**
- Use auto-detection logic (lines 144-193) for scanning COM ports on Windows
- Use serial connection setup with jacobsa/go-serial library (lines 72-86, 196-211)
- Use read loop pattern with bufio.Reader (lines 214-255)
- Use response sending pattern via conn.Write() (lines 456-472)
- Use reconnection logic on connection failure (lines 228-243)

**Config watching and reload logic (pkg/deej/config.go)**
- Keep existing viper-based config file watching (WatchConfigFileChanges method)
- Keep existing config reload notification to consumers
- Add serial_connection_info fields alongside existing slider_mapping and other configs

**Deej controller interfaces (pkg/deej/deej_components_controller.go)**
- SerialIO must implement both DeejSlidersController (Start, Stop, SubscribeToSliderMoveEvents) and DeejButtonsController (Start, Stop, setMuteButtonClickEventConsumer, setToggleOutputDeviceEventConsumer)
- Use same Start() method for both interfaces (single instance serves both)
- Use same Stop() method for both interfaces

## Out of Scope
- Modifying ESP32 firmware code (separate effort required to migrate from WiFi to USB serial)
- Implementing timeout-based error LED blinking on ESP32 (simplified: just handle reconnection)
- Supporting Linux or macOS COM port auto-detection (Windows-only for now)
- Adding web-based configuration UI
- Implementing serial port monitoring/diagnostics tools
- Supporting multiple simultaneous deej devices on different COM ports
- Adding rate limiting or throttling beyond existing noise reduction
- Changing the pipe-delimited protocol format
- Supporting different baud rates via config (hardcode 115200 for ESP32)
- Implementing hardware handshaking (RTS/CTS)
- Adding serial buffer size configuration
- Creating migration scripts for existing users
