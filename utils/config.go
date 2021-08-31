package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	RootPath         string
	IncludePatterns  []string
	ExcludePatterns  []string
	Delimiter        string
	Interval         int
	FilesPlaceholder string
	Commands         []string
}

func LoadConfig(configPath string) (config Config, err error) {
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return config, err
	}

	// Create new json decoder that does not allow any unknown fields in the config file
	dec := json.NewDecoder(bytes.NewReader(configData))
	dec.DisallowUnknownFields()

	if err := dec.Decode(&config); err != nil {
		return config, fmt.Errorf("Error while unmarshalling config file. Error: `%v`", err)
	}

	// Format include & exclude patterns into valid regex
	FormatPatternSlice(config.ExcludePatterns)
	FormatPatternSlice(config.IncludePatterns)
	return config, nil
}

// FormatPatternSlice Modifies all members of a pattern slice to legal regex (in-place)
func FormatPatternSlice(patterns []string) {
	for i, pattern := range patterns {
		patterns[i] = FormatPattern(pattern)
	}
}

func FormatPattern(pattern string) string {
	// Escape all "." characters
	pattern = strings.Replace(pattern, ".", "\\.", -1)

	// Convert match all operator "*" to valid regex ".*"
	pattern = strings.Replace(pattern, "*", ".*", -1)

	return pattern
}
