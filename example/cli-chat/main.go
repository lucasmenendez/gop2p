package main

import (
	"bufio"
	"crypto/rand"
	"flag"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/lucasmenendez/gop2p/pkg/message"
	"github.com/lucasmenendez/gop2p/pkg/node"
	"github.com/lucasmenendez/gop2p/pkg/peer"
)

func getOptions() int {
	minSafePort, maxSafePort := 49152, 65535
	limit := new(big.Int).SetInt64(int64(maxSafePort - minSafePort))
	r, _ := rand.Int(rand.Reader, limit)
	randomSafePort := int(r.Int64()) + minSafePort
	selfPortFlag := flag.Int("self", randomSafePort, "self node port")
	flag.Parse()

	return *selfPortFlag
}

func printInputs(client *node.Node) {
	// Create a logger and pass as argument to the goroutine
	logger := log.New(os.Stdout, "", 0)
	logger.Printf("[INFO] started on %s\n", client.Self.Hostname())
	go func(logger *log.Logger) {
		for {
			select {
			// Catch messages
			case msg := <-client.Inbox:
				logger.Printf("[MSG] (%s) -> %s\n", msg.From, string(msg.Data))
			// Catch errors
			case err := <-client.Error:
				if err != nil {
					logger.Println("[ERROR]:", err.Error())
				}
			}
		}
	}(logger)
}

func main() {
	// Get parsed current node and entry point node ports from cmd flags
	selfPort := getOptions()

	// Start current node on provided port
	selfPeer, _ := peer.Me(selfPort, true)
	client := node.New(selfPeer)
	client.Start()

	// Launch a goroutine to handle new messages and errors
	printInputs(client)

	// Listen for user commands ('connect', 'disconnect', 'exit') and messages.
	// To do that, start reading stdin in a while-true loop.
	reader := bufio.NewReader(os.Stdin)
	for {
		// Read every line writted by the user and clean the text
		prompt, _ := reader.ReadBytes('\n')
		prompt = prompt[:len(prompt)-1]

		args := strings.Fields(string(prompt))
		if len(args) > 0 {
			// Catch some commands or send the input as a message
			switch args[0] {
			case "connect":
				if len(args) > 1 {
					if port, err := strconv.Atoi(args[1]); err == nil {
						p, _ := peer.Me(port, true)
						client.Connection <- p
					}
				}
			case "disconnect":
				close(client.Connection)
			case "dm":
				if len(args) > 2 {
					if port, err := strconv.Atoi(args[1]); err == nil {
						p, _ := peer.Me(port, true)
						data := []byte(args[2])
						msg := new(message.Message).SetFrom(client.Self).SetData(data).SetTo(p)
						client.Outbox <- msg
					}
				}
			case "exit":
				client.Stop()
				return
			default:
				msg := new(message.Message).SetFrom(client.Self).SetData(prompt)
				client.Outbox <- msg
			}
		}

	}
}
