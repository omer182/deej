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
#define SLIDER_2_PIN 32
#define SLIDER_3_PIN 33
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

  // // Visually indicate that the system is ready.
  // util::sequentialLEDOn(MUTE_BUTTON_0_LED_PIN, MUTE_BUTTON_1_LED_PIN,
  //                       AUDIO_DEVICE_SELECTOR_BUTTON_DEV_1_LED_PIN, 300);

  // 2. Send initial slider values to sync backend
  std::string initial_sliders = "Sliders";
  for (int i = 0; i < sliders->size(); i++) {
    auto [changed, value] = sliders->at(i)->getValue();
    initial_sliders += "|";
    initial_sliders += std::to_string(value);
  }
  serial_api->sendSliders(initial_sliders);

  // 3. Send initial mute states (all unmuted by default, unless slider below threshold)
  // Check if slider 0 (speakers session, which is active by default) is below threshold
  auto [changed_0, value_0] = sliders->at(0)->getValue();
  if (value_0 < 400) {
    // Slider 0 below threshold, send mute for button 0
    serial_api->sendMuteButton(0, true);
  } else {
    // Slider 0 above threshold, ensure button 0 is unmuted
    serial_api->sendMuteButton(0, false);
  }

  // Send initial state for mic button (always unmuted at start)
  serial_api->sendMuteButton(1, false);
}

void loop() {
  static int last_sent_slider_values[5] = {-1, -1, -1, -1, -1};
  static bool previous_auto_mute_state[2] = {false, false};  // Track previous mute state for sliders 0 and 1
  static bool waitingForConnection = true;  // Connection state indicator
  const int SLIDER_CHANGE_THRESHOLD = 50;  // Only send after 50+ units change
  const int MUTE_THRESHOLD = 400;  // Below this value = muted

  // PRIORITY 0: Connection status indication
  // While waiting for backend connection, only blink LED and check for "Connected" message
  if (waitingForConnection) {
    // Blink MUTE_BUTTON_0 LED every 500ms (using millis to avoid blocking)
    digitalWrite(MUTE_BUTTON_0_LED_PIN, (millis() / 500) % 2 ? HIGH : LOW);

    // Check for "Connected" message
    if (Serial.available()) {
      String incoming = Serial.readStringUntil('\n');
      incoming.trim();

      if (incoming == "Connected") {
        // Backend connected - stop blinking and show connection sequence
        waitingForConnection = false;
        digitalWrite(MUTE_BUTTON_0_LED_PIN, LOW);  // Turn off blink LED
        util::sequentialLEDOn(MUTE_BUTTON_0_LED_PIN, MUTE_BUTTON_1_LED_PIN,
                              AUDIO_DEVICE_SELECTOR_BUTTON_DEV_1_LED_PIN, 300);
                                // Initialize state and sync with backend
        // 1. Set default output device to speakers (device 0)
        audio_device_selector->setActiveDevice(0);
        serial_api->sendSwitchOutput(0);  // Notify backend of initial device
      }
    }

    // Skip all normal processing while waiting for connection
    delay(50);
    return;
  }

  // Check for unsolicited messages from backend (e.g., "Connected")
  // NOTE: This consumes the message from the serial buffer. In the current design,
  // "Connected" is the only unsolicited message and is sent right at startup before
  // any request/response exchanges happen, so there's no conflict with SerialApi.
  // If we add more unsolicited messages in the future, we'll need a message router.
  if (Serial.available()) {
    String incoming = Serial.readStringUntil('\n');
    incoming.trim();
    // Ignore any other unsolicited messages during normal operation
  }

  // PRIORITY 1: Check mute buttons FIRST (most critical for user responsiveness)
  // Track which buttons changed and their new states
  std::vector<int> changed_buttons;  // indices of buttons that changed
  std::vector<bool> changed_states;  // new states for changed buttons

  for (int i = 0; i < mute_buttons->size(); i++) {
    auto [changed, value] = mute_buttons->at(i)->getValue();
    if (changed) {
      changed_buttons.push_back(i);
      changed_states.push_back(value);
    }
  }

  // PRIORITY 2: Check device switcher button
  bool device_changed = false;
  int new_device = 0;
  if (auto [changed, value] = audio_device_selector->getValue(); changed) {
    device_changed = true;
    new_device = value;
  }

  // PRIORITY 3: Read slider values with threshold-based change detection
  std::string sliders_data = "Sliders";
  bool sliders_changed = false;
  std::vector<int> auto_mute_action(mute_buttons->size(), -1);  // -1=no action, 0=unmute, 1=mute

  for (int i = 0; i < sliders->size(); i++) {
    auto [changed, value] = sliders->at(i)->getValue();
    sliders_data += "|";
    sliders_data += std::to_string(value);

    // Check if change is significant enough to send (reduce jitter/spam)
    bool significant_change = false;
    if (last_sent_slider_values[i] == -1) {
      // First read, always send
      significant_change = true;
    } else if (abs(value - last_sent_slider_values[i]) >= SLIDER_CHANGE_THRESHOLD) {
      // Change >= 50 units, send it
      significant_change = true;
    } else if ((value == 0 || value == 4095) && last_sent_slider_values[i] != value) {
      // At extremes (0 or max) AND value changed, send for auto-mute
      significant_change = true;
    }

    if (significant_change) {
      sliders_changed = true;
      last_sent_slider_values[i] = value;
    }

    // Auto-mute/unmute control based on slider position
    // CRITICAL: Only control mute if this slider's session is the active session
    // AND only when crossing the mute threshold boundary
    if (changed && sliders->at(i)->hasMuteButton()) {
      auto mute_btn = sliders->at(i)->getMuteButton();
      if (mute_btn.has_value()) {
        // Check if this slider's session matches the active device
        int active_device = audio_device_selector->getActiveDevice();
        if (mute_btn->session == active_device) {
          // Calculate current mute state based on threshold
          bool current_mute_state = (value < MUTE_THRESHOLD);

          // Only trigger action if mute state actually changed
          if (current_mute_state != previous_auto_mute_state[i]) {
            previous_auto_mute_state[i] = current_mute_state;

            // Find which mute button index this is
            for (int j = 0; j < mute_buttons->size(); j++) {
              if (mute_buttons->at(j) == mute_btn->button) {
                auto_mute_action[j] = current_mute_state ? 1 : 0;  // 1=mute, 0=unmute
                break;
              }
            }
          }
        }
      }
    }
  }

  // Apply auto-mute actions - add to changed buttons list
  for (int i = 0; i < mute_buttons->size(); i++) {
    if (auto_mute_action[i] != -1) {
      // Check if this button is already in changed list (button was pressed)
      bool already_changed = false;
      for (size_t j = 0; j < changed_buttons.size(); j++) {
        if (changed_buttons[j] == i) {
          // Override with auto-mute state
          changed_states[j] = (auto_mute_action[i] == 1);
          already_changed = true;
          break;
        }
      }
      if (!already_changed) {
        changed_buttons.push_back(i);
        changed_states.push_back(auto_mute_action[i] == 1);
      }
    }
  }

  // SEND MESSAGES IN PRIORITY ORDER

  // 1. Send mute button changes immediately (highest priority)
  // Send individual mute events for each changed button
  for (size_t i = 0; i < changed_buttons.size(); i++) {
    bool success = serial_api->sendMuteButton(changed_buttons[i], changed_states[i]);
    if (success) {
      // Backend confirmed - update button's internal state for LED
      mute_buttons->at(changed_buttons[i])->setActiveSessionMuteState(changed_states[i]);
    }
  }

  // 2. Send device switch changes immediately (high priority)
  if (device_changed) {
    // Device already switched locally in getValue(), just notify backend
    serial_api->sendSwitchOutput(new_device);
    // No need to call setActiveDevice again - already done in getValue()

    // Check if the newly active device's slider is below mute threshold
    // If so, send mute command to align Windows state with slider position
    if (new_device < sliders->size()) {
      auto [changed, value] = sliders->at(new_device)->getValue();
      if (value < MUTE_THRESHOLD) {
        // Find the mute button that controls this device
        if (sliders->at(new_device)->hasMuteButton()) {
          auto mute_btn = sliders->at(new_device)->getMuteButton();
          if (mute_btn.has_value() && mute_btn->session == new_device) {
            // Find the button index
            for (int j = 0; j < mute_buttons->size(); j++) {
              if (mute_buttons->at(j) == mute_btn->button) {
                bool success = serial_api->sendMuteButton(j, true);
                if (success) {
                  mute_buttons->at(j)->setActiveSessionMuteState(true);
                  previous_auto_mute_state[new_device] = true;
                }
                break;
              }
            }
          }
        }
      } else {
        // Slider above threshold, ensure unmuted
        if (sliders->at(new_device)->hasMuteButton()) {
          auto mute_btn = sliders->at(new_device)->getMuteButton();
          if (mute_btn.has_value() && mute_btn->session == new_device) {
            for (int j = 0; j < mute_buttons->size(); j++) {
              if (mute_buttons->at(j) == mute_btn->button) {
                bool success = serial_api->sendMuteButton(j, false);
                if (success) {
                  mute_buttons->at(j)->setActiveSessionMuteState(false);
                  previous_auto_mute_state[new_device] = false;
                }
                break;
              }
            }
          }
        }
      }
    }
  }

  // 3. Send slider changes only when significant (lower priority, reduced spam)
  if (sliders_changed) {
    serial_api->sendSliders(sliders_data);
  }

  delay(50);
}
