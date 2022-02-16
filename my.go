package main

import (
	"context"
	"sync"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/libp2p/go-libp2p-core/routing"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"

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

	var kdht *dht.IpfsDHT
	var err error

	h, err := libp2p.New(libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
		kdht, err = dht.New(ctx, h)
		return kdht, err
	}),
		libp2p.EnableAutoRelay(),
	)

	logger.Info(h.Addrs())
	logger.Info(h.ID())

	if err = kdht.Bootstrap(ctx); err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	for _, addr := range dht.DefaultBootstrapPeers {
		pi, _ := peer.AddrInfoFromP2pAddr(addr)
		// We ignore errors as some bootstrap peers may be down
		// and that is fine.
		wg.Add(1)
		h.Connect(ctx, *pi)
		wg.Done()
	}
	wg.Wait()

	logger.Info("Announcing ourselves...")
	routingDiscovery := discovery.NewRoutingDiscovery(kdht)
	discovery.Advertise(ctx, routingDiscovery, meetingPoint)
	logger.Info("Successfully announced!")

	logger.Info("Searching for other peers...")
	peerChan, err := routingDiscovery.FindPeers(ctx, meetingPoint)

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
