package main

import (
	"bufio"
	"flag"
	"log"
	"os"

	"github.com/lucasmenendez/gop2p"
	"github.com/lucasmenendez/gop2p/pkg/message"
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

func main() {
	selfPort, entryPort := getOptions()

	client := gop2p.StartLocalNode(selfPort)
	defer client.Wait()

	if entryPort > 0 {
		client.Connect <- peer.Me(entryPort)
	}

	var logger = log.New(os.Stdout, "", 0)
	go func() {
		for {
			select {
			case msg := <-client.Inbox:
				logger.Printf("[%s] -> %s\n", msg.From.String(), string(msg.Data))
			case err := <-client.Error:
				logger.Println("/ERROR/:", err)
			}
		}
	}()

	reader := bufio.NewReader(os.Stdin)
	for {
		prompt, _ := reader.ReadBytes('\n')
		prompt = prompt[:len(prompt)-1]

		if string(prompt) == "exit" {
			close(client.Leave)
			return
		}

		var msg = new(message.Message).SetFrom(client.Self).SetData(prompt)
		client.Outbox <- msg
	}
}
