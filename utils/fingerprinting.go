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
	// Fingerprint user supplied GET
	respGet, errGet := HttpRequest(userSettings.Url.String(), "GET", "", userSettings)
	if errGet != nil {
		log.Printf("Failed to fingerprint the target: %v", errGet)
		os.Exit(1)
	}

	WriteRespMemory(respGet, FingerprintStore)

	randomString := RandomStringGen(12)
	nonExistentUrl := userSettings.Url.Scheme + "://" + userSettings.Url.Host + "/" + randomString

	respNonex, errNonEx := HttpRequest(nonExistentUrl, "GET", "", userSettings)
	if errNonEx != nil {
		log.Printf("Nonexistent URL /%s fingerprint error: %v", randomString, errNonEx)
	} else {
		WriteRespMemory(respNonex, FingerprintStore)
	}

	// fingerprint random at root path, ex: /admin/api -> /admin/RANDOM
	rootPath := filepath.Dir(userSettings.Url.Path)
	rootPath = strings.ReplaceAll(rootPath, "\\", "/")
	if rootPath != "" && rootPath != "/" && rootPath != "\\" {
		rootPathRandom := rootPath + "/" + randomString
		nonExistentPathUrl := userSettings.Url.Scheme + "://" + userSettings.Url.Host + rootPathRandom
		respNonexPth, errNonExPth := HttpRequest(nonExistentPathUrl, "GET", "", userSettings)
		if errNonExPth != nil {
			log.Printf("Nonexistent URL %s fingerprint error: %v", rootPathRandom, errNonExPth)
		} else {
			WriteRespMemory(respNonexPth, FingerprintStore)
		}
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
