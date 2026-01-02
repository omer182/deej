# Product Mission

## Pitch
deej is a **serial-first hardware volume mixer** that helps streamers, gamers, and power users achieve instant, tactile control over Windows audio by providing physical sliders and buttons connected via USB that control individual application volumes, mute states, and output device switching.

## Users

### Primary Customers
- **Streamers & Content Creators**: Need split-second audio control during live streams without interrupting their workflow
- **Gamers**: Want to balance game audio, Discord, and music without alt-tabbing or breaking immersion
- **Audio Enthusiasts**: Prefer tactile hardware controls over software interfaces for precision audio management

### User Personas

**Live Streamer** (22-32)
- **Role:** Content creator managing complex audio setup while streaming
- **Context:** Simultaneously running game audio, Discord voice chat, music, browser alerts, and streaming software - each needs independent volume control
- **Pain Points:** Windows volume mixer requires multiple clicks and window switching. Adjusting Discord during intense gameplay breaks focus. Network-based solutions add unnecessary WiFi configuration when hardware is already USB-connected for power.
- **Goals:** Instant physical control over every audio source. Visual confirmation of mute states (critical for mic muting). One-button switching between headphones (for gaming) and speakers (for editing).

**Competitive Gamer** (18-28)
- **Role:** PC gamer optimizing for competitive advantage
- **Context:** Needs to quickly adjust voice comms volume relative to game sounds during matches without losing focus
- **Pain Points:** Opening volume mixer during gameplay causes deaths. Can't quickly mute mic between rounds. Switching from headset to speakers after gaming session requires navigating Windows settings.
- **Goals:** Zero-latency volume adjustments during gameplay. Physical mute button for instant mic control. Reliable USB connection that doesn't drop during critical moments.

## The Problem

### Audio Control Breaks Flow State
Windows volume mixer requires opening a separate interface, searching for the application, and making adjustments with a mouse - a workflow that completely breaks immersion during gaming or focus during content creation. For users managing 5+ simultaneous audio sources (game, Discord, Spotify, OBS, browser), this becomes a constant source of frustration.

**Our Solution:** Physical sliders and buttons connected via direct USB serial provide immediate hardware control over individual Windows audio sessions with zero software interaction required.

### Unreliable Feedback on Critical Mute States
Accidentally leaving a microphone unmuted during private conversations or forgetting to unmute before speaking are common pain points. Software indicators are easily missed when focused on other tasks.

**Our Solution:** Hardware LEDs directly linked to Windows Core Audio mute states provide always-visible confirmation of mic and output mute status, synchronized in real-time with the backend.

### Output Device Switching Requires Multiple Clicks
Streamers frequently switch between monitoring devices (headphones for gaming, speakers for editing). Windows requires opening Sound settings, navigating to Playback devices, and selecting the target device.

**Our Solution:** Single button press toggles between configured output devices with LED confirmation of the active device, all controlled through Windows Core Audio API.

## Differentiators

### Serial-First Architecture
Unlike the original deej (serial-only, no feedback) or network-based forks (WiFi/UDP complexity), this implementation is designed from the ground up for **USB serial communication at 115200 baud**. The hardware is already connected via USB for power - using WiFi adds unnecessary network configuration, latency, and failure modes. Serial-first design provides:
- Zero-configuration connectivity (auto-detected COM port)
- Sub-50ms latency for volume changes
- Bidirectional communication for LED state synchronization
- No network dependencies or firewall issues

This results in plug-and-play simplicity with professional-grade reliability.

### True Windows Core Audio Integration
Unlike volume-to-zero workarounds, deej implements **Windows Core Audio Session API muting** - the same API used by Windows itself. This provides:
- Proper mute state synchronization across all Windows audio interfaces
- Automatic mute when sliders reach zero (with backend confirmation)
- Visual feedback in Windows volume mixer showing "Muted" status
- Session-aware mute control (separate mute states for speakers vs headphones)

This results in predictable behavior that matches user expectations from Windows native controls.

### Hardware State Synchronization
Unlike one-way controllers that can desynchronize from actual system state, deej implements **bidirectional serial protocol** where:
- ESP32 sends input changes to PC
- PC backend sends actual state back to ESP32
- LEDs update only after backend confirmation
- State stays synchronized even if backend operations fail or timeout

This results in LEDs that always reflect true Windows audio state, preventing confusion from desynchronized hardware indicators.

## Key Features

### Core Features
- **5 Analog Sliders:** Smooth 12-bit ADC volume control (0-4095 range) for individual Windows audio sessions - map to specific applications (Discord, Spotify, games) or output devices (speakers, headphones)
- **USB Serial Communication:** Direct 115200 baud connection with auto-COM-port detection, providing reliable sub-50ms latency without network configuration
- **Auto-Mute on Zero:** When slider reaches 0, automatically triggers Windows Core Audio mute (not just 0% volume) with backend confirmation

### Control Features
- **Master Mute Button:** Mutes current active output device (speakers OR headphones depending on active session) with LED indicating mute state of active session only
- **Mic Mute Button:** Mutes microphone input device with dedicated LED indicator
- **Output Device Toggle:** One-button switching between configured audio devices (speakers/headphones) with dual LEDs showing active device - includes long-press ESP32 reset

### Advanced Features
- **Session-Aware Mute States:** Each output device (session 0 = speakers, session 1 = headphones) maintains independent mute state - switching devices shows correct LED for newly active device
- **Bidirectional State Sync:** PC backend sends actual mute states and active device index back to ESP32 for LED updates, ensuring hardware always reflects true Windows state
- **Flexible YAML Mapping:** Configure slider-to-application bindings, mute button targets, and output device names through simple YAML config with hot-reload support
