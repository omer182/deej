#ifndef LIB_INPUT_COMPONENTS_SLIDER_H
#define LIB_INPUT_COMPONENTS_SLIDER_H

#include <mute_button.h>

#include <optional>
#include <tuple>

namespace lib {
namespace input_components {

class Slider {
 public:
  struct SessionMuteButton {
    MuteButton* button;
    int session;
  };

  Slider(int slider_index, int gpioPinNumber)
      : Slider(slider_index, gpioPinNumber, std::nullopt) {}

  Slider(int slider_index, int gpioPinNumber,
         std::optional<SessionMuteButton> session_mute_button)
      : _slider_index(slider_index),
        _gpioPinNumber(gpioPinNumber),
        _session_mute_button(session_mute_button),
        _previous_value(-1),
        _previous_led_mute_state(false) {
    this->getValue();  // Force update new state.
  }

  std::tuple<bool, int> getValue();

  // Returns true if this slider has an associated mute button
  inline bool hasMuteButton() const { return _session_mute_button.has_value(); }

  // Returns the mute button and session index if available
  inline std::optional<SessionMuteButton> getMuteButton() const {
    return _session_mute_button;
  }

  const int _slider_index;

 private:
  const int _gpioPinNumber;
  int _previous_value;
  const std::optional<SessionMuteButton> _session_mute_button;
  bool _previous_led_mute_state;
};

}  // namespace input_components
}  // namespace lib

#endif