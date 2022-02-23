package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"time"

	mrand "math/rand"

	"github.com/libp2p/go-libp2p"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	noise "github.com/libp2p/go-libp2p-noise"
	libp2ptls "github.com/libp2p/go-libp2p-tls"
)

func IfThen(condition bool, a interface{}) interface{} {
	if condition {
		return a
	}
	return nil
}

func makeHost(ctx context.Context, listenPort int, randseed int64, peerAddr string) (host.Host, *discovery.RoutingDiscovery) {
	privateKey, _ := generateKeyPair(randseed)

	opts := []libp2p.Option{
		// Use the keypair we generated
		libp2p.Identity(privateKey),
		// Multiple listen addresses
		libp2p.ListenAddrStrings(
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", listenPort),
			fmt.Sprintf("/ip6/::/tcp/%d", listenPort),
			fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic", listenPort),
			fmt.Sprintf("/ip6/::/udp/%d/quic", listenPort),
		),
		// support TLS connections
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		// support noise connections
		libp2p.Security(noise.ID, noise.New),
		// support any other default transports (TCP)
		libp2p.DefaultTransports,
		// Let's prevent our peer from having too many
		// connections by attaching a connection manager.
		libp2p.ConnectionManager(connmgr.NewConnManager(
			100,         // Lowwater
			400,         // HighWater,
			time.Minute, // GracePeriod
		)),
		// Let this host use the DHT to find other hosts
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			idht, err := dht.New(ctx, h)
			return idht, err
		}),
	}

	// Attempt to open ports using uPNP for NATed hosts.
	IfThen(prod, append(opts, libp2p.NATPortMap()))

	// Let this host use relays and advertise itself on relays if
	// it finds it is behind NAT. Use libp2p.Relay(options...) to
	// enable active relays and more.
	IfThen(prod, append(opts, libp2p.EnableAutoRelay()))

	// If you want to help other peers to figure out if they are behind
	// NATs, you can launch the server-side of AutoNAT too (AutoRelay
	// already runs the client)
	//
	// This service is highly rate-limited and should not cause any
	// performance issues.
	IfThen(prod, append(opts, libp2p.EnableNATService()))

	//var idht *dht.IpfsDHT

	ha, err := libp2p.New(opts...)

	if err != nil {
		panic(err)
	}

	if peerAddr != "" {
		logger.Info("Bootstrap multi address=", peerAddr)
		pi, err := peer.AddrInfoFromString(peerAddr)
		if err != nil {
			panic(err)
		}
		ha.Connect(ctx, *pi)
	}

	/*
		logger.Info("Announcing ourselves...")
		routingDiscovery := discovery.NewRoutingDiscovery(idht)
		discovery.Advertise(ctx, routingDiscovery, meetingPoint)
		logger.Info("Successfully announced!")
	*/
	return ha, routingDiscovery
}

func generateKeyPair(randseed int64) (crypto.PrivKey, crypto.PubKey) {
	// If the seed is zero, use real cryptographic randomness. Otherwise, use a
	// deterministic randomness source to make generated keys stay the same
	// across multiple runs
	var r io.Reader
	if randseed == 0 {
		r = rand.Reader
	} else {
		r = mrand.New(mrand.NewSource(randseed))
	}

	// Generate a key pair for this host. We will use it at least
	// to obtain a valid host ID.
	privateKey, publicKey, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		panic(err)
	}
	return privateKey, publicKey
}
