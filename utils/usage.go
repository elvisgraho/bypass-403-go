package utils

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"
)

type UserSettings struct {
	FilterSize  []string
	FilterCode  []string
	Timeout     time.Duration
	UserHeader  string
	UserHeaders []string
	Url         url.URL
}

func UserInput() UserSettings {
	// Define flags
	var inputUrl string
	var userHeadersFile string
	var userSettings UserSettings
	var inputFilterSize string
	var inputFilterCode string

	flag.StringVar(&inputUrl, "u", "", "Target URL (mandatory)")
	flag.StringVar(&userSettings.UserHeader, "h", "", "User header (optional), specify multiple times")
	flag.StringVar(&userHeadersFile, "hfile", "", "File containing user headers (optional), one header per line")
	flag.StringVar(&inputFilterSize, "fs", "", "Filter size (optional). -fs 0,200")
	flag.StringVar(&inputFilterCode, "fc", "", "Filter size (optional). -fc 301,307")
	flag.DurationVar(&userSettings.Timeout, "t", 0, "Timeout (optional) ex: 50ms")

	// Parse flags
	flag.Parse()

	// Check if mandatory flag is provided
	if inputUrl == "" || len(os.Args) == 1 {
		PrintUsage()
		os.Exit(1)
	}

	// Read headers from file if provided
	if userHeadersFile != "" {
		headersFromFile, err := ReadHeadersFromFile(userHeadersFile)
		if err != nil {
			log.Fatal("Error reading headers from file:", err)
		}
		userSettings.UserHeaders = append(userSettings.UserHeaders, headersFromFile...)
	}

	// Append individual header provided by -h flag
	if userSettings.UserHeader != "" {
		userSettings.UserHeaders = append(userSettings.UserHeaders, userSettings.UserHeader)
	}

	// parse url
	parsedURL, err := url.Parse(inputUrl)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
		os.Exit(1)
	}
	userSettings.Url = *parsedURL

	// parse filters
	userSettings.FilterSize, err = ParseStringInputSplit(inputFilterSize)
	if err != nil {
		fmt.Println("Error parsing filter size:", err)
		os.Exit(1)
	}
	userSettings.FilterCode, err = ParseStringInputSplit(inputFilterCode)
	if err != nil {
		fmt.Println("Error parsing filter code:", err)
		os.Exit(1)
	}

	return userSettings
}

func PrintUsage() {
	fmt.Println("bypass-403-go - Bypass 403 Forbidden requests for specific endpoints.")
	fmt.Println("Usage: bypass-403-go -u <URL> [-h <header>] [-hfile <header_file>]")
	fmt.Println("Flags:")
	fmt.Println("  -u <URL>             : Target URL (mandatory), https://example.com/admin")
	fmt.Println("  -h <header>          : User header (optional), e.g., 'Cookie: ...'")
	fmt.Println("  -hfile <header_file> : File containing user headers (optional), one header per line")
	fmt.Println("  -fs numbers : Supresses output with the desired size.")
	fmt.Println("  -fc numbers : Supresses output with the desired response code. Ex. -fc 301,307")
	fmt.Println("  -t  duration : Timeout between requests in. Ex. -t 50ms")
	fmt.Println("Example:")
	fmt.Println("  bypass-403-go -u https://example.com/secret -h 'Cookie: lol'")
	fmt.Println("  bypass-403-go -u https://example.com/secret -hfile headers.txt")
	fmt.Println("  bypass-403-go -u https://example.com/secret -hfile headers.txt -fs 42")
}

func ParseStringInputSplit(input string) ([]string, error) {
	if input == "" {
		return []string{}, nil
	}
	splitValues := strings.Split(input, ",")
	for i, v := range splitValues {
		splitValues[i] = strings.TrimSpace(v)
		if splitValues[i] == "" {
			return nil, errors.New("invalid input format")
		}
	}
	return splitValues, nil
}
