# Product Mission

## Pitch
deej is an **open-source hardware volume mixer** that helps content creators, gamers, and streamers gain precise, wphysical control over PC audio by providing real-time volume adjustment for individual applications and audio devices through tactile sliders and buttons.

## Users

### Primary Customers
- **Content Creators & Streamers**: Individuals who need quick, precise audio control during live streaming or recording sessions
- **Gamers**: PC gamers who want to balance game audio, voice chat, and music without alt-tabbing
- **DIY PC Enthusiasts**: Tech-savvy users who enjoy building custom hardware solutions for their setups

### User Personas

**Solo Streamer/Gamer** (20-35)
- **Role:** Content creator managing personal gaming/streaming setup
- **Context:** Uses multiple audio sources simultaneously (game, Discord, music, browser) while creating content or gaming
- **Pain Points:** Software volume mixers require switching windows, breaking immersion and workflow. Network-based solutions add unnecessary complexity when the controller is physically connected via USB.
- **Goals:** Instant, tactile audio control without leaving the current application. Simple, reliable USB connection without network configuration overhead.

## The Problem

### Audio Control Interrupts Workflow
Adjusting application volumes on Windows requires opening the volume mixer, finding the right app, and using a mouse - an interruption that breaks immersion during gaming or focus during content creation. For users running multiple audio sources, this becomes a constant distraction.

**Our Solution:** Physical sliders and buttons connected via USB provide immediate, tactile control over individual application volumes, mute states, and output device switching without ever leaving your current task.

## Differentiators

### Direct USB Serial Communication
Unlike the original deej (serial-only) or other forks (network-based), this fork migrates to **USB serial with bidirectional communication**. This provides LED feedback for mute states and device selection while maintaining the simplicity of a direct USB connection - eliminating network configuration complexity.

This results in a more reliable connection, visual feedback on hardware state, and simplified setup for users who already have their ESP32 connected via USB for power.

### True Mute Functionality
Unlike solutions that simply set volume to 0%, deej implements **Windows Core Audio muting** - the same behavior as the system mute button. This prevents audio from playing entirely and provides proper visual feedback in Windows audio controls.

### Physical Output Device Switching
Instead of navigating Windows sound settings, a single button press toggles between configured audio devices (headphones, speakers, etc.). This is especially valuable for streamers who switch between monitoring devices or gamers alternating between headset and speakers.

## Key Features

### Core Features
- **5 Analog Sliders:** Independent volume control for any Windows audio session - control individual apps (Discord, Chrome, games) or system targets (master, mic, system sounds, currently active app)
- **USB Serial Communication:** Direct USB connection at 115200 baud with auto-detection of COM port, providing reliable communication without network configuration
- **Bidirectional LED Feedback:** Real-time visual indication of mute states and active output device through hardware LEDs

### Control Features
- **2 Mute Buttons:** True Windows Core Audio muting (not volume=0) for quick audio cutoff with proper system integration
- **Output Device Toggle:** One-button switching between configured audio output devices (headphones, speakers, etc.)
- **Flexible Mapping:** YAML-based configuration maps sliders/buttons to specific apps, audio devices, or system targets

### Advanced Features
- **Smart Session Targeting:** Control unmapped apps collectively, target currently active window audio, or bind to specific hardware devices by name
- **ESP32-Based Controller:** Hardware built on affordable ESP32/ESP8266 microcontrollers programmed via Arduino framework
- **Windows Core Audio Integration:** Deep integration with Windows WCA API for professional-grade audio control and session management
