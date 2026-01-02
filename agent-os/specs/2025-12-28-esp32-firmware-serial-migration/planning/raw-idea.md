# ESP32 Firmware Serial Migration - Raw Idea

## Feature Description

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

## Context

The PC backend is already complete and tested - it's ready to receive this serial protocol. The existing firmware is in firmware/esp32-5-sliders-3-buttons/ and uses WiFi/UDP/TCP which needs to be completely replaced.

## Date Initiated

2025-12-28
