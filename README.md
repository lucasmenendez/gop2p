[![GoDoc](https://godoc.org/github.com/lucasmenendez/gop2p?status.svg)](https://godoc.org/github.com/lucasmenendez/gop2p) [![Go Report Card](https://goreportcard.com/badge/github.com/lucasmenendez/gop2p)](https://goreportcard.com/report/github.com/lucasmenendez/gop2p)

# gop2p
Simple *Peer-to-Peer* protocol implementation in pure Go. Uses HTTP client and server to communicate over internet to knowed network members.

## Download
```bash
go get github.com/lucasmenendez/gop2p@latest
```

## Docs & example
Checkout [GoDoc Documentation](https://godoc.org/github.com/lucasmenendez/gop2p).

Also, it is available a simple **example** that implments a CLI Chat [here](example/cli-chat/).

### Workflow explained

gop2p implements the following functional workflow:
 1. **Connect to the network**: The client `gop2p.Node` know a entry point of the desired network (other `gop2p.Node` that is already connected). The entry point response with the current network `gop2p.Node`'s and updates its members `gop2p.Node` list. The client `gop2p.Node` broadcast a connection request to every `gop2p.Node` received from entry point.
 2. **Broadcasting**: The client `gop2p.Node` prepares and broadcast a `gop2p.Message` to every network `gop2p.Node`.
 3. **Disconnect**: The client `gop2p.Node` broadcast a disconnection request to every network `gop2p.Node`. This `gop2p.Node`'s updates its current network members list unregistering the client `gop2p.Node`.


```mermaid
sequenceDiagram
participant Client (gop2p.Node)
participant Network entrypoint (gop2p.Node)
participant Network nodes (gop2p.Node)

Note over Client (gop2p.Node),Network entrypoint (gop2p.Node): 1. Connect to the network
Client (gop2p.Node) ->> Network entrypoint (gop2p.Node): Send connection request to knowed Network entrypoint
Network entrypoint (gop2p.Node) -->> Network entrypoint (gop2p.Node): Register Client as new member
Network entrypoint (gop2p.Node) -->> Client (gop2p.Node): Response with the current Node list of the network
Client (gop2p.Node) -->> Client (gop2p.Node): Register all the received Node's
Client (gop2p.Node) ->> Network nodes (gop2p.Node): Send connection request
Network nodes (gop2p.Node) -->> Network nodes (gop2p.Node): Register Client as new member

Note over Client (gop2p.Node),Network nodes (gop2p.Node): 2. Broadcasting message
Client (gop2p.Node) -->> Client (gop2p.Node): Create the message
Client (gop2p.Node) ->> Network nodes (gop2p.Node): Broadcast message request to current network Node's

Network nodes (gop2p.Node) -->> Network nodes (gop2p.Node): Handle received Client message

Note over Client (gop2p.Node),Network nodes (gop2p.Node): 3. Disconnection from the network
Client (gop2p.Node) -->> Client (gop2p.Node): Create the disconnect request
Client (gop2p.Node) ->> Network nodes (gop2p.Node): Broadcast disconnect request to current network Node's

Network nodes (gop2p.Node) -->> Network nodes (gop2p.Node): Unregister Client from current Node's network
```