# Spec Requirements: ESP32 Firmware Serial Migration

## Initial Description
Complete rewrite of ESP32 firmware to migrate from WiFi/UDP/TCP network communication to USB Serial communication. The firmware needs to:

1. Remove all WiFi, UDP, and TCP code
2. Implement Serial communication at 115200 baud over USB
3. Send hardware input data to PC:
   - Slider values (5 sliders, 12-bit ADC 0-4095): "Sliders|val1|val2|val3|val4|val5\n"
   - Mute button requests (2 buttons): "MuteButtons|bool1|bool2\n"
   - Output device switch requests: "SwitchOutput|index\n"
4. Receive and parse PC responses:
   - Acknowledgment: "OK\n"
   - Mute states: "MuteState|bool1|bool2\n"
   - Active output device: "OutputDevice|index\n"
5. Update LEDs based on actual PC state:
   - 2 mute button LEDs (show actual mute state from PC)
   - 2 output device LEDs (show active device from PC)
6. Maintain all existing hardware functionality (5 analog sliders, 2 mute buttons with LEDs, 1 device switch button with 2 LEDs)

The PC backend is already complete and tested - it's ready to receive this serial protocol. The existing firmware is in firmware/esp32-5-sliders-3-buttons/ and uses WiFi/UDP/TCP which needs to be completely replaced.

## Requirements Discussion

### First Round Questions

**Q1:** For the file structure, should we keep the existing firmware structure (lib/input_components/ classes for Slider, MuteButton, AudioDeviceSelector) and just replace the WiFi/UDP/TCP communication layer with Serial, or do you prefer a simpler, more monolithic main.cpp approach?
**Answer:** Keep existing firmware structure (lib/input_components/ classes) and just replace WiFi/UDP/TCP communication with Serial. Minimal changes - just swap out the communication layer.

**Q2:** For serial communication, should we use Arduino's built-in Serial API (Serial.begin(115200), Serial.println(), Serial.readStringUntil('\n')), which is standard and works on all ESP32s, or do you have a preference for a different serial library?
**Answer:** Use Arduino's built-in Serial API (Serial.begin(115200), Serial.println(), Serial.readStringUntil('\n')) - it works perfectly on ESP32.

**Q3:** For ADC reading and sending, should we send raw analogRead() values (12-bit ADC: 0-4095 on ESP32) or apply local smoothing/filtering before transmission? The PC backend can handle either approach.
**Answer:** Send raw analogRead() values as-is (12-bit ADC 0-4095), no local filtering/smoothing.

**Q4:** For LED updates, should we update the LEDs immediately upon receiving PC responses (MuteState, OutputDevice), or implement any visual feedback for connection status (e.g., blink on startup, flash on timeout)?
**Answer:** Update LEDs immediately upon receiving PC responses. Skip visual feedback for connection status (no blinking) for now.

**Q5:** For button handling, should we reuse the existing debouncing logic from lib/input_components/MuteButton and AudioDeviceSelector classes, or implement simplified debouncing in the main loop?
**Answer:** Reuse existing button handling logic from lib/input_components folder if it's better quality.

**Q6:** For error handling, if the ESP32 doesn't receive a PC response within a timeout period (e.g., 100ms), should we retry, log to Serial, or just continue with the next loop iteration? Should we provide any visual indication (LED blink) for communication errors?
**Answer:** No visual error indication for now - just silently continue on timeout.

**Q7:** For main loop timing, what delay should we use between sending slider updates? The current WiFi firmware uses 150ms. Serial is faster, so we could go lower (e.g., 50ms), or keep it at 150ms to reduce PC processing load?
**Answer:** Use 50ms loop delay (faster than current 150ms) since serial is faster than network.

**Q8:** For startup behavior, should the ESP32 wait for a handshake from the PC before starting to send data, or immediately start sending after Serial.begin()? The PC backend already handles reconnection gracefully.
**Answer:** Option B - ESP32 immediately starts sending data after Serial.begin(). No handshake waiting. PC backend already handles reconnection gracefully.

### Existing Code to Reference

**Similar Features Identified:**
- Feature: Input Components Library - Path: `lib/input_components/`
  - Classes: Slider, MuteButton, AudioDeviceSelector
  - These classes provide button debouncing logic and input handling patterns

- Feature: Current WiFi-Based Firmware - Path: `firmware/esp32-5-sliders-3-buttons/`
  - Main loop structure to maintain
  - Existing hardware pin definitions and initialization

**Components to Potentially Reuse:**
- `lib/input_components/Slider` class for analog slider handling
- `lib/input_components/MuteButton` class for button debouncing and LED control
- `lib/input_components/AudioDeviceSelector` class for device toggle button and LED management

**Backend Logic to Reference:**
- Replace `VolumeControllerApi` (UDP) with new `SerialApi` class
- Replace `BackendStateApi` (TCP) with same `SerialApi` class (unified communication)
- Keep similar main loop structure from current firmware

### Follow-up Questions
None required - all requirements are clear and comprehensive.

## Visual Assets

### Files Provided:
No visual files found.

### Visual Insights:
No visual assets provided.

## Requirements Summary

### Functional Requirements
- Remove all WiFi, UDP, and TCP communication code from existing firmware
- Implement USB Serial communication at 115200 baud using Arduino's built-in Serial API
- Send hardware input data to PC in pipe-delimited format with newline termination:
  - Slider values: `Sliders|val1|val2|val3|val4|val5\n` (raw 12-bit ADC values 0-4095)
  - Mute button states: `MuteButtons|bool1|bool2\n`
  - Output device switch requests: `SwitchOutput|index\n`
- Receive and parse PC responses:
  - Acknowledgment: `OK\n`
  - Mute states for LED feedback: `MuteState|bool1|bool2\n`
  - Active output device for LED feedback: `OutputDevice|index\n`
- Update hardware LEDs immediately upon receiving PC state updates:
  - 2 mute button LEDs (reflect actual PC mute state)
  - 2 output device LEDs (reflect active PC output device)
- Maintain 50ms main loop delay for faster responsiveness than network-based approach
- Start sending data immediately after `Serial.begin()` without waiting for handshake
- Silently continue operation on communication timeout (no error LED indication)

### Reusability Opportunities
- Reuse existing `lib/input_components/` classes for input handling:
  - `Slider` class for analog input reading
  - `MuteButton` class for button debouncing and LED control
  - `AudioDeviceSelector` class for device toggle button and LED management
- Maintain similar main loop structure from `firmware/esp32-5-sliders-3-buttons/`
- Replace `VolumeControllerApi` (UDP) and `BackendStateApi` (TCP) with single unified `SerialApi` class
- Keep existing hardware pin definitions and initialization patterns

### Scope Boundaries

**In Scope:**
- Complete replacement of WiFi/UDP/TCP communication with USB Serial
- Sending all hardware input states to PC (sliders, buttons)
- Receiving and processing PC responses for LED feedback
- Immediate LED updates based on actual PC state
- 50ms loop timing for responsive control
- Reuse of existing input component classes
- No-handshake startup (immediate data transmission)

**Out of Scope:**
- Visual feedback for connection status (LED blinking on startup/errors)
- Visual error indication on communication timeout
- Local ADC filtering or smoothing (raw values sent to PC)
- Interactive handshake protocol with PC on startup
- Retry logic on communication timeout
- Serial logging of errors or debug information

### Technical Considerations
- **Platform:** ESP32 microcontroller using Arduino framework
- **Serial Library:** Arduino's built-in Serial API (Serial.begin, Serial.println, Serial.readStringUntil)
- **Baud Rate:** 115200 (matching PC backend configuration)
- **ADC Resolution:** 12-bit (0-4095) raw values from analogRead()
- **Protocol Format:** Pipe-delimited text with newline termination
- **Loop Timing:** 50ms delay between iterations
- **PC Backend Status:** Already implemented and tested, ready to receive this protocol
- **Existing Firmware Location:** `firmware/esp32-5-sliders-3-buttons/` (WiFi/UDP/TCP version to be replaced)
- **Component Library:** `lib/input_components/` (classes to be reused)
- **Communication Pattern:** Unidirectional sending (ESP32 → PC) with optional PC responses for state synchronization (PC → ESP32)
- **Startup Behavior:** Immediate transmission after Serial initialization, no blocking wait for PC
