[![Go Report Card](https://goreportcard.com/badge/github.com/Luzifer/tasmota-config)](https://goreportcard.com/report/github.com/Luzifer/tasmota-config)
![](https://badges.fyi/github/license/Luzifer/tasmota-config)
![](https://badges.fyi/github/downloads/Luzifer/tasmota-config)
![](https://badges.fyi/github/latest-release/Luzifer/tasmota-config)
![](https://knut.in/project-status/tasmota-config)

# Luzifer / tasmota-config

`tasmota-config` is a helper to configure [Tasmota](https://tasmota.github.io/docs/) devices in code:

- Settings defined in the config are fetched through MQTT
- If the setting does not match an update is issued as `BackLog` command
