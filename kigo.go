package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

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

// PatternsInPath Checks if any of the provided patterns is found in the path.
func PatternsInPath(patterns []string, path string) bool {
	for _, pattern := range patterns {
		match, err := regexp.MatchString(pattern, path)
		if err != nil {
			log.Fatal(err)
		}

		if match {
			return true
		}
	}
	return false
}

func GetFileHash(path string) (uint64, error) {
	var fileHash uint64 = 0

	// this reads the whole file in memory
	contents, err := os.ReadFile(path)
	if err != nil {
		return fileHash, err
	}

	h := fnv.New64a()
	h.Write(contents)
	fileHash = h.Sum64()

	return fileHash, nil
}

type FilesHash map[string]uint64

func ComputeChanges(filesHash FilesHash, rootPath string, excludePatterns, includePatterns []string) ([]string, error) {
	var changedFiles []string

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		// Ignore directories, excluded patterns and patterns not present in the include filter
		if info.IsDir() || PatternsInPath(excludePatterns, path) || !PatternsInPath(includePatterns, path) {
			return nil
		}

		hash, err := GetFileHash(path)
		if err != nil {
			return fmt.Errorf("Error occured while trying to get hash for: %s: %v", path, err)
		}

		elem, ok := filesHash[path]
		// if file is not in the hash map or the hashes don't match => register a file change
		if !ok || elem != hash {
			changedFiles = append(changedFiles, path)
			// update hash for changed file only.
			// there is no need to update hash for all files
			filesHash[path] = hash
		}

		return nil
	})

	return changedFiles, err
}

func ExecuteCommands(commands []string) {
	for _, cmd := range commands {
		// split the command into parts (a part is any whitespace separated chain of chars)
		command := strings.Fields(cmd)
		executable := command[0]
		args := command[1:len(command)]

		out, err := exec.Command(executable, args...).CombinedOutput()
		log.Printf("|> %s: \n%s\n", cmd, out)

		if err != nil {
			log.Printf("Error while executing: `%s`\n", cmd)
			log.Println(err)
			log.Println("Interrupting further execution.")
			break
		}
	}
}

type Config struct {
	RootPath        string
	IncludePatterns []string
	ExcludePatterns []string
	Commands        []string
}

func LoadConfig(configPath string) (config Config, err error) {
	configFile, err := os.Open(configPath)
	if err != nil {
		return config, fmt.Errorf("Failed to open config file. Error: `%v`", err)
	}
	defer configFile.Close()

	configData, err := io.ReadAll(configFile)
	if err != nil {
		return config, fmt.Errorf("Failed while reading config file. Error: `%v`", err)
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

func main() {
	// todo make this a cli argument with default value
	configPath := "config.json"

	config, err := LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	FilesHash := make(FilesHash)

	for {
		// todo check for keypresses and exit gracefully
		time.Sleep(1000)

		// todo provide directly config instead of individual parameters
		changedFiles, err := ComputeChanges(FilesHash, config.RootPath, config.ExcludePatterns, config.IncludePatterns)
		if err != nil {
			log.Fatal(err)
		}

		if len(changedFiles) == 0 {
			continue
		}

		fmt.Println(changedFiles)
		// todo add support for operations on changed files
		ExecuteCommands(config.Commands)
	}
}
