package gop2p

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
)

// baseUri contains node address template
const baseUri string = "http://%s:%s%s"

// connectPath contains node route where listen for node join.
const connectPath string = "/connect"

// disconnectPath contains node route where listen for node leave.
const disconnectPath string = "/disconnect"

// broadcastPath contains node route where listen for messages.
const broadcastPath string = "/broadcast"

// contentType contains HTTP Header Content-Type default option.
const contentType string = "text/plain"

// peeraddress contains default environment variable that contains address.
const peeraddress string = "PEER_ADDRESS"

// peerport contains default environment variable that contains port.
const peerport string = "PEER_PORT"

// listener type involves http handler.
type listener func(http.ResponseWriter, *http.Request)

type network struct {
	address string
	port    string
	node    *Node
	client  *http.Client
	server  *http.Server
}

func newNetwork(n *Node) *network {
	return &network{
		address: n.Self.address,
		port:    n.Self.port,
		node:    n,
		client:  &http.Client{},
	}
}

// startListen function initializes HTTP Server node assigning to each route
// its listener function.
func (n *network) start() {
	var s *http.ServeMux = http.NewServeMux()
	s.HandleFunc(connectPath, n.connectListener())
	s.HandleFunc(disconnectPath, n.disconnectListener())
	s.HandleFunc(broadcastPath, n.messageListener())

	var host string = fmt.Sprintf("%s:%s", n.address, n.port)
	n.server = &http.Server{Addr: host, Handler: s}
	if e := http.ListenAndServe(host, s); e != nil {
		n.node.log("Error initializing server: %s", e.Error())
		n.node.disconnect <- true
	}
}

// connectEmitter function send connection request to bootnode, waits for node
// response with member list as body, and send the same request to each member.
func (n *network) connectEmitter(p Peer) {
	var (
		e    error
		req  *http.Request
		res  *http.Response
		boot string = fmt.Sprintf(baseUri, p.address, p.port, connectPath)
	)
	if req, e = http.NewRequest(http.MethodGet, boot, nil); e != nil {
		n.node.log("Error sending connect: %s", e.Error())
		n.node.disconnect <- true
		return
	}

	req.Header.Add(peeraddress, n.address)
	req.Header.Add(peerport, n.port)
	if res, e = n.client.Do(req); e != nil {
		n.node.log("Error sending connect: %s", e.Error())
		n.node.disconnect <- true
		return
	}
	defer res.Body.Close()

	var body []byte
	if body, e = ioutil.ReadAll(res.Body); e != nil {
		n.node.log("Error sending connect: %s", e.Error())
		n.node.disconnect <- true
		return
	}

	var rgx *regexp.Regexp = regexp.MustCompile("((.+):(.+))")
	var hosts [][][]byte = rgx.FindAllSubmatch(body, -1)
	for _, host := range hosts {
		var a, p string = string(host[2]), string(host[3])
		n.node.join <- Peer{p, a}

		var (
			req *http.Request
			res *http.Response
			uri string = fmt.Sprintf(baseUri, a, p, connectPath)
		)
		if req, e = http.NewRequest(http.MethodGet, uri, nil); e != nil {
			n.node.log("Error sending connect: %s", e.Error())
		}

		req.Header.Add(peeraddress, n.address)
		req.Header.Add(peerport, n.port)
		if res, e = n.client.Do(req); e != nil {
			n.node.log("Error sending connect: %s", e.Error())
		}
		defer res.Body.Close()

		var body []byte
		if body, e = ioutil.ReadAll(res.Body); e != nil {
			n.node.log("Error sending connect: %s", e.Error())
			n.node.disconnect <- true
			return
		}

		for _, h := range rgx.FindAllSubmatch(body, -1) {
			var a, p string = string(h[2]), string(h[3])
			n.node.join <- Peer{p, a}
		}
	}
	n.node.join <- p
}

// connectListener function return new connection http handler to listen for
// new nodes connection. Create new member with headers info.
func (n *network) connectListener() listener {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			var a, p string = r.Header.Get(peeraddress), r.Header.Get(peerport)
			n.node.join <- Peer{p, a}

			var members []byte
			for _, m := range n.node.Members {
				var member string = fmt.Sprintf("%s:%s\n", m.address, m.port)
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

// disconnectEmitter function emit disconnect request to all current members,
// making http call to disconnectPath.
func (n *network) disconnectEmitter() {
	var e error
	for _, m := range n.node.Members {
		var (
			req *http.Request
			uri string = fmt.Sprintf(baseUri, m.address, m.port, disconnectPath)
		)

		if req, e = http.NewRequest(http.MethodDelete, uri, nil); e != nil {
			n.node.log("Error sending disconnect: %s", e.Error())
		}

		req.Header.Add(peeraddress, n.address)
		req.Header.Add(peerport, n.port)

		if _, e = n.client.Do(req); e != nil {
			n.node.log("Error sending disconnect: %s", e.Error())
		}
	}
}

// disconnectListener function return the disconnect request handler, that send
// emitter peer via leave channel.
func (n *network) disconnectListener() listener {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			var a, p string = r.Header.Get(peeraddress), r.Header.Get(peerport)
			n.node.leave <- Peer{p, a}
		}
	}
}

// messageEmitter function sends byte array as message to all current members,
// making http call to broadcastPath.
func (n *network) messageEmitter(message []byte) {
	var e error
	for _, m := range n.node.Members {
		var (
			req *http.Request
			uri string = fmt.Sprintf(baseUri, m.address, m.port, broadcastPath)
		)

		var body *bytes.Buffer = bytes.NewBuffer(message)
		if req, e = http.NewRequest(http.MethodPost, uri, body); e != nil {
			n.node.log("Error sending message: %s", e.Error())
		}

		req.Header.Add(peeraddress, n.address)
		req.Header.Add(peerport, n.port)

		if _, e = n.client.Do(req); e != nil {
			n.node.log("Error sending message: %s", e.Error())
		}
	}
}

// messageListener function returns message handler function, that listens to
// new messages broadcast and send it to inbox channel.
func (n *network) messageListener() listener {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			defer r.Body.Close()

			if body, e := ioutil.ReadAll(r.Body); e != nil {
				n.node.log("Error receiving message: %s", e.Error())
				n.node.disconnect <- true
			} else {
				n.node.inbox <- body
			}
		}
	}
}
