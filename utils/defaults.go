package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

type DefaultConfig struct {
	tag    string
	config Config
}

type DefaultConfigs struct {
	defaults []DefaultConfig
}

func ParseDefaultConfigs(defaults []byte) (configs []DefaultConfig, err error) {
	// Create new json decoder that does not allow any unknown fields in the config file
	dec := json.NewDecoder(bytes.NewReader(defaults))
	dec.DisallowUnknownFields()

	var config DefaultConfigs
	if err := dec.Decode(&config); err != nil {
		return configs, fmt.Errorf("Error while unmarshalling default configs file. Error: `%v`", err)
	}

	return config.defaults, nil
}

func WriteConfigToFile(config Config, outFilePath string) error {
	serializedData, err := json.MarshalIndent(config, "", "\t")
	if err != nil {
		return fmt.Errorf("Could not serialize config data. Error: `%v`", err)
	}

	if err = os.WriteFile(outFilePath, serializedData, 0666); err != nil {
		return fmt.Errorf("Could not write config to file. Error: `%v`", err)
	}
	return nil
}

func CreateDefaultConfig(defaults []byte, tag, outFilePath string) (err error) {
	defaultConfigs, err := ParseDefaultConfigs(defaults)
	if err != nil {
		return err
	}

	for _, defaultConfig := range defaultConfigs {
		if defaultConfig.tag != tag {
			continue
		}

		return WriteConfigToFile(defaultConfig.config, outFilePath)
	}

	return fmt.Errorf("Could not find default config for tag=`%s`", tag)
}
