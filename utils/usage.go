package utils

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

type UserSettings struct {
	FilterSize  int
	Timeout     time.Duration
	UserHeader  string
	UserHeaders []string
}

func UserInput() (string, UserSettings) {
	// Define flags
	var url string
	var userHeadersFile string
	var userSettings UserSettings

	flag.StringVar(&url, "u", "", "Target URL (mandatory)")
	flag.StringVar(&userSettings.UserHeader, "h", "", "User header (optional), specify multiple times")
	flag.StringVar(&userHeadersFile, "hfile", "", "File containing user headers (optional), one header per line")
	flag.IntVar(&userSettings.FilterSize, "fs", 0, "Filter size (optional)")
	flag.DurationVar(&userSettings.Timeout, "t", 0, "Timeout (optional) ex: 50ms")

	// Parse flags
	flag.Parse()

	// Check if mandatory flag is provided
	if url == "" || len(os.Args) == 1 {
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

	return url, userSettings
}

func PrintUsage() {
	fmt.Println("bypass-403-go - Bypass 403 Forbidden requests for specific endpoints.")
	fmt.Println("Usage: bypass-403-go -u <URL> [-h <header>] [-hfile <header_file>]")
	fmt.Println("Flags:")
	fmt.Println("  -u <URL>             : Target URL (mandatory), https://example.com/admin")
	fmt.Println("  -h <header>          : User header (optional), e.g., 'Cookie: ...'")
	fmt.Println("  -hfile <header_file> : File containing user headers (optional), one header per line")
	fmt.Println("  -fs int : Supresses output with the desired size.")
	fmt.Println("  -t  duration : Timeout between requests in. Ex. -t 50ms")
	fmt.Println("Example:")
	fmt.Println("  bypass-403-go -u https://example.com/secret -h 'Cookie: lol'")
	fmt.Println("  bypass-403-go -u https://example.com/secret -hfile headers.txt")
	fmt.Println("  bypass-403-go -u https://example.com/secret -hfile headers.txt -fs 42")
}
