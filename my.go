package main

import (
	"context"

	//"github.com/libp2p/go-libp2p"
	//"github.com/libp2p/go-libp2p-core/host"

	"github.com/libp2p/go-libp2p-core/protocol"
	//"github.com/libp2p/go-libp2p-core/routing"

	"github.com/ipfs/go-log/v2"
)

//var logger = log.Logger("rendezvous")
var meetingPoint = "kmrn"
var protocolID = "/chat/1.1.0"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.SetAllLoggers(log.LevelInfo)
	log.SetLogLevel("rendezvous", "info")

	h, routingDiscovery, _ := makeHost(ctx, 63636)
	logger.Info(h.Addrs())
	logger.Info(h.ID())

	logger.Info("Searching for other peers...")
	peerChan, _ := routingDiscovery.FindPeers(ctx, meetingPoint)

	h.SetStreamHandler(protocol.ID(protocolID), handleStream)

	for peer := range peerChan {
		if peer.ID == h.ID() {
			continue
		}
		logger.Info("Found peer:", peer.Addrs)

		stream, err := h.NewStream(ctx, peer.ID, protocol.ID(protocolID))
		if err != nil {
			logger.Warning("Connection failed:", err)
			continue
		} else {
			handleStream(stream)
		}
	}

	select {}
}
