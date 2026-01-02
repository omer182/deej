#include <Arduino.h>
#include <audio_device_selector.h>
#include <mute_button.h>
#include <serial_api.h>
#include <slider.h>
#include <string.h>
#include <util.h>

#include <string>
#include <vector>

#define SLIDER_0_PIN 34
#define SLIDER_1_PIN 35
#define SLIDER_2_PIN 33
#define SLIDER_3_PIN 32
#define SLIDER_4_PIN 36
#define MUTE_BUTTON_0_PIN 14
#define MUTE_BUTTON_0_LED_PIN 12
#define MUTE_BUTTON_1_PIN 4
#define MUTE_BUTTON_1_LED_PIN 21
#define AUDIO_DEVICE_SELECTOR_BUTTON_PIN 5
#define AUDIO_DEVICE_SELECTOR_BUTTON_DEV_0_LED_PIN 18
#define AUDIO_DEVICE_SELECTOR_BUTTON_DEV_1_LED_PIN 19

using lib::api::SerialApi;
using lib::input_components::AudioDeviceSelector;
using lib::input_components::MuteButton;
using lib::input_components::Slider;

SerialApi *serial_api = nullptr;
std::vector<MuteButton *> *mute_buttons = nullptr;
std::vector<Slider *> *sliders = nullptr;
AudioDeviceSelector *audio_device_selector = nullptr;

void setup() {
  Serial.begin(115200);

  mute_buttons = new std::vector<MuteButton *>();
  // Controls two sessions (speakers and headphones).
  MuteButton *output_devices_mute_button =
      new MuteButton(0, MUTE_BUTTON_0_PIN, MUTE_BUTTON_0_LED_PIN, 2);
  MuteButton *mic_mute_button =
      new MuteButton(1, MUTE_BUTTON_1_PIN, MUTE_BUTTON_1_LED_PIN);
  mute_buttons->push_back(output_devices_mute_button);
  mute_buttons->push_back(mic_mute_button);

  sliders = new std::vector<Slider *>();
  sliders->push_back(new Slider(0, SLIDER_0_PIN,
                                std::make_optional<Slider::SessionMuteButton>(
                                    {output_devices_mute_button, 0})));
  sliders->push_back(new Slider(1, SLIDER_1_PIN,
                                std::make_optional<Slider::SessionMuteButton>(
                                    {output_devices_mute_button, 1})));
  sliders->push_back(new Slider(2, SLIDER_2_PIN));
  sliders->push_back(new Slider(3, SLIDER_3_PIN));
  sliders->push_back(new Slider(4, SLIDER_4_PIN));

  audio_device_selector = new AudioDeviceSelector(
      AUDIO_DEVICE_SELECTOR_BUTTON_PIN,
      AUDIO_DEVICE_SELECTOR_BUTTON_DEV_0_LED_PIN,
      AUDIO_DEVICE_SELECTOR_BUTTON_DEV_1_LED_PIN, output_devices_mute_button,
      []() { esp_restart(); });

  serial_api = new SerialApi();

  // Visually indicate that the system is ready.
  util::sequentialLEDOn(MUTE_BUTTON_0_LED_PIN, MUTE_BUTTON_1_LED_PIN,
                        AUDIO_DEVICE_SELECTOR_BUTTON_DEV_0_LED_PIN,
                        AUDIO_DEVICE_SELECTOR_BUTTON_DEV_1_LED_PIN, 300);
}

void loop() {
  // Read and send slider values
  std::string sliders_data = "Sliders";
  bool sliders_changed = false;
  std::vector<bool> auto_mute_triggered(mute_buttons->size(), false);

  for (int i = 0; i < sliders->size(); i++) {
    auto [changed, value] = sliders->at(i)->getValue();
    sliders_changed |= changed;
    sliders_data += "|";
    sliders_data += std::to_string(value);

    // If slider hit 0 and has a mute button, trigger auto-mute
    if (changed && value == 0 && sliders->at(i)->hasMuteButton()) {
      auto mute_btn = sliders->at(i)->getMuteButton();
      if (mute_btn.has_value()) {
        // Find which mute button index this is
        for (int j = 0; j < mute_buttons->size(); j++) {
          if (mute_buttons->at(j) == mute_btn->button) {
            auto_mute_triggered[j] = true;
            break;
          }
        }
      }
    }
  }
  if (sliders_changed) {
    serial_api->sendSliders(sliders_data);
  }

  // Read and send mute button values, update LEDs based on PC response
  std::vector<bool> mute_buttons_state(mute_buttons->size());
  bool mute_buttons_changed = false;
  for (int i = 0; i < mute_buttons->size(); i++) {
    auto [changed, value] = mute_buttons->at(i)->getValue();
    mute_buttons_changed |= changed;
    mute_buttons_state[i] = value;

    // If auto-mute was triggered by slider hitting 0, force mute state
    if (auto_mute_triggered[i]) {
      mute_buttons_changed = true;
      mute_buttons_state[i] = true;  // Force mute on
    }
  }

  if (mute_buttons_changed) {
    const auto updated_state = serial_api->sendMuteButtons(mute_buttons_state);

    // Update LEDs if we got a valid response
    if (!updated_state.empty() && updated_state.size() == mute_buttons->size()) {
      for (int i = 0; i < mute_buttons->size(); i++) {
        mute_buttons->at(i)->setActiveSessionMuteState(updated_state[i]);
      }
    }
    // Silently continue on timeout or invalid response
  }

  // Read and send audio device selector, update LEDs based on PC response
  if (auto [changed, value] = audio_device_selector->getValue(); changed) {
    const int updated_device = serial_api->sendSwitchOutput(value);

    // Update LEDs if we got a valid response
    if (updated_device >= 0) {
      audio_device_selector->setActiveDevice(updated_device);
    }
    // Silently continue on timeout or invalid response
  }

  delay(50);
}
