package main

import (
	"fmt"
	"path/filepath"
	"os"
	"strings"
	"regexp"
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

// var Files [string]int


func main() {
	fmt.Println("Hello world")

	root := "./"
	
	exclude_patterns := []string{"*.exe", "*.git/*"}	
	for i, pattern := range exclude_patterns {
		exclude_patterns[i] = FormatPattern(pattern)
	}
	
	fmt.Println(exclude_patterns)
	
	var files []string
	
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && !PatternsInPath(exclude_patterns, path) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		panic(fmt.Errorf("%v", err))
	}
	
	fmt.Println(files)
}
