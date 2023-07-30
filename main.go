package main

import (
	"errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"strconv"
)

func main() {
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

	nodeID, err := getNodeID()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	node := NewNode(nodeID)
	go node.Listen()

	node.Elect()

	/*go func() {
		time.Sleep(5 * time.Second)
		for true {
			time.Sleep(3 * time.Second)
			log.Info().Msgf("I'm %s and my leader is %s", node.ID, node.LeaderID)
		}
	}()*/

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

func getNodeID() (string, error) {
	if len(os.Args) < 2 {
		return "", errors.New("node id required")
	}

	nodeID := os.Args[1]
	return nodeID, nil
}
