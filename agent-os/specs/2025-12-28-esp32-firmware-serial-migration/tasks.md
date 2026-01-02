# Task Breakdown: ESP32 Firmware Serial Migration

## Overview
Total Tasks: 29 tasks across 5 task groups

This task breakdown migrates the ESP32 firmware from WiFi/UDP/TCP network communication to USB Serial communication while maintaining all existing hardware functionality and reusing existing input component classes.

## Task List

### Firmware Cleanup & Preparation

#### Task Group 1: Remove Network Communication Code
**Dependencies:** None

- [x] 1.0 Complete network code removal
  - [x] 1.1 Remove WiFi setup code from main.cpp
    - Delete WiFi.h include directive
    - Remove WiFi.begin() and WiFi.setHostname() calls
    - Delete all WiFi-related variables (ssid, password)
    - Remove network status checking code
  - [x] 1.2 Delete UDP-based VolumeControllerApi files
    - Remove lib/api/volume_controller_api.h
    - Remove lib/api/volume_controller_api.cc
    - Update platformio.ini or build configuration if needed
  - [x] 1.3 Delete TCP-based BackendStateApi files
    - Remove lib/api/backend_state_api.h
    - Remove lib/api/backend_state_api.cc
    - Remove lib/api/backend_state.h (network state structure)
  - [x] 1.4 Remove network-related variables and includes
    - Delete server_address variable
    - Delete udp_server_port and tcp_server_port variables
    - Remove any remaining network library includes
    - Remove maybeApplyBackendState() function from main.cpp
  - [x] 1.5 Remove network error handling code
    - Delete all util::blink2Leds calls for network errors
    - Delete all util::blinkLed calls for communication errors
    - Remove network connection status LED indicators
    - Keep only sequentialLEDOn for startup sequence
  - [x] 1.6 Verify compilation after network code removal
    - Compile firmware to identify any missing dependencies
    - Ensure no network-related symbols remain undefined
    - Confirm clean build with no network references

**Acceptance Criteria:**
- All WiFi/UDP/TCP code removed from firmware
- No network library includes or dependencies remain
- Firmware compiles successfully without network code
- lib/api/ directory contains no network communication files

### Serial Communication Infrastructure

#### Task Group 2: Implement SerialApi Class
**Dependencies:** Task Group 1

- [x] 2.0 Complete SerialApi implementation
  - [x] 2.1 Write 2-8 focused tests for SerialApi class
    - Limit to 2-8 highly focused tests maximum
    - Test only critical serial behaviors (e.g., message format validation, basic parsing, timeout handling)
    - Skip exhaustive coverage of all edge cases
    - Note: If hardware-based testing is preferred, these can be integration tests with real serial port
  - [x] 2.2 Create SerialApi class structure
    - Create lib/api/serial_api.h header file
    - Create lib/api/serial_api.cc implementation file
    - Define SerialApi class with private Serial instance access
    - Add timeout constant (e.g., 100ms) as class member
  - [x] 2.3 Implement sendSliders() method
    - Signature: void sendSliders(const std::string& data)
    - Build message format: "Sliders|val1|val2|val3|val4|val5\n"
    - Use Serial.println() for automatic newline termination
    - Read acknowledgment response: "OK\n" (optional, timeout silently if missing)
  - [x] 2.4 Implement sendMuteButtons() method
    - Signature: std::vector<bool> sendMuteButtons(const std::vector<bool>& states)
    - Build message format: "MuteButtons|bool1|bool2\n"
    - Send using Serial.println()
    - Parse response: "MuteState|bool1|bool2\n"
    - Return parsed boolean vector for LED updates
  - [x] 2.5 Implement sendSwitchOutput() method
    - Signature: int sendSwitchOutput(int deviceIndex)
    - Build message format: "SwitchOutput|index\n"
    - Send using Serial.println()
    - Parse response: "OutputDevice|index\n"
    - Return parsed device index for LED updates
  - [x] 2.6 Implement non-blocking response reading
    - Use Serial.available() check before reading
    - Use Serial.readStringUntil('\n') for line reading
    - Implement timeout handling with millis() comparison
    - Return default/unchanged values on timeout (no error indication)
  - [x] 2.7 Implement response parsing helpers
    - Create parseResponse() helper to split pipe-delimited strings
    - Create parseBool() helper to convert string to boolean
    - Create parseInt() helper to convert string to integer
    - Handle malformed responses gracefully (silently continue)
  - [x] 2.8 Ensure SerialApi tests pass
    - Run ONLY the 2-8 tests written in 2.1
    - Verify critical message formatting and parsing works
    - Do NOT run entire test suite at this stage

**Acceptance Criteria:**
- The 2-8 tests written in 2.1 pass
- SerialApi class compiles and links successfully
- All three send methods implemented with correct message formats
- Response parsing works for OK, MuteState, and OutputDevice messages
- Timeout handling works without visual error indication

### Main Firmware Logic

#### Task Group 3: Update main.cpp for Serial Communication
**Dependencies:** Task Group 2

- [x] 3.0 Complete main.cpp serial integration
  - [x] 3.1 Write 2-8 focused tests for main loop integration
    - Limit to 2-8 highly focused tests maximum
    - Test only critical main loop behaviors (e.g., slider data sending, button event handling, LED updates)
    - Skip exhaustive testing of all component interactions
    - Note: These will likely be integration tests with mock serial or real hardware
  - [x] 3.2 Initialize Serial communication in setup()
    - Add Serial.begin(115200) to setup() function
    - Position after hardware component initialization
    - Remove any network initialization code
    - Keep existing GPIO pin setup and component initialization
  - [x] 3.3 Instantiate SerialApi in main.cpp
    - Create global SerialApi instance or pointer
    - Initialize after Serial.begin() in setup()
    - Replace VolumeControllerApi and BackendStateApi references
  - [x] 3.4 Update slider value sending logic
    - Build slider data string by iterating sliders vector
    - Call getValue() on each slider to get tuple<bool, int>
    - Build "Sliders|val1|val2|val3|val4|val5\n" format
    - Call serialApi.sendSliders() when any slider changed
    - Keep existing change detection logic
  - [x] 3.5 Update mute button handling
    - Detect mute button changes using getValue() on each MuteButton
    - Build "MuteButtons|bool1|bool2\n" message
    - Call serialApi.sendMuteButtons() when any button changed
    - Parse response to get actual PC mute states
    - Call setActiveSessionMuteState() on each MuteButton with response values
  - [x] 3.6 Update audio device selector handling
    - Detect device switch press using AudioDeviceSelector.getValue()
    - Build "SwitchOutput|index\n" message with device index
    - Call serialApi.sendSwitchOutput() when device changed
    - Parse response to get actual active device index
    - Call setActiveDevice() on AudioDeviceSelector with response value
  - [x] 3.7 Update main loop timing
    - Change delay from 150ms to 50ms at end of loop()
    - Verify loop structure: read inputs → send changes → update LEDs
    - Maintain existing component iteration patterns
  - [x] 3.8 Verify startup behavior
    - Confirm no handshake waiting after Serial.begin()
    - Maintain existing sequentialLEDOn startup sequence
    - Start sending data immediately in first loop iteration
    - Remove any connection status LED indicators
  - [x] 3.9 Ensure main loop integration tests pass
    - Run ONLY the 2-8 tests written in 3.1
    - Verify critical main loop workflows function correctly
    - Do NOT run entire test suite at this stage

**Acceptance Criteria:**
- The 2-8 tests written in 3.1 pass
- Serial communication initialized at 115200 baud
- Slider values sent in correct format with raw 12-bit ADC values
- Mute button states sent and LEDs updated from PC responses
- Audio device switching works with LED updates from PC responses
- Main loop runs at 50ms intervals
- Firmware starts sending data immediately without handshake

### Hardware Integration & Testing

#### Task Group 4: Hardware Testing with PC Backend
**Dependencies:** Task Group 3

- [ ] 4.0 Complete hardware integration testing
  - [ ] 4.1 Flash firmware to ESP32 hardware
    - Build firmware for target ESP32 board
    - Flash using platformio run --target upload or Arduino IDE
    - Verify successful upload and ESP32 boot
  - [ ] 4.2 Verify serial connection with PC backend
    - Connect ESP32 via USB to PC
    - Start deej backend with correct COM port configuration
    - Verify backend logs show successful connection
    - Confirm ESP32 begins sending data immediately
  - [ ] 4.3 Test slider volume control
    - Move each of 5 sliders through full range
    - Verify PC volume changes in real-time
    - Confirm smooth control with no lag or stuttering
    - Verify raw 12-bit ADC values (0-4095) are sent correctly
  - [ ] 4.4 Test mute button functionality
    - Press each of 2 mute buttons
    - Verify PC mute states toggle correctly
    - Confirm LEDs update to match actual PC mute state
    - Test rapid button presses to verify debouncing works
  - [ ] 4.5 Test output device switching
    - Press audio device selector button
    - Verify PC audio output device switches
    - Confirm device LEDs show active device correctly
    - Test multiple device switches in sequence
  - [ ] 4.6 Test LED state synchronization
    - Manually change volume on PC (keyboard, mouse)
    - Verify ESP32 LEDs reflect actual PC state
    - Test with different mute states on PC
    - Verify device LEDs match active audio device
  - [ ] 4.7 Test USB reconnection behavior
    - Unplug ESP32 USB cable during operation
    - Replug USB cable
    - Verify ESP32 reconnects and continues working
    - Confirm no manual reset or firmware reflash needed
  - [ ] 4.8 Test startup sequence variations
    - Test: Start ESP32 before deej backend
    - Verify ESP32 works when backend starts later
    - Test: Start deej backend before ESP32
    - Verify ESP32 connects immediately on USB plug-in
  - [ ] 4.9 Verify all acceptance criteria from spec
    - All 5 sliders control PC volume smoothly
    - Both mute buttons toggle correctly with LED feedback
    - Output device switching works with LED indication
    - USB reconnection seamless without reset
    - No WiFi/UDP/TCP code in firmware
    - 115200 baud serial communication reliable
    - Immediate data transmission without PC handshake wait

**Acceptance Criteria:**
- All 5 sliders control PC volume smoothly with no lag
- Both mute buttons toggle correctly and LEDs show actual PC mute state
- Output device switching works and LEDs indicate active device correctly
- USB reconnection works seamlessly without firmware reset
- ESP32 starts and sends data immediately without waiting for PC
- All spec acceptance criteria verified on real hardware

### Final Verification & Cleanup

#### Task Group 5: Code Quality & Documentation Review
**Dependencies:** Task Group 4

- [ ] 5.0 Complete final verification and cleanup
  - [ ] 5.1 Review code for consistency with standards
    - Verify consistent naming conventions (snake_case for variables/functions)
    - Check for meaningful variable/function names
    - Ensure small, focused functions following Single Responsibility Principle
    - Confirm no dead code or commented-out blocks remain
  - [ ] 5.2 Verify error handling compliance
    - Confirm timeout handling is graceful (silent continuation)
    - Verify no error messages exposed to users via LEDs
    - Check that resources are properly managed (Serial instance)
    - Ensure no retry logic on communication failures (per spec)
  - [ ] 5.3 Code cleanup pass
    - Remove any unused imports or includes
    - Delete any temporary debug code or Serial.println() logging
    - Ensure consistent indentation throughout
    - Verify DRY principle - no duplicated logic
  - [ ] 5.4 Review test coverage for critical workflows
    - Verify tests from groups 2.1, 3.1 cover core serial communication
    - Ensure hardware tests from 4.x cover all user-facing functionality
    - Confirm no unnecessary edge case tests were added
    - Total test count should be minimal (6-24 tests + hardware verification)
  - [ ] 5.5 Final compilation and verification
    - Clean build of entire firmware project
    - Verify binary size is reasonable (check for bloat)
    - Confirm no compiler warnings or errors
    - Document any platform-specific notes if needed
  - [ ] 5.6 Create brief implementation notes (optional)
    - Document serial protocol format for future reference
    - Note any deviations from original plan (if any)
    - List any known limitations or future improvements
    - Keep documentation minimal and focused

**Acceptance Criteria:**
- Code follows consistent style and naming conventions
- No dead code, unused imports, or debug logging remains
- Error handling is graceful and complies with spec (silent continuation)
- Test coverage is minimal but covers critical workflows
- Firmware compiles cleanly with no warnings
- All spec requirements are met and verified

## Execution Order

Recommended implementation sequence:
1. **Firmware Cleanup & Preparation** (Task Group 1) - Remove all network code
2. **Serial Communication Infrastructure** (Task Group 2) - Implement SerialApi class
3. **Main Firmware Logic** (Task Group 3) - Integrate serial communication into main loop
4. **Hardware Integration & Testing** (Task Group 4) - Test with real ESP32 and PC backend
5. **Final Verification & Cleanup** (Task Group 5) - Code quality review and final checks

## Key Technical Notes

**Serial Communication Protocol:**
- Outgoing: "Sliders|val1|val2|val3|val4|val5\n", "MuteButtons|bool1|bool2\n", "SwitchOutput|index\n"
- Incoming: "OK\n", "MuteState|bool1|bool2\n", "OutputDevice|index\n"
- Baud Rate: 115200
- Format: Pipe-delimited text with newline termination

**Reused Components:**
- lib/input_components/Slider class (getValue() method)
- lib/input_components/MuteButton class (getValue(), setActiveSessionMuteState() methods)
- lib/input_components/AudioDeviceSelector class (getValue(), setActiveDevice() methods)

**Error Handling Strategy:**
- Silent continuation on timeout (no visual feedback)
- No retry logic
- No serial logging of errors
- Graceful degradation when PC is not responding

**Testing Philosophy:**
- Minimal test writing during development (2-8 tests per group)
- Focus on critical behaviors only
- Hardware integration testing is primary validation method
- Total expected tests: 6-24 unit/integration tests + comprehensive hardware verification
