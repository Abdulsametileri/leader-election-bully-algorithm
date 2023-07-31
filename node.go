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
	LeaderID string
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
		log.Error().Msgf("Error dialing rpc dial %s", err.Error())
		time.Sleep(50 * time.Millisecond)
		goto retry
	}
	return client
}

func (node *Node) CommunicateWithPeer(RPCClient *rpc.Client, args Message) (Message, error) {
	var reply Message

	err := RPCClient.Call("Node.HandleMessage", args, &reply)
	if err != nil {
		log.Error().Msgf("Error calling HandleMessage %s", err.Error())
	}

	return reply, err
}

func (node *Node) HandleMessage(args Message, reply *Message) error {
	reply.FromPeerID = node.ID

	if args.Type == ELECTION {
		reply.Type = ALIVE
	} else if args.Type == ELECTED {
		log.Debug().Msgf("%s peers %s", node.ID, node.Peers.ToIDs())
		node.SetLeader(args.FromPeerID)
		log.Info().Msgf("Election is done. %s has a new leader %s", node.ID, args.FromPeerID)
		reply.Type = OK
		node.eventBus.Emit(event.LeaderElected, args.FromPeerID)
	} else if args.Type == PING {
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
		electedMessage := Message{FromPeerID: node.ID, Type: ELECTED}
		node.BroadcastMessage(electedMessage)
		node.SetLeader(node.ID)
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

func (node *Node) PingLeaderContinuously(eventName string, payload any) {
ping:
	leader := node.Peers.Get(node.LeaderID)
	pingMessage := Message{FromPeerID: node.ID, Type: PING}
	reply, err := node.CommunicateWithPeer(leader.RPCClient, pingMessage)
	if err != nil {
		log.Info().Msgf("Leader is down, new election about to start!")
		node.Peers.Delete(node.LeaderID)
		node.LeaderID = ""
		node.Elect()
		return
	}

	if reply.IsPongMessage() {
		log.Info().Msgf("Leader %s sent PONG message", reply.FromPeerID)
		time.Sleep(3 * time.Second)
		goto ping
	}
}

func (node *Node) IsRankHigherThan(id string) bool {
	return strings.Compare(node.ID, id) == 1
}

func (node *Node) SetLeader(leaderID string) {
	node.LeaderID = leaderID
}

func (node *Node) IsItself(id string) bool {
	return node.ID == id
}
