package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/rconfig/v2"
)

var (
	cfg = struct {
		Config             string        `flag:"config,c" default:"config.yaml" description:"Configuration file with settings"`
		Device             string        `flag:"device,d" default:"" description:"Limit execution to specific device by name"`
		DryRun             bool          `flag:"dry-run,n" default:"false" description:"Do not execute write actions, just print changes"`
		LogLevel           string        `flag:"log-level" default:"info" description:"Log level (debug, info, warn, error, fatal)"`
		MQTTBroker         string        `flag:"mqtt-broker" default:"tcp://localhost:1883" description:"MQTT Broker to connect to"`
		MQTTCommandTimeout time.Duration `flag:"mqtt-command-timeout" default:"2s" description:"How long to wait for commands to succeed"`
		MQTTPassword       string        `flag:"mqtt-password" default:"" description:"Credentials for MQTT-Broker"`
		MQTTUsername       string        `flag:"mqtt-username" default:"" description:"Credentials for MQTT-Broker"`
		VersionAndExit     bool          `flag:"version" default:"false" description:"Prints current version and exits"`
	}{}

	version = "dev"
)

func init() {
	rconfig.AutoEnv(true)
	if err := rconfig.ParseAndValidate(&cfg); err != nil {
		log.Fatalf("Unable to parse commandline options: %s", err)
	}

	if cfg.VersionAndExit {
		fmt.Printf("tasmota-config %s\n", version)
		os.Exit(0)
	}

	if l, err := log.ParseLevel(cfg.LogLevel); err != nil {
		log.WithError(err).Fatal("Unable to parse log level")
	} else {
		log.SetLevel(l)
	}
}

func main() {
	mqttConfig := mqtt.NewClientOptions().
		AddBroker(cfg.MQTTBroker).
		SetPassword(cfg.MQTTPassword).
		SetUsername(cfg.MQTTUsername)
	mqttClient := mqtt.NewClient(mqttConfig)

	if err := checkToken(mqttClient.Connect()); err != nil {
		log.WithError(err).Fatal("Unable to connect to broker due to error or timeout")
	}

	config, err := loadConfig(cfg.Config)
	if err != nil {
		log.WithError(err).Fatal("Unable to load config file")
	}

	for devName, devConfig := range config.Devices {
		if cfg.Device != "" && devName != cfg.Device {
			log.WithField("name", devName).Trace("Skipping device as requested")
			continue
		}

		if err := processDevice(mqttClient, config, devName, devConfig); err != nil {
			log.WithField("name", devName).WithError(err).Error("Unable to process device")
		}
	}
}

func processDevice(mqttClient mqtt.Client, config *configFile, devName string, devConfig deviceConfig) error {
	log.WithField("name", devName).Info("Starting device config")

	var (
		responses = make(chan []byte, 30)
		updates   []string
	)
	defer close(responses)

	if err := checkToken(mqttClient.Subscribe(
		devConfig.constructTopic(config.StatPrefix, "RESULT"),
		1,
		func(c mqtt.Client, m mqtt.Message) { responses <- m.Payload() },
	)); err != nil {
		return errors.Wrap(err, "Unable to subscribe to topic due to error or timeout")
	}
	defer mqttClient.Unsubscribe(devConfig.constructTopic(config.StatPrefix, "RESULT"))

	for setName, setValue := range mergeSettings(config.Settings, devConfig.Settings) {
		if err := checkToken(mqttClient.Publish(
			devConfig.constructTopic(config.CommandPrefix, setName),
			1, false,
			"",
		)); err != nil {
			return errors.Wrap(err, "Unable to send request command")
		}

		resp, err := extractSettingValue(setName, responses)
		if err != nil {
			return errors.Wrap(err, "Unable to extract settings value")
		}

		if resp == setValue {
			log.WithFields(log.Fields{
				"name":     devName,
				"expected": setValue,
				"setting":  setName,
			}).Debug("Value is fine")
			continue
		}

		log.WithFields(log.Fields{
			"actual":   fmt.Sprintf("%#v (%T)", resp, resp),
			"expected": fmt.Sprintf("%#v (%T)", setValue, setValue),
			"name":     devName,
			"setting":  setName,
		}).Warn("Value needs adjustment")

		updates = append(updates, fmt.Sprintf("%s %v", setName, setValue))
	}

	if len(updates) == 0 {
		log.WithField("name", devName).Info("Device looks good, nothing to do")
		return nil
	}

	if cfg.DryRun {
		log.WithField("name", devName).Infof("Device needs %d updates but requested dry-run", len(updates))
		return nil
	}

	log.WithField("name", devName).Infof("Requesting %d updates", len(updates))

	log.WithField("name", devName).Tracef("Sending BackLog: %q", strings.Join(updates, "; "))

	if err := checkToken(mqttClient.Publish(
		devConfig.constructTopic(config.CommandPrefix, "BackLog"),
		1, false,
		strings.Join(updates, "; "),
	)); err != nil {
		return errors.Wrap(err, "Unable to send BackLog command")
	}

	return nil
}

func checkToken(tok mqtt.Token) error {
	if !tok.WaitTimeout(cfg.MQTTCommandTimeout) {
		return errors.New("Command timed out")
	}

	return errors.Wrap(tok.Error(), "Command errored")
}

func mergeSettings(global, local map[string]interface{}) map[string]interface{} {
	var out = map[string]interface{}{}

	if global != nil {
		for k, v := range global {
			out[k] = v
		}
	}

	if local != nil {
		for k, v := range local {
			out[k] = v
		}
	}

	return out
}
