package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"

	"github.com/go-openapi/strfmt"

	apiclient "github.com/elakito/swagsock/examples/greeter-client/client"
	"github.com/elakito/swagsock/swagsock"
)

//
// This is not the final picture of wsclient but demonstrates how the greeter requests can be sent over websocket and
// gives a hint to how it can be integrated to the current client that only supports http.
//
func main() {
	target_path := flag.String("target_path", "ws://localhost:8091/samples/greeter", "target_path")
	flag.Parse()

	u, err := url.Parse(*target_path)
	if err != nil {
		log.Fatalf("failed to parse the target url %s: %v", *target_path, err)
	}
	//transport := httpclient.New(u.Host, u.Path, []string{u.Scheme})
	transport := swagsock.NewTransport(u.String())
	defer transport.Close()

	// create the API client, with the transport
	client := apiclient.New(transport, strfmt.Default)
	fmt.Printf("demo_client invoking some operations at target_path %s ...\n", *target_path)

	perform(client, true)
}
