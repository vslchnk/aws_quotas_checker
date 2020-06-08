package utils

import (
	"fmt"
	"os"
)

// checks if value exist in the string slice
func Find(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}

	return false
}

// exits with error code 1 and returns error message
func ExitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

// merges maps with string keys and int values
func MergeMaps(maps ...map[string]int) *map[string][]int {
	res := make(map[string][]int)
	for _, m := range maps {
		for k, v := range m {
			res[k] = append(res[k], v)
		}
	}

	return &res
}

// merges maps with unique keys
func MergeMapsUniqueKeys(maps ...map[string]int) *map[string]int {
	res := make(map[string]int)
	for _, m := range maps {
		for k, v := range m {
			res[k] = v
		}
	}

	return &res
}

// checks if key exists in map
func CheckKeyMap(m map[string]*[]string, key string) bool {
	_, ok := m[key]

	return ok
}
