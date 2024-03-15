package utils

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func UserInput() (string, []string, int) {
	// Define flags
	var url string
	var userHeader string
	var userHeadersFile string
	var filterSize int

	flag.StringVar(&url, "u", "", "Target URL (mandatory)")
	flag.StringVar(&userHeader, "h", "", "User header (optional), specify multiple times")
	flag.StringVar(&userHeadersFile, "hfile", "", "File containing user headers (optional), one header per line")
	flag.IntVar(&filterSize, "fs", 0, "Filter size (optional)")

	// Parse flags
	flag.Parse()

	// Check if mandatory flag is provided
	if url == "" || len(os.Args) == 1 {
		PrintUsage()
		os.Exit(1)
	}

	// Read headers from file if provided
	var userHeaders []string
	if userHeadersFile != "" {
		headersFromFile, err := ReadHeadersFromFile(userHeadersFile)
		if err != nil {
			log.Fatal("Error reading headers from file:", err)
		}
		userHeaders = append(userHeaders, headersFromFile...)
	}

	// Append individual header provided by -h flag
	if userHeader != "" {
		userHeaders = append(userHeaders, userHeader)
	}

	return url, userHeaders, filterSize
}

func PrintUsage() {
	fmt.Println("403-bypass-go - Bypass 403 Forbidden requests for specific endpoints.")
	fmt.Println("Usage: 403-bypass-go -u <URL> [-h <header>] [-hfile <header_file>]")
	fmt.Println("Flags:")
	fmt.Println("  -u <URL>             : Target URL (mandatory), https://example.com/admin")
	fmt.Println("  -h <header>          : User header (optional), e.g., 'Cookie: ...'")
	fmt.Println("  -hfile <header_file> : File containing user headers (optional), one header per line")
	fmt.Println("  -fs int : Supresses output with the desired size.")
	fmt.Println("Example:")
	fmt.Println("  403-bypass-go -u https://example.com/secret -h 'Cookie: lol'")
	fmt.Println("  403-bypass-go -u https://example.com/secret -hfile headers.txt")
	fmt.Println("  403-bypass-go -u https://example.com/secret -hfile headers.txt -fs 42")
}
