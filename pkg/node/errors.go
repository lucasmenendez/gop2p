package node

import (
	"fmt"
)

const (
	CONNECTION_ERR = iota
	PARSING_ERR    = iota
	INTERNAL_ERR   = iota
)

type NodeErr struct {
	ErrCode int
	Text    string
	Trace   error
}

func (err *NodeErr) Error() string {
	tag := "internal error"
	if err.ErrCode == CONNECTION_ERR {
		tag = "connection error"
	} else if err.ErrCode == PARSING_ERR {
		tag = "parsing error"
	}

	text := err.Text

	if err.Trace != nil {
		text = fmt.Sprintf("%s: %v", text, err.Trace)
	}

	return fmt.Sprintf("%s: %s", tag, text)
}

func ConnErr(text string, err error) *NodeErr {
	return &NodeErr{CONNECTION_ERR, text, err}
}

func ParseErr(text string, err error) *NodeErr {
	return &NodeErr{PARSING_ERR, text, err}
}

func InternalErr(text string, err error) *NodeErr {
	return &NodeErr{INTERNAL_ERR, text, err}
}
