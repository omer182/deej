#include "serial_api.h"

#include <sstream>

namespace lib {
namespace api {

void SerialApi::sendSliders(const std::string& data) {
  Serial.println(data.c_str());

  // Optionally read "OK\n" response but timeout silently if not received
  readResponse();
}

bool SerialApi::sendMuteButton(int button_index, bool state) {
  // Build message format: "MuteButton|index|state\n"
  std::string message = "MuteButton|" + std::to_string(button_index) + "|" +
                        (state ? "1" : "0");
  Serial.println(message.c_str());

  // Parse response: "OK\n"
  std::string response = readResponse();

  // Return true if we got "OK", false on timeout/error
  return (response == "OK");
}

bool SerialApi::sendSwitchOutput(int device_index) {
  // Build message format: "SwitchOutput|index\n"
  std::string message = "SwitchOutput|" + std::to_string(device_index);
  Serial.println(message.c_str());

  // Parse response: "OK\n"
  std::string response = readResponse();

  // Return true if we got "OK", false on timeout/error
  return (response == "OK");
}

std::string SerialApi::readResponse() {
  unsigned long start_time = millis();

  // Wait for data to be available or timeout
  while (!Serial.available()) {
    if (millis() - start_time > _timeout_ms) {
      return "";  // Timeout - return empty string
    }
    delay(1);  // Small delay to prevent busy-waiting
  }

  // Read the response
  String response = Serial.readStringUntil('\n');
  return std::string(response.c_str());
}

std::vector<std::string> SerialApi::parseResponse(const std::string& response) {
  std::vector<std::string> parts;
  std::stringstream ss(response);
  std::string part;

  while (std::getline(ss, part, '|')) {
    parts.push_back(part);
  }

  return parts;
}

bool SerialApi::parseBool(const std::string& str) {
  if (str.empty()) {
    return false;
  }
  // Handle both "1"/"0" and "true"/"false"
  return (str == "1" || str == "true" || str == "True");
}

int SerialApi::parseInt(const std::string& str) {
  if (str.empty()) {
    return -1;
  }
  try {
    return std::stoi(str);
  } catch (...) {
    return -1;
  }
}

}  // namespace api
}  // namespace lib
