[![Go Report Card](https://goreportcard.com/badge/github.com/Luzifer/tasmota-config)](https://goreportcard.com/report/github.com/Luzifer/tasmota-config)
![](https://badges.fyi/github/license/Luzifer/tasmota-config)
![](https://badges.fyi/github/downloads/Luzifer/tasmota-config)
![](https://badges.fyi/github/latest-release/Luzifer/tasmota-config)
![](https://knut.in/project-status/tasmota-config)

# Luzifer / tasmota-config

`tasmota-config` is a helper to configure [Tasmota](https://tasmota.github.io/docs/) devices in code:

- Settings defined in the config are fetched through MQTT
- If the setting does not match an update is issued as `BackLog` command

## Example config

```yaml
---

settings:
  TelePeriod: 30
  Timezone: +00:00

devices:

  bedroom:
    topic: bedroom
    settings:
      DeviceName: Bedroom Sensor
      Module: 0
      Template: '{"NAME":"DevRoom 1 Mov 1 BME","GPIO":[0,0,0,0,6,5,0,0,0,0,9,0,0],"FLAG":0,"BASE":19}'

  fridge:
    topic: fridge
    settings:
      DeviceName: Fridge
      Module: 6
      PowerCal: 13769
      VoltageCal: 2127

  phonecharge:
    topic: phonecharge
    settings:
      DeviceName: Phone Charger
      LedState: 0  # Don't shine on my during night
      Module: 8

...
```
