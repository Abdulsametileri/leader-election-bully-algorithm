package main

import (
	"errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/rpc"
	"os"
	"os/signal"
	"strconv"
	"time"
)

func main() {
	setLogConfigurations()

	nodeID, err := getNodeID()
	if err != nil {
		log.Fatal().Err(err)
	}

	node := NewNode(nodeID)

	listener, err := node.NewListener()
	if err != nil {
		log.Fatal().Err(err)
	}
	defer listener.Close()

	rpcServer := rpc.NewServer()
	rpcServer.Register(node)

	// FIXME: Accept bitmeden İstek atma durumları var
	go rpcServer.Accept(listener)

	node.ConnectToPeers()

	log.Info().Msgf("%s is aware of own peers %s", node.ID, node.Peers.ToIDs())

	warmupTime := 5 * time.Second
	time.Sleep(warmupTime)
	node.Elect()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

func setLogConfigurations() {
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
		return file + ":" + strconv.Itoa(line)
	}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.With().Caller().Logger()
}

func getNodeID() (string, error) {
	if len(os.Args) < 2 {
		return "", errors.New("node id required")
	}

	nodeID := os.Args[1]
	return nodeID, nil
}
