package main

import (
	"github.com/rs/zerolog/log"
	"leader-election/event"
	"net"
	"net/rpc"
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
	Peers    *Peers
	eventBus event.Bus
}

func NewNode(nodeID string) *Node {
	node := &Node{
		ID:       nodeID,
		Addr:     nodeAddressByID[nodeID],
		Peers:    NewPeers(),
		eventBus: event.NewBus(),
	}

	node.eventBus.Subscribe(event.LeaderElected, node.PingLeaderContinuously)

	return node
}

func (node *Node) NewListener() (net.Listener, error) {
	addr, err := net.Listen("tcp", node.Addr)
	return addr, err
}

func (node *Node) ConnectToPeers() {
	for peerID, peerAddr := range nodeAddressByID {
		if node.IsItself(peerID) {
			continue
		}

		rpcClient := node.connect(peerAddr)
		pingMessage := Message{FromPeerID: node.ID, Type: PING}
		reply, _ := node.CommunicateWithPeer(rpcClient, pingMessage)

		if reply.IsPongMessage() {
			log.Debug().Msgf("%s got pong message from %s", node.ID, peerID)
			node.Peers.Add(peerID, rpcClient)
		}
	}
}

func (node *Node) connect(peerAddr string) *rpc.Client {
retry:
	client, err := rpc.Dial("tcp", peerAddr)
	if err != nil {
		log.Debug().Msgf("Error dialing rpc dial %s", err.Error())
		time.Sleep(50 * time.Millisecond)
		goto retry
	}
	return client
}

func (node *Node) CommunicateWithPeer(RPCClient *rpc.Client, args Message) (Message, error) {
	var reply Message

	err := RPCClient.Call("Node.HandleMessage", args, &reply)
	if err != nil {
		log.Debug().Msgf("Error calling HandleMessage %s", err.Error())
	}

	return reply, err
}

func (node *Node) HandleMessage(args Message, reply *Message) error {
	reply.FromPeerID = node.ID

	switch args.Type {
	case ELECTION:
		reply.Type = ALIVE
	case ELECTED:
		leaderID := args.FromPeerID
		log.Info().Msgf("Election is done. %s has a new leader %s", node.ID, leaderID)
		node.eventBus.Emit(event.LeaderElected, leaderID)
		reply.Type = OK
	case PING:
		reply.Type = PONG
	}

	return nil
}

func (node *Node) Elect() {
	isHighestRankedNodeAvailable := false

	peers := node.Peers.ToList()
	for i := range peers {
		peer := peers[i]

		if node.IsRankHigherThan(peer.ID) {
			continue
		}

		log.Debug().Msgf("%s send ELECTION message to peer %s", node.ID, peer.ID)
		electionMessage := Message{FromPeerID: node.ID, Type: ELECTION}

		reply, _ := node.CommunicateWithPeer(peer.RPCClient, electionMessage)

		if reply.IsAliveMessage() {
			isHighestRankedNodeAvailable = true
		}
	}

	if !isHighestRankedNodeAvailable {
		leaderID := node.ID
		electedMessage := Message{FromPeerID: leaderID, Type: ELECTED}
		node.BroadcastMessage(electedMessage)
		log.Info().Msgf("%s is a new leader", node.ID)
	}
}

func (node *Node) BroadcastMessage(args Message) {
	peers := node.Peers.ToList()
	for i := range peers {
		peer := peers[i]
		node.CommunicateWithPeer(peer.RPCClient, args)
	}
}

func (node *Node) PingLeaderContinuously(_ string, payload any) {
	leaderID := payload.(string)

ping:
	leader := node.Peers.Get(leaderID)
	if leader == nil {
		log.Error().Msgf("%s, %s, %s", node.ID, leaderID, node.Peers.ToIDs())
		return
	}

	pingMessage := Message{FromPeerID: node.ID, Type: PING}
	reply, err := node.CommunicateWithPeer(leader.RPCClient, pingMessage)
	if err != nil {
		log.Info().Msgf("Leader is down, new election about to start!")
		node.Peers.Delete(leaderID)
		node.Elect()
		return
	}

	if reply.IsPongMessage() {
		log.Debug().Msgf("Leader %s sent PONG message", reply.FromPeerID)
		time.Sleep(3 * time.Second)
		goto ping
	}
}

func (node *Node) IsRankHigherThan(id string) bool {
	return strings.Compare(node.ID, id) == 1
}

func (node *Node) IsItself(id string) bool {
	return node.ID == id
}
