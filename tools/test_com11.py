#!/usr/bin/env python3
"""Quick test to verify COM11 can be opened"""
import serial
import sys

try:
    print("Attempting to open COM11...")
    ser = serial.Serial('COM11', 115200, timeout=1)
    print(f"SUCCESS: COM11 opened successfully")
    print(f"Port: {ser.port}")
    print(f"Baud: {ser.baudrate}")
    ser.close()
    print("Port closed successfully")
    sys.exit(0)
except Exception as e:
    print(f"ERROR: Failed to open COM11")
    print(f"Error: {e}")
    sys.exit(1)
