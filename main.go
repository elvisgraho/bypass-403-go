package main

import (
	"embed"
	"log"

	"github.com/elvisgraho/bypass-403-go/utils"
)

//go:embed payloads/*.txt
var payloadFiles embed.FS

func main() {
	// Set log output format to include timestamps
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Println("Started bypass-403-go")

	// Parse user input
	url, userHeaders, filterSize := utils.UserInput()

	// Parse payloads
	payloads, err := utils.ParsePayloads(&payloadFiles)
	if err != nil {
		log.Fatal(err)
	}

	// HTTP Method Attack
	if payloads["methods.txt"] != nil {
		utils.HttpMethodAttack(url, payloads["methods.txt"], userHeaders, filterSize)
	}

	// Attacking with headers and ip's
	if payloads["ip.txt"] != nil && payloads["headers.txt"] != nil {
		combinedList := utils.CombineLists(payloads["headers.txt"], payloads["ip.txt"])
		utils.HeaderAttack(url, combinedList, userHeaders, filterSize)
	}

	// Appending to the url path
	if payloads["url_after.txt"] != nil {
		utils.UrlAfterAttack(url, payloads["url_after.txt"], userHeaders, filterSize)
	}

	// Prepending to the url path
	if payloads["url_before.txt"] != nil {
		utils.UrlBeforeAttack(url, payloads["url_before.txt"], userHeaders, filterSize)
	}

	if payloads["ports.txt"] != nil {
		utils.XForwardedPortsAttack(url, payloads["ports.txt"], userHeaders, filterSize)
	}

	log.Println("Finished bypass-403-go")
}
