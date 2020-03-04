package sambridge

import (
	"bufio"
	"errors"
	"log"
	"net"
	"strings"
)

type SAMBridge struct {
	Options       SamOptions
	Conn          net.Conn
	namingHandler *namingHandler
	destHandler   *destHandler

	SessionReply chan SessionReply
	StreamReply  chan StreamReply
}

type SamOptions struct {
	Address  string
	User     string
	Password string
}

const handshakeReply string = "HELLO VERSION MIN=3.0 MAX=3.3\n"

func (s *SAMBridge) Start() {
	s.Options.Address = "127.0.0.1:7656"
	err := s.dialSAMBridge()
	if err != nil {
		panic(err.Error())
	}
	r := bufio.NewReader(s.Conn)
	for {
		line, _ := r.ReadString('\n')
		if line != "" {
			log.Printf("DEBUG line: %s", line)
		}
		if strings.Contains(line, " ") {
			first := strings.Fields(line)[0]
			switch first {
			case "HELLO":
				go s.handshakeReply(line)
			case "SESSION":
				s.sessionReply(line)
			case "STREAM":
				// TODO: Wait for an "OK" STREAM result and as soon as that happens break out of the loop so we can panic on unknown sam lines
				s.streamReply(line)
			case "DEST":
				s.destReply(line)
			case "NAMING":
				s.namingReply(line)
			case "PING":
				go s.pingReply(line)
			default:
				//panic("Unknown SAM line encountered")
			}
		}
	}
}

func (s *SAMBridge) dialSAMBridge() error {
	c, err := net.Dial("tcp", s.Options.Address)
	s.Conn = c
	s.sendHandshake()
	return err
}

func (s SAMBridge) Send(line string) error {
	if s.Conn != nil {
		var sb strings.Builder
		sb.WriteString(line)
		if !strings.HasSuffix(line, "\n") {
			sb.WriteString("\n")
		}
		w := bufio.NewWriter(s.Conn)
		_, err := w.WriteString(sb.String())

		err = w.Flush()
		return err
	}
	return errors.New("Conn was closed")
}

func (s SAMBridge) sendHandshake() error {
	return s.Send(handshakeReply)
}

func (s SAMBridge) handshakeReply(reply string) error {
	if !strings.HasPrefix(reply, "HELLO REPLY RESULT=OK") {
		return errHandshakeFailed
	}
	return nil
}
