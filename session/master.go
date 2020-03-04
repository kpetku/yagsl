package session

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/kpetku/yagsl/sambridge"
)

const masterSessionStyle string = "MASTER"

var errMasterSessionFailed = errors.New("Failed to start MasterSession")

type MasterSession struct {
	Style       string
	ID          string
	Destination string
	Transient   bool

	Sam *sambridge.SAMBridge
	*Session

	Ready chan bool

	Subsessions []Subsession

	// i2cp and streaming options go here
}

func NewMasterSession(hostAndPort string, destination string) (MasterSession, error) {
	sb := new(sambridge.SAMBridge)
	sb.Options.Address = hostAndPort
	go sb.Start(hostAndPort)
	ms := MasterSession{Style: "MASTER"}
	ms.Ready = make(chan bool, 1)
	if destination == "" {
		destination = "TRANSIENT"
		ms.Transient = true
	}
	ms.Sam = sb
	ms.Sam.SessionReply = make(chan sambridge.SessionReply)
	go ms.newSessionHandler()

	ms.Sam.StreamReply = make(chan sambridge.StreamReply)
	go ms.newStreamHandler()

	ms.ID = "master" + strconv.FormatInt(time.Now().UTC().Unix(), 10)
	ms.Destination = destination

	time.Sleep(time.Second * 1)
	ms.Sam.Send(fmt.Sprintf("SESSION CREATE STYLE=MASTER ID=%s DESTINATION=%s\n", ms.ID, ms.Destination))
	for ready := range ms.Ready {
		if ready {
			return ms, nil
		}
		return ms, errMasterSessionFailed
	}
	return ms, nil
}

func (ms *MasterSession) NewSubsession(fromPort string, toPort string, listenPort string) Subsession {
	ss := Subsession{}

	ss.ID = "sub" + strconv.FormatInt(time.Now().UTC().Unix(), 10)
	if listenPort != "" {
		ms.Sam.Send(fmt.Sprintf("SESSION ADD STYLE=STREAM ID=%s FROM_PORT=%s LISTEN_PORT=%s SILENT=FALSE\n", ss.ID, fromPort, listenPort))
	} else {
		ms.Sam.Send(fmt.Sprintf("SESSION ADD STYLE=STREAM ID=%s FROM_PORT=%s SILENT=FALSE\n", ss.ID, fromPort))
	}

	ms.Subsessions = append(ms.Subsessions, ss)

	return ss
}

func (ms *MasterSession) newSessionHandler() {
	var sessionReply sambridge.SessionReply
	for reply := range ms.Sam.SessionReply {
		sessionReply = reply
		break
	}
	if sessionReply.Err != nil {
		ms.Ready <- false
	}
	ms.Destination = sessionReply.Destination
	ms.Ready <- true
}

func (ms *MasterSession) newStreamHandler() {
	var streamReply sambridge.StreamReply
	for reply := range ms.Sam.StreamReply {
		streamReply = reply
		break
	}
	if streamReply.Err != nil {
		log.Printf("StreamReply err: %s", streamReply.Err.Error())
	} else {
		log.Printf("StreamReply success!")
	}
}
