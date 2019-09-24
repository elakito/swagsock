package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/elakito/swagsock/swagsock"
	httpclient "github.com/go-openapi/runtime/client"

	apiclient "github.com/elakito/swagsock/examples/greeter-client/client"
	apioperations "github.com/elakito/swagsock/examples/greeter-client/client/operations"
	apimodels "github.com/elakito/swagsock/examples/greeter-client/models"
)

var (
	commandList = [][]string{
		{"ping", "Ping", "ping"},
		{"ping_a", "Ping in async-mode", "ping_a"},
		{"echo", "Echo the message", "echo text"},
		{"echo_a", "Echo the message in async-mode", "echo_a text"},
		{"greet", "Greet the person", "greet name text"},
		{"greet_a", "Greet the person in async-mode", "greet_a name text"},
		{"status", "Display the greeting status", "status name"},
		{"status_a", "Display the greeting status in async-mode", "status_a name"},
		{"summary", "Display the greeting summary", "summary"},
		{"summary_a", "Display the greeting summary in async-mode", "summary_a"},
		{"subscribe_a", "Subscribe to the greeting events in async-mode", "subscribe_a"},
		{"unsubscribe_a", "Unsubscribe from the greeting events in async-mode", "unsubscribe_a sid"},
		{"help", "Display this help message", "help"},
		{"quit", "Quit the application", "quit"},
	}
)

func main() {
	target_path := flag.String("target_path", "ws://localhost:8091/samples/greeter", "target_path")
	disable := flag.Bool("disable", false, "disable swaggersocket")
	flag.Parse()

	targetpath := *target_path
	if *disable {
		// ensure the target_path begins with http*
		if strings.HasPrefix(targetpath, "ws") {
			targetpath = "http" + targetpath[2:]
		}
	} else {
		// ensure the target_path begins with ws*
		if strings.HasPrefix(targetpath, "http") {
			targetpath = "ws" + targetpath[4:]
		}
	}

	u, err := url.Parse(targetpath)
	if err != nil {
		log.Fatalf("failed to parse the target url %s: %v", targetpath, err)
	}

	var transport runtime.ClientTransport
	if *disable {
		transport = httpclient.New(u.Host, u.Path, []string{u.Scheme})
	} else {
		transport = swagsock.NewTransport(u.String())
		defer transport.(swagsock.ClientTransport).Close()
	}

	// create the API client, with the transport
	client := apiclient.New(transport, strfmt.Default)
	fmt.Printf("demo_client (swaggersocket %v) invoking operations at target %s ...\n", !*disable, targetpath)

	perform(client, !*disable)
}

// the common code to run after the client is setup
func perform(client *apiclient.GreeterDemo, genasync bool) {
	reader := bufio.NewReader(os.Stdin)
	var prompt = "user: "
	var user string
	var cmdmap map[string]func(args ...string)
	for {
		if user == "" {
			fmt.Println("Choose user ...")
		}
		fmt.Print(prompt)
		cmd, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				os.Exit(0)
			}
			fmt.Fprintln(os.Stderr, err)
		}
		cmd = strings.TrimSuffix(cmd, "\n")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		if len(cmd) == 0 {
			continue
		}
		if user == "" {

			user = cmd
			prompt = fmt.Sprintf("%s> ", user)
			cmdmap = createCommands(user, client, prompt, genasync)
		} else {
			cmdargs := strings.Fields(cmd)
			if cmdf, ok := cmdmap[cmdargs[0]]; ok {
				cmdf(cmdargs...)
			} else {
				fmt.Fprintf(os.Stdout, "Unknown command: %s\n", cmd)
			}
		}
	}
}

// creates the available commands for the client
func createCommands(user string, client *apiclient.GreeterDemo, prompt string, genasync bool) map[string]func(args ...string) {
	cmdmap := make(map[string]func(args ...string))

	println := func(reqid string, data interface{}) {
		var pdata interface{}
		switch data.(type) {
		case string, []byte:
			pdata = data
		default:
			pdata, _ = json.Marshal(data)
		}
		if reqid == "" {
			reqid = "*"
		}
		if reqid != "*" {
			fmt.Println()
		}
		fmt.Printf("res%s: %s\n", reqid, pdata)
		if reqid != "*" {
			fmt.Printf(prompt)
		}
	}

	usage := func(name string) {
		for _, cl := range commandList {
			if name == cl[0] {
				fmt.Printf("usage: %s\n", cl[2])
				break
			}
		}
	}

	checkargs := func(name string, args []string, q bool) bool {
		if q {
			fmt.Printf("invalid usage: %v\n", strings.Join(args, " "))
			usage(name)
		}
		return q
	}

	// ping
	cmdmap["ping"] = func(args ...string) {
		if checkargs("ping", args, len(args) != 1) {
			return
		}
		pingOK, err := client.Operations.Ping(nil)
		if err != nil {
			fmt.Printf("failed to invoke: %v", err)
			return
		}
		println("", pingOK.Payload)
	}

	// echo
	cmdmap["echo"] = func(args ...string) {
		if checkargs("echo", args, len(args) < 2) {
			return
		}
		echoParams := apioperations.NewEchoParams().WithBody(strings.Join(args[1:], " "))
		echoOK, err := client.Operations.Echo(echoParams)
		if err != nil {
			fmt.Printf("failed to invoke: %v", err)
			return
		}
		println("", echoOK.Payload)
	}

	// greet
	cmdmap["greet"] = func(args ...string) {
		if checkargs("greet", args, len(args) < 3) {
			return
		}
		greetBody := &apimodels.Greeting{Name: args[1], Text: strings.Join(args[2:], " ")}
		greetParams := apioperations.NewGreetParams().WithName(user).WithBody(greetBody)
		greetOK, err := client.Operations.Greet(greetParams)
		if err != nil {
			fmt.Print("failed to invoke: %v", err)
			return
		}
		println("", greetOK.Payload)
	}

	// greetStatus
	cmdmap["status"] = func(args ...string) {
		if checkargs("status", args, len(args) != 1) {
			return
		}
		getGreetStatusParams := apioperations.NewGetGreetStatusParams().WithName(user)
		getGreetStatusOK, err := client.Operations.GetGreetStatus(getGreetStatusParams)
		if err != nil {
			fmt.Print("failed to invoke: %v", err)
			return
		}
		println("", getGreetStatusOK.Payload)
	}

	// greetSummary
	cmdmap["summary"] = func(args ...string) {
		if checkargs("summary", args, len(args) != 1) {
			return
		}
		getGreetSummaryOK, err := client.Operations.GetGreetSummary(nil)
		if err != nil {
			fmt.Print("failed to invoke: %v", err)
			return
		}
		println("", getGreetSummaryOK.Payload)
	}

	if genasync {
		// ping
		cmdmap["ping_a"] = func(args ...string) {
			if checkargs("ping_a", args, len(args) != 1) {
				return
			}
			_, err := client.Operations.PingAsync(nil, func(reqid string, r *apioperations.PingOK, e error) {
				println(reqid, r.Payload)
			}, swagsock.SubmitAsyncOptionNone)
			if err != nil {
				fmt.Printf("failed to invoke: %v", err)
				return
			}
		}

		// echo
		cmdmap["echo_a"] = func(args ...string) {
			if checkargs("echo_a", args, len(args) < 2) {
				return
			}
			echoParams := apioperations.NewEchoParams().WithBody(strings.Join(args[1:], " "))
			_, err := client.Operations.EchoAsync(echoParams, func(reqid string, r *apioperations.EchoOK, e error) {
				println(reqid, r.Payload)
			}, swagsock.SubmitAsyncOptionNone)
			if err != nil {
				fmt.Printf("failed to invoke: %v", err)
				return
			}
		}

		// greet
		cmdmap["greet_a"] = func(args ...string) {
			if checkargs("greet_a", args, len(args) < 3) {
				return
			}
			greetBody := &apimodels.Greeting{Name: args[1], Text: strings.Join(args[2:], " ")}
			greetParams := apioperations.NewGreetParams().WithName(user).WithBody(greetBody)
			_, err := client.Operations.GreetAsync(greetParams, func(reqid string, r *apioperations.GreetOK, e error) {
				println(reqid, r.Payload)
			}, swagsock.SubmitAsyncOptionNone)
			if err != nil {
				fmt.Print("failed to invoke: %v", err)
				return
			}
		}

		// greetStatus
		cmdmap["status_a"] = func(args ...string) {
			if checkargs("status_a", args, len(args) != 1) {
				return
			}
			getGreetStatusParams := apioperations.NewGetGreetStatusParams().WithName(user)
			_, err := client.Operations.GetGreetStatusAsync(getGreetStatusParams, func(reqid string, r *apioperations.GetGreetStatusOK, e error) {
				println(reqid, r.Payload)
			}, swagsock.SubmitAsyncOptionNone)
			if err != nil {
				fmt.Print("failed to invoke: %v", err)
				return
			}
		}

		// greetSummary
		cmdmap["summary_a"] = func(args ...string) {
			if checkargs("summary_a", args, len(args) != 1) {
				return
			}
			_, err := client.Operations.GetGreetSummaryAsync(nil, func(reqid string, r *apioperations.GetGreetSummaryOK, e error) {
				println(reqid, r.Payload)
			}, swagsock.SubmitAsyncOptionNone)
			if err != nil {
				fmt.Print("failed to invoke: %v", err)
				return
			}
		}

		// subscribe
		cmdmap["subscribe_a"] = func(args ...string) {
			if checkargs("subscribe_a", args, len(args) != 1) {
				return
			}
			subscribeParams := apioperations.NewSubscribeParams().WithName(user)
			_, err := client.Operations.SubscribeAsync(subscribeParams, func(reqid string, r *apioperations.SubscribeOK, e error) {
				println(reqid, r.Payload)
			}, swagsock.SubmitAsyncOptionSubscribe)
			if err != nil {
				fmt.Print("failed to invoke: %v", err)
				return
			}
		}

		// unsubscribe
		cmdmap["unsubscribe_a"] = func(args ...string) {
			if checkargs("unsubscribe_a", args, len(args) != 2) {
				return
			}
			unsubscribeParams := apioperations.NewUnsubscribeParams().WithSid(args[1])
			_, err := client.Operations.UnsubscribeAsync(unsubscribeParams, func(reqid string, r *apioperations.UnsubscribeOK, e error) {
				println(reqid, r.Payload)
			}, swagsock.SubmitAsyncOptionUnsubscribe(args[1]))
			if err != nil {
				fmt.Print("failed to invoke: %v", err)
				return
			}
		}
	}

	// console commands wihout any server-side
	cmdmap["help"] = func(args ...string) {
		fmt.Println("Available commands")
		for _, cargs := range commandList {
			if strings.HasSuffix(cargs[0], "_a") && !genasync {
				continue
			}
			fmt.Printf("%s%s%s\n", cargs[0], "                    "[0:20-len(cargs[0])], cargs[1])
		}
	}

	cmdmap["quit"] = func(args ...string) {
		os.Exit(0)
	}
	return cmdmap
}
