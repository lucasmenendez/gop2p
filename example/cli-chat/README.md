# gop2p example: CLI-Chat
Simple example that implements the CLI prompt as input message to broadcast to all members into the network.

##  Run example

This example allows to the use to execute a `gop2p.Node` in two modes:
 - **Host mode**: Listen for ever for entry connections and allows to broadcast messages via cli prompt. To run it execute: 

    ```sh
    go run example/cli-chat/main.go --self=<host_port>
    ```


 - **Client mode**: Connect to a Host mode `gop2p.Node`, but also allows to broadcast messages via cli prompt. To run it execute: 
 
    ```sh
    go run example/cli-chat/main.go --self=<client_port> --entry=<host_port>
    ```