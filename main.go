package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"

	golog "github.com/ipfs/go-log/v2"
	ma "github.com/multiformats/go-multiaddr"
)

func main1() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// LibP2P code uses golog to log messages. They log with different
	// string IDs (i.e. "swarm"). We can control the verbosity level for
	// all loggers with:
	golog.SetAllLoggers(golog.LevelInfo) // Change to INFO for extra info

	// Parse options from the command line
	listenF := flag.Int("l", 0, "wait for incoming connections")
	targetF := flag.String("d", "", "target peer to dial")
	flag.Parse()

	if *listenF == 0 {
		log.Fatal("Please provide a port to bind on with -l")
	}

	// Make a host that listens on the given multiaddress
	ha, err := makeHost(ctx, *listenF)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Hello World, my hosts ID is %s\n", ha.ID())
	//defer ha.Close()

	if *targetF == "" {
		startListener(ctx, ha, *listenF)
		// Run until canceled.
		<-ctx.Done()
	} else {
		runSender(ctx, ha, *targetF)
	}
}

func getHostAddress(ha host.Host) string {
	// Build host multiaddress
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", ha.ID().Pretty()))

	// Now we can build a full multiaddress to reach this host
	// by encapsulating both addresses:
	addr := ha.Addrs()[0]
	return addr.Encapsulate(hostAddr).String()
}

func startListener(ctx context.Context, ha host.Host, listenPort int) {
	fullAddr := getHostAddress(ha)
	log.Printf("I am %s\n", fullAddr)

	// Set a stream handler on host A. /echo/1.0.0 is
	// a user-defined protocol name.
	ha.SetStreamHandler("/echo/1.0.0", func(s network.Stream) {
		log.Println("listener received new stream")
		if err := doEcho(s); err != nil {
			log.Println(err)
			s.Reset()
		} else {
			s.Close()
		}
	})

	log.Println("listening for connections")
	log.Printf("Now run \"./echo -l %d -d %s\" on a different terminal\n", listenPort+1, fullAddr)
}

func runSender(ctx context.Context, ha host.Host, targetPeer string) {
	fullAddr := getHostAddress(ha)
	log.Printf("I am %s\n", fullAddr)

	// Set a stream handler on host A. /echo/1.0.0 is
	// a user-defined protocol name.
	ha.SetStreamHandler("/echo/1.0.0", func(s network.Stream) {
		log.Println("sender received new stream")
		if err := doEcho(s); err != nil {
			log.Println(err)
			s.Reset()
		} else {
			s.Close()
		}
	})

	// The following code extracts target's the peer ID from the
	// given multiaddress
	ipfsaddr, err := ma.NewMultiaddr(targetPeer)
	if err != nil {
		log.Println(err)
		return
	}

	pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		log.Println(err)
		return
	}

	peerid, err := peer.Decode(pid)
	if err != nil {
		log.Println(err)
		return
	}

	// Decapsulate the /ipfs/<peerID> part from the target
	// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
	targetPeerAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", pid))
	targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)

	// We have a peer ID and a targetAddr so we add it to the peerstore
	// so LibP2P knows how to contact it
	ha.Peerstore().AddAddr(peerid, targetAddr, peerstore.PermanentAddrTTL)

	log.Println("sender opening stream")
	// make a new stream from host B to host A
	// it should be handled on host A by the handler we set above because
	// we use the same /echo/1.0.0 protocol
	s, err := ha.NewStream(context.Background(), peerid, "/echo/1.0.0")
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("sender saying hello")
	_, err = s.Write([]byte("Hello, world!\n"))
	if err != nil {
		log.Println(err)
		return
	}

	out, err := ioutil.ReadAll(s)
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("read reply: %q\n", out)
}

// doEcho reads a line of data a stream and writes it back
func doEcho(s network.Stream) error {
	buf := bufio.NewReader(s)
	str, err := buf.ReadString('\n')
	if err != nil {
		return err
	}

	log.Printf("read: %s", str)
	_, err = s.Write([]byte(str))
	return err
}
