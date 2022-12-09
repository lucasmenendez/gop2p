package main

import (
	"bufio"
	"os"
	"strconv"

	"github.com/lucasmenendez/gop2p"
)

func parseArgs() (selfPort, entryPort int) {
	if len(os.Args) <= 1 {
		return 5000, -1
	}

	var err error
	if selfPort, err = strconv.Atoi(os.Args[1]); err != nil {
		selfPort = 5001
	}

	entryPort = 5000
	if len(os.Args) > 2 {
		if entryPort, err = strconv.Atoi(os.Args[2]); err != nil {
			entryPort = 5000
		}
	}

	return
}

func catchStdin(cb func(msg []byte), exit func()) {
	reader := bufio.NewReader(os.Stdin)
	for {
		msg, _ := reader.ReadBytes('\n')
		msg = msg[:len(msg)-1]

		if string(msg) == "exit" {
			exit()
			break
		}

		cb(msg)
	}
}

func main() {
	selfPort, entryPort := parseArgs()

	client := gop2p.NewNode(selfPort)
	client.Logger.Printf("[%s] node initialized\n", client.Self)
	defer client.Wait()

	if entryPort > 0 {
		entryPeer := gop2p.Me(entryPort)
		client.Connect(entryPeer)
	}

	go func() {
		for {
			var msg = <-client.Inbox
			client.Logger.Printf("[%s] message from %s: '%s'\n",
				client.Self.String(), msg.From.String(), string(msg.Data))
		}
	}()

	catchStdin(func(data []byte) {
		var msg = new(gop2p.Message).SetFrom(client.Self).SetData(data)
		client.Outbox <- msg
		client.Logger.Printf("[%s] message sended: '%s'\n",
			client.Self.String(), string(data))
	}, func() {
		close(client.Leave)
	})
}
