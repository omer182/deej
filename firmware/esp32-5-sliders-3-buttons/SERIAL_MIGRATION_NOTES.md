# ESP32 Firmware Serial Migration - Implementation Notes

## Overview
The ESP32 firmware has been successfully migrated from WiFi/UDP/TCP network communication to USB Serial communication at 115200 baud.

## Changes Made

### Removed Files
- `lib/api/volume_controller_api.h` (UDP-based)
- `lib/api/volume_controller_api.cc` (UDP-based)
- `lib/api/backend_state_api.h` (TCP-based)
- `lib/api/backend_state_api.cc` (TCP-based)
- `lib/api/backend_state.h` (network state structure)

### New Files
- `lib/api/serial_api.h` - Serial communication API header
- `lib/api/serial_api.cc` - Serial communication API implementation

### Modified Files
- `src/main.cpp` - Removed WiFi setup, integrated SerialApi

## Serial Protocol

### Outgoing Messages (ESP32 → PC)
- Slider values: `Sliders|val1|val2|val3|val4|val5\n`
  - Values are raw 12-bit ADC readings (0-4095)
- Mute buttons: `MuteButtons|bool1|bool2\n`
  - Boolean values sent as "1" or "0"
- Output device switch: `SwitchOutput|index\n`
  - Index is the device number

### Incoming Messages (PC → ESP32)
- Acknowledgment: `OK\n`
- Mute state: `MuteState|bool1|bool2\n`
- Output device: `OutputDevice|index\n`

## Key Implementation Details

### SerialApi Class
- **Timeout:** 100ms for PC responses
- **Error Handling:** Silently continues on timeout (no visual feedback)
- **Non-blocking:** Uses `Serial.available()` check before reading
- **Response Parsing:** Pipe-delimited string parsing with helper methods

### Main Loop
- **Timing:** 50ms delay (reduced from 150ms for faster responsiveness)
- **Startup:** Immediate data transmission after `Serial.begin(115200)`, no handshake required
- **LED Updates:** Based on actual PC state from responses

### Hardware Components (Unchanged)
- 5 analog sliders (GPIO 33, 35, 32, 36, 34)
- 2 mute buttons with LEDs (GPIO 4/12, 14/21)
- 1 audio device selector with 2 LEDs (GPIO 5/18/19)

## Binary Size
- RAM: 8.3% (27,356 bytes / 327,680 bytes)
- Flash: 39.8% (521,085 bytes / 1,310,720 bytes)

## Testing Recommendations

### Hardware Testing Required
1. Flash firmware to ESP32 (`pio run --target upload`)
2. Connect ESP32 via USB to PC
3. Start deej backend with correct COM port
4. Test all 5 sliders for smooth volume control
5. Test both mute buttons for correct LED feedback
6. Test output device switching
7. Test USB reconnection (unplug/replug)
8. Test startup order variations (ESP32 first vs. PC backend first)

## Known Limitations
- No visual error indication on communication timeout
- No retry logic on communication failure
- Relies on PC backend for state persistence
