package sambridge

import (
	"fmt"
	"strings"
	"sync"
)

type DestReply struct {
	Pub  string
	Priv string
	err  error
}

type destHandler struct {
	incoming chan DestReply
	m        sync.Mutex
}

func (m *SAMBridge) DestGenerate(signatureType string) (DestReply, error) {
	var destReply DestReply
	dh := new(destHandler)
	dh.incoming = make(chan DestReply)
	m.destHandler = dh

	if signatureType == "" {
		signatureType = "DSA_SHA1"
	}
	m.Send(fmt.Sprintf("DEST GENERATE SIGNATURE_TYPE=%s", signatureType))
	for reply := range m.destHandler.incoming {
		destReply = reply
		break
	}
	if destReply.err != nil {
		return destReply, destReply.err
	}
	return destReply, nil
}

func (m SAMBridge) destReply(reply string) {
	dest := DestReply{}
	fields := strings.Fields(reply)
	dest.Pub = strings.TrimPrefix(fields[2], "PUB=")
	dest.Priv = strings.TrimPrefix(fields[3], "PRIV=")
	dest.err = nil
	m.destHandler.incoming <- dest
}

func (s SAMBridge) pingReply(reply string) error {
	s.Send(fmt.Sprintf("PONG %s", strings.TrimLeft(reply, "PING ")))
	return nil
}

func (m *SAMBridge) newDestHandler() {
	if m.destHandler == nil {
		dh := destHandler{}
		dh.incoming = make(chan DestReply)
		m.destHandler = &dh
	}
}
