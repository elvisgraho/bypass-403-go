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
	url, userSettings := utils.UserInput()

	// Parse payloads
	payloads, err := utils.ParsePayloads(&payloadFiles)
	if err != nil {
		log.Fatal(err)
	}

	// HTTP Method Attack
	if payloads["methods.txt"] != nil {
		utils.HttpMethodAttack(url, payloads["methods.txt"], userSettings)
	}

	// Attacking with headers and ip's
	if payloads["ip.txt"] != nil && payloads["headers.txt"] != nil {
		combinedList := utils.CombineLists(payloads["headers.txt"], payloads["ip.txt"])
		utils.HeaderAttack(url, combinedList, userSettings)
	}

	// Appending to the url path
	if payloads["url_after.txt"] != nil {
		utils.UrlAfterAttack(url, payloads["url_after.txt"], userSettings)
	}

	// Prepending to the url path
	if payloads["url_before.txt"] != nil {
		utils.UrlBeforeAttack(url, payloads["url_before.txt"], userSettings)
	}

	// X forwarded Ports Attack
	if payloads["ports.txt"] != nil {
		utils.XForwardedPortsAttack(url, payloads["ports.txt"], userSettings)
	}

	// Attacking with other headers
	if payloads["other_headers.txt"] != nil {
		utils.HeaderAttack(url, payloads["other_headers.txt"], userSettings)
	}

	// Attacking with user agents
	if payloads["user_agents.txt"] != nil {
		utils.HeaderAttack(url, payloads["user_agents.txt"], userSettings)
	}

	log.Println("Finished bypass-403-go")
}
