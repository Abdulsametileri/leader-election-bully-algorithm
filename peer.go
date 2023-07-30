package main

import (
	"io"
	"sync"
)

type Peer struct {
	ID        string
	WriteSock io.WriteCloser
}

type Peers struct {
	*sync.RWMutex
	peerByID map[string]*Peer
}

func NewPeerList() *Peers {
	return &Peers{
		RWMutex:  &sync.RWMutex{},
		peerByID: make(map[string]*Peer),
	}
}

func (p *Peers) Add(ID string, conn io.WriteCloser) {
	p.Lock()
	defer p.Unlock()

	p.peerByID[ID] = &Peer{ID: ID, WriteSock: conn}
}

func (p *Peers) Delete(ID string) {
	p.Lock()
	defer p.Unlock()

	delete(p.peerByID, ID)
}

func (p *Peers) Find(ID string) bool {
	p.RLock()
	defer p.RUnlock()

	_, ok := p.peerByID[ID]
	return ok
}

func (p *Peers) Get(ID string) *Peer {
	p.RLock()
	defer p.RUnlock()

	val := p.peerByID[ID]
	return val
}

func (p *Peers) GetAll() []Peer {
	p.RLock()
	defer p.RUnlock()

	peers := make([]Peer, 0, len(p.peerByID))
	for _, peer := range p.peerByID {
		peers = append(peers, *peer)
	}

	return peers
}
