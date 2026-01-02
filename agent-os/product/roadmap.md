# Product Roadmap

This roadmap focuses on **rewriting the ESP32 firmware only** with a serial-first architecture. The Go backend at `pkg/deej/` already implements all required functionality (serial communication, mute buttons, device switching, Windows Audio API integration) and will be reused as-is from the existing codebase.

## Firmware Rewrite Phases

### Phase 1: Core Serial Communication & Sliders

1. [x] ESP32 Serial Protocol Foundation - Create new firmware project with clean serial-first design: setup Serial at 115200 baud in setup(), implement 50ms main loop with non-blocking Serial.available() checks, send slider values in format "Sliders|val0|val1|val2|val3|val4\n" using 12-bit ADC (0-4095), implement zero-deadband (~100) to prevent jitter at bottom, read backend "OK\n" responses (non-blocking, ignore timeouts) `M`

2. [x] Slider Hardware Integration - Wire 5 analog sliders to ESP32 ADC GPIO pins (GPIO 32-36), implement simple change detection (send only when value != previous_value), apply value inversion (4095 - rawValue) so slider-top = high volume, verify smooth volume control with backend volume normalization (existing backend handles 0-4095 → 0.0-1.0) `S`

### Phase 2: Button Components

3. [x] Master Mute Button - Create MuteButton class with GPIO input (pull-up) + LED output, implement 40ms debounce logic, support multi-session mode (session 0 = speakers, session 1 = headphones), send "MuteButtons|bool0|bool1\n" on button press (toggle local state, send to backend, wait for confirmation), parse backend response to update LED state (format: "MuteState|bool0|bool1\n"), only update LED after backend confirms actual mute state `M`

4. [x] Mic Mute Button - Add second MuteButton instance for microphone, simpler single-session implementation, reuse existing MuteButton class architecture, backend already handles mic mute via existing serial protocol handlers `S`

5. [x] Output Device Switcher - Create AudioDeviceSelector class with button GPIO + 2 LEDs (one per device), implement debouncing with long-press detection (2 seconds = ESP32 reset via esp_restart()), send "SwitchOutput|deviceIndex\n" on button press, parse backend response (format: "OutputDevice|index\n"), update LEDs so only active device LED is on, switch master mute button's active session when device changes `M`

### Phase 3: Auto-Mute Integration

6. [x] Auto-Mute on Slider Zero - In main loop, detect when sliders 0-1 hit 0 (after zero-deadband applied), trigger corresponding mute button state change (set mute=true for that session), send mute button update to backend, LED turns on automatically via backend response, when slider moves above 0, trigger unmute (mute=false), backend already implements Windows Core Audio mute (not just 0% volume) `M`

7. [x] Session-Aware LED Feedback - Ensure slider zero-detection updates the correct session's mute state (slider 0 → master mute session 0, slider 1 → master mute session 1), master mute button LED always shows state of currently active session (speakers or headphones), when switching devices, LED updates to show new session's mute state, all LED updates driven by backend confirmation to prevent desync `S`

### Phase 4: Testing & Polish

8. [ ] Hardware Testing & Validation - Test all 5 sliders with full 0-4095 range (verify inversion, zero-deadband, smooth movement), test both mute buttons (verify toggle, LED feedback, backend state sync), test device switcher (verify device toggle, LED switching, long-press reset), test auto-mute (slider to zero → LED on + Windows shows muted, slider up → LED off + unmuted), verify USB reconnection handling `M`

## Backend Status (No Changes Required)

The existing Go backend already provides:
- ✅ Serial port auto-detection and connection handling
- ✅ Message parsing for Sliders, MuteButtons, SwitchOutput commands
- ✅ Windows Core Audio session control and muting
- ✅ Audio device switching via `SetAudioDeviceByID()`
- ✅ YAML configuration for slider/device mappings
- ✅ Config hot-reload with fsnotify
- ✅ Bidirectional serial responses (OK, MuteState, OutputDevice)

**Reference**: Use existing firmware at `firmware/esp32-5-sliders-3-buttons/` as reference for GPIO pins, hardware architecture, and class structure patterns. The rewrite focuses on simplifying the architecture by removing TCP/UDP abstractions while keeping proven patterns.

> Notes
> - Backend handlers already tested and working - firmware rewrite only
> - Firmware should be serial-first with no network protocol remnants
> - Reuse proven class structures (Slider, MuteButton, AudioDeviceSelector) with cleaner implementation
> - All LED state updates must wait for backend confirmation to prevent desync
