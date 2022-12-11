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
	var logger = log.New(os.Stdout, "", 0)
	go func(logger *log.Logger) {
		for {
			select {
			case msg := <-client.Inbox:
				logger.Printf("[%s] -> %s\n", msg.From.String(), string(msg.Data))
			case err := <-client.Error:
				logger.Println("/ERROR/:", err.Error())
			}
		}
	}(logger)
}

func handlePrompt(client *node.Node, entryPoint *peer.Peer) {
	reader := bufio.NewReader(os.Stdin)
	for {
		prompt, _ := reader.ReadBytes('\n')
		prompt = prompt[:len(prompt)-1]

		switch string(prompt) {
		case "connect":
			if entryPoint != nil {
				client.Connect <- entryPoint
			}
		case "disconnect":
			if client.IsConnected() {
				close(client.Leave)
			}
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
	selfPort, entryPort := getOptions()

	var entryPoint *peer.Peer = nil
	if entryPort > 0 {
		entryPoint = peer.Me(entryPort)
	}

	client := gop2p.StartLocalNode(selfPort)
	defer client.Wait()

	printInputs(client)
	handlePrompt(client, entryPoint)
}
