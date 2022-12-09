[![GoDoc](https://godoc.org/github.com/lucasmenendez/gop2p?status.svg)](https://godoc.org/github.com/lucasmenendez/gop2p) [![Go Report Card](https://goreportcard.com/badge/github.com/lucasmenendez/gop2p)](https://goreportcard.com/report/github.com/lucasmenendez/gop2p)

# gop2p
Simple *Peer-to-Peer* protocol implementation in pure Go. Uses HTTP client and server to communicate over internet to knowed network members.

## Download
```bash
go get github.com/lucasmenendez/gop2p@latest
```

## Docs & example
- Checkout [GoDoc Documentation](https://godoc.org/github.com/lucasmenendez/gop2p).
- Also, it is available a simple **example** that implments a CLI Chat [here](example/cli-chat/).

### Workflow explained

gop2p implements the following functional workflow:

```mermaid
sequenceDiagram
participant Client (Node)
participant Network entrypoint (Node)
participant Network peers (Node)

Note over Client (Node),Network entrypoint (Node): 1. Connect to the network
Client (Node) ->> Network entrypoint (Node): Send connection request to knowed Network entrypoint
Network entrypoint (Node) -->> Network entrypoint (Node): Register Client as new member
Network entrypoint (Node) -->> Client (Node): Response with the current Node list of the network
Client (Node) -->> Client (Node): Register all the received Node's
Client (Node) ->> Network peers (Node): Send connection request
Network peers (Node) -->> Network peers (Node): Register Client as new member

Note over Client (Node),Network peers (Node): 2. Broadcasting message
Client (Node) -->> Client (Node): Create the message
Client (Node) ->> Network peers (Node): Broadcast message request to current network Node's

Network peers (Node) -->> Network peers (Node): Handle received Client message

Note over Client (Node),Network peers (Node): 3. Disconnection from the network
Client (Node) -->> Client (Node): Create the disconnect request
Client (Node) ->> Network peers (Node): Broadcast disconnect request to current network Node's

Network peers (Node) -->> Network peers (Node): Unregister Client from current Node's network
```

#### 1. Connect to the network
The client `Node` know a entry point of the desired network (other `Node` that is already connected). The entry point response with the current network `Node`'s and updates its members `Node` list. The client `Node` broadcast a connection request to every `Node` received from entry point.

```go
package main

import (
	"log"

	"github.com/lucasmenendez/gop2p"
	"github.com/lucasmenendez/gop2p/pkg/message"
	"github.com/lucasmenendez/gop2p/pkg/peer"
)

func main() {
    // -> init
    node := gop2p.StartLocalNode(5001) // Local Node
    // node := gop2p.StartNode("0.0.0.0", 5001) // Remote Node
    defer node.Wait()

    // -> connect
    node.Connect <- peer.Me(5000) // Local entry point Node
    // node.Connect <- peer.New("192.68.1.43", 5000) // Remote entry point Node

    // -> print incoming messages and erros
    var logger = log.New(os.Stdout, "", 0)
    go func() {
        for {
            select {
            case msg := <-node.Inbox:
                logger.Printf("[%s] -> %s\n", msg.From.String(), string(msg.Data))
            case err := <-node.Error:
                logger.Fatalln(err)
            }
        }
    }()

    // ...
}
```

#### 2. Broadcasting 
The client `Node` prepares and broadcast a `gop2p.Message` to every network `Node`.

```go
package main

import (
	"log"

	"github.com/lucasmenendez/gop2p"
	"github.com/lucasmenendez/gop2p/pkg/message"
	"github.com/lucasmenendez/gop2p/pkg/peer"
)

func main() {
    // ...

    var data = []byte("Hello network!")
    var msg = new(message.Message).SetFrom(node.Self).SetData(data)
    node.Outbox <- msg

    // ...
}
```

#### 3. Disconnect 
The client `Node` broadcast a disconnection request to every network `Node`. This `Node`'s updates its current network members list unregistering the client `Node`.

```go
package main

import (
	"log"

	"github.com/lucasmenendez/gop2p"
	"github.com/lucasmenendez/gop2p/pkg/message"
	"github.com/lucasmenendez/gop2p/pkg/peer"
)

func main() {
    // ...
    close(node.Leave)
}
```