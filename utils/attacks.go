package utils

import (
	"fmt"
	"log"
)

func HttpMethodAttack(url string, methods []string, userHeaders []string, filterSize int) {
	for _, method := range methods {
		// Send HTTP request for each method
		resp, err := HttpRequest(url, method, "", userHeaders)
		if err != nil {
			return
		}
		// Handle the HTTP response
		HandleHTTPResponse(resp, "", filterSize, false)
	}
}

func HeaderAttack(url string, headers []string, userHeaders []string, filterSize int) {
	for _, header := range headers {
		// Send HTTP request for each method
		resp, err := HttpRequest(url, "GET", header, userHeaders)
		if err != nil {
			log.Printf("failed to create HTTP request: %v", err)
			continue
		}

		// Handle the HTTP response
		HandleHTTPResponse(resp, header, filterSize, true)
	}
}

func UrlAfterAttack(url string, payloadList []string, userHeaders []string, filterSize int) {
	// example: https://t.com/admin..;/
	for _, payload := range payloadList {
		newUrl := fmt.Sprintf("%s%s", url, payload)
		// Send HTTP request for each method
		resp, err := HttpRequest(newUrl, "GET", "", userHeaders)
		if err != nil {
			return
		}
		// Handle the HTTP response
		HandleHTTPResponse(resp, "", filterSize, false)
	}
}

func UrlBeforeAttack(url string, payloadList []string, userHeaders []string, filterSize int) {
	// example: https://t.com/./admin
	splitUrl, _ := SplitUrl(url)

	for _, payload := range payloadList {
		newUrl := fmt.Sprintf("%s/%s%s", splitUrl[0], payload, splitUrl[1])

		// Send HTTP request for each method
		resp, err := HttpRequest(newUrl, "GET", "", userHeaders)
		if err != nil {
			return
		}
		// Handle the HTTP response
		HandleHTTPResponse(resp, "", filterSize, false)
	}
}

func XForwardedPortsAttack(url string, portsList []string, userHeaders []string, filterSize int) {
	// port bypass X-Forwarded-Port: 8080
	for _, port := range portsList {
		newHeader := fmt.Sprintf("%s: %s", "X-Forwarded-Port", port)

		// Send HTTP request for each method
		resp, err := HttpRequest(url, "GET", newHeader, userHeaders)
		if err != nil {
			return
		}
		// Handle the HTTP response
		HandleHTTPResponse(resp, newHeader, filterSize, true)
	}
}
