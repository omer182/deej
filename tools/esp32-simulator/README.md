# ESP32 Deej Simulator

Test your deej serial communication without physical ESP32 hardware!

## 🎯 Purpose

This simulator allows you to test the serial communication between deej and your ESP32 controller without needing the actual hardware. It's perfect for:

- Testing the PC backend implementation
- Debugging serial protocol
- Verifying slider/button behavior
- Development without hardware

## 📦 What's Included

### 1. **serial_bridge.py** - Python Serial Simulator
A command-line tool that creates a virtual ESP32 on a COM port.

### 2. **simulator.html** - Web UI (Optional)
A beautiful web interface showing what the simulator is doing.

## 🚀 Quick Start

### Prerequisites

1. **Python 3.x** installed
2. **PySerial** library:
   ```bash
   pip install pyserial
   ```

3. **Virtual Serial Port** (for Windows):
   - Download **com0com**: https://sourceforge.net/projects/com0com/
   - Install and create a pair: COM10 ↔ COM11
   - Or use any two free COM ports

### Step 1: Configure Your Virtual Ports

If using com0com:
1. Install com0com
2. Run "Setup Command Prompt" as Administrator
3. Create a port pair:
   ```
   install PortName=COM10 PortName=COM11
   ```
4. COM10 will be for the simulator, COM11 for deej (or vice versa)

### Step 2: Update deej config.yaml

Edit your `config.yaml`:

```yaml
serial_connection_info:
  com_port: "COM11"  # The OTHER port in the pair
  # or use "auto" to let deej find it
  baud_rate: 115200
```

### Step 3: Run the Simulator

```bash
cd tools/esp32-simulator
python serial_bridge.py --port COM10
```

You should see:
```
╔════════════════════════════════════════════════════════════╗
║         ESP32 Deej Simulator - Serial Bridge              ║
╚════════════════════════════════════════════════════════════╝

✓ Connected to COM10 at 115200 baud
  Deej should auto-detect this port!

============================================================
ESP32 Simulator Running
============================================================

Interactive Commands:
  s [0-4] [value]  - Set slider value (0-4095)
  m [0-1]          - Toggle mute button
  d                - Toggle output device
  a                - Send all slider values
  r                - Random slider movement simulation
  q                - Quit

Waiting for deej to connect...
```

### Step 4: Start Deej

In another terminal:
```bash
cd c:\Users\rio18\Documents\Projects\deej
go build
.\deej.exe
```

Deej should auto-detect COM11 and connect!

### Step 5: Test the Communication

In the simulator terminal, try these commands:

**Test Sliders:**
```
> s 0 4095    # Set slider 0 to maximum (4095)
> s 1 0       # Set slider 1 to minimum (0)
> s 2 2048    # Set slider 2 to middle
> a           # Send all current slider values
```

**Test Mute Buttons:**
```
> m 0         # Toggle mute button 0
> m 1         # Toggle mute button 1
```

**Test Device Switching:**
```
> d           # Switch output device
```

**Random Movement:**
```
> r           # Simulate 10 random slider movements
```

## 📊 Example Session

```
> s 0 4095
→ Sliders|4095|2048|2048|2048|2048
← OK

> m 0
→ MuteButtons|true|false
← MuteState|true|false
  LED states: Mute0=True, Mute1=False

> d
→ SwitchOutput|1
← OutputDevice|1
  Current device: 1

> r
Simulating random slider movement...
→ Sliders|3251|2048|2048|2048|2048
← OK
→ Sliders|3251|891|2048|2048|2048
← OK
... (8 more movements)
Simulation complete
```

## 🎨 Web UI (Optional)

Open `simulator.html` in your browser to see a visual representation of the simulator.

**Note:** The web UI is currently for visualization only. The actual serial communication happens through the Python script.

## 🔧 Troubleshooting

### "Access Denied" or "Port in use"
- Close any other programs using the COM port
- Make sure the port numbers are correct
- Try running as Administrator

### "Could not auto-detect"
- Make sure the simulator is running FIRST
- Then start deej
- Check that deej is looking at the right port range (COM3-COM16)
- Try specifying the exact port in config.yaml instead of "auto"

### No response from deej
- Check that both programs are using the correct port pair
- Verify baud rate matches (115200)
- Look at deej's logs for connection messages

### Sliders not affecting audio
- Make sure your `slider_mapping` in config.yaml is correct
- Check that the target applications are running
- Verify Windows audio sessions are active

## 📝 Protocol Reference

The simulator implements the deej serial protocol:

**ESP32 → PC:**
- `Sliders|val0|val1|val2|val3|val4\n` - Slider values (0-4095)
- `MuteButtons|bool0|bool1\n` - Mute states (true/false)
- `SwitchOutput|index\n` - Switch to device index
- `GetCurrentOutputDevice\n` - Query current device

**PC → ESP32:**
- `OK\n` - Slider acknowledgment
- `MuteState|bool0|bool1\n` - Actual mute states
- `OutputDevice|index\n` - Current device index
- `ERROR\n` - Error occurred

## 🎯 What to Test

1. **Slider Movement**
   - Move each slider and verify audio changes
   - Test minimum (0), maximum (4095), and middle values
   - Verify noise reduction works

2. **Mute Buttons**
   - Toggle mute and verify LED feedback
   - Check actual mute state vs requested state
   - Test both buttons independently

3. **Device Switching**
   - Switch between output devices
   - Verify LED indicates current device
   - Check audio actually switches

4. **Error Handling**
   - Disconnect simulator during operation
   - Verify deej attempts reconnection
   - Check toast notifications appear

5. **Config Reload**
   - Edit config.yaml while running
   - Verify connection updates
   - Test slider value reset

## 🚀 Advanced Usage

### Custom Port and Baud Rate

```bash
python serial_bridge.py --port COM20 --baud 9600
```

### Automated Testing Script

Create a test script:
```python
import time
from serial_bridge import ESP32Simulator

sim = ESP32Simulator('COM10', 115200)
if sim.connect():
    # Test sequence
    for i in range(5):
        sim.slider_values[i] = 4095
        sim.send_sliders()
        time.sleep(1)
    sim.disconnect()
```

## 📚 Next Steps

Once you've verified everything works with the simulator:

1. Update your ESP32 firmware to use Serial instead of WiFi
2. Flash the firmware to your ESP32
3. Connect ESP32 to PC via USB
4. Update config.yaml to use auto-detection or the ESP32's COM port
5. Enjoy your physical hardware!

## 💡 Tips

- Use `auto` for com_port during development for easier testing
- Keep the simulator running to test reconnection logic
- Watch the deej logs to see what it's receiving
- Use verbose mode in deej for detailed logging: set DEEJ_DEBUG=1

---

**Happy Testing!** 🎉
