package utils

import (
	"bytes"
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"math/rand"
	"net"
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
	rawURL, err := ParseRawURL(url)
	if err != nil {
		return nil, err
	}

	headers := []RawHeader{}
	headerIndex := map[string]int{}

	if header != "" {
		headers, err = appendRawHeader(headers, headerIndex, header)
		if err != nil {
			return nil, err
		}
	}

	for _, userHeader := range userSettings.UserHeaders {
		headers, err = appendRawHeader(headers, headerIndex, userHeader)
		if err != nil {
			return nil, err
		}
	}

	if _, exists := headerIndex["user-agent"]; !exists {
		headers = append(headers, RawHeader{Name: "User-Agent", Value: GetRandomUserAgent()})
		headerIndex["user-agent"] = len(headers) - 1
	}

	if userSettings.RequestDelay != 0 {
		time.Sleep(userSettings.RequestDelay)
	}

	return doRawHTTPRequest(rawURL, method, headers, userSettings.RequestTimeout)
}

type RawHeader struct {
	Name  string
	Value string
}

func appendRawHeader(headers []RawHeader, headerIndex map[string]int, rawHeader string) ([]RawHeader, error) {
	colonIndex := bytes.IndexByte([]byte(rawHeader), ':')
	if colonIndex == -1 {
		return nil, fmt.Errorf("invalid header format: %s", rawHeader)
	}

	headerName := string(bytes.TrimSpace([]byte(rawHeader[:colonIndex])))
	headerValue := string(bytes.TrimSpace([]byte(rawHeader[colonIndex+1:])))
	key := strings.ToLower(headerName)

	if index, exists := headerIndex[key]; exists {
		headers[index].Name = headerName
		headers[index].Value = headerValue
		return headers, nil
	}

	headers = append(headers, RawHeader{Name: headerName, Value: headerValue})
	headerIndex[key] = len(headers) - 1
	return headers, nil
}

func doRawHTTPRequest(rawURL RawURL, method string, headers []RawHeader, requestTimeout time.Duration) (*http.Response, error) {
	address := rawURL.Host
	if rawURL.Port() == "" {
		if rawURL.Scheme == "https" {
			address += ":443"
		} else {
			address += ":80"
		}
	}

	var conn net.Conn
	var err error
	if requestTimeout <= 0 {
		requestTimeout = 15 * time.Second
	}
	dialer := &net.Dialer{Timeout: requestTimeout}

	switch strings.ToLower(rawURL.Scheme) {
	case "https":
		conn, err = tls.DialWithDialer(dialer, "tcp", address, &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         rawURL.Hostname(),
		})
	case "http":
		conn, err = dialer.Dial("tcp", address)
	default:
		return nil, fmt.Errorf("unsupported scheme: %s", rawURL.Scheme)
	}
	if err != nil {
		return nil, err
	}
	_ = conn.SetDeadline(time.Now().Add(requestTimeout))

	var requestBuffer bytes.Buffer
	fmt.Fprintf(&requestBuffer, "%s %s HTTP/1.1\r\n", method, rawURL.RequestURI())

	hasHost := false
	hasConnection := false
	for _, rawHeader := range headers {
		switch strings.ToLower(rawHeader.Name) {
		case "host":
			hasHost = true
		case "connection":
			hasConnection = true
		}
		fmt.Fprintf(&requestBuffer, "%s: %s\r\n", rawHeader.Name, rawHeader.Value)
	}

	if !hasHost {
		fmt.Fprintf(&requestBuffer, "Host: %s\r\n", rawURL.Host)
	}
	if !hasConnection {
		requestBuffer.WriteString("Connection: close\r\n")
	}
	requestBuffer.WriteString("\r\n")

	if _, err := conn.Write(requestBuffer.Bytes()); err != nil {
		conn.Close()
		return nil, err
	}

	req := &http.Request{
		Method:     method,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Host:       rawURL.Host,
		URL:        rawURL.ParsedForDisplay(),
		RequestURI: rawURL.RequestURI(),
		Close:      true,
	}

	for _, rawHeader := range headers {
		req.Header.Set(rawHeader.Name, rawHeader.Value)
	}
	if !hasHost {
		req.Header.Set("Host", rawURL.Host)
	}
	if !hasConnection {
		req.Header.Set("Connection", "close")
	}

	resp, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		conn.Close()
		return nil, err
	}
	resp.Request = req
	return resp, nil
}

func HandleHTTPResponse(resp *http.Response, additionalOutString string, userSettings UserSettings, show300 bool) {
	WriteRespMemory(resp, RespMemoryStore)

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

	PrintRespInformation(resp, additionalOutString, userSettings, show300)
	// Close the response body
	defer resp.Body.Close()
}

// getRandomUserAgent selects a random user agent from the list
func GetRandomUserAgent() string {
	return userAgents[rand.Intn(len(userAgents))]
}

func PrintRespInformation(resp *http.Response, additionalOutString string, userSettings UserSettings, show300 bool) {
	var stringToPrint string
	var doesNotMatchFingerpting = DoesNotMatchFingerprint(resp)
	doShow400 := userSettings.DoShow400 || doesNotMatchFingerpting

	if resp.StatusCode >= 200 && resp.StatusCode < 300 && doesNotMatchFingerpting {
		// Successful response, do not show if the same as fingerprint
		stringToPrint = fmt.Sprintf("\x1b[32m%s %s %s. Length: %d. %s\x1b[0m\n", resp.Request.Method, requestURLString(resp.Request), resp.Status, resp.ContentLength, additionalOutString)
	} else if (resp.StatusCode >= 400 && resp.StatusCode < 500) && doShow400 {
		// try to read body
		stringToPrint = fmt.Sprintf("\x1b[31m%d Error. Length: %d. %s %s %s\x1b[0m\n", resp.StatusCode, resp.ContentLength, resp.Request.Method, requestURLString(resp.Request), additionalOutString)
	} else if (resp.StatusCode >= 300 && resp.StatusCode < 400) && show300 {
		// print out 300 resp
		stringToPrint = fmt.Sprintf("\x1b[33m%s %s %s. Length: %d. %s\x1b[0m\n", resp.Request.Method, requestURLString(resp.Request), resp.Status, resp.ContentLength, additionalOutString)
	} else {
		// Error response
		// fmt.Printf("Error performing %s request. Status: %s\n", resp.Request.Method, resp.Status)
	}

	if stringToPrint != "" {
		PrintRespHtml(resp, userSettings, stringToPrint)
	}
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
			if len(body) > 150 {
				fmt.Printf("%s...\n", body[:150])
			} else {
				fmt.Printf("%s\n", body)
			}
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
		if n.FirstChild == nil {
			return ""
		}
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
