package sambridge

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

const (
	resultOk    string = "OK"
	invalidKey  string = "INVALID_KEY"
	keyNotFound string = "KEY_NOT_FOUND"
)

var (
	errKeyNotFound         = errors.New("Key not found")
	errUnknownNamingLookup = errors.New("Unknown naming lookup error")
)

type reply struct {
	name  string
	value string
	err   error
}

type namingHandler struct {
	incoming chan reply
	m        sync.Mutex
}

func (m *SAMBridge) Lookup(name string) (string, error) {
	var lookupReply reply
	m.newNamingHandler()

	go m.Send(fmt.Sprintf("NAMING LOOKUP NAME=%s\n", name))
	for reply := range m.namingHandler.incoming {
		lookupReply = reply
		break
	}
	if lookupReply.err != nil {
		return name, lookupReply.err
	}
	return lookupReply.value, nil
}

func (m *SAMBridge) namingReply(line string) {
	reply := reply{}

	fields := strings.Fields(line)
	m.namingHandler.m.Lock()
	reply.name = strings.SplitN(fields[3], "=", 2)[1]
	switch strings.SplitN(fields[2], "=", 2)[1] { // result
	case resultOk:
		reply.err = nil
		reply.value = strings.SplitN(fields[4], "=", 2)[1]
	case invalidKey:
		reply.err = errInvalidKey
	case keyNotFound:
		reply.err = errKeyNotFound
	default:
		reply.err = errUnknownNamingLookup
	}
	m.namingHandler.incoming <- reply
	m.namingHandler.m.Unlock()
}

func (m *SAMBridge) newNamingHandler() {
	if m.namingHandler == nil {
		t := new(namingHandler)
		t.incoming = make(chan reply)
		m.namingHandler = t
	}
}
