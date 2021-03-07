package main

import (
	"fmt"
	"hash/fnv"
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

func main() {
	// parameters
	root := "./"
	excludePatterns := []string{"*.exe", "*.git*"}
	includePatterns := []string{"*.go"}

	// Format include & exclude patterns into valid regex
	FormatPatternSlice(excludePatterns)
	FormatPatternSlice(includePatterns)

	commands := []string{"clear", "gofmt -w kigo.go"}

	FilesHash := make(FilesHash)

	for {
		time.Sleep(1000)

		changedFiles, err := ComputeChanges(FilesHash, root, excludePatterns, includePatterns)
		if err != nil {
			log.Fatal(err)
		}

		if len(changedFiles) == 0 {
			continue
		}

		fmt.Println(changedFiles)
		ExecuteCommands(commands)
	}
}
