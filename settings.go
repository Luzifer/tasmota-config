package main

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type settingExtractor func([]byte) (interface{}, error)

var extractors = map[string]settingExtractor{
	"currentcal": func(p []byte) (interface{}, error) { return extractFloatToInt("CurrentCal", p) },
	"devicename": func(p []byte) (interface{}, error) { return extractGenericJSONValue("DeviceName", p) },
	"ledstate":   func(p []byte) (interface{}, error) { return extractFloatToInt("LedState", p) },
	"module":     extractModule,
	"otaurl":     func(p []byte) (interface{}, error) { return extractGenericJSONValue("OtaUrl", p) },
	"powercal":   func(p []byte) (interface{}, error) { return extractFloatToInt("PowerCal", p) },
	"teleperiod": func(p []byte) (interface{}, error) { return extractFloatToInt("TelePeriod", p) },
	"timezone":   func(p []byte) (interface{}, error) { return extractGenericJSONValue("Timezone", p) },
	"topic":      func(p []byte) (interface{}, error) { return extractGenericJSONValue("Topic", p) },
	"voltagecal": func(p []byte) (interface{}, error) { return extractFloatToInt("VoltageCal", p) },
}

func extractSettingValue(setting string, payloadChan chan []byte) (interface{}, error) {
	e, ok := extractors[strings.ToLower(setting)]
	if !ok {
		// Default extractor: Full value
		e = func(in []byte) (interface{}, error) { return string(in), nil }
	}

	var deadline = time.NewTimer(cfg.MQTTCommandTimeout)
	for {
		select {

		case payload := <-payloadChan:
			return e(payload)

		case <-deadline.C:
			return nil, errors.New("Read timed out")

		}
	}
}

func extractGenericJSONValue(setting string, payload []byte) (interface{}, error) {
	var data = map[string]interface{}{}
	if err := json.Unmarshal(payload, &data); err != nil {
		return nil, errors.Wrap(err, "Unable to map payload into map[string]interface{}")
	}

	if _, ok := data[setting]; !ok {
		return nil, errors.New("Unable to find requested value")
	}

	return data[setting], nil
}

func extractFloatToInt(setting string, payload []byte) (interface{}, error) {
	v, err := extractGenericJSONValue(setting, payload)
	if err != nil {
		return nil, err
	}

	if _, ok := v.(float64); !ok {
		return nil, errors.Errorf("Expected float value, got %T in %s", v, setting)
	}

	return int(v.(float64)), nil
}

func extractModule(payload []byte) (interface{}, error) {
	var v = &struct {
		Module map[string]string `json:"Module"`
	}{}
	if err := json.Unmarshal(payload, v); err != nil {
		return nil, err
	}

	var values []string
	for k := range v.Module {
		values = append(values, k)
	}

	if len(values) != 1 {
		return nil, errors.New("Unexpected number of module definitions found")
	}

	return strconv.Atoi(values[0])
}
