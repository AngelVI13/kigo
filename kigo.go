package main

import (
	"flag"
	"fmt"
	"github.com/AngelVI13/kigo/utils"
	"github.com/jwalton/gchalk"
	"hash/fnv"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

import _ "embed"

//go:embed defaults.json
var DEFAULT_CONFIGS string

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

func ComputeChanges(filesHash FilesHash, config *utils.Config) ([]string, error) {
	var changedFiles []string

	err := filepath.Walk(config.RootPath, func(path string, info os.FileInfo, err error) error {
		// Ignore directories, excluded patterns and patterns not present in the include filter
		if info.IsDir() ||
			PatternsInPath(config.ExcludePatterns, path) ||
			!PatternsInPath(config.IncludePatterns, path) {
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

var CommandStyle = gchalk.WithBold().Green
var ErrorStyle = gchalk.WithBold().Red

func ExecuteCommands(config *utils.Config, changedFiles []string) {
	log.Println(gchalk.WithBold().Blue("Running commands..."))

	for _, cmd := range config.Commands {
		files := strings.Join(changedFiles, " ")
		cmd = strings.ReplaceAll(cmd, config.FilesPlaceholder, files)

		// split the command into parts (a part is any whitespace separated chain of chars)
		command := strings.Fields(cmd)
		executable := command[0]
		args := command[1:len(command)]

		out, err := exec.Command(executable, args...).CombinedOutput()

		commandLine := fmt.Sprintf("%s %s:", config.Delimiter, cmd)
		log.Printf(CommandStyle(commandLine))

		if len(out) > 0 {
			log.Printf("\n%s\n", out)
		} else {
			log.Println(CommandStyle("...ok"))
		}

		if err != nil {
			log.Println(err)
			log.Printf(ErrorStyle(fmt.Sprintf("Error while executing: `%s`\n", cmd)))
			log.Printf(ErrorStyle(fmt.Sprintf("Interrupting further execution.\n\n")))
			break
		}
	}
}

func getSleepDuration(configInterval int, defaultInterval time.Duration) time.Duration {
	// Compute sleep duration in seconds. Minimum sleep is 1s.
	sleepDuration := time.Duration(configInterval) * time.Second
	if sleepDuration < defaultInterval {
		sleepDuration = defaultInterval
	}
	return sleepDuration
}

const ChangedFilesPlaceholder = "<files>"

func Run(config *utils.Config) {
	FilesHash := make(FilesHash)
	sleepDuration := getSleepDuration(config.Interval, time.Second)

	// Set default placeholder for changed files if not provided
	if config.FilesPlaceholder == "" {
		config.FilesPlaceholder = ChangedFilesPlaceholder
	}

	// the first iteration of the loop will mark all files as changed
	// since we are just building the file hash. This will lead to all
	// commands being executed even though no actual file changes have happened.
	// Disable command execution on startup iteration.
	isStartup := true

	for {
		// todo check for keypresses and exit gracefully

		time.Sleep(sleepDuration)

		changedFiles, err := ComputeChanges(FilesHash, config)
		if err != nil {
			log.Fatal(err)
		}

		if isStartup {
			isStartup = false
			continue
		}

		if len(changedFiles) == 0 {
			continue
		}

		ExecuteCommands(config, changedFiles)
	}
}

// todo 1. add tests
// todo 2. split logic into packages
// todo 3. update readme
func main() {
	var configPath = flag.String("config", "config.json", "Path to config file. (ex. `config.json`)")
	flag.Parse()

	config, err := utils.LoadConfig(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	Run(&config)
}
