package main

import (
	"fmt"
	"path/filepath"
	"os"
	"strings"
	"regexp"
	"hash/fnv"
	"io/ioutil"
)


func FormatPattern(pattern string) string {
	// Escape all "." characters
	pattern =	strings.Replace(pattern, ".", "\\.", -1)
	
	// Convert match all operator "*" to valid regex ".*"
	pattern =	strings.Replace(pattern, "*", ".*", -1)
			
	return pattern
}

// PatternsInPath Checks if any of the provided patterns is found in the path.
func PatternsInPath(patterns []string, path string) bool {	
	for _, pattern := range patterns {
		match, err := regexp.MatchString(pattern, path)
		if err != nil {
			panic(fmt.Errorf("%v", err))
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
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return fileHash, err
	}
	
	h := fnv.New64a()
	h.Write(contents)
	fileHash = h.Sum64()
	
	return fileHash, nil
}

// var Files [string]int


func main() {
	fmt.Println("Hello world")

	// parameters
	root := "./"	
	exclude_patterns := []string{"*.exe", "*.git/*"}

	// Format exclude patterns into valid regex	
	for i, pattern := range exclude_patterns {
		exclude_patterns[i] = FormatPattern(pattern)
	}
	
	fmt.Println(exclude_patterns)
	
	var files []string
	
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && !PatternsInPath(exclude_patterns, path) {
			
			hash, err := GetFileHash(path)
			if err != nil {
				panic(fmt.Errorf("%v", err))
			}
			fmt.Println(hash, path)
			
			files = append(files, path)
			

		}
		return nil
	})
	if err != nil {
		panic(fmt.Errorf("%v", err))
	}
	
	fmt.Println(files)
}
