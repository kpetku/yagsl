package session

import (
	"log"
	"strings"
	"sync"
)

type Session struct {
	Style       string
	ID          string
	Destination string
	Options     SessionOptions
}

type SessionOptions struct {
	Port          string
	Host          string
	FromPort      string
	ToPort        string
	ListenPort    string
	ListenProtcol string
	Header        bool
	// i2cp and streaming options go here
}

type sessionHandler struct {
	incoming chan reply
	m        sync.Mutex
}

type reply struct {
	full string

	name  string
	value string
	err   error
}

func (m *Session) SessionCreate() {
}

func (s Session) sessionReply(reply string) error {
	fields := strings.Fields(reply)
	if strings.HasPrefix(fields[1], "RESULT=") {
		log.Printf("Session status result: %s", fields[1:])
	} else {
		log.Printf("Got SESSION reply: %s", reply)
	}
	return nil
}
