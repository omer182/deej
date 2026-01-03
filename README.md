## About This Fork

This is a fork of the original [deej project by omriharel](https://github.com/omriharel/deej), with significant enhancements and modernizations. Special thanks to **[tomerhh](https://github.com/tomerhh/deej)** for the initial UDP/network-based fork that inspired many of the improvements in this version.

### Key Changes in This Fork:

1. **USB Serial Communication** - Migrated from UDP to USB Serial for improved reliability, lower latency, and simpler setup
2. **Hardware Mute Buttons** - True audio muting (not just volume to 0%) with LED feedback
3. **Output Device Switching** - Toggle between audio devices (speakers/headphones) with a button
4. **Individual Button Events** - More efficient protocol with per-button messages instead of array updates
5. **Optimistic UI Updates** - Immediate LED feedback on ESP32 before PC confirmation
6. **Connection Status Indicator** - Blinking LED shows when waiting for PC connection
7. **Auto-Mute on Zero** - Automatically mute when slider reaches bottom position

**Acknowledgments:**
- Original deej concept and implementation: [omriharel](https://github.com/omriharel)
- UDP/network fork foundation: [tomerhh](https://github.com/tomerhh/deej)
- USB Serial migration and enhancements: This fork

# deej (with USB Serial support)

![deej](assets/deejudp-logo.png)

deej is an **open-source hardware volume mixer** for Windows and Linux PCs. It lets you use real-life sliders (like a DJ!) to **seamlessly control the volumes of different apps** (such as your music player, the game you're playing and your voice chat session) without having to stop what you're doing.

**This fork uses USB Serial communication** for direct connection between your ESP32-based hardware controller and your PC. This provides lower latency and more reliable communication compared to UDP/Wi-Fi solutions.

For thorough documentation on the basics, please check out [the README of the original project](https://github.com/omriharel/deej).

**[Download the latest release](https://github.com/omer182/deej/releases)**

## Configuration
In `config.yaml` edit the following properties:

### Serial Port

```yaml
# settings for the serial connection
com_port: COM5  # Adjust to match your ESP32's COM port
baud_rate: 115200
```

### Sliders
an index based list of volume targets that will be controlled from the deej board.
See notes below on target names.

```yaml
slider_mapping:
  0: "Headphones (HyperX Cloud Flight Wireless Headset)"
  1: "Speakers (Realtek(R) Audio)"
  2: chrome.exe
  3: discord.exe
```

### Mute buttons
an index based list of targets that will be muted from the deej board.
See notes below on target names.

```yaml
mute_button_mapping:
  0: master
  1: mic
```

### Output device toggeling
an index based list of device names that will be available to choose from the deej board.
See notes below on target names.

```yaml
available_output_device:
  0: "Speakers (Realtek(R) Audio)"
  1: "Headphones (HyperX Cloud III Wireless)"
```

### Notes on target names
To get device names on windows, write this in a PowerShell terminal (be sure to select an output device):
```powershell
 Get-CimInstance Win32_PnPEntity | ? { $_.PNPClass -eq "AudioEndpoint" } | Select-Object -Property PNPDeviceID, Name | ForEach-Object { Write-Host "$($_.Name)" }
```


Additionally:
* process names are **not** case sensitive
* you can use 'master' to indicate the master channel (i.e. the currently selected channel in the mixer)
* you can indicate a list of process names to create a group and control them together
* you can use 'mic' to control your mic input level (uses the default recording device)
* you can use 'deej.unmapped' to control all apps that aren't bound to any slider (this ignores master, system, mic and device-targeting sessions)
* windows only - you can use 'deej.current' to control the currently active app (whether full-screen or not)
* windows only - you can use a device's full name, i.e. "Speakers (Realtek High Definition Audio)", to bind it. this works for both output and input devices
* windows only - you can use 'system' to control the "system sounds" volume
* important: slider indexes start at 0, regardless of which analog pins you're using!


## The Serial Communication Protocol

The deej serial protocol is simple and efficient: each message consists of a command name, followed by values separated by a `|` pipe character, terminated with a newline (`\n`). The backend responds with `OK\n` for successful operations.

### Available Commands

#### Sliders
Sends all slider values in a single message. Values range from 0-4095 (12-bit ADC).

**Format:** `Sliders|<value0>|<value1>|...|<valueN>\n`

**Example:** If you have 5 sliders with varying positions:
```text
Sliders|0|2048|4095|1024|0
```
The backend does not send a response for slider updates (fire-and-forget for performance).

#### MuteButton (Individual)
Sends a single mute button event. The backend responds with `OK\n` on success.

**Format:** `MuteButton|<button_index>|<state>\n`
- `button_index`: 0-based index of the mute button
- `state`: `1` for muted, `0` for unmuted

**Example:** To mute button 0:
```text
MuteButton|0|1
```
**Response:** `OK\n`

#### SwitchOutput
Switches the active output device. The backend responds with `OK\n` on success.

**Format:** `SwitchOutput|<device_index>\n`
- `device_index`: 0-based index from `available_output_device` in config

**Example:** To switch to device 1:
```text
SwitchOutput|1
```
**Response:** `OK\n`

### Protocol Benefits
- **Individual events**: Only changed buttons send data (reduces serial traffic)
- **Acknowledgment**: `OK` responses ensure critical operations succeeded
- **Optimistic UI**: ESP32 updates LEDs immediately, backend confirms asynchronously
- **Low latency**: USB Serial provides <10ms round-trip time

### Building the Controller

This fork includes complete ESP32 firmware with support for:
- 5 analog sliders (12-bit ADC resolution)
- 2 mute buttons with LED feedback
- 1 output device selector with dual-LED indication
- Auto-mute based on slider position (threshold: 400/4095)
- Optimistic UI updates (LEDs respond instantly to button presses)

See the firmware implementation [here](https://github.com/omer182/deej/tree/serial-migration/firmware/esp32-serial-first)

**Hardware Requirements:**
- ESP32 development board
- 5x 10kΩ linear potentiometers
- 3x momentary push buttons
- 4x LEDs with appropriate resistors (220Ω recommended)
- USB cable for serial communication

**Flashing the Firmware:**
1. Install [PlatformIO](https://platformio.org/)
2. Open the `firmware/esp32-serial-first` folder in PlatformIO
3. Connect your ESP32 via USB
4. Run `pio run --target upload`
5. The device will appear as a COM port (check Device Manager on Windows)
## License

deej is released under the [MIT license](./LICENSE).
