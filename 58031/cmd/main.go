package main

import (
	"flag"
	"log"
	nethttp "net/http"
	kvhttp "raft-kv/internal/http"
	"raft-kv/internal/raft"
	"raft-kv/internal/store"
	"strings"
)

type NodeConfig struct {
	ID       string
	HTTPAddr string
	Peers    map[string]string
	DataDir  string
}

func main() {
	nodeID := flag.String("id", "node1", "Node ID")
	httpAddr := flag.String("http", "127.0.0.1:8001", "HTTP listen address")
	peersStr := flag.String("peers", "node2=127.0.0.1:8002,node3=127.0.0.1:8003", "Comma-separated peer addresses (id=addr)")
	dataDir := flag.String("data", "./data", "Data directory")
	flag.Parse()

	peers := make(map[string]string)
	if *peersStr != "" {
		for _, peer := range strings.Split(*peersStr, ",") {
			parts := strings.SplitN(peer, "=", 2)
			if len(parts) == 2 {
				peers[parts[0]] = parts[1]
			}
		}
	}

	applyCh := make(chan raft.ApplyMsg, 100)

	raftNode := raft.NewNode(*nodeID, *httpAddr, peers, applyCh)

	kvStore, err := store.NewKVStore(*dataDir + "/" + *nodeID + ".json")
	if err != nil {
		log.Fatalf("Failed to create KV store: %v", err)
	}

	httpServer := kvhttp.NewServer(raftNode, kvStore, applyCh)

	mux := nethttp.NewServeMux()
	kvhttp.SetupRoutes(mux, httpServer)

	raftNode.Start()

	log.Printf("Node %s starting on %s", *nodeID, *httpAddr)
	log.Printf("Peers: %v", peers)
	log.Printf("Data dir: %s", *dataDir)

	if err := kvhttp.StartServer(*httpAddr, mux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
