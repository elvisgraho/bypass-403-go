package utils

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
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
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // Ignore certificate errors
			},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

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

	// Set user defined headers if provided
	for _, userHeader := range userSettings.UserHeaders {
		splitHeader := strings.SplitN(userHeader, ":", 2)
		if len(splitHeader) != 2 {
			return nil, fmt.Errorf("invalid header format: %s", userHeader)
		}
		key := strings.TrimSpace(splitHeader[0])
		value := strings.TrimSpace(splitHeader[1])
		req.Header.Set(key, value)
	}

	if req.Header.Get("User-Agent") == "" {
		// User-Agent header is not set, set it to the selected user agent
		userAgent := GetRandomUserAgent()
		req.Header.Set("User-Agent", userAgent)
	}

	// wait before request based on user flag
	if userSettings.Timeout != 0 {
		time.Sleep(userSettings.Timeout)
	}

	// Perform the HTTP request
	return client.Do(req)
}

func HandleHTTPResponse(resp *http.Response, additionalOutString string, userSettings UserSettings) {
	for _, size := range userSettings.FilterSize {
		if resp.ContentLength == int64(size) {
			defer resp.Body.Close()
			return
		}
	}

	for _, code := range userSettings.FilterCode {
		if resp.StatusCode == code {
			defer resp.Body.Close()
			return
		}
	}

	PrintRespInformation(resp, additionalOutString, userSettings)
	// Close the response body
	defer resp.Body.Close()
}

// getRandomUserAgent selects a random user agent from the list
func GetRandomUserAgent() string {
	return userAgents[rand.Intn(len(userAgents))]
}

func PrintRespInformation(resp *http.Response, additionalOutString string, userSettings UserSettings) {
	var stringToPrint string

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Successful response
		stringToPrint = fmt.Sprintf("\x1b[32m%s %s %s. Length: %d. %s\x1b[0m\n", resp.Request.Method, resp.Request.URL, resp.Status, resp.ContentLength, additionalOutString)
	} else if (resp.StatusCode >= 400 || resp.StatusCode < 500) && userSettings.DoShow400 {
		// try to read body
		stringToPrint = fmt.Sprintf("\x1b[31m%d Error. Length: %d. %s %s %s\x1b[0m\n", resp.StatusCode, resp.ContentLength, resp.Request.Method, resp.Request.URL, additionalOutString)
	} else if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		// print out 300 resp
		stringToPrint = fmt.Sprintf("\x1b[33m%s %s %s. Length: %d. %s\x1b[0m\n", resp.Request.Method, resp.Request.URL, resp.Status, resp.ContentLength, additionalOutString)
	} else {
		// Error response
		// fmt.Printf("Error performing %s request. Status: %s\n", resp.Request.Method, resp.Status)
	}

	PrintRespHtml(resp, userSettings, stringToPrint)
}

func PrintRespHtml(resp *http.Response, userSettings UserSettings, stringToPrint string) {
	// try to read body
	body, err := io.ReadAll(resp.Body)
	if err == nil {
		var bodyString string
		if len(body) > 0 {
			bodyString = string(body)
		} else {
			fmt.Print(stringToPrint)
			return
		}

		// user filters out this string
		if userSettings.FilterRespString != "" && strings.Contains(bodyString, userSettings.FilterRespString) {
			// user filtered this out
			return
		}

		// print resposne before the body/title
		fmt.Print(stringToPrint)

		// try to grab only title if possible
		title := FindTitle(body)
		if title != "" {
			fmt.Printf("<title>%s</title>\n", title)
		} else if len(body) > 0 {
			fmt.Printf("%s\n", body)
		}
	}

}

func FindTitle(body []byte) string {
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		fmt.Println("Error parsing HTML:", err)
		return ""
	}

	return findTitle(doc)
}

func findTitle(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "title" {
		return n.FirstChild.Data
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		title := findTitle(c)
		if title != "" {
			return title
		}
	}
	return ""
}
