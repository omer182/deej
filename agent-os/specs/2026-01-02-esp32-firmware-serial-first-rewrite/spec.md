# Specification: ESP32 Firmware Serial-First Rewrite

## Goal
Completely rewrite the ESP32 firmware from scratch with a clean serial-first architecture, removing all TCP/UDP/WiFi remnants from the previous implementation while maintaining all existing functionality.

## User Stories
- As a deej user, I want my volume sliders to control Windows audio smoothly without jitter or lag
- As a deej user, I want mute buttons that stay synchronized with Windows audio state through LED feedback

## Specific Requirements

**Serial Communication Foundation**
- Initialize Serial at 115200 baud in setup() using Serial.begin(115200)
- Implement non-blocking 50ms main loop using delay(50)
- Check for incoming data using Serial.available() before reading
- Send pipe-delimited messages terminated with newline character
- Parse backend responses using string splitting on pipe delimiter
- Handle 100ms timeout gracefully when backend doesn't respond (continue operation, don't block)
- Remove all WiFi, TCP, UDP includes and initialization code from previous firmware

**Slider Hardware and ADC Reading**
- Wire 5 analog sliders to GPIO pins 32, 33, 34, 35, 36 (ADC1 channels)
- Read 12-bit ADC values (0-4095 range) using analogRead()
- Apply value inversion formula: percentValue = 4095 - rawValue so slider top = high volume
- Implement zero deadband threshold of 100 to prevent jitter when slider at bottom
- Detect changes by comparing current value to previous value (send only when changed)
- Send slider data in format "Sliders|val0|val1|val2|val3|val4\n"
- Store previous slider values to enable change detection
- Handle backend "OK\n" response (optional, timeout silently if not received)

**Master Mute Button Multi-Session Architecture**
- GPIO 14 for button input configured with INPUT_PULLUP mode
- GPIO 12 for LED output (LOW = LED on, HIGH = LED off due to pull-up wiring)
- Support 2 independent sessions: session 0 for speakers, session 1 for headphones
- Track active session (switched by AudioDeviceSelector)
- Implement 40ms debounce using delay in while loop checking digitalRead()
- Toggle button's internal state on press, send to backend for confirmation
- Send format "MuteButtons|bool0|bool1\n" where bool is "1" or "0"
- Parse backend response "MuteState|bool0|bool1\n" to get actual Windows state
- Update LED only after receiving backend confirmation
- LED shows mute state of currently active session only

**Mic Mute Button Single-Session**
- GPIO 4 for button input with INPUT_PULLUP
- GPIO 21 for LED output
- Simpler single-session implementation (always session 0)
- Reuse same debounce and LED update logic as master mute button
- Participate in same "MuteButtons" message (second boolean value)

**Audio Device Switcher**
- GPIO 5 for button input with INPUT_PULLUP
- GPIO 18 for device 0 LED (speakers)
- GPIO 19 for device 1 LED (headphones)
- Implement same 40ms debounce as mute buttons
- Detect long-press: count debounce iterations, if exceeds 20 iterations (2 seconds), call esp_restart()
- Toggle device index on short press (XOR operation: selected_device ^ 1)
- Send format "SwitchOutput|deviceIndex\n"
- Parse backend response "OutputDevice|index\n"
- Update LEDs so only active device LED is HIGH, inactive is LOW
- When device changes, call setActiveSession() on master mute button to switch displayed mute state

**Auto-Mute on Slider Zero**
- In main loop, check if slider 0 or slider 1 value equals 0 after zero-deadband applied
- If slider 0 hits 0, set mute state to true for session 0 (speakers)
- If slider 1 hits 0, set mute state to true for session 1 (headphones)
- Force mute_buttons_changed flag to true to trigger backend update
- Send updated mute state to backend and wait for confirmation
- When slider moves above 0, detect change and trigger unmute (set mute state to false)
- Backend implements actual Windows Core Audio mute, not just 0% volume
- only mute the state if the slider that hits 0 is the active session slider

**Session-Aware LED State Management**
- Master mute button maintains separate ButtonState for each session (is_pressed, led_state)
- ButtonState.is_pressed tracks button toggle state from backend confirmation
- ButtonState.led_state tracks slider-triggered mute state (auto-mute on zero)
- LED turns on when either is_pressed OR led_state is true for active session
- When AudioDeviceSelector changes active session, call updateLedState() to show new session's state
- Slider class can call setLedState(session, muted) to update led_state without affecting is_pressed
- All LED updates wait for backend confirmation to prevent desync

**Class Structure and Organization**
- Slider class: constructor takes slider_index, GPIO pin, optional SessionMuteButton struct
- SessionMuteButton struct contains pointer to MuteButton and session index
- Slider.getValue() returns tuple of (changed: bool, value: int)
- MuteButton class: constructor takes button_index, button_gpio_pin, led_gpio_pin, optional controlled_sessions count
- MuteButton maintains map of session index to ButtonState
- MuteButton.getValue() returns tuple (changed: bool, mute_state: bool)
- AudioDeviceSelector class: constructor takes button_gpio_pin, dev_0_led_pin, dev_1_led_pin, pointer to multi-session MuteButton, longpress callback lambda
- SerialApi class: provides sendSliders(), sendMuteButtons(), sendSwitchOutput() methods with timeout handling

**Startup and Initialization**
- Begin Serial at 115200 baud before any other initialization
- Create MuteButton instances first (master output mute with 2 sessions, mic mute with 1 session)
- Create Slider instances, passing SessionMuteButton structs to sliders 0 and 1
- Create AudioDeviceSelector with reference to master mute button and esp_restart lambda
- Create SerialApi instance last
- Show visual startup indication by sequentially lighting all LEDs for 300ms each
- No handshake protocol required - start sending data immediately in loop()

## Existing Code to Leverage

**Current Firmware Class Architecture**
- Reuse proven OOP structure with Slider, MuteButton, AudioDeviceSelector classes in lib/input_components namespace
- Reuse SerialApi class in lib/api namespace for message formatting and response parsing
- Keep same GPIO pin assignments: sliders on 32-36, mute buttons on 14/12 and 4/21, device selector on 5/18/19
- Maintain C++17 standard with std::tuple, std::optional, std::vector, std::map
- Keep PlatformIO build configuration for ESP32-DevKit with Arduino framework

**GPIO Configuration Patterns**
- Reuse INPUT_PULLUP configuration for all button inputs (active LOW when pressed)
- Reuse OUTPUT configuration for LEDs with inverted logic (LOW = LED on, HIGH = LED off)
- Keep analogRead() for 12-bit ADC without additional configuration
- Maintain pinMode() calls in constructor initialization

**Debounce and Timing Logic**
- Reuse 40ms debounce delay in while loop for mute buttons
- Reuse long-press detection counting method (100ms delay * 20 iterations = 2 seconds)
- Keep 50ms main loop delay for responsive slider reading
- Maintain 100ms timeout for serial responses in SerialApi

**Serial Message Parsing Helpers**
- Reuse parseResponse() method using std::stringstream and std::getline with pipe delimiter
- Reuse parseBool() helper accepting both "1"/"0" and "true"/"false" strings
- Reuse parseInt() helper with try-catch for std::stoi error handling
- Keep readResponse() method with millis() timeout tracking

**State Management Patterns**
- Reuse ButtonState struct with is_pressed and led_state booleans
- Maintain previous_value tracking in Slider class for change detection
- Keep std::map<int, ButtonState> for multi-session mute button state
- Reuse updateLedState() method that combines is_pressed and led_state with OR logic

## Out of Scope
- No WiFi initialization or network stack includes
- No TCP/UDP socket creation or connection handling
- No network-based volume controller API
- No backend state polling over network protocols
- No WiFi credential configuration or SSID scanning
- No network reconnection logic or ping/keep-alive mechanisms
- No HTTP server or web-based configuration interface
- No mDNS or service discovery protocols
- No OTA (over-the-air) firmware updates
- No retry logic for serial communication failures beyond simple timeout handling
