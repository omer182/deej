#ifndef LIB_API_SERIAL_API_H
#define LIB_API_SERIAL_API_H

#include <Arduino.h>

#include <string>
#include <vector>

namespace lib {
namespace api {

class SerialApi {
 public:
  SerialApi() : _timeout_ms(100) {}

  // Send slider values and optionally wait for "OK\n" response
  void sendSliders(const std::string& data);

  // Send mute button states and return actual PC mute states
  std::vector<bool> sendMuteButtons(const std::vector<bool>& states);

  // Send output device switch request and return actual active device index
  int sendSwitchOutput(int device_index);

 private:
  const int _timeout_ms;

  // Helper to read response with timeout
  std::string readResponse();

  // Helper to parse pipe-delimited response
  std::vector<std::string> parseResponse(const std::string& response);

  // Helper to convert string to boolean
  bool parseBool(const std::string& str);

  // Helper to convert string to integer
  int parseInt(const std::string& str);
};

}  // namespace api
}  // namespace lib

#endif
