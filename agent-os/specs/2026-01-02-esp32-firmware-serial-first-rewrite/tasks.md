# Task Breakdown: ESP32 Firmware Serial-First Rewrite

## Overview
Total Task Groups: 7
Total Sub-Tasks: ~45

This is a complete firmware rewrite from scratch with a clean serial-first architecture, removing all WiFi/TCP/UDP code while maintaining all existing functionality.

## Task List

### Task Group 1: Project Setup & Serial Communication Foundation

**Dependencies:** None

- [x] 1.0 Complete project setup and serial foundation
  - [x] 1.1 Create new PlatformIO project structure
    - Create new directory `firmware/esp32-serial-first/`
    - Copy `platformio.ini` from `firmware/esp32-5-sliders-3-buttons/`
    - Verify platform = espressif32, board = esp32dev, framework = arduino
    - Set build_flags = -std=gnu++17 (reuse C++17 standard)
    - Configure monitor_speed = 115200
    - Set upload_port to appropriate COM port
  - [x] 1.2 Create lib directory structure
    - Create `lib/api/` for SerialApi class
    - Create `lib/input_components/` for hardware component classes
    - Create `lib/utils/` for utility functions
    - Verify PlatformIO can find library directories
  - [x] 1.3 Create minimal main.cpp with serial initialization
    - Create `src/main.cpp` with setup() and loop() functions
    - Add Serial.begin(115200) in setup()
    - Add delay(50) in loop() for 50ms non-blocking cycle
    - Reference pattern: `firmware/esp32-5-sliders-3-buttons/src/main.cpp` lines 35-37
  - [x] 1.4 Create SerialApi class header
    - Create `lib/api/serial_api.h`
    - Declare SerialApi class with _timeout_ms = 100
    - Add sendSliders(), sendMuteButtons(), sendSwitchOutput() method declarations
    - Add private helpers: readResponse(), parseResponse(), parseBool(), parseInt()
    - Reference: `firmware/esp32-5-sliders-3-buttons/lib/api/serial_api.h`
  - [x] 1.5 Implement SerialApi class
    - Create `lib/api/serial_api.cc`
    - Implement sendSliders() with Serial.println() and optional readResponse()
    - Implement readResponse() with millis() timeout tracking (100ms)
    - Implement parseResponse() using std::stringstream and std::getline with '|' delimiter
    - Implement parseBool() accepting "1"/"0" and "true"/"false"
    - Implement parseInt() with try-catch for std::stoi
    - Reference: `firmware/esp32-5-sliders-3-buttons/lib/api/serial_api.cc`
  - [x] 1.6 Test serial communication with backend
    - Build and upload firmware to ESP32
    - Connect to backend via USB serial
    - Send test "Sliders|0|0|0|0|0\n" message from setup()
    - Verify backend receives message in deej logs
    - Verify Serial.available() and readResponse() work with backend "OK\n" response

**Acceptance Criteria:**
- PlatformIO project builds successfully with no compilation errors
- ESP32 successfully initializes Serial at 115200 baud
- SerialApi can send pipe-delimited messages terminated with \n
- SerialApi can read and parse backend responses with 100ms timeout
- Basic bidirectional serial communication with backend verified

---

### Task Group 2: Slider Component & Volume Control

**Dependencies:** Task Group 1

- [x] 2.0 Complete slider hardware integration
  - [x] 2.1 Create Slider class header
    - Create `lib/input_components/slider.h`
    - Declare Slider class with constructor taking slider_index, gpio_pin, optional SessionMuteButton
    - Add SessionMuteButton struct with MuteButton* pointer and int session
    - Add getValue() method returning std::tuple<bool, int> for (changed, value)
    - Add hasMuteButton() and getMuteButton() helper methods
    - Add private members: _slider_index, _gpioPinNumber, _previous_value, _session_mute_button
    - Reference: `firmware/esp32-5-sliders-3-buttons/lib/input_components/slider.h`
  - [x] 2.2 Implement Slider class with ADC reading
    - Create `lib/input_components/slider.cc`
    - Implement getValue() using analogRead() for 12-bit ADC (0-4095 range)
    - Apply inversion formula: percentValue = 4095 - rawValue (top = high volume)
    - Implement ZERO_THRESHOLD = 100 to prevent jitter at bottom
    - Implement change detection by comparing to _previous_value
    - Reference: `firmware/esp32-5-sliders-3-buttons/lib/input_components/slider.cc` lines 22-41
  - [x] 2.3 Wire up GPIO pins for 5 sliders
    - Add #define macros in main.cpp: SLIDER_0_PIN 34, SLIDER_1_PIN 35, SLIDER_2_PIN 33, SLIDER_3_PIN 32, SLIDER_4_PIN 36
    - Create std::vector<Slider*> in setup()
    - Instantiate 5 Slider objects (without SessionMuteButton for now)
    - Reference: `firmware/esp32-5-sliders-3-buttons/src/main.cpp` lines 12-16, 47-56
  - [x] 2.4 Implement slider reading in main loop
    - Build "Sliders|val0|val1|val2|val3|val4" message string
    - Iterate through sliders vector calling getValue()
    - Track sliders_changed flag with OR operation
    - Append values to message string with pipe delimiter
    - Reference: `firmware/esp32-5-sliders-3-buttons/src/main.cpp` lines 73-82
  - [x] 2.5 Send slider data to backend via SerialApi
    - Call serial_api->sendSliders(sliders_data) when sliders_changed is true
    - Add #include <string> for std::string and std::to_string
    - Verify only changed values trigger serial send
  - [x] 2.6 Test slider functionality with backend
    - Upload firmware and connect to backend
    - Move each slider and verify Windows volume changes
    - Test slider at top position (value ~4095, high volume)
    - Test slider at bottom position (value 0, muted)
    - Verify zero deadband prevents jitter at bottom
    - Verify smooth volume control without lag

**Acceptance Criteria:**
- All 5 sliders successfully read ADC values (0-4095 range)
- Value inversion works correctly (slider top = high volume)
- Zero deadband of 100 prevents jitter at bottom position
- Change detection prevents unnecessary serial sends
- Windows volume responds smoothly to slider movements
- Message format matches "Sliders|val0|val1|val2|val3|val4\n"

---

### Task Group 3: Master Mute Button with Multi-Session Support

**Dependencies:** Task Group 2

- [x] 3.0 Complete master mute button implementation
  - [x] 3.1 Create MuteButton class header
    - Create `lib/input_components/mute_button.h`
    - Declare ButtonState struct with is_pressed and led_state booleans
    - Declare MuteButton class with constructors (single and multi-session)
    - Add constructor parameters: button_index, button_gpio_pin, led_gpio_pin, optional controlled_sessions
    - Add getValue() method returning std::tuple<bool, bool>
    - Add setActiveSessionMuteState(), setActiveSession(), setLedState(), updateLedState() methods
    - Add private members: _button_gpio_pin, _led_gpio_pin, _active_session, std::map<int, ButtonState>
    - Reference: `firmware/esp32-5-sliders-3-buttons/lib/input_components/mute_button.h`
  - [x] 3.2 Implement MuteButton GPIO and LED setup
    - Create `lib/input_components/mute_button.cc`
    - Configure pinMode(_button_gpio_pin, INPUT_PULLUP) in constructor
    - Configure pinMode(_led_gpio_pin, OUTPUT) in constructor
    - Set initial LED state digitalWrite(_led_gpio_pin, HIGH) for LED off
    - Initialize _buttons_states map with ButtonState for each session
    - Reference: `firmware/esp32-5-sliders-3-buttons/lib/input_components/mute_button.cc`
  - [x] 3.3 Implement 40ms debounce logic
    - Implement getValue() checking digitalRead(_button_gpio_pin) == LOW for button press
    - Add while loop with digitalRead() and delay(40) for debounce
    - Toggle ButtonState.is_pressed for active session on button press
    - Return tuple with (changed: true, new_mute_state) when toggled
    - Reference existing debounce pattern from previous firmware
  - [x] 3.4 Implement session-aware LED state management
    - Implement updateLedState() combining is_pressed OR led_state for active session
    - Implement setActiveSession() to switch displayed mute state
    - Implement setLedState(session, muted) to update led_state for auto-mute
    - Implement setActiveSessionMuteState(mute_state) to update is_pressed from backend
    - LED on when either is_pressed OR led_state is true (digitalWrite LOW = LED on)
  - [x] 3.5 Wire up master mute button GPIO pins
    - Add #define macros: MUTE_BUTTON_0_PIN 14, MUTE_BUTTON_0_LED_PIN 12
    - Create MuteButton instance with 2 controlled sessions (speakers and headphones)
    - Add to std::vector<MuteButton*> in setup()
    - Reference: `firmware/esp32-5-sliders-3-buttons/src/main.cpp` lines 17-18, 38-44
  - [x] 3.6 Implement mute button protocol in main loop
    - Iterate through mute_buttons vector calling getValue()
    - Build std::vector<bool> for mute states
    - Track mute_buttons_changed flag
    - Call serial_api->sendMuteButtons(mute_buttons_state)
    - Parse backend response "MuteState|bool0|bool1\n"
    - Update LED states with setActiveSessionMuteState() after backend confirmation
    - Reference: `firmware/esp32-5-sliders-3-buttons/src/main.cpp` lines 103-127
  - [x] 3.7 Implement SerialApi mute button methods
    - Implement sendMuteButtons() building "MuteButtons|bool0|bool1\n" message
    - Use "1" for true, "0" for false in message
    - Parse response expecting "MuteState|bool0|bool1\n"
    - Return std::vector<bool> with backend confirmation (empty on timeout)
    - Reference: `firmware/esp32-5-sliders-3-buttons/lib/api/serial_api.cc` lines 15-37
  - [x] 3.8 Test master mute button with backend
    - Upload firmware and connect to backend
    - Press mute button and verify LED toggles after backend confirmation
    - Verify Windows audio mutes and unmutes correctly
    - Test both sessions independently (will need AudioDeviceSelector for full test)
    - Verify LED updates only after backend confirmation (no desync)

**Acceptance Criteria:**
- Master mute button detects presses with 40ms debounce
- Multi-session state tracked in std::map<int, ButtonState>
- LED shows mute state of currently active session only
- LED combines both is_pressed and led_state with OR logic
- Backend confirmation required before LED updates
- Message format matches "MuteButtons|bool0|bool1\n"
- Backend response "MuteState|bool0|bool1\n" parsed correctly

---

### Task Group 4: Mic Mute Button

**Dependencies:** Task Group 3

- [x] 4.0 Complete mic mute button implementation
  - [x] 4.1 Wire up mic mute button GPIO pins
    - Add #define macros: MUTE_BUTTON_1_PIN 4, MUTE_BUTTON_1_LED_PIN 21
    - Create second MuteButton instance with 1 controlled session (single-session)
    - Add to mute_buttons vector in setup()
    - Reference: `firmware/esp32-5-sliders-3-buttons/src/main.cpp` lines 19-20, 42-44
  - [x] 4.2 Integrate mic mute into main loop
    - Verify mic mute button included in mute_buttons vector iteration
    - Verify mic mute state included in "MuteButtons|bool0|bool1\n" message
    - Second boolean value represents mic mute state
    - LED update logic reuses same pattern as master mute button
  - [x] 4.3 Test mic mute button with backend
    - Upload firmware and connect to backend
    - Press mic mute button and verify LED toggles
    - Verify microphone mutes and unmutes in Windows
    - Test both mute buttons independently
    - Verify both LEDs update correctly from backend responses

**Acceptance Criteria:**
- Mic mute button uses same debounce and LED logic as master mute
- Simpler single-session implementation (always session 0)
- Participates in same "MuteButtons" message (second boolean)
- LED feedback works correctly for mic mute state
- Both mute buttons work independently without interference

---

### Task Group 5: Audio Device Switcher

**Dependencies:** Task Group 3

- [x] 5.0 Complete audio device switcher implementation
  - [x] 5.1 Create AudioDeviceSelector class header
    - Create `lib/input_components/audio_device_selector.h`
    - Declare AudioDeviceSelector class with constructor parameters
    - Constructor takes: button_gpio_pin, dev_0_led_pin, dev_1_led_pin, MuteButton* pointer, longpress callback lambda
    - Add getValue() method returning std::tuple<bool, int>
    - Add setActiveDevice(int) and getActiveDevice() methods
    - Add private members: _button_gpio_pin, _dev_0_led_pin, _dev_1_led_pin, _selected_device, _multi_session_mute_button
    - Reference: `firmware/esp32-5-sliders-3-buttons/lib/input_components/audio_device_selector.h`
  - [x] 5.2 Implement AudioDeviceSelector GPIO setup
    - Create `lib/input_components/audio_device_selector.cc`
    - Configure pinMode for button with INPUT_PULLUP
    - Configure pinMode for both device LEDs with OUTPUT
    - Initialize both LEDs to HIGH (off state)
    - Call setActiveDevice(0) to show initial device selection
  - [x] 5.3 Implement device toggle with debounce
    - Implement getValue() checking digitalRead() == LOW for button press
    - Add 40ms debounce using same while loop pattern as mute button
    - Toggle device index using XOR: _selected_device = _selected_device ^ 1
    - Return tuple (changed: true, new_device_index) on toggle
  - [x] 5.4 Implement long-press reset detection
    - Count debounce iterations in while loop
    - If iterations exceed 20 (40ms * 20 = 800ms approximating 2 seconds with delays)
    - Call _on_longpress_override_callback() to trigger esp_restart()
    - Use delay(100) between iterations for long-press counting
    - Reference existing long-press pattern from previous firmware
  - [x] 5.5 Implement LED state management
    - Implement setActiveDevice(int selected_device)
    - Update _selected_device member
    - Set active device LED to HIGH (on), inactive to LOW (off)
    - Call _multi_session_mute_button->setActiveSession(selected_device)
    - Call _multi_session_mute_button->updateLedState() to show new session mute state
  - [x] 5.6 Wire up device switcher GPIO pins
    - Add #define macros: AUDIO_DEVICE_SELECTOR_BUTTON_PIN 5
    - Add AUDIO_DEVICE_SELECTOR_BUTTON_DEV_0_LED_PIN 18
    - Add AUDIO_DEVICE_SELECTOR_BUTTON_DEV_1_LED_PIN 19
    - Create AudioDeviceSelector with pointer to master mute button
    - Pass lambda []() { esp_restart(); } for long-press callback
    - Reference: `firmware/esp32-5-sliders-3-buttons/src/main.cpp` lines 21-23, 58-62
  - [x] 5.7 Implement device switch protocol in main loop
    - Call audio_device_selector->getValue() to get (changed, value)
    - Call serial_api->sendSwitchOutput(value) when changed
    - Parse backend response "OutputDevice|index\n"
    - Call audio_device_selector->setActiveDevice(updated_device) on valid response
    - Reference: `firmware/esp32-5-sliders-3-buttons/src/main.cpp` lines 130-138
  - [x] 5.8 Implement SerialApi device switch methods
    - Implement sendSwitchOutput() building "SwitchOutput|deviceIndex\n" message
    - Parse response expecting "OutputDevice|index\n"
    - Return device index as int (return -1 on timeout/error)
    - Reference: `firmware/esp32-5-sliders-3-buttons/lib/api/serial_api.cc` lines 39-54
  - [x] 5.9 Test device switcher with backend
    - Upload firmware and connect to backend
    - Short press button and verify device toggles (speakers <-> headphones)
    - Verify device LEDs update correctly (only active LED on)
    - Verify master mute button LED shows correct session mute state after switch
    - Test long press (hold for 2+ seconds) and verify ESP32 restarts

**Acceptance Criteria:**
- Device selector toggles between 2 devices using XOR operation
- 40ms debounce prevents multiple toggles from single press
- Long-press (2 seconds) triggers esp_restart()
- Device LEDs show only active device (matches reference implementation polarity)
- setActiveSession() called on master mute button when device changes
- Master mute LED updates to show new session's mute state
- Message format matches "SwitchOutput|deviceIndex\n"
- Backend response "OutputDevice|index\n" parsed correctly

---

### Task Group 6: Auto-Mute Integration

**Dependencies:** Task Groups 2, 3, 5

- [x] 6.0 Complete auto-mute on slider zero
  - [x] 6.1 Connect sliders to mute button sessions
    - Update Slider(0) instantiation to pass SessionMuteButton{output_devices_mute_button, 0}
    - Update Slider(1) instantiation to pass SessionMuteButton{output_devices_mute_button, 1}
    - Verify sliders 2, 3, 4 have no SessionMuteButton (std::nullopt)
    - Reference: `firmware/esp32-5-sliders-3-buttons/src/main.cpp` lines 48-56
  - [x] 6.2 Implement slider zero detection in Slider class
    - Update Slider::getValue() to check if percentValue == 0
    - When zero detected and _session_mute_button.has_value() is true
    - Call _session_mute_button->button->setLedState(_session_mute_button->session, true)
    - When slider moves above zero, call setLedState(session, false)
    - Reference: `firmware/esp32-5-sliders-3-buttons/lib/input_components/slider.cc` lines 34-36
  - [x] 6.3 Implement auto-mute trigger in main loop with active session check
    - Add std::vector<bool> auto_mute_triggered tracking
    - In slider loop, detect when value == 0 and hasMuteButton() is true
    - CRITICAL: Only trigger auto-mute if slider's session matches active device session
    - Set auto_mute_triggered flag for corresponding mute button
    - Force mute_buttons_changed = true when auto-mute triggered
    - Force mute_buttons_state[i] = true to ensure mute sent to backend
    - Reference: `firmware/esp32-5-sliders-3-buttons/src/main.cpp` lines 76-96, 110-114
  - [x] 6.4 Test auto-mute functionality
    - Upload firmware and connect to backend
    - Move slider 0 to bottom (session 0 / speakers)
    - Verify mute button LED turns on
    - Verify Windows audio mutes for speakers
    - Move slider 0 back up and verify auto-unmute
    - Repeat test for slider 1 (session 1 / headphones)
    - Switch active device and verify correct session auto-mutes

**Acceptance Criteria:**
- Slider 0 hitting zero triggers mute for session 0 (speakers) ONLY when session 0 is active
- Slider 1 hitting zero triggers mute for session 1 (headphones) ONLY when session 1 is active
- Auto-mute sends "MuteButtons" message with mute=true
- Backend implements actual Windows Core Audio mute (not 0% volume)
- Moving slider above zero triggers auto-unmute
- LED shows combined state of button press and slider-triggered mute
- Auto-mute respects session independence (doesn't affect other session)

---

### Task Group 7: Startup Sequence & Final Integration Testing

**Dependencies:** Task Groups 1-6

- [x] 7.0 Complete startup and final integration
  - [x] 7.1 Create util.h for LED startup sequence
    - Create `lib/utils/util.h`
    - Implement sequentialLEDOn() taking 4 LED pins and delay_ms
    - Store previous LED states with digitalRead()
    - Turn all LEDs HIGH (off), then sequentially turn on and off each LED
    - Restore previous LED states after sequence
    - Reference: `firmware/esp32-5-sliders-3-buttons/lib/utils/util.h` lines 8-36
  - [x] 7.2 Implement startup initialization sequence
    - Order initialization in setup(): Serial first, then components
    - Create MuteButton instances first (master with 2 sessions, mic with 1)
    - Create Slider instances with SessionMuteButton structs for sliders 0 and 1
    - Create AudioDeviceSelector with master mute button pointer and esp_restart lambda
    - Create SerialApi instance last
    - Call util::sequentialLEDOn() with all 4 LED pins and 300ms delay
    - Reference: `firmware/esp32-5-sliders-3-buttons/src/main.cpp` lines 35-69
  - [x] 7.3 Add necessary includes to main.cpp
    - #include <Arduino.h>
    - #include <audio_device_selector.h>
    - #include <mute_button.h>
    - #include <serial_api.h>
    - #include <slider.h>
    - #include <util.h>
    - #include <string> and <vector> for STL containers
    - Add using declarations for lib::api and lib::input_components namespaces
    - Reference: `firmware/esp32-5-sliders-3-buttons/src/main.cpp` lines 1-28
  - [x] 7.4 Test complete integration with all components
    - Upload firmware and connect to backend
    - Verify startup LED sequence plays on boot
    - Test all 5 sliders control Windows volume
    - Test master mute button toggles and shows LED feedback
    - Test mic mute button toggles and shows LED feedback
    - Test device switcher toggles between speakers and headphones
    - Verify device LEDs show active device correctly
    - Test auto-mute on slider 0 and slider 1 hitting zero
    - Verify master mute LED shows correct session when switching devices
  - [x] 7.5 Test edge cases and reconnection scenarios
    - Disconnect USB cable and reconnect - verify firmware continues operation
    - Rapidly press mute buttons - verify debounce prevents multiple toggles
    - Rapidly move sliders - verify change detection prevents excess serial traffic
    - Rapidly toggle device selector - verify state stays synchronized
    - Test long-press device selector - verify ESP32 restarts after 2 seconds
  - [x] 7.6 Validate against spec success criteria
    - Verify no WiFi/TCP/UDP includes or initialization in code
    - Verify Serial.begin(115200) used instead of network setup
    - Verify 50ms loop delay for responsive operation
    - Verify all GPIO pins match spec (sliders 32-36, buttons/LEDs as specified)
    - Verify all serial message formats match spec exactly
    - Verify 100ms serial timeout implemented correctly
    - Verify all 8 feature areas from roadmap implemented
  - [x] 7.7 Performance testing and optimization
    - Monitor serial output for excessive message frequency
    - Verify slider change detection works (no spam when not moving)
    - Verify LED updates happen only on state changes
    - Check memory usage is reasonable (no memory leaks from std::vector/map)
    - Verify loop() executes within 50ms consistently

**Acceptance Criteria:**
- All components initialized in correct order with no race conditions
- Startup LED sequence provides visual confirmation of successful boot
- All 5 sliders, 2 mute buttons, and device switcher work correctly
- Auto-mute on slider zero works for both sessions independently
- Session-aware LED feedback shows correct mute state for active device
- Edge cases handled gracefully (reconnection, rapid inputs, timeouts)
- No WiFi/network code present in firmware
- All serial protocols match backend expectations exactly
- Performance is smooth with no lag or jitter
- Success criteria from spec fully satisfied

---

## Execution Order

Recommended implementation sequence:

1. **Task Group 1** - Project Setup & Serial Communication Foundation
   - Establishes build system, serial communication, and message passing infrastructure
   - Critical foundation for all subsequent work

2. **Task Group 2** - Slider Component & Volume Control
   - Adds core functionality of reading ADC and controlling volume
   - Independent feature that can be tested immediately

3. **Task Group 3** - Master Mute Button with Multi-Session Support
   - Implements multi-session architecture that device switcher depends on
   - Complex state management needed before other features

4. **Task Group 4** - Mic Mute Button
   - Simple addition reusing master mute button patterns
   - Can be done quickly after Task Group 3

5. **Task Group 5** - Audio Device Switcher
   - Depends on master mute button multi-session support
   - Integrates with master mute button for session switching

6. **Task Group 6** - Auto-Mute Integration
   - Requires sliders and mute buttons to be fully functional
   - Connects slider zero detection to mute button state

7. **Task Group 7** - Startup Sequence & Final Integration Testing
   - Brings all components together in correct initialization order
   - Comprehensive testing of all features working in harmony

## Notes

- **Reference Code Location**: All references point to `firmware/esp32-5-sliders-3-buttons/` which contains the proven implementation patterns to reuse
- **Clean Architecture**: This rewrite removes all WiFi/TCP/UDP code - grep for "WiFi", "TCP", "UDP" to verify none present
- **Serial-First**: USB Serial communication only, no network stack
- **Backend Unchanged**: Go backend at `pkg/deej/` already handles serial protocol correctly
- **Testing Strategy**: Each task group ends with component-level testing before moving to next group
- **GPIO Pin Assignments**: Reuse exact pin numbers from existing firmware (verified in hardware)
- **C++17 Standard**: Use std::tuple, std::optional, std::vector, std::map as in existing code
- **No Test Suite**: This is embedded firmware - testing is hardware validation with backend connection
- **Size Estimates**: Most task groups are M (medium) except Task Group 1 (L - large setup) and Task Group 7 (L - comprehensive testing)
