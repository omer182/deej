#include "slider.h"

#include <Arduino.h>

#include <optional>
#include <string>
#include <tuple>

namespace lib {
namespace input_components {
namespace {
// Threshold for treating slider as zero (to avoid jitter at bottom)
constexpr int ZERO_THRESHOLD = 400;

bool valueIsChanged(int new_val, int old_val) {
  // Always detect changes - serial bandwidth is plentiful
  // Only skip if exactly the same value
  return (old_val == -1 || new_val != old_val);
}
}  // namespace

std::tuple<bool, int> Slider::getValue() {
  int rawValue = analogRead(_gpioPinNumber);
  // Invert the value: bottom = 4095, top = 0
  int percentValue = 4095 - rawValue;

  // Apply zero threshold to avoid jitter at bottom
  if (percentValue < ZERO_THRESHOLD) {
    percentValue = 0;
  }

  if (valueIsChanged(percentValue, _previous_value)) {
    _previous_value = percentValue;
    if (_session_mute_button.has_value()) {
      _session_mute_button->button->setLedState(_session_mute_button->session,
                                                percentValue == 0);
    }
    return std::tuple(true, percentValue);
  }
  return std::tuple(false, percentValue);
}

}  // namespace input_components
}  // namespace lib
