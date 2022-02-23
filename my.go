package main

import (
	"context"
	"flag"

	"github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	discovery "github.com/libp2p/go-libp2p-discovery"
)

//var logger = log.Logger("rendezvous")
var meetingPoint = "kmrn"
var protocolID = "/chat/1.1.0"
var routingDiscovery *discovery.RoutingDiscovery
var h host.Host
var logger = log.Logger("mychat")
var prod = false

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.SetAllLoggers(log.LevelInfo)
	log.SetLogLevel("rendezvous", "info")

	// Parse options from the command line
	p2pPort := flag.Int("l", -1, "P2P port")
	seed := flag.Int64("s", -1, "Random seed")
	webPort := flag.String("w", "", "Web port")
	peerAddr := flag.String("b", "", "Bootstrap peer address")
	flag.Parse()

	if *p2pPort == -1 {
		logger.Fatal("Please provide p2p port to bind on with -l")
	}
	if *seed == -1 {
		logger.Fatal("Please provide randon seed with -s")
	}
	if *webPort == "" {
		logger.Fatal("Please provide web port to bind on with -s")
	}
	/*
		if *peerAddr == "0" {
			logger.Fatal("Please provide bootstrap peers with -b")
		}
	*/
	var err error
	h, routingDiscovery = makeHost(ctx, *p2pPort, *seed, *peerAddr)
	if err != nil {
		panic(err)
	}

	logger.Info("My Host Id = ", h.ID())
	logger.Info("My Host Address", h.Addrs())

	web(*webPort)
}

func findMembers(ctx context.Context) {
	var members = make(map[string]peer.AddrInfo)
	logger.Info("Searching for other peers...")
	peerChan, _ := routingDiscovery.FindPeers(ctx, meetingPoint)

	for p := range peerChan {
		if p.ID == h.ID() {
			continue
		}
		logger.Info("Found peer:", p.ID)

		if len(p.Addrs) > 0 {
			members[p.ID.String()] = p
		}
	}
}
