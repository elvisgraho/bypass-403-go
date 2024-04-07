package main

import (
	"embed"
	"log"

	"github.com/elvisgraho/bypass-403-go/utils"
)

//go:embed all:payloads
var payloadFiles embed.FS

func main() {
	// Set log output format to include timestamps
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Println("Started bypass-403-go")

	// Parse user input
	userSettings := utils.UserInput()

	// Parse payloads
	payloads, err := utils.ParsePayloads(&payloadFiles)
	if err != nil {
		log.Fatal(err)
	}

	// fingerprint server
	utils.FingerprintRequests(userSettings)

	// HTTP Method Attack
	if !userSettings.DoSkipMethodsAttack && payloads["methods.txt"] != nil {
		utils.HttpMethodAttack(userSettings, payloads["methods.txt"])
	}

	// Prepending to the url path
	if !userSettings.DoSkipUrlAttacks && payloads["url_before.txt"] != nil {
		utils.UrlBeforeAttack(userSettings, payloads["url_before.txt"])
	}

	// Appending to the url path
	if !userSettings.DoSkipUrlAttacks && payloads["url_after.txt"] != nil {
		utils.UrlAfterAttack(userSettings, payloads["url_after.txt"])
	}

	// X forwarded Ports Attack
	if payloads["ports.txt"] != nil {
		utils.XForwardedPortsAttack(userSettings, payloads["ports.txt"])
	}

	// Attacking with full headers
	if payloads["headers_full.txt"] != nil {
		utils.HeaderAttack(userSettings, payloads["headers_full.txt"])
	}

	// Attacking with user_agents
	if !userSettings.DoSkipAgentAttacks && payloads["user_agents.txt"] != nil {
		utils.HeaderAttack(userSettings, payloads["user_agents.txt"])
	}

	// Attacking with headers path /admin
	if payloads["headers_path.txt"] != nil {
		combinedList := utils.CombineLists(payloads["headers_path.txt"], []string{userSettings.Url.Path})
		utils.HeaderAttack(userSettings, combinedList)
	}

	// Attacking with headers and ip's
	log.Println("Starting the last attack: headers_ip + ip.")
	if payloads["ip.txt"] != nil && payloads["headers_ip.txt"] != nil {
		combinedList := utils.CombineLists(payloads["headers_ip.txt"], payloads["ip.txt"])
		utils.HeaderAttack(userSettings, combinedList)
	}

	log.Println("Finished bypass-403-go")
}
