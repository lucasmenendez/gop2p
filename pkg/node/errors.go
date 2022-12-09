package node

import (
	"fmt"

	"github.com/lucasmenendez/gop2p/pkg/message"
)

const (
	CONNECTION_ERR = iota
	PARSING_ERR    = iota
)

type NodeErr struct {
	ErrCode int
	Text    string
	Trace   error
	Message *message.Message
}

func (err *NodeErr) Error() string {
	var tag = "internal error"
	if err.ErrCode == CONNECTION_ERR {
		tag = "connection error"
	} else if err.ErrCode == PARSING_ERR {
		tag = "parsing error"
	}

	var text = err.Text
	if err.Message != nil {
		text = fmt.Sprintf("%s (msg %s)", text, err.Message.String())
	}

	if err.Trace != nil {
		text = fmt.Sprintf("%s: %v", text, err.Trace)
	}

	return fmt.Sprintf("%s: %s\n", tag, text)
}

func ConnErr(text string, err error, msg *message.Message) *NodeErr {
	return &NodeErr{CONNECTION_ERR, text, err, msg}
}

func ParseErr(text string, err error, msg *message.Message) *NodeErr {
	return &NodeErr{PARSING_ERR, text, err, msg}
}
