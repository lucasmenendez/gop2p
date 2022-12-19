# gop2p example: CLI-Chat
Simple example that implements the CLI prompt as input message to broadcast to all members into the network.

##  Run example

This example allows to the use to execute a `gop2p.Node`:

1. Start the `cli-chat`:
    ```sh
    $ go run example/cli-chat/main.go
    ```

2. Connect to a other peer using the command `connect <port>`:
   ```sh 
   $ go run example/cli-chat/main.go
   [INFO] started on http://localhost:61731
   connect 53214
   ```

3. Broadcast a message typing the content and pressing enter:
   ```sh
   $ go run example/cli-chat/main.go
   [INFO] started on http://localhost:61731
   connect 53214
   hello world
   ```

4. Disconnect from the network with the command `disconnect` (you can connect again repeating the from step 1):
   ```sh
   $ go run example/cli-chat/main.go
   [INFO] started on http://localhost:61731
   connect 53214
   hello world
   disconnect
   ```

5. Exit from the `cli-chat` with the command `exit`:
   ```sh
   $ go run example/cli-chat/main.go
   [INFO] started on http://localhost:61731
   connect 53214
   hello world
   disconnect
   exit
   $
   ```