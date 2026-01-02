# Specification: ESP32 Firmware Serial Migration

## Goal
Migrate ESP32 firmware from WiFi/UDP/TCP network communication to USB Serial communication at 115200 baud while maintaining all existing hardware functionality and reusing existing input component classes.

## User Stories
- As a deej user, I want to connect my ESP32 device via USB instead of WiFi so that I have a more reliable, lower-latency connection to my PC
- As a developer, I want to reuse existing hardware component classes so that I minimize code changes and maintain code quality

## Specific Requirements

**Remove All Network Communication Code**
- Delete all WiFi setup code from main.cpp (WiFi.h includes, WiFi.begin, WiFi.setHostname)
- Remove lib/api/volume_controller_api.h and lib/api/volume_controller_api.cc (UDP-based)
- Remove lib/api/backend_state_api.h and lib/api/backend_state_api.cc (TCP-based)
- Remove lib/api/backend_state.h (network state structure)
- Remove all network-related variables (ssid, password, server_address, udp_server_port, tcp_server_port)

**Implement USB Serial Communication**
- Use Arduino's built-in Serial API (Serial.begin, Serial.println, Serial.readStringUntil)
- Initialize serial communication at 115200 baud in setup()
- Start sending data immediately after Serial.begin() with no handshake wait
- Read incoming responses using Serial.readStringUntil('\n') with non-blocking approach
- Handle partial or missing responses gracefully by continuing to next loop iteration

**Create New SerialApi Class**
- Create new unified SerialApi class in lib/api/serial_api.h and lib/api/serial_api.cc
- Replace both VolumeControllerApi and BackendStateApi with this single class
- Implement sendSliders(const std::string& data) method to send slider values
- Implement sendMuteButtons(const std::vector<bool>& states) method to send mute requests and receive response
- Implement sendSwitchOutput(int deviceIndex) method to send device switch request and receive response
- Parse incoming responses: "OK\n", "MuteState|bool1|bool2\n", "OutputDevice|index\n"
- Return parsed state values to calling code for LED updates
- Use non-blocking Serial.available() check before reading responses
- Implement timeout handling (silently continue on timeout, no error indication)

**Send Hardware Input Data**
- Send slider data: "Sliders|val1|val2|val3|val4|val5\n" where values are raw 12-bit ADC (0-4095)
- Build slider string by iterating through sliders vector and calling getValue() on each
- Send mute button data: "MuteButtons|bool1|bool2\n" when any mute button state changes
- Send output device switch: "SwitchOutput|index\n" when audio device selector button is pressed
- Use Serial.println() to send data with automatic newline termination
- Send slider data every loop iteration if any slider changed (maintain existing change detection)

**Receive and Parse PC Responses**
- Receive acknowledgment: "OK\n" after slider updates
- Receive mute state: "MuteState|bool1|bool2\n" after sending mute button request
- Receive output device: "OutputDevice|index\n" after sending device switch request
- Parse pipe-delimited responses using string splitting
- Extract boolean values and integer indices from response data
- Handle malformed responses by logging to Serial and continuing (no visual feedback)

**Update Hardware LEDs Based on PC State**
- Update mute button LEDs immediately upon receiving "MuteState" response
- Call setActiveSessionMuteState() on each MuteButton instance with actual PC state
- Update output device LEDs immediately upon receiving "OutputDevice" response
- Call setActiveDevice() on AudioDeviceSelector instance with actual PC device index
- Maintain existing LED control logic from input component classes
- No visual error indication for communication failures (silently continue)

**Maintain Existing Hardware Component Structure**
- Keep lib/input_components/Slider class with existing getValue() and change detection
- Keep lib/input_components/MuteButton class with existing button debouncing and LED control
- Keep lib/input_components/AudioDeviceSelector class with existing device toggle and LED management
- Maintain all existing GPIO pin definitions (SLIDER_0-4_PIN, MUTE_BUTTON_0-1_PIN/LED_PIN, AUDIO_DEVICE_SELECTOR pins)
- Keep same hardware initialization sequence in setup()
- Remove maybeApplyBackendState() function (replaced by immediate serial data transmission)

**Main Loop Timing and Structure**
- Change main loop delay from 150ms to 50ms for faster responsiveness
- Keep same loop structure: read sliders, read mute buttons, read device selector
- Send slider data when any slider changed (existing behavior)
- Send mute button data when any button changed and update LEDs based on response
- Send device switch when audio selector changed and update LEDs based on response
- Remove error LED blink sequences (util::blink2Leds, util::blinkLed calls)

**Startup Behavior**
- Initialize Serial at 115200 baud in setup()
- Initialize hardware components (sliders, mute buttons, audio device selector)
- Maintain existing visual startup sequence (sequentialLEDOn for ready indication)
- Start sending data immediately in first loop() call without waiting for PC handshake
- No special connection status indication beyond existing startup LED sequence

**Error Handling Strategy**
- Silently continue operation if no response received within timeout
- No visual LED indication for communication errors or timeouts
- No serial logging of communication errors (keep Serial for data protocol only)
- Remove all error LED blink sequences from main loop
- Maintain existing component error handling (button debouncing logic unchanged)

## Visual Design
No visual assets provided.

## Existing Code to Leverage

**lib/input_components/Slider class (slider.h, slider.cc)**
- Reuse getValue() method that returns tuple<bool changed, int value> for change detection
- Maintains _previous_value for automatic change detection
- Reads raw analogRead() values (12-bit ADC 0-4095)
- Supports optional SessionMuteButton linkage for slider-to-button associations
- Keep all existing functionality unchanged

**lib/input_components/MuteButton class (mute_button.h, mute_button.cc)**
- Reuse getValue() method with button debouncing logic
- Reuse setActiveSessionMuteState(bool) method to update LED based on PC response
- Reuse setActiveSession(int) and setLedState(int, bool) for multi-session support
- Maintains internal ButtonState map for debouncing and LED control
- Keep all existing functionality unchanged

**lib/input_components/AudioDeviceSelector class (audio_device_selector.h, audio_device_selector.cc)**
- Reuse getValue() method for device toggle button press detection
- Reuse setActiveDevice(int) method to update LEDs based on PC response
- Manages two device LEDs (_dev_0_led_pin, _dev_1_led_pin)
- Handles long-press callback for ESP restart (keep unchanged)
- Keep all existing functionality unchanged

**Main loop structure from main.cpp**
- Maintain vector-based iteration for sliders->at(i)->getValue() pattern
- Maintain structured data building: "Sliders" + "|" + value pattern
- Maintain change detection with boolean flags (sliders_changed, mute_buttons_changed)
- Keep same component initialization sequence in setup()
- Keep same GPIO pin definitions (#define SLIDER_0_PIN 33, etc.)

**Serial communication protocol from PC backend (pkg/deej/serial.go)**
- PC expects: "Sliders|val1|val2|...\n", "MuteButtons|bool1|bool2\n", "SwitchOutput|index\n"
- PC sends: "OK\n", "MuteState|bool1|bool2\n", "OutputDevice|index\n"
- PC validates lines with regex: ^\w+(\|\w+)*$
- PC normalizes 12-bit ADC values (0-4095) to 0.0-1.0 float internally
- PC handles reconnection gracefully, no handshake required from ESP32

## Out of Scope
- Visual feedback for connection status (LED blinking on startup beyond existing sequence)
- Visual error indication on communication timeout (no LED error blinks)
- Local ADC filtering or smoothing (send raw analogRead() values)
- Interactive handshake protocol with PC on startup (send data immediately)
- Retry logic on communication timeout (silently continue to next loop)
- Serial logging of debug information or errors (Serial used only for protocol)
- Changes to input component classes internal logic (debouncing, LED control)
- Changes to hardware pin definitions or GPIO configuration
- WiFi fallback mode or dual communication support
- Configuration of baud rate at runtime (hardcoded 115200)
- Persistent storage of device state (rely on PC for state)

## Testing Strategy

**End-to-End Testing with Real Hardware:**
1. Flash firmware to ESP32 and connect via USB to PC
2. Start deej backend and verify COM port connection
3. Test slider movements → verify PC volume changes in real-time
4. Test mute button presses → verify mute states toggle and LEDs update correctly
5. Test output device switching → verify audio device switches and LEDs show active device
6. Test PC state synchronization → manually change volume/mute on PC, verify LEDs reflect actual state
7. Test reconnection → unplug/replug USB, verify ESP32 reconnects and continues working
8. Test startup sequence → start ESP32 before deej, verify it works when deej starts later

**Serial Protocol Validation:**
- Use Serial Monitor to verify outgoing message format matches spec (pipe-delimited, newline-terminated)
- Monitor incoming PC responses to verify correct parsing
- Verify all message types: Sliders, MuteButtons, SwitchOutput, OK, MuteState, OutputDevice

**Component-Level Testing:**
- Verify each slider sends correct 12-bit ADC values (0-4095 range)
- Verify mute button debouncing works correctly
- Verify LED states match PC responses exactly
- Verify 50ms loop timing maintains responsiveness

**Acceptance Criteria:**
- ✅ All 5 sliders control PC volume smoothly with no lag
- ✅ Both mute buttons toggle correctly and LEDs show actual PC mute state
- ✅ Output device switching works and LEDs indicate active device
- ✅ USB reconnection works seamlessly without firmware reset
- ✅ No WiFi/UDP/TCP code remains in firmware
- ✅ Serial communication at 115200 baud works reliably
- ✅ Firmware starts and sends data immediately without waiting for PC
