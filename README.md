[![GoDoc](https://godoc.org/github.com/lucasmenendez/gop2p?status.svg)](https://godoc.org/github.com/lucasmenendez/gop2p) [![Go Report Card](https://goreportcard.com/badge/github.com/lucasmenendez/gop2p)](https://goreportcard.com/report/github.com/lucasmenendez/gop2p)

# gop2p
Simple *Peer-to-Peer* protocol implementation in pure Go.

## Download
```bash
go get github.com/lucasmenendez/gop2p
```

## Example
```go
package main

import (
    "github.com/lucasmenendez/gop2p"
    "time"
)

func main() {
	main := gop2p.InitNode(5001, true)
	defer main.Wait()

	main.OnMessage(func(message []byte) {
		fmt.Printf("\t\t-> %s\n", string(message))
	})

    go func() {
        entry := main.Self
        //_entry := gop2p.CreatePeer("localhost", 5001)
        
        time.Sleep(time.Second)
        node := gop2p.InitNode(5002, true)
        node.Connect(entry)
        time.Sleep(2 * time.Second)
		node.Broadcast([]byte("Hola"))
        time.Sleep(2 * time.Second)
        node.Disconnect()
    }()
    go func() {
        entry := main.Self
        //_entry := gop2p.CreatePeer("localhost", 5001)
        
        time.Sleep(1 * time.Second)
        node := gop2p.InitNode(5003, true)
        node.Connect(entry)
        time.Sleep(2 * time.Second)
        node.Disconnect()
    }()

	time.Sleep(6 * time.Second)
	main.Disconnect()
}
```
