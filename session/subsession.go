package session

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/kpetku/yagsl/sambridge"
)

type Subsession struct {
	*Session
	ID    string
	Style string
	Sam   *sambridge.SAMBridge
	Ready chan (bool)
}

var (
	errSubSessionFailed = errors.New("the subsession failed :(")
)

func NewStreamConnect(ID string, destination string, fromPort string) (Subsession, error) {
	sb := new(sambridge.SAMBridge)
	go sb.Start()
	ss := Subsession{
		Style: "STREAM",
		ID:    ID,
	}
	ss.Sam = sb
	ss.Sam.SessionReply = make(chan sambridge.SessionReply)
	go ss.newSessionHandler()
	// TODO: Replace this handshake pause with a Ready callback to prevent a data race
	time.Sleep(time.Second * 1)
	// TODO: Refactor this sam.Send into a stream package w/a streamHandler and it's own connect message in that section
	err := ss.Sam.Send(fmt.Sprintf("STREAM CONNECT ID=%s DESTINATION=%s FROM_PORT=%s TO_PORT=0 SILENT=FALSE\n", ID, destination, fromPort))
	return ss, err
}

func NewStreamAccept(ID string) (Subsession, error) {
	sb := new(sambridge.SAMBridge)
	go sb.Start()
	ss := Subsession{
		Style: "STREAM",
		ID:    ID,
	}
	ss.Sam = sb
	ss.Sam.SessionReply = make(chan sambridge.SessionReply)
	go ss.newSessionHandler()
	// TODO: Replace this handshake pause with a Ready callback to prevent a data race
	time.Sleep(time.Second * 1)
	// TODO: Refactor this sam.Send into a stream package w/a streamHandler and it's own connect message in that section
	err := ss.Sam.Send(fmt.Sprintf("STREAM ACCEPT ID=%s SILENT=FALSE\n", ID))
	return ss, err
}

func (ss *Subsession) newSessionHandler() {
	var sessionReply sambridge.SessionReply
	for reply := range ss.Sam.SessionReply {
		sessionReply = reply
		break
	}
	if sessionReply.Err != nil {
		ss.Ready <- false
	}
	ss.Destination = sessionReply.Destination
	ss.Ready <- true
}

func (ss *Subsession) Stop() error {
	ss.Sam.Send(fmt.Sprintf("SESSION REMOVE ID=%s", ss.ID))
	return nil
}

func (s *Subsession) Dial(network, addr string) (net.Conn, error) {
	return s.Sam.Conn, nil
}
