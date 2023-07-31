package main

const (
	ConnectionBufferSize = 128
)

type Message struct {
	FromPeerID string
	Type       MessageType
}

func (m *Message) IsAliveMessage() bool {
	return m.Type == ALIVE
}

func (m *Message) IsPongMessage() bool {
	return m.Type == PONG
}

type MessageType uint32

const (
	PING MessageType = iota + 1
	PONG
	ELECTION
	ALIVE
	ELECTED
	OK = 6
)

func FromValue(val uint32) MessageType {
	switch val {
	case ELECTION.ToValue():
		return ELECTION
	case ALIVE.ToValue():
		return ALIVE
	case ELECTED.ToValue():
		return ELECTED
	}
	panic("Unknown Value")
}

func (m MessageType) ToValue() uint32 {
	return uint32(m)
}
