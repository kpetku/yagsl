package sambridge

import (
	"errors"
	"strings"
)

type SessionReply struct {
	Destination string
	Err         error
	ID          string
	Message     string
}

const (
	duplicatedDest = "DUPLICATED_DEST"
	duplicatedID   = "DUPLICATED_ID"
)

var (
	errHandshakeFailed = errors.New("Handshake failed")
	errInvalidKey      = errors.New("invalid key error")
	errDuplicatedDest  = errors.New("the destination is not a valid private destination key")
	errDuplicatedID    = errors.New("the destination is already in use")

	errUnknownSession = errors.New("Unknown session error")
)

func (s SAMBridge) sessionReply(line string) {
	reply := SessionReply{}
	fields := strings.Fields(line)

	switch strings.SplitN(fields[2], "=", 2)[1] { // result
	case resultOk:
		reply.Err = nil
		if strings.HasPrefix(fields[3], "DESTINATION") {
			reply.Destination = strings.SplitN(fields[3], "=", 2)[1]
		}
		if strings.HasPrefix(fields[3], "ID") {
			reply.ID = strings.SplitN(fields[3], "=", 2)[1]
		}
		if len(fields) > 4 {
			if strings.HasPrefix(fields[4], "MESSAGE") {
				reply.Message = strings.SplitN(fields[4], "=", 2)[1]
			}
		}
	case duplicatedID:
		reply.Err = errDuplicatedID
	case duplicatedDest:
		reply.Err = errDuplicatedDest
	case invalidKey:
		reply.Err = errInvalidKey
	default:
		reply.Err = errUnknownSession
	}
	s.SessionReply <- reply
}
