# gop2p
Simple *Peer-to-Peer* protocol implementation in pure Go.

## Download
```bash
go get github.com/lucasmenendez/gop2p
```

## Demo
![demo.gif](demo.gif)

## Example
```go
package main

import (
    "github.com/lucasmenendez/gop2p"
    "time"
)

func main() {
    main := gop2p.InitNode("main", 5001)
    defer main.Wait()
    
    go func() {
        _main := main.Self
        //_main := gop2p.Me("main", 5001)
        
        time.Sleep(time.Second)
        node := gop2p.InitNode("peer", 5002)
        node.Connect(_main)
        defer node.Wait()
        
        time.Sleep(2 * time.Second)
        node.Broadcast("Hello network!")
        
        time.Sleep(2 * time.Second)
        node.Broadcast("Hello again network!")
        time.Sleep(2 * time.Second)
        node.Leave()
        time.Sleep(time.Second)
        
        return
    }()
    
    time.Sleep(20 * time.Second)
    main.Broadcast("Are you there?")
    time.Sleep(time.Second)
    main.Broadcast("Hi?")
    time.Sleep(time.Second)
    main.Leave()
}
```