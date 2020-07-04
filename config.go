package main

import (
	"os"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type configFile struct {
	CommandPrefix string `yaml:"command_prefix"`
	StatPrefix    string `yaml:"stat_prefix"`

	Settings map[string]interface{} `yaml:"settings"`

	Devices map[string]deviceConfig `yaml:"devices"`
}

type deviceConfig struct {
	Topic    string                 `yaml:"topic"`
	Settings map[string]interface{} `yaml:"settings"`
}

func (d deviceConfig) constructTopic(prefix, suffix string) string {
	return strings.Join([]string{
		strings.Trim(prefix, "/"),
		strings.Trim(d.Topic, "/"),
		suffix,
	}, "/")
}

func loadConfig(p string) (*configFile, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to open config file")
	}
	defer f.Close()

	var out = &configFile{
		CommandPrefix: "cmnd",
		StatPrefix:    "stat",
	}
	return out, errors.Wrap(yaml.NewDecoder(f).Decode(out), "Unable to decode config file")
}
