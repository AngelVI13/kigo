package main

import (
	"fmt"
	"hash/fnv"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

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

func main() {
	// parameters
	root := "./"
	excludePatterns := []string{"*.exe", "*.git*"}
	includePatterns := []string{"*.go"}
	// commands := []string{"echo Hello"}

	// Format include & exclude patterns into valid regex
	for i, pattern := range excludePatterns {
		excludePatterns[i] = FormatPattern(pattern)
	}
	for i, pattern := range includePatterns {
		includePatterns[i] = FormatPattern(pattern)
	}

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
	}
}
