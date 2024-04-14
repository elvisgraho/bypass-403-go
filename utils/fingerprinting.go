package utils

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type RespMemory struct {
	respCode       int
	respLength     int64
	occurenceCount int
	requests       []string
}

var RespMemoryStore = map[int][]RespMemory{}
var FingerprintStore = map[int][]RespMemory{}

func FingerprintRequests(userSettings UserSettings) {
	randomString12 := RandomStringGen(12)
	randomMethod := RandomStringGen(5)
	escapeTest := "/..%2f..%2f..%2f..%2f..%2f..%2f..%2f..%2f..%2f"

	// Fingerprint user supplied GET
	respGet, errGet := HttpRequest(userSettings.Url.String(), "GET", "", userSettings)
	if errGet != nil {
		log.Printf("Failed to fingerprint the target: %v", errGet)
		os.Exit(1)
	} else {
		WriteRespMemory(respGet, FingerprintStore)
	}

	// fingerprint non existent method
	respNonMethod, errNonMethod := HttpRequest(userSettings.Url.String(), randomMethod, "", userSettings)
	if errNonMethod != nil {
		log.Printf("Failed to fingerprint the target: %v", errNonMethod)
	} else {
		WriteRespMemory(respNonMethod, FingerprintStore)
	}

	// fingerptint non existent url
	nonExistentUrl := userSettings.Url.Scheme + "://" + userSettings.Url.Host + "/" + randomString12
	respNonex, errNonEx := HttpRequest(nonExistentUrl, "GET", "", userSettings)
	if errNonEx != nil {
		log.Printf("Nonexistent URL /%s fingerprint error: %v", randomString12, errNonEx)
	} else {
		WriteRespMemory(respNonex, FingerprintStore)
	}

	// get path -> /admin/api/v1
	rootPath := filepath.Dir(userSettings.Url.Path)
	rootPath = strings.ReplaceAll(rootPath, "\\", "/")
	if rootPath != "" && rootPath != "/" && rootPath != "\\" {
		// fingerprint random at root path, ex: /admin/api/v1 -> /admin/api/RANDOM
		rootPathRandom := rootPath + "/" + randomString12
		nonExistentPathUrl := userSettings.Url.Scheme + "://" + userSettings.Url.Host + rootPathRandom
		respNonexPath, errNonExPath := HttpRequest(nonExistentPathUrl, "GET", "", userSettings)
		if errNonExPath != nil {
			log.Printf("Nonexistent URL %s fingerprint error: %v", rootPathRandom, errNonExPath)
		} else {
			WriteRespMemory(respNonexPath, FingerprintStore)
		}

		// fingerptint non existent after path, ex: /admin/api/v1/RANDOM
		nonExistentUrlPath := userSettings.Url.String() + "/" + randomString12
		respNonexPathAfter, errNonExPathAfter := HttpRequest(nonExistentUrlPath, "GET", "", userSettings)
		if errNonExPathAfter != nil {
			log.Printf("Nonexistent URL /%s fingerprint error: %v", randomString12, errNonExPathAfter)
		} else {
			WriteRespMemory(respNonexPathAfter, FingerprintStore)
		}

		// check if there two have different responces, if yes, tell the user
		notifyOnProxyDetect(respNonexPath, respNonexPathAfter)
	}

	// try to fingerpring 400
	error400Url := userSettings.Url.Scheme + "://" + userSettings.Url.Host + escapeTest
	resp400, err400 := HttpRequest(error400Url, "GET", "", userSettings)
	if err400 != nil {
		log.Printf("Failed to fingerprint the target: %v", err400)
	} else if resp400.StatusCode == 400 {
		WriteRespMemory(resp400, FingerprintStore)
	}

	// try to fingerpring 400 after path
	error400UrlAfter := userSettings.Url.String() + escapeTest
	resp400after, err400after := HttpRequest(error400UrlAfter, "GET", "", userSettings)
	if err400after != nil {
		log.Printf("Failed to fingerprint the target: %v", err400after)
	} else if resp400after.StatusCode == 400 {
		WriteRespMemory(resp400after, FingerprintStore)
	}
}

func WriteRespMemory(resp *http.Response, respStore map[int][]RespMemory) {
	// memorizes responses
	foundStore, exists := respStore[resp.StatusCode]

	if exists {
		if resp.StatusCode >= 300 && resp.StatusCode < 400 {
			// 300 codes have varied resp length (ignore variation)
			respStore[resp.StatusCode][0].occurenceCount += 1
			return
		}

		foundResp, exists, index := findRespByLength(foundStore, resp.ContentLength)
		if exists {
			// update existend response
			foundResp.occurenceCount += 1
			foundResp.requests = append(foundResp.requests, requestToString(resp.Request))
			respStore[resp.StatusCode][index] = foundResp
			return
		}
	}

	// add response to memory
	newReqMemory := RespMemory{
		respCode:       resp.StatusCode,
		respLength:     resp.ContentLength,
		occurenceCount: 1,
		requests:       []string{requestToString(resp.Request)},
	}
	respStore[resp.StatusCode] = append(respStore[resp.StatusCode], newReqMemory)
}

func findRespByLength(responses []RespMemory, respLength int64) (RespMemory, bool, int) {
	for i, resp := range responses {
		if resp.respLength == respLength {
			return resp, true, i
		}
	}
	return RespMemory{}, false, 0
}

func requestToString(req *http.Request) string {
	var sb strings.Builder

	// Write request line
	fmt.Fprintf(&sb, "%s %s %s\r\n", req.Method, req.URL, req.Proto)

	// Write headers
	req.Header.Write(&sb)
	sb.WriteString("\r\n")

	// Write body (if present)
	if req.Body != nil {
		sb.WriteString("[Body content]")
	}

	return sb.String()
}

func DoesNotMatchFingerprint(resp *http.Response) bool {
	foundFingerprint, exists := FingerprintStore[resp.StatusCode]

	if exists {
		_, exists, _ := findRespByLength(foundFingerprint, resp.ContentLength)
		if exists {
			// the response exists in fingerprint
			return false
		}
	}

	return true
}

// Function to generate a random string of given length
func RandomStringGen(length int) string {
	// 	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const charset = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func notifyOnProxyDetect(resp1 *http.Response, resp2 *http.Response) {
	if resp1.ContentLength != resp2.ContentLength {
		// different responses
		fmt.Print("\x1b[32mProxy Detected!\x1b[0m\n")
		fmt.Printf("Req 1 Content Length: %d\n", resp1.ContentLength)
		fmt.Printf("%s", requestToString(resp1.Request))
		fmt.Printf("Reg 2 Content Length: %d\n", resp2.ContentLength)
		fmt.Printf("%s", requestToString(resp2.Request))
	}
}
