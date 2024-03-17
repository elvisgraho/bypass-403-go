package utils

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"time"
)

// List of user agents
var userAgents = []string{
	"Mozilla/5.0 (Linux; Android 10; BLA-L29) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3865.116 Mobile Safari/537.36 EdgA/45.06.4.5043",
	"Mozilla/5.0 (Linux; U; Android 9; en-us; CPH1861 Build/PPR1.180610.011) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/53.0.2785.134 Mobile Safari/537.36 OppoBrowser/15.5.0.9",
	"Mozilla/5.0 (Linux; Android 10; POCOPHONE F1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.111 Mobile Safari/537.36",
	"Mozilla/5.0 (Linux; Android 10; SM-G988U) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.96 Mobile Safari/537.36",
	"Mozilla/5.0 (Linux; Android 9; SM-J330F Build/PPR1.180610.011) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.101 Mobile Safari/537.36 YaApp_Android/10.91 YaSearchBrowser/10.91",
	"Mozilla/5.0 (Linux; Android 8.1.0; SM-A260G Build/OPR6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.152 Mobile Safari/537.36",
	"Mozilla/5.0 (Linux; Android 10; M2004J19C) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.99 Mobile Safari/537.36",
	"Mozilla/5.0 (Linux; Android 9; SAMSUNG SM-J415G) AppleWebKit/537.36 (KHTML, like Gecko) SamsungBrowser/15.0 Chrome/90.0.4430.210 Mobile Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36 OPR/72.0.3815.465",
	"Mozilla/5.0 (Linux; Android 10; SM-J810G) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.101 Mobile Safari/537.36",
}

func HttpRequest(url, method string, header string, userSettings UserSettings) (*http.Response, error) {
	// Create a new HTTP client
	client := &http.Client{}

	// Create a new request
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	// Set headers from input
	if header != "" {
		// Find the index of the first occurrence of ":"
		colonIndex := bytes.IndexByte([]byte(header), ':')
		if colonIndex == -1 {
			return nil, fmt.Errorf("invalid header format: %s", header)
		}

		// Convert the header name to a byte slice and trim spaces
		headerName := bytes.TrimSpace([]byte(header[:colonIndex]))
		// Convert the header value to a byte slice and trim spaces
		headerValue := bytes.TrimSpace([]byte(header[colonIndex+1:]))

		// Set the header in the request
		req.Header.Set(string(headerName), string(headerValue))
	}

	// Set user-defined headers if provided
	for _, userHeader := range userSettings.UserHeaders {
		splitHeader := bytes.Split([]byte(userHeader), []byte(":"))
		if len(splitHeader) != 2 {
			return nil, fmt.Errorf("invalid header format: %s", userHeader)
		}
		req.Header.Set(string(bytes.TrimSpace(splitHeader[0])), string(bytes.TrimSpace(splitHeader[1])))
	}

	if req.Header.Get("User-Agent") == "" {
		// User-Agent header is not set, set it to the selected user agent
		userAgent := getRandomUserAgent()
		req.Header.Set("User-Agent", userAgent)
	}

	// wait before request based on user flag
	if userSettings.Timeout != 0 {
		time.Sleep(userSettings.Timeout)
	}

	// Perform the HTTP request
	return client.Do(req)
}

func HandleHTTPResponse(resp *http.Response, additionalOutString string, userSettings UserSettings, doStop404 bool) {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Successful response
		if userSettings.FilterSize != 0 && resp.ContentLength == int64(userSettings.FilterSize) {
			// 200, but we filter for the size
		} else {
			// print out
			fmt.Printf("\x1b[32m%s %s %s. Length: %d. %s\x1b[0m\n", resp.Request.Method, resp.Request.URL, resp.Status, resp.ContentLength, additionalOutString)
		}
	} else if resp.StatusCode == 404 && doStop404 {
		fmt.Printf("\x1b[31m404 Error. %s %s. %s\x1b[0m\n", resp.Request.Method, resp.Request.URL, additionalOutString)
		os.Exit(1)
		defer resp.Body.Close()
	} else {
		// Error response
		// fmt.Printf("Error performing %s request. Status: %s\n", resp.Request.Method, resp.Status)
	}

	// Close the response body
	defer resp.Body.Close()
}

func SplitUrl(userURL string) ([]string, error) {
	parsedURL, err := url.Parse(userURL)
	if err != nil {
		return nil, err
	}

	// Extract base URL
	baseURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)

	// Extract endpoint
	endpoint := parsedURL.Path
	if endpoint == "" {
		endpoint = "/"
	} else if endpoint[0] == '/' {
		// Remove the first character
		endpoint = endpoint[1:]
	}

	return []string{baseURL, endpoint}, nil
}

// getRandomUserAgent selects a random user agent from the list
func getRandomUserAgent() string {
	return userAgents[rand.Intn(len(userAgents))]
}
