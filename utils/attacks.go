package utils

import (
	"fmt"
	"log"
)

func HttpMethodAttack(url string, methods []string, userSettings UserSettings) {
	for _, method := range methods {
		// Send HTTP request for each method
		resp, err := HttpRequest(url, method, "", userSettings)
		if err != nil {
			return
		}
		// Handle the HTTP response
		HandleHTTPResponse(resp, "", userSettings, false)
	}
}

func HeaderAttack(url string, headers []string, userSettings UserSettings) {
	for _, header := range headers {
		// Send HTTP request for each method
		resp, err := HttpRequest(url, "GET", header, userSettings)
		if err != nil {
			log.Printf("failed to create HTTP request: %v", err)
			continue
		}

		// Handle the HTTP response
		HandleHTTPResponse(resp, header, userSettings, true)
	}
}

func UrlAfterAttack(url string, payloadList []string, userSettings UserSettings) {
	// example: https://t.com/admin..;/
	for _, payload := range payloadList {
		newUrl := fmt.Sprintf("%s%s", url, payload)
		// Send HTTP request for each method
		resp, err := HttpRequest(newUrl, "GET", "", userSettings)
		if err != nil {
			return
		}
		// Handle the HTTP response
		HandleHTTPResponse(resp, "", userSettings, false)
	}
}

func UrlBeforeAttack(url string, payloadList []string, userSettings UserSettings) {
	// example: https://t.com/./admin
	splitUrl, _ := SplitUrl(url)

	for _, payload := range payloadList {
		newUrl := fmt.Sprintf("%s/%s%s", splitUrl[0], payload, splitUrl[1])

		// Send HTTP request for each method
		resp, err := HttpRequest(newUrl, "GET", "", userSettings)
		if err != nil {
			return
		}
		// Handle the HTTP response
		HandleHTTPResponse(resp, "", userSettings, false)
	}
}

func XForwardedPortsAttack(url string, portsList []string, userSettings UserSettings) {
	// port bypass X-Forwarded-Port: 8080
	for _, port := range portsList {
		newHeader := fmt.Sprintf("%s: %s", "X-Forwarded-Port", port)

		// Send HTTP request for each method
		resp, err := HttpRequest(url, "GET", newHeader, userSettings)
		if err != nil {
			return
		}
		// Handle the HTTP response
		HandleHTTPResponse(resp, newHeader, userSettings, true)
	}
}
