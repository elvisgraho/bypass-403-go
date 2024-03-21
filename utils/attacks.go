package utils

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func SingleHeaderAttack(userSettings UserSettings, header string, method string) {
	// Send HTTP request for each method
	resp, err := HttpRequest(userSettings.Url.String(), method, header, userSettings)
	if err != nil {
		AttackHttpErrorHandling(err)
		return
	}

	// Handle the HTTP response
	HandleHTTPResponse(resp, header, userSettings)
}

func HttpMethodAttack(userSettings UserSettings, methods []string) {
	for _, method := range methods {
		// Send HTTP request for each method
		resp, err := HttpRequest(userSettings.Url.String(), method, "", userSettings)
		if err != nil {
			AttackHttpErrorHandling(err)
			return
		}
		// Handle the HTTP response
		HandleHTTPResponse(resp, "", userSettings)
	}
}

func HeaderAttack(userSettings UserSettings, headers []string) {
	for _, header := range headers {
		SingleHeaderAttack(userSettings, header, "GET")
	}
}

func UrlAfterAttack(userSettings UserSettings, payloadList []string) {
	// example: https://t.com/admin..;/
	for _, payload := range payloadList {
		newUrl := fmt.Sprintf("%s%s", userSettings.Url.String(), payload)

		// Send HTTP request for each method
		resp, err := HttpRequest(newUrl, "GET", "", userSettings)
		if err != nil {
			AttackHttpErrorHandling(err)
			return
		}
		// Handle the HTTP response
		HandleHTTPResponse(resp, "", userSettings)
	}
}

func UrlBeforeAttack(userSettings UserSettings, payloadList []string) {
	// example: https://t.com/./admin
	for _, payload := range payloadList {
		hostWithProtocol := userSettings.Url.Scheme + "://" + userSettings.Url.Host
		pathWithoutSlash := strings.TrimPrefix(userSettings.Url.Path, "/")
		newUrl := fmt.Sprintf("%s/%s%s", hostWithProtocol, payload, pathWithoutSlash)

		// Send HTTP request for each method
		resp, err := HttpRequest(newUrl, "GET", "", userSettings)
		if err != nil {
			AttackHttpErrorHandling(err)
			return
		}

		// Handle the HTTP response
		HandleHTTPResponse(resp, "", userSettings)
	}
}

func XForwardedPortsAttack(userSettings UserSettings, portsList []string) {
	// port bypass X-Forwarded-Port: 8080
	for _, port := range portsList {
		newHeader := fmt.Sprintf("%s: %s", "X-Forwarded-Port", port)
		SingleHeaderAttack(userSettings, newHeader, "GET")
	}
}

func AttackHttpErrorHandling(err error) {
	log.Printf("failed to create HTTP request: %v", err)
	if dnsErr, ok := err.(*net.DNSError); ok && dnsErr.IsNotFound {
		os.Exit(1)
	}
}
