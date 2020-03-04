package sambridge

import (
	"errors"
	"log"
	"strings"
)

type StreamReply struct {
	Err     error
	Message string
}

const (
	invalidID     = "INVALID_ID"
	i2pErr        = "I2P_ERROR"
	cantReachPeer = "CANT_REACH_PEER"
	timeout       = "TIMEOUT"
)

var (
	errInvalidID     = errors.New("invalid ID")
	errI2PErr        = errors.New("generic I2P error")
	errCantReachPeer = errors.New("can't reach peer")
	errTimeout       = errors.New("timeout")
	errUnknownStream = errors.New("unknown stream error")
)

func (s SAMBridge) streamReply(line string) {
	log.Printf("DEBUG STREAMREPLY: %s", line)
	reply := StreamReply{}
	fields := strings.Fields(line)

	switch strings.SplitN(fields[2], "=", 2)[1] { // result
	case resultOk:
		reply.Err = nil
	case invalidID:
		reply.Err = errInvalidID
	case i2pErr:
		reply.Err = errI2PErr
	case cantReachPeer:
		reply.Err = errCantReachPeer
	case invalidKey:
		reply.Err = errInvalidKey
	case timeout:
		reply.Err = errTimeout
	default:
		reply.Err = errUnknownStream
	}

	if len(fields) > 3 {
		if strings.HasPrefix(fields[3], "MESSAGE") {
			reply.Message = strings.SplitN(fields[3], "=", 2)[1]
		}
	}

	s.StreamReply <- reply
}
