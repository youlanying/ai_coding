package http

import (
	"encoding/gob"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"raft-kv/internal/raft"
	"raft-kv/internal/store"
	"strings"
	"time"
)

type Server struct {
	raftNode *raft.Node
	kvStore  *store.KVStore
	applyCh   chan raft.ApplyMsg
}

func NewServer(raftNode *raft.Node, kvStore *store.KVStore, applyCh chan raft.ApplyMsg) *Server {
	s := &Server{
		raftNode: raftNode,
		kvStore:  kvStore,
		applyCh:   applyCh,
	}
	go s.applyLoop()
	return s
}

func (s *Server) applyLoop() {
	for msg := range s.applyCh {
		if msg.CommandValid {
			cmd, ok := msg.Command.(raft.KVCommand)
			if !ok {
				continue
			}
			switch cmd.Type {
			case raft.CommandSet:
				s.kvStore.Set(cmd.Key, cmd.Value)
			case raft.CommandDelete:
				s.kvStore.Delete(cmd.Key)
			}
		}
	}
}

func (s *Server) HandleVote(w http.ResponseWriter, r *http.Request) {
	var args raft.RequestVoteArgs
	decoder := gob.NewDecoder(r.Body)
	if err := decoder.Decode(&args); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	reply := s.raftNode.RequestVote(args)

	encoder := gob.NewEncoder(w)
	encoder.Encode(reply)
}

func (s *Server) HandleAppendEntries(w http.ResponseWriter, r *http.Request) {
	var args raft.AppendEntriesArgs
	decoder := gob.NewDecoder(r.Body)
	if err := decoder.Decode(&args); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	reply := s.raftNode.AppendEntries(args)

	encoder := gob.NewEncoder(w)
	encoder.Encode(reply)
}

type SetRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type GetResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Found bool   `json:"found"`
}

type StatusResponse struct {
	NodeID string `json:"node_id"`
	State  string `json:"state"`
	Term   int    `json:"term"`
	Leader bool   `json:"leader"`
}

func (s *Server) HandleGet(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/get/")
	if key == "" {
		http.Error(w, "key is required", http.StatusBadRequest)
		return
	}

	value, found := s.kvStore.Get(key)
	resp := GetResponse{
		Key:   key,
		Value: value,
		Found: found,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) HandleSet(w http.ResponseWriter, r *http.Request) {
	_, isLeader := s.raftNode.GetState()
	if !isLeader {
		http.Error(w, "not leader", http.StatusServiceUnavailable)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req SetRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Key == "" {
		http.Error(w, "key is required", http.StatusBadRequest)
		return
	}

	cmd := raft.KVCommand{
		Type:  raft.CommandSet,
		Key:   req.Key,
		Value: req.Value,
	}

	_, _, ok := s.raftNode.SubmitCommand(cmd)
	if !ok {
		http.Error(w, "failed to submit command", http.StatusInternalServerError)
		return
	}

	time.Sleep(100 * time.Millisecond)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "key": req.Key})
}

func (s *Server) HandleDelete(w http.ResponseWriter, r *http.Request) {
	_, isLeader := s.raftNode.GetState()
	if !isLeader {
		http.Error(w, "not leader", http.StatusServiceUnavailable)
		return
	}

	key := strings.TrimPrefix(r.URL.Path, "/delete/")
	if key == "" {
		http.Error(w, "key is required", http.StatusBadRequest)
		return
	}

	cmd := raft.KVCommand{
		Type: raft.CommandDelete,
		Key:  key,
	}

	_, _, ok := s.raftNode.SubmitCommand(cmd)
	if !ok {
		http.Error(w, "failed to submit command", http.StatusInternalServerError)
		return
	}

	time.Sleep(100 * time.Millisecond)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "key": key})
}

func (s *Server) HandleStatus(w http.ResponseWriter, r *http.Request) {
	term, isLeader := s.raftNode.GetState()
	resp := StatusResponse{
		NodeID: s.raftNode.GetID(),
		State:  s.raftNode.GetStateString(),
		Term:   term,
		Leader: isLeader,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) HandleList(w http.ResponseWriter, r *http.Request) {
	data := s.kvStore.GetAll()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func SetupRoutes(mux *http.ServeMux, server *Server) {
	mux.HandleFunc("/raft/vote", server.HandleVote)
	mux.HandleFunc("/raft/append", server.HandleAppendEntries)
	mux.HandleFunc("/get/", server.HandleGet)
	mux.HandleFunc("/set", server.HandleSet)
	mux.HandleFunc("/delete/", server.HandleDelete)
	mux.HandleFunc("/status", server.HandleStatus)
	mux.HandleFunc("/list", server.HandleList)
}

func StartServer(addr string, mux *http.ServeMux) error {
	log.Printf("Server starting on %s", addr)
	return http.ListenAndServe(addr, mux)
}
