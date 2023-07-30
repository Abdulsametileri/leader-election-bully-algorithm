package main

const (
	ConnectionBufferSize = 128
)

type MessageType uint32

const (
	ELECTION MessageType = iota + 1
	ALIVE
	ELECTED
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
