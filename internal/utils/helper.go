package utils

import (
	"os"
	"strconv"
	"strings"
)

// WriteStringToFile writes a string to a file
func WriteStringToFile(filename, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}

// ReadStringFromFile reads a string from a file
func ReadStringFromFile(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(content)), nil
}

// WriteIntToFile writes an integer to a file
func WriteIntToFile(filename string, value int) error {
	return WriteStringToFile(filename, strconv.Itoa(value))
}

// ReadIntFromFile reads an integer from a file
func ReadIntFromFile(filename string) (int, error) {
	content, err := ReadStringFromFile(filename)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(content)
}

// FileExists checks if a file exists
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// EnsureDir ensures directory exists
func EnsureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}
