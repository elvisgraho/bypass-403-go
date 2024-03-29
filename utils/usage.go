package utils

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type UserSettings struct {
	FilterSize          []int
	FilterCode          []int
	FilterRespString    string
	Timeout             time.Duration
	UserHeader          string
	UserHeaders         []string
	Url                 url.URL
	DoSkipUrlAttacks    bool
	DoSkipMethodsAttack bool
	DoSkipAgentAttacks  bool
	DoShow400           bool
}

func UserInput() UserSettings {
	// Define flags
	var inputUrl string
	var userHeadersFile string
	var userSettings UserSettings
	var inputFilterSize string
	var inputFilterCode string

	flag.StringVar(&inputUrl, "u", "", "Target URL (mandatory)")
	flag.StringVar(&userSettings.UserHeader, "h", "", "User header ex: \"Cookie: test\"")
	flag.StringVar(&userHeadersFile, "hfile", "", "File containing user headers, one header per line.")
	flag.StringVar(&inputFilterSize, "fs", "", "Filter response content length. -fs 0,200.")
	flag.StringVar(&inputFilterCode, "fc", "", "Filter response code. -fc 301,307.")
	flag.StringVar(&userSettings.FilterRespString, "fr", "", "Filter specific message in the response.")
	flag.BoolVar(&userSettings.DoSkipUrlAttacks, "skipUrl", false, "Skip attacks that change url.")
	flag.BoolVar(&userSettings.DoSkipMethodsAttack, "skipMethod", false, "Skip attacks that change request method.")
	flag.BoolVar(&userSettings.DoSkipAgentAttacks, "skipAgent", false, "Skip attacks that change Agent header.")
	flag.BoolVar(&userSettings.DoShow400, "show400", false, "Show all 400 errors.")
	flag.DurationVar(&userSettings.Timeout, "t", 0, "Timeout ex: 50ms.")

	// Parse flags
	err := flag.CommandLine.Parse(os.Args[1:])

	// Check for errors in flag parsing
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		PrintUsage()
		os.Exit(1)
	}

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
		log.Fatal("Error parsing URL:", err)
	}
	userSettings.Url = *parsedURL

	// parse filters
	userSettings.FilterSize, err = ParseInputStringToInt(inputFilterSize)
	if err != nil {
		log.Fatal("Error parsing filter size:", err)
	}

	userSettings.FilterCode, err = ParseInputStringToInt(inputFilterCode)
	if err != nil {
		log.Fatal("Error parsing filter code:", err)
	}

	return userSettings
}

func PrintUsage() {
	fmt.Println("bypass-403-go - Bypass 403 Forbidden requests for specific endpoints.")
	fmt.Println("Usage: bypass-403-go -u <URL> [-h <header>] [-hfile <header_file>]")
	fmt.Println("Flags:")
	fmt.Println("  -u <URL>             : Target URL (mandatory), https://example.com/admin")
	fmt.Println("  -h <header>          : User header, e.g., 'Cookie: ...'")
	fmt.Println("  -hfile <header_file> : File containing user headers, one header per line.")
	fmt.Println("  -fs numbers  : Supresses output with the desired content length.")
	fmt.Println("  -fc numbers  : Supresses output with the desired response code ex. -fc 301,307.")
	fmt.Println("  -fr string   : Supresses output with the desired response ex. -fr \"Request unsuccessful.\".")
	fmt.Println("  -skipUrl     : Skip attacks that change url.")
	fmt.Println("  -skipMethod  : Skip attacks that change request method.")
	fmt.Println("  -skipAgent   : Skip attacks that change Agent header.")
	fmt.Println("  -show400     : Show all 400 errors.")
	fmt.Println("  -t  duration : Timeout between requests in ex. -t 50ms.")
	fmt.Println("Example:")
	fmt.Println("  bypass-403-go -u https://example.com/secret -h 'Cookie: lol'")
	fmt.Println("  bypass-403-go -u https://example.com/secret -fs -1 -skipUrl -skipAgent")
}

func ParseInputStringToInt(input string) ([]int, error) {
	if input == "" {
		return []int{}, nil
	}
	var returnInts []int

	splitValues := strings.Split(input, ",")
	for i, v := range splitValues {
		splitValues[i] = strings.TrimSpace(v)
		if splitValues[i] == "" {
			return nil, errors.New("invalid input format")
		}
		intValue, err := strconv.Atoi(v)
		returnInts = append(returnInts, intValue)
		if err != nil {
			fmt.Printf("\x1b[31m %s\x1b[0m\n", err)
			os.Exit(1)
		}
	}

	return returnInts, nil
}
