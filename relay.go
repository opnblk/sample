package main

import (
	"context"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"

	relayv1 "github.com/libp2p/go-libp2p/p2p/protocol/circuitv1/relay"

	"github.com/ipfs/go-log/v2"
	swarm "github.com/libp2p/go-libp2p-swarm"
	ma "github.com/multiformats/go-multiaddr"
)

//var logger = log.Logger("rendezvous")

func main4() {
	log.SetAllLoggers(log.LevelInfo)
	log.SetLogLevel("rendezvous", "info")
	run()
}

func run() {

	// Create three libp2p hosts, enable relay client capabilities on all
	// of them.

	// Tell the host use relays
	h1, err := libp2p.New(libp2p.EnableRelay())
	if err != nil {
		logger.Errorf("Failed to create h1: %v", err)
		return
	}

	// Tell the host to relay connections for other peers (The ability to *use*
	// a relay vs the ability to *be* a relay)
	h2, err := libp2p.New(libp2p.DisableRelay())
	if err != nil {
		logger.Errorf("Failed to create h2: %v", err)
		return
	}
	_, err = relayv1.NewRelay(h2)
	if err != nil {
		logger.Errorf("Failed to instantiate h2 relay: %v", err)
		return
	}

	// Zero out the listen addresses for the host, so it can only communicate
	// via p2p-circuit for our example
	h3, err := libp2p.New(libp2p.ListenAddrs(), libp2p.EnableRelay())
	if err != nil {
		logger.Errorf("Failed to create h3: %v", err)
		return
	}

	h2info := peer.AddrInfo{
		ID:    h2.ID(),
		Addrs: h2.Addrs(),
	}

	// Connect both h1 and h3 to h2, but not to each other
	if err := h1.Connect(context.Background(), h2info); err != nil {
		logger.Errorf("Failed to connect h1 and h2: %v", err)
		return
	}
	if err := h3.Connect(context.Background(), h2info); err != nil {
		logger.Errorf("Failed to connect h3 and h2: %v", err)
		return
	}

	// Now, to test things, let's set up a protocol handler on h3
	h3.SetStreamHandler("/cats", func(s network.Stream) {
		logger.Errorf("Meow! It worked!")
		s.Close()
	})

	_, err = h1.NewStream(context.Background(), h3.ID(), "/cats")
	if err == nil {
		logger.Errorf("Didnt actually expect to get a stream here. What happened?")
		return
	}
	logger.Errorf("Okay, no connection from h1 to h3: %v", err)
	logger.Errorf("Just as we suspected")

	// Creates a relay address to h3 using h2 as the relay
	relayaddr, err := ma.NewMultiaddr("/p2p/" + h2.ID().Pretty() + "/p2p-circuit/ipfs/" + h3.ID().Pretty())
	if err != nil {
		logger.Errorf(err.Error())
		return
	}
	logger.Info("Relay address = ", relayaddr)
	// Since we just tried and failed to dial, the dialer system will, by default
	// prevent us from redialing again so quickly. Since we know what we're doing, we
	// can use this ugly hack (it's on our TODO list to make it a little cleaner)
	// to tell the dialer "no, its okay, let's try this again"
	h1.Network().(*swarm.Swarm).Backoff().Clear(h3.ID())

	h3relayInfo := peer.AddrInfo{
		ID:    h3.ID(),
		Addrs: []ma.Multiaddr{relayaddr},
	}
	if err := h1.Connect(context.Background(), h3relayInfo); err != nil {
		logger.Errorf("Failed to connect h1 and h3: %v", err)
		return
	}

	// Woohoo! we're connected!
	s, err := h1.NewStream(context.Background(), h3.ID(), "/cats")
	if err != nil {
		logger.Errorf("huh, this should have worked: ", err)
		return
	}

	s.Read(make([]byte, 1)) // block until the handler closes the stream
}
