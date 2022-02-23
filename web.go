package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/libp2p/go-libp2p-core/peer"
)

func web(port string) {
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/peers", peers)
	http.HandleFunc("/addr", addr)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		//serveWs(hub, w, r)
	})
	http.ListenAndServe(":"+port, nil)
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "home.html")
}

func peers(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, h.ID().String()+"\n")
	logger.Info("Searching for other members...")
	members := h.Peerstore().Peers()
	for _, p := range members {
		fmt.Fprintf(w, p.String()+"\n")
	}
}

func addr(w http.ResponseWriter, req *http.Request) {
	pid := req.URL.Query().Get("pid")
	pi := h.Peerstore().PeerInfo(peer.ID(pid))
	fmt.Fprintf(w, pi.String())
}
