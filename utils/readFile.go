package utils

import (
	"bufio"
	"embed"
	"fmt"
	"os"
	"strings"
)

func ReadHeadersFromFile(filename string) ([]string, error) {
	// Read file contents
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Split file contents by newline
	lines := strings.Split(string(content), "\n")

	// Filter out empty lines
	var headers []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			headers = append(headers, line)
		}
	}

	return headers, nil
}

func ParsePayloads(payloadFiles *embed.FS) (map[string][]string, error) {
	// List all files in the payloads directory
	files, err := os.ReadDir("payloads")
	if err != nil {
		return nil, err
	}

	// Map to store payloads
	payloads := make(map[string][]string)

	// Read each file and store its contents line by line
	for _, file := range files {
		filename := file.Name()
		filePath := fmt.Sprintf("payloads/%s", filename)

		// Open the file
		f, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		// Create a scanner to read the file line by line
		scanner := bufio.NewScanner(f)

		// Slice to store lines of the file
		var lines []string

		// Read each line and append it to the lines slice
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		// Check for scanner errors
		if err := scanner.Err(); err != nil {
			return nil, err
		}

		// Store the lines slice in the payloads map
		payloads[filename] = lines
	}

	return payloads, nil
}

func CombineLists(list1, list2 []string) []string {
	// iterate list 2 first
	var combinedList []string

	for _, val2 := range list2 {
		for _, val1 := range list1 {
			combinedList = append(combinedList, fmt.Sprintf("%s: %s", val1, val2))
		}
	}

	return combinedList
}
