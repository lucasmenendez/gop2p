package gop2p

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"regexp"
)

// baseUri contains node address template
const baseUri string = "http://%s:%s%s"

const connectPath string = "/connect"

// joinPath contains node route where listen for node join.
const joinPath string = "/join"

// leavePath contains node route where listen for node leave.
const disconnectPath string = "/disconnect"

// broadcastPath contains node route where listen for messages.
const broadcastPath string = "/broadcast"

// contentType contains HTTP Header Content-Type default option.
const contentType string = "text/plain"

const peerAddress string = "PEER_ADDRESS"
const peerPort string = "PEER_PORT"

// listener type involves http handler.
type listener func(http.ResponseWriter, *http.Request)

type network struct {
	address string
	port string
	node *Node
	client *http.Client
	server *http.Server
}

func newNetwork(n *Node) *network {
	return &network{
		address: n.Self.Address,
		port: n.Self.Port,
		node: n,
		client: &http.Client{},
	}
}

// startListen function initializes HTTP Server node assigning to each route
// its listener function.
func (n *network) start() {
	var s *http.ServeMux = http.NewServeMux()
	s.HandleFunc(connectPath, n.connectListener())

	var host string = fmt.Sprintf("%s:%s", n.address, n.port)
	n.server = &http.Server{Addr: host, Handler: s}
	go func() {
		if e := http.ListenAndServe(host, s); e != nil {
			n.node.log("Error initializing server: %s", e.Error())
			n.node.exit <- true
		}
	}()
}


func (n *network) connectEmitter(p Peer) {
	var (
		e error
		req *http.Request
		res *http.Response
		boot string = fmt.Sprintf(baseUri, p.Address, p.Port, joinPath)
	)

	if req, e = http.NewRequest(http.MethodGet, boot, nil); e != nil {
		n.node.log("Error sending join: %s", e.Error())
		n.node.exit <- true
	}

	req.Header.Add(peerAddress, n.address)
	req.Header.Add(peerPort, n.port)
	
	if res, e = n.client.Do(req); e != nil {
		n.node.log("Error sending join: %s", e.Error())
		n.node.exit <- true
	}

	defer res.Body.Close()
	var body []byte
	if body, e = ioutil.ReadAll(res.Body); e != nil {
		n.node.log("Error sending join: %s", e.Error())
		n.node.exit <- true
	}

	var rgx *regexp.Regexp = regexp.MustCompile("((.+):(.+))")
	var hosts [][][]byte = rgx.FindAllSubmatch(body, -1)
	for _, host := range hosts {
		var address, port string = string(host[2]), string(host[3])
		var p Peer = Peer{ port, address }
		n.node.join <- p
	}
}

func (n *network) connectListener() listener {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			var address, port string = r.Header.Get(peerAddress), r.Header.Get(peerPort)
			var p Peer = Peer{ port, address }
			n.node.join <- p

			var members []byte
			for _, m := range n.node.Members {
				var member string = fmt.Sprintf("%s:%s\n", m.Address, m.Port)
				members = append(members, []byte(member)...)
			}

			w.Header().Set("Content-Type", contentType)
			w.Write(members)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("405 - Method not allowed!"))
		}
	}
}

/*
// leaveEmitter function send to all node members list a message that contains
// Peer information to their leavePath of each node. Then, communicates to the
// eventLoop that Peer leave has been broadcasting.
func leaveEmitter(n *Node) {
	// MUST SHUTDOWN SERVER STUPID!
	if j, e := json.Marshal(n.Self); e == nil {
		for _, p := range n.Members {
			var u string = fmt.Sprintf(baseUri, p.Address, p.Port, leavePath)

			var r *http.Response
			var b *bytes.Buffer = bytes.NewBuffer(j)
			if r, e = http.Post(u, contentType, b); e != nil {
				n.log("Error sending join: %s", e.Error())
				n.leave <- p
			}

			r.Body.Close()
		}
	}
}

// outboxEmitter function send message struct to all node members list to node
// broadcast path.
func outboxEmitter(n *Node, m Message) {
	var c *http.Client = &http.Client{}
	for _, p := range n.Members {
		var h string = fmt.Sprintf(baseUri, p.Address, p.Port, broadcastPath)

		var e error
		var req *http.Request
		req, e := http.NewRequest(http.MethodPost, h, bytes.NewBuffer(m.Content)
		if e != nil {
			n.log("Error connecting: %s", e.Error())
			n.leave <- p
		}

		var port string = fmt.Sprintf("%d", n.Selft.Port)
		req.Header.Add("PEERPORT", port)
		if e = client.Do(req); e != nil {
			n.log("Error connecting: %s", e.Error())
			n.leave <- p
		}
	}

}

// joinListener function listens for new nodes in the network and communicate it
// to the eventLoop.
func joinListener(n *Node) listener {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var p Peer = Peer{}

			d := json.NewDecoder(r.Body)
			if e := d.Decode(&p); e != nil {
				n.log("Error decoding Peer joins: %s", e.Error())
				return
			}

			n.join <- p
			n.update <- true
			e := json.NewEncoder(w)
			e.Encode(<-n.sync)
		}
	}
}

// leaveListener function listens for Peer network leaves and communicate it to
// the eventLoop.
func leaveListener(n *Node) listener {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var p Peer = Peer{}

			d := json.NewDecoder(r.Body)
			if e := d.Decode(&p); e != nil {
				n.log("‼️ Error decoding Peer leaves: %s", e.Error())
				return
			}

			n.leave <- p
		}
	}
}

// inbocListener function listens for new message broadcasted and communicate it
// to the eventLoop.
func inboxListener(n *Node) listener {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var peer Peer = Peer{
				Port: port,
				Address: //HERE
			}
			var msg Message = Message{}

			d := json.NewDecoder(r.Body)
			if e := d.Decode(&msg); e != nil {
				n.log("‼️ Error decoding incoming message: %s", e.Error())
				return
			}

			n.inbox <- msg
		}
	}
}
*/