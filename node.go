package main

import (
	"encoding/binary"
	"github.com/rs/zerolog/log"
	"io"
	"net"
	"strings"
	"time"
)

// nodeAddressByID: It includes nodes currently in cluster
var nodeAddressByID = map[string]string{
	"node-01": "node-01:6001",
	"node-02": "node-02:6002",
	"node-03": "node-03:6003",
	"node-04": "node-04:6004",
}

type Node struct {
	ID       string
	Addr     string
	LeaderID string
	PeerList *Peers
	aliveCh  chan string
}

func NewNode(nodeID string) *Node {
	return &Node{
		ID:       nodeID,
		Addr:     nodeAddressByID[nodeID],
		PeerList: NewPeerList(),
		aliveCh:  make(chan string),
	}
}

func (node *Node) connect(peerID string) error {
	writeSock, err := net.Dial("tcp", nodeAddressByID[peerID])
	if err != nil {
		return err
	}

	log.Debug().Msgf("Node %s is connected to peer %s", node.ID, peerID)
	node.PeerList.Add(peerID, writeSock)
	return nil
}

func (node *Node) Listen() {
	addr, err := net.Listen("tcp", node.Addr)
	if err != nil {
		log.Fatal().Err(err)
	}

	for {
		conn, err := addr.Accept()
		if err != nil {
			log.Error().Msgf("inbound connection error: %v", err)
			continue
		}

		go node.receive(conn)
	}
}

func (node *Node) receive(conn io.ReadCloser) {
	for {
		buf := make([]byte, ConnectionBufferSize)
		_, err := conn.Read(buf)
		if err != nil {
			log.Error().Err(err)
			conn.Close()
			break
		}

		mType := binary.LittleEndian.Uint32(buf[0:])
		mLen := binary.LittleEndian.Uint32(buf[4:])
		fromNode := string(buf[8 : 8+mLen])

		log.Debug().Msgf("Message comes from %s, len %d in type %d", fromNode, mLen, mType)

		switch FromValue(mType) {
		case ELECTION:
			node.Send(fromNode, ALIVE)
		case ALIVE:
			node.aliveCh <- fromNode
		case ELECTED:
			log.Info().Msgf("%s has new leader and it's id is %s", node.ID, fromNode)
			node.SetLeader(fromNode)
		}
	}
}

func (node *Node) Elect() {
	peers := node.PeerList.GetAll()

	for i := range peers {
		peer := &peers[i]

		if node.IsHigherThan(peer.ID) {
			continue
		}

		node.Send(peer.ID, ELECTION)
	}

	select {
	case aliveNode := <-node.aliveCh:
		log.Debug().Msgf("ALIVE message has came to %s from %s", node.ID, aliveNode)
	case <-time.After(3 * time.Second):
		log.Info().Msgf("%s is making itself a leader", node.ID)
		node.SetLeader(node.ID)
		node.Broadcast(ELECTED)
	}
}

func (node *Node) Broadcast(mType MessageType) {
	peers := node.PeerList.GetAll()
	for i := range peers {
		peer := &peers[i]
		node.Send(peer.ID, mType)
	}
}

func (node *Node) Send(peer string, mType MessageType) {
	if !node.PeerList.Find(peer) {
		node.connect(peer)
	}

	buf := make([]byte, ConnectionBufferSize)
	binary.LittleEndian.PutUint32(buf[0:], mType.ToValue())
	binary.LittleEndian.PutUint32(buf[4:], uint32(len(node.ID)))
	copy(buf[8:], node.ID)

	node.PeerList.Get(peer).WriteSock.Write(buf)
}

func (node *Node) IsHigherThan(id string) bool {
	return strings.Compare(node.ID, id) == 1
}

func (node *Node) SetLeader(leaderID string) {
	node.LeaderID = leaderID
}
