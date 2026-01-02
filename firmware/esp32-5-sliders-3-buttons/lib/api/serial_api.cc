#include "serial_api.h"

#include <sstream>

namespace lib {
namespace api {

void SerialApi::sendSliders(const std::string& data) {
  Serial.println(data.c_str());

  // Optionally read "OK\n" response but timeout silently if not received
  readResponse();
}

std::vector<bool> SerialApi::sendMuteButtons(const std::vector<bool>& states) {
  // Build message format: "MuteButtons|bool1|bool2\n"
  std::string message = "MuteButtons";
  for (const auto& state : states) {
    message += "|";
    message += (state ? "1" : "0");
  }
  Serial.println(message.c_str());

  // Parse response: "MuteState|bool1|bool2\n"
  std::string response = readResponse();
  std::vector<std::string> parts = parseResponse(response);

  std::vector<bool> result;
  if (!parts.empty() && parts[0] == "MuteState") {
    for (size_t i = 1; i < parts.size(); i++) {
      result.push_back(parseBool(parts[i]));
    }
  }

  // Return empty vector on error (caller should handle gracefully)
  return result;
}

int SerialApi::sendSwitchOutput(int device_index) {
  // Build message format: "SwitchOutput|index\n"
  std::string message = "SwitchOutput|" + std::to_string(device_index);
  Serial.println(message.c_str());

  // Parse response: "OutputDevice|index\n"
  std::string response = readResponse();
  std::vector<std::string> parts = parseResponse(response);

  if (parts.size() >= 2 && parts[0] == "OutputDevice") {
    return parseInt(parts[1]);
  }

  // Return -1 on error (caller should handle gracefully)
  return -1;
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
