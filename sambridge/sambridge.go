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

func (s *SAMBridge) Start(hostAndPort string) {
	s.Options.Address = hostAndPort
	err := s.dialSAMBridge()
	if err != nil {
		panic(err.Error())
	}
	scanner := bufio.NewScanner(s.Conn)
	for {
		if ok := scanner.Scan(); !ok {
			break
		}
		log.Printf("DEBUG line: %s", scanner.Text())
		if strings.Contains(scanner.Text(), " ") {
			first := strings.Fields(scanner.Text())[0]
			switch first {
			case "HELLO":
				go s.handshakeReply(scanner.Text())
			case "SESSION":
				s.sessionReply(scanner.Text())
			case "STREAM":
				// TODO: Wait for an "OK" STREAM result and as soon as that happens break out of the loop so we can panic on unknown sam lines
				s.streamReply(scanner.Text())
			case "DEST":
				s.destReply(scanner.Text())
			case "NAMING":
				s.namingReply(scanner.Text())
			case "PING":
				go s.pingReply(scanner.Text())
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
