#!/usr/bin/env python3
"""
ESP32 Deej Serial Bridge
Creates a virtual COM port pair and bridges between the deej app and the web simulator.
"""

import serial
import time
import sys
from threading import Thread

class ESP32Simulator:
    def __init__(self, port='COM10', baud=115200):
        """
        Initialize the ESP32 simulator on a specific COM port.

        Args:
            port: COM port to simulate (e.g., 'COM10')
            baud: Baud rate (default: 115200)
        """
        self.port = port
        self.baud = baud
        self.ser = None
        self.running = False
        self.slider_values = [2048, 2048, 2048, 2048, 2048]
        self.mute_states = [False, False]
        self.device_index = 0

    def connect(self):
        """Connect to the virtual COM port"""
        try:
            self.ser = serial.Serial(self.port, self.baud, timeout=1)
            self.running = True
            print(f"✓ Connected to {self.port} at {self.baud} baud")
            print(f"  Deej should auto-detect this port!")
            print(f"\n{'='*60}")
            print(f"ESP32 Simulator Running")
            print(f"{'='*60}")
            return True
        except Exception as e:
            print(f"✗ Failed to connect to {self.port}: {e}")
            print(f"\nMake sure:")
            print(f"  1. You have pyserial installed: pip install pyserial")
            print(f"  2. {self.port} is not in use")
            print(f"  3. You have a virtual serial port tool (see README)")
            return False

    def disconnect(self):
        """Disconnect from COM port"""
        self.running = False
        if self.ser:
            self.ser.close()
            print(f"\n✓ Disconnected from {self.port}")

    def send_sliders(self):
        """Send current slider values"""
        message = f"Sliders|{self.slider_values[0]}|{self.slider_values[1]}|{self.slider_values[2]}|{self.slider_values[3]}|{self.slider_values[4]}\n"
        self.ser.write(message.encode())
        print(f"→ {message.strip()}")

    def send_mute_buttons(self):
        """Send current mute button states"""
        message = f"MuteButtons|{'true' if self.mute_states[0] else 'false'}|{'true' if self.mute_states[1] else 'false'}\n"
        self.ser.write(message.encode())
        print(f"→ {message.strip()}")

    def send_switch_output(self):
        """Send output device switch request"""
        message = f"SwitchOutput|{self.device_index}\n"
        self.ser.write(message.encode())
        print(f"→ {message.strip()}")

    def read_responses(self):
        """Read and process responses from deej"""
        while self.running:
            try:
                if self.ser.in_waiting > 0:
                    line = self.ser.readline().decode().strip()
                    if line:
                        print(f"← {line}")
                        self.process_response(line)
            except Exception as e:
                if self.running:
                    print(f"Error reading: {e}")
            time.sleep(0.01)

    def process_response(self, response):
        """Process response from deej"""
        parts = response.split('|')

        if parts[0] == "OK":
            pass  # Slider acknowledgment

        elif parts[0] == "MuteState":
            if len(parts) >= 3:
                self.mute_states[0] = parts[1] == 'true'
                self.mute_states[1] = parts[2] == 'true'
                print(f"  LED states: Mute0={self.mute_states[0]}, Mute1={self.mute_states[1]}")

        elif parts[0] == "OutputDevice":
            if len(parts) >= 2:
                self.device_index = int(parts[1])
                print(f"  Current device: {self.device_index}")

        elif parts[0] == "ERROR":
            print(f"  ⚠️  Error from deej!")

    def interactive_mode(self):
        """Run interactive command mode"""
        # Start response reader thread
        reader_thread = Thread(target=self.read_responses, daemon=True)
        reader_thread.start()

        print(f"\nInteractive Commands:")
        print(f"  s [0-4] [value]  - Set slider value (0-4095)")
        print(f"  m [0-1]          - Toggle mute button")
        print(f"  d                - Toggle output device")
        print(f"  a                - Send all slider values")
        print(f"  r                - Random slider movement simulation")
        print(f"  q                - Quit")
        print(f"\nWaiting for deej to connect...\n")

        # Initial slider send
        time.sleep(1)
        self.send_sliders()

        while self.running:
            try:
                cmd = input("> ").strip().lower()

                if cmd == 'q':
                    break

                elif cmd == 'a':
                    self.send_sliders()

                elif cmd.startswith('s '):
                    parts = cmd.split()
                    if len(parts) == 3:
                        idx = int(parts[1])
                        val = int(parts[2])
                        if 0 <= idx < 5 and 0 <= val <= 4095:
                            self.slider_values[idx] = val
                            self.send_sliders()
                        else:
                            print("Invalid range (slider: 0-4, value: 0-4095)")
                    else:
                        print("Usage: s [slider] [value]")

                elif cmd.startswith('m '):
                    parts = cmd.split()
                    if len(parts) == 2:
                        idx = int(parts[1])
                        if 0 <= idx < 2:
                            self.mute_states[idx] = not self.mute_states[idx]
                            self.send_mute_buttons()
                        else:
                            print("Invalid mute button (0-1)")
                    else:
                        print("Usage: m [button]")

                elif cmd == 'd':
                    self.device_index = (self.device_index + 1) % 2
                    self.send_switch_output()

                elif cmd == 'r':
                    print("Simulating random slider movement...")
                    import random
                    for _ in range(10):
                        idx = random.randint(0, 4)
                        val = random.randint(0, 4095)
                        self.slider_values[idx] = val
                        self.send_sliders()
                        time.sleep(0.2)
                    print("Simulation complete")

                else:
                    print("Unknown command. Type 'q' for quit.")

            except KeyboardInterrupt:
                print("\n^C detected")
                break
            except Exception as e:
                print(f"Error: {e}")

    def run(self):
        """Main run loop"""
        if self.connect():
            try:
                self.interactive_mode()
            finally:
                self.disconnect()


def main():
    """Main entry point"""
    import argparse

    parser = argparse.ArgumentParser(description='ESP32 Deej Serial Simulator')
    parser.add_argument('--port', default='COM10', help='COM port to use (default: COM10)')
    parser.add_argument('--baud', type=int, default=115200, help='Baud rate (default: 115200)')

    args = parser.parse_args()

    print(f"""
╔════════════════════════════════════════════════════════════╗
║         ESP32 Deej Simulator - Serial Bridge              ║
╚════════════════════════════════════════════════════════════╝

This tool simulates an ESP32 controller over a serial port.

Configuration:
  Port: {args.port}
  Baud: {args.baud}

Instructions:
  1. Make sure deej is configured with:
     serial_connection_info:
       com_port: "{args.port}"  (or "auto" to detect)
       baud_rate: {args.baud}

  2. Start deej.exe
  3. Use commands below to simulate hardware

""")

    simulator = ESP32Simulator(args.port, args.baud)
    simulator.run()


if __name__ == '__main__':
    main()
