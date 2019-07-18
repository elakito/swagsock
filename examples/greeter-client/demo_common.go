package main

import (
	"encoding/json"
	"fmt"
	"log"

	apiclient "github.com/elakito/swagsock/examples/greeter-client/client"
	apioperations "github.com/elakito/swagsock/examples/greeter-client/client/operations"
	apimodels "github.com/elakito/swagsock/examples/greeter-client/models"
)

// the common code to run after the client is setup
// currently, only invokes each operation
func perform(client *apiclient.GreeterDemo) {
	// ping
	pingOK, err := client.Operations.Ping(nil)
	if err != nil {
		log.Fatalf("failed to invoke: %v", err)
	}
	// print out the json representaiton of the pong object
	jsonpayload, _ := json.Marshal(pingOK.Payload)
	fmt.Printf("Ping-Response: %s\n", jsonpayload)

	// echo
	echoBody := "hi"
	echoParams := apioperations.NewEchoParams().WithBody(echoBody)
	echoOK, err := client.Operations.Echo(echoParams)
	if err != nil {
		log.Fatalf("failed to invoke: %v", err)
	}
	jsonpayload, _ = json.Marshal(echoOK.Payload)
	fmt.Printf("Echo-Response: %s\n", jsonpayload)

	// greet
	greetBody := &apimodels.Greeting{Name: "foo", Text: "hola"}
	greetParams := apioperations.NewGreetParams().WithName("bar").WithBody(greetBody)
	greetOK, err := client.Operations.Greet(greetParams)
	if err != nil {
		log.Fatalf("failed to invoke: %v", err)
	}
	jsonpayload, _ = json.Marshal(greetOK.Payload)
	fmt.Printf("Greet-Response: %s\n", jsonpayload)

	// greetStatus
	getGreetStatusParams := apioperations.NewGetGreetStatusParams().WithName("bar")
	getGreetStatusOK, err := client.Operations.GetGreetStatus(getGreetStatusParams)
	if err != nil {
		log.Fatalf("failed to invoke: %v", err)
	}
	jsonpayload, _ = json.Marshal(getGreetStatusOK.Payload)
	fmt.Printf("GetGreetStatus-Response: %s\n", jsonpayload)

	// greetSummary
	getGreetSummaryOK, err := client.Operations.GetGreetSummary(nil)
	if err != nil {
		log.Fatalf("failed to invoke: %v", err)
	}
	jsonpayload, _ = json.Marshal(getGreetSummaryOK.Payload)
	fmt.Printf("GetGreetSummary-Response: %s\n", jsonpayload)
}

func doPing() {

}
