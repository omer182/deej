#!/usr/bin/env python3
"""
Simple serial communication test for deej.
This script simulates an ESP32 controller sending data to deej over serial.
"""

import serial
import time
import sys

def test_serial_communication(port='COM4', baud=115200):
    """
    Test serial communication with deej by sending mock ESP32 data.

    Args:
        port: COM port to use (e.g., 'COM4')
        baud: Baud rate (default: 115200)
    """
    print(f"Opening serial port {port} at {baud} baud...")

    try:
        ser = serial.Serial(port, baud, timeout=1)
        time.sleep(2)  # Wait for serial connection to stabilize
        print(f"Connected to {port}")

        # Test 1: Send slider data
        print("\n=== Test 1: Sending slider data ===")
        slider_values = [2048, 4095, 0, 2048, 1024]  # 5 sliders with various values
        slider_msg = f"Sliders|{slider_values[0]}|{slider_values[1]}|{slider_values[2]}|{slider_values[3]}|{slider_values[4]}\n"
        print(f"Sending: {slider_msg.strip()}")
        ser.write(slider_msg.encode())

        # Wait for response
        time.sleep(0.5)
        if ser.in_waiting:
            response = ser.readline().decode().strip()
            print(f"Received: {response}")
            if response == "OK":
                print("✓ Slider data acknowledged!")
            else:
                print(f"✗ Unexpected response: {response}")
        else:
            print("✗ No response received")

        # Test 2: Request mute state
        print("\n=== Test 2: Requesting mute button state ===")
        mute_msg = "MuteButtons|true|false\n"
        print(f"Sending: {mute_msg.strip()}")
        ser.write(mute_msg.encode())

        time.sleep(0.5)
        if ser.in_waiting:
            response = ser.readline().decode().strip()
            print(f"Received: {response}")
            if response.startswith("MuteState|"):
                print("✓ Mute state response received!")
            else:
                print(f"✗ Unexpected response: {response}")
        else:
            print("✗ No response received")

        # Test 3: Switch output device
        print("\n=== Test 3: Switching output device ===")
        switch_msg = "SwitchOutput|1\n"
        print(f"Sending: {switch_msg.strip()}")
        ser.write(switch_msg.encode())

        time.sleep(0.5)
        if ser.in_waiting:
            response = ser.readline().decode().strip()
            print(f"Received: {response}")
            if response.startswith("OutputDevice|"):
                print("✓ Output device response received!")
            else:
                print(f"✗ Unexpected response: {response}")
        else:
            print("✗ No response received")

        # Test 4: Send continuous slider updates
        print("\n=== Test 4: Sending continuous slider updates ===")
        print("Sending 10 slider updates with changing values...")
        for i in range(10):
            # Simulate slider movement
            val = int(2048 + 1000 * (i / 10))
            slider_msg = f"Sliders|{val}|{val}|{val}|{val}|{val}\n"
            print(f"  [{i+1}/10] Sending: {slider_msg.strip()}")
            ser.write(slider_msg.encode())

            time.sleep(0.3)
            if ser.in_waiting:
                response = ser.readline().decode().strip()
                print(f"         Response: {response}")

        print("\n=== Test complete! ===")
        ser.close()
        print(f"Closed connection to {port}")

    except serial.SerialException as e:
        print(f"Error: Could not open serial port {port}")
        print(f"Details: {e}")
        print(f"\nMake sure:")
        print(f"  1. The port {port} exists")
        print(f"  2. No other program is using {port}")
        print(f"  3. You have permission to access {port}")
        sys.exit(1)
    except KeyboardInterrupt:
        print("\n\nTest interrupted by user")
        ser.close()
        sys.exit(0)

if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser(description='Test serial communication with deej')
    parser.add_argument('--port', default='COM4', help='Serial port to use (default: COM4)')
    parser.add_argument('--baud', type=int, default=115200, help='Baud rate (default: 115200)')

    args = parser.parse_args()

    print("=" * 60)
    print("deej Serial Communication Test")
    print("=" * 60)

    test_serial_communication(args.port, args.baud)
