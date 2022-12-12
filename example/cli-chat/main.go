package main

import (
	"bufio"
	"flag"
	"log"
	"os"

	"github.com/lucasmenendez/gop2p"
	"github.com/lucasmenendez/gop2p/pkg/message"
	"github.com/lucasmenendez/gop2p/pkg/node"
	"github.com/lucasmenendez/gop2p/pkg/peer"
)

func getOptions() (int, int) {
	var selfPortFlag = flag.Int("self", 5000, "self node port")
	var entryPortFlag = flag.Int("entry", 5000, "entrypoint node port")
	flag.Parse()

	var selfPort, entryPort = *selfPortFlag, *entryPortFlag
	if selfPort == entryPort {
		entryPort = -1
	}

	return selfPort, entryPort
}

func printInputs(client *node.Node) {
	// Create a logger and pass as argument to the goroutine
	var logger = log.New(os.Stdout, "", 0)
	go func(logger *log.Logger) {
		for {
			select {
			// Catch messages
			case msg := <-client.Inbox:
				logger.Printf("[%s] -> %s\n", msg.From.String(), string(msg.Data))
			// Catch errors
			case err := <-client.Error:
				logger.Println("/ERROR/:", err.Error())
			}
		}
	}(logger)
}

func handlePrompt(client *node.Node, entryPoint *peer.Peer) {
	// Start reading stdin in a while-true loop
	var reader = bufio.NewReader(os.Stdin)
	for {
		// Read every line writted by the user and clean the text
		prompt, _ := reader.ReadBytes('\n')
		prompt = prompt[:len(prompt)-1]

		// Catch some commands or send the input as a message
		switch string(prompt) {
		case "connect":
			if entryPoint != nil {
				client.Connect <- entryPoint
			}
		case "disconnect":
			close(client.Leave)
		case "exit":
			client.Stop()
			return
		default:
			var msg = new(message.Message).SetFrom(client.Self).SetData(prompt)
			client.Outbox <- msg
		}
	}
}

func main() {
	// Get parsed current node and entry point node ports from cmd flags
	selfPort, entryPort := getOptions()

	// If entry point node port is not setted, the default value will be 0,
	// which is not a valid as port value so peer.Me function will be return nil
	var entryPoint = peer.Me(entryPort)

	// Start current node on provided port
	client := gop2p.StartLocalNode(selfPort)
	// Keep running
	defer client.Wait()

	// Launch a goroutine to handle new messages and errors
	printInputs(client)
	// Listen for user commands ('connect', 'disconnect', 'exit') and messages
	handlePrompt(client, entryPoint)
}
