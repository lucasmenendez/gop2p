package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/lucasmenendez/gop2p"
)

func getOptions() (int, int) {
	var selfPortFlag = flag.Int("-self", 5000, "self node port")
	var entryPortFlag = flag.Int("-entry", 5000, "entrypoint node port")
	flag.Parse()

	var selfPort, entryPort = *selfPortFlag, *entryPortFlag
	if selfPort == entryPort {
		entryPort = -1
	}

	return selfPort, entryPort
}

func main() {
	selfPort, entryPort := getOptions()

	client := gop2p.NewNode(selfPort)
	defer client.Wait()

	if entryPort > 0 {
		entryPeer := gop2p.Me(entryPort)
		client.Connect(entryPeer)
	}

	go func() {
		for {
			select {
			case msg := <-client.Inbox:
				fmt.Printf("[%s] -> %s\n", msg.From.String(), string(msg.Data))
			case err := <-client.Error:
				fmt.Println("/ERROR/:", err)
			}
		}
	}()

	reader := bufio.NewReader(os.Stdin)
	for {
		prompt, _ := reader.ReadBytes('\n')
		prompt = prompt[:len(prompt)-1]

		if string(prompt) == "exit" {
			break
		}

		var msg = new(gop2p.Message).SetFrom(client.Self).SetData(prompt)
		client.Outbox <- msg
	}
}
