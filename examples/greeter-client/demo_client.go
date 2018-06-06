package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"

	"github.com/go-openapi/strfmt"
	apiclient "github.com/elakito/swagsock/examples/greeter-client/client"
	apioperations "github.com/elakito/swagsock/examples/greeter-client/client/operations"
	apimodels "github.com/elakito/swagsock/examples/greeter-client/models"
	httpclient "github.com/go-openapi/runtime/client"
)

func main() {
	target_path := flag.String("target_path", "http://localhost:8091/samples/greeter", "target_path")
	flag.Parse()

	u, err := url.Parse(*target_path)
	if err != nil {
		log.Fatalf("failed to parse the target url %s: %v", *target_path, err)
	}
	transport := httpclient.New(u.Host, u.Path, []string{u.Scheme})

	// create the API client, with the transport
	client := apiclient.New(transport, strfmt.Default)
	fmt.Printf("demo_client invoking some operations at target_path %s ...\n", *target_path)

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
	echoParams := apioperations.NewEchoParams().WithBody(&echoBody)
	echoOK, err := client.Operations.Echo(echoParams)
	if err != nil {
		log.Fatalf("failed to invoke: %v", err)
	}
	jsonpayload, _ = json.Marshal(echoOK.Payload)
	fmt.Printf("Echo-Response: %s\n", jsonpayload)

	// greet
	greetBody := &apimodels.Body{Name: "foo", Text: "hola"}
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
