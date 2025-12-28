# Raw Feature Idea

## Feature Description

Convert the current network-based (TCP/UDP) communication system to USB serial communication. The ESP32 is connected to the PC via USB cable all the time, so there's no need for WiFi networking. The system needs bidirectional communication for LED feedback (so the ESP32 knows the actual mute/device state from the PC).

## Key Requirements

- USB serial connection (115200 baud)
- Auto-detect COM port
- Replace all TCP/UDP code with serial
- Bidirectional protocol for LED state feedback
- Support: 5 sliders, 2 mute buttons, 1 device switch button
- Error handling with LED blink feedback

## Context

This is a fork of the original deej project. The fork added network support, but now wants to simplify back to serial like the original, while keeping the enhanced features (mute buttons, device switching) that require bidirectional communication.

## Original User Request

The user wants to migrate the deej project from TCP/UDP network communication to USB serial communication.

The ESP32 is connected to the PC via USB cable all the time, so there's no need for WiFi networking. The system needs bidirectional communication for LED feedback (so the ESP32 knows the actual mute/device state from the PC).
