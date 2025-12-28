# Product Roadmap

## Current State
This is a fork of the original deej project that added UDP/TCP network support, mute button functionality, and output device switching. The codebase now uses direct USB serial communication, eliminating the unnecessary network complexity while the hardware is already connected via USB for power.

## Development Roadmap

1. [x] Serial Communication Migration - Replace UDP/TCP network protocol with direct USB serial communication at 115200 baud, including COM port auto-detection and removal of all network-related code paths from the Go backend and ESP32 firmware `L`

2. [x] Bidirectional Serial Protocol - Implement PC-to-ESP32 communication to send mute button states and active output device index back to the hardware for LED feedback display `M`

3. [ ] Hardware LED Indicators - Update ESP32 firmware to receive state data from PC and control LEDs that indicate which mute buttons are active and which output device is currently selected `S`

4. [x] Serial Connection Resilience - Add automatic reconnection logic when USB device disconnects/reconnects, including proper COM port re-enumeration and graceful degradation when hardware is unplugged `M`

5. [ ] Configuration Validation - Implement comprehensive YAML config validation with clear error messages for invalid slider mappings, device names, and button configurations to prevent runtime failures `S`

6. [ ] Startup Auto-Detection - Enhance COM port detection to handle multiple serial devices, filtering for deej-specific device identification (vendor ID, initial handshake, or device descriptor) `M`

7. [ ] Session Recovery - Add logic to restore volume levels and mute states after PC sleep/wake cycles or audio device reconnection events `M`

8. [ ] Slider Calibration - Create a calibration mode accessible via config or button combination that allows users to set min/max values for analog sliders to compensate for hardware variance `S`

> Notes
> - Order items by technical dependencies and product architecture
> - Each item should represent an end-to-end (frontend + backend) functional and testable feature
