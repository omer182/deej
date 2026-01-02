#include "audio_device_selector.h"

#include <Arduino.h>

namespace lib {
namespace input_components {

std::tuple<bool, int> AudioDeviceSelector::getValue() {
  if (digitalRead(_button_gpio_pin) == LOW) {
    // Debounce if needed.
    int debounce_count = 0;
    while (digitalRead(_button_gpio_pin) == LOW) {
      debounce_count++;
      if (debounce_count > 20) {  // 2 seconds == 2000 ms == 20 * delay(100)
        _on_longpress_override_callback();
        return std::tuple(false, _selected_device);
      }
      delay(100);
    }
    // Toggle the device immediately
    int new_device = _selected_device ^ 1;
    // Update internal state and LEDs immediately (don't wait for backend confirmation)
    setActiveDevice(new_device);
    return std::tuple(true, new_device);
  }
  return std::tuple(false, _selected_device);
}

void AudioDeviceSelector::setActiveDevice(int selected_device) {
  _selected_device = selected_device;
  if (_selected_device == 0) {
    digitalWrite(_dev_0_led_pin, HIGH);  // Device 0 active = green LED ON (HIGH)
    digitalWrite(_dev_1_led_pin, LOW);   // Device 1 inactive = blue LED OFF (LOW)
  } else {
    digitalWrite(_dev_0_led_pin, LOW);   // Device 0 inactive = green LED OFF (LOW)
    digitalWrite(_dev_1_led_pin, HIGH);  // Device 1 active = blue LED ON (HIGH)
  }
  _multi_session_mute_button->setActiveSession(selected_device);
  _multi_session_mute_button->updateLedState();  // Update mute LED for new session
}

}  // namespace input_components
}  // namespace lib
