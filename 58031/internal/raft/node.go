package raft

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"
)

func init() {
	gob.Register(KVCommand{})
}

func NewNode(id string, httpAddr string, peers map[string]string, applyCh chan ApplyMsg) *Node {
	n := &Node{
		id:         id,
		state:      Follower,
		term:       0,
		voteFor:    "",
		logs:       []LogEntry{{Term: 0, Index: 0}},
		peers:      peers,
		httpAddr:   httpAddr,
		nextIndex:  make(map[string]int),
		matchIndex: make(map[string]int),
		applyCh:    applyCh,
		stopCh:     make(chan struct{}),
	}

	for peerID := range peers {
		n.nextIndex[peerID] = 1
		n.matchIndex[peerID] = 0
	}

	return n
}

func (n *Node) Start() {
	n.electionTimer = time.NewTimer(randomElectionTimeout())
	n.heartbeatTimer = time.NewTimer(heartbeatInterval)
	n.heartbeatTimer.Stop()

	go n.run()
}

func (n *Node) Stop() {
	close(n.stopCh)
}

func (n *Node) run() {
	for {
		select {
		case <-n.stopCh:
			return
		case <-n.electionTimer.C:
			n.startElection()
		case <-n.heartbeatTimer.C:
			if n.state == Leader {
				n.sendHeartbeats()
				n.heartbeatTimer.Reset(heartbeatInterval)
			}
		}
	}
}

func randomElectionTimeout() time.Duration {
	return electionTimeoutMin + time.Duration(rand.Int63n(int64(electionTimeoutMax-electionTimeoutMin)))
}

func (n *Node) startElection() {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.state = Candidate
	n.term++
	n.voteFor = n.id
	votes := 1

	lastLogIndex := len(n.logs) - 1
	lastLogTerm := n.logs[lastLogIndex].Term

	var wg sync.WaitGroup
	for peerID, peerAddr := range n.peers {
		wg.Add(1)
		go func(id, addr string) {
			defer wg.Done()
			args := RequestVoteArgs{
				Term:         n.term,
				CandidateID:  n.id,
				LastLogIndex: lastLogIndex,
				LastLogTerm:  lastLogTerm,
			}
			reply := n.sendRequestVote(addr, args)
			n.mu.Lock()
			defer n.mu.Unlock()
			if reply.Term > n.term {
				n.becomeFollower(reply.Term)
				return
			}
			if reply.VoteGranted && n.state == Candidate && n.term == args.Term {
				votes++
				if votes > (len(n.peers)+1)/2 {
					n.becomeLeader()
				}
			}
		}(peerID, peerAddr)
	}

	n.electionTimer.Reset(randomElectionTimeout())
}

func (n *Node) becomeFollower(term int) {
	n.term = term
	n.state = Follower
	n.voteFor = ""
	n.heartbeatTimer.Stop()
	n.electionTimer.Reset(randomElectionTimeout())
}

func (n *Node) becomeLeader() {
	n.state = Leader
	n.electionTimer.Stop()
	n.heartbeatTimer.Reset(heartbeatInterval)

	for peerID := range n.peers {
		n.nextIndex[peerID] = len(n.logs)
		n.matchIndex[peerID] = 0
	}

	n.sendHeartbeats()
}

func (n *Node) sendHeartbeats() {
	for peerID, peerAddr := range n.peers {
		go func(id, addr string) {
			n.mu.Lock()
			prevLogIndex := n.nextIndex[id] - 1
			prevLogTerm := n.logs[prevLogIndex].Term
			entries := n.logs[n.nextIndex[id]:]
			args := AppendEntriesArgs{
				Term:         n.term,
				LeaderID:     n.id,
				PrevLogIndex: prevLogIndex,
				PrevLogTerm:  prevLogTerm,
				Entries:      entries,
				LeaderCommit: n.commitIndex,
			}
			n.mu.Unlock()

			reply := n.sendAppendEntries(addr, args)
			n.handleAppendEntriesReply(id, reply, args)
		}(peerID, peerAddr)
	}
}

func (n *Node) handleAppendEntriesReply(peerID string, reply AppendEntriesReply, args AppendEntriesArgs) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if reply.Term > n.term {
		n.becomeFollower(reply.Term)
		return
	}

	if n.state != Leader || n.term != args.Term {
		return
	}

	if reply.Success {
		n.nextIndex[peerID] = args.PrevLogIndex + len(args.Entries) + 1
		n.matchIndex[peerID] = n.nextIndex[peerID] - 1
		n.updateCommitIndex()
	} else {
		n.nextIndex[peerID]--
		if n.nextIndex[peerID] < 1 {
			n.nextIndex[peerID] = 1
		}
	}
}

func (n *Node) updateCommitIndex() {
	for i := len(n.logs) - 1; i > n.commitIndex; i-- {
		if n.logs[i].Term == n.term {
			matchCount := 1
			for peerID := range n.peers {
				if n.matchIndex[peerID] >= i {
					matchCount++
				}
			}
			if matchCount > (len(n.peers)+1)/2 {
				n.commitIndex = i
				n.applyCommitted()
				break
			}
		}
	}
}

func (n *Node) applyCommitted() {
	for n.lastApplied < n.commitIndex {
		n.lastApplied++
		entry := n.logs[n.lastApplied]
		n.applyCh <- ApplyMsg{
			CommandValid: true,
			Command:      entry.Command,
			CommandIndex: entry.Index,
		}
	}
}

func (n *Node) RequestVote(args RequestVoteArgs) RequestVoteReply {
	n.mu.Lock()
	defer n.mu.Unlock()

	reply := RequestVoteReply{Term: n.term, VoteGranted: false}

	if args.Term < n.term {
		return reply
	}

	if args.Term > n.term {
		n.becomeFollower(args.Term)
	}

	lastLogIndex := len(n.logs) - 1
	lastLogTerm := n.logs[lastLogIndex].Term

	logOk := args.LastLogTerm > lastLogTerm ||
		(args.LastLogTerm == lastLogTerm && args.LastLogIndex >= lastLogIndex)

	if (n.voteFor == "" || n.voteFor == args.CandidateID) && logOk {
		n.voteFor = args.CandidateID
		reply.VoteGranted = true
		n.electionTimer.Reset(randomElectionTimeout())
	}

	return reply
}

func (n *Node) AppendEntries(args AppendEntriesArgs) AppendEntriesReply {
	n.mu.Lock()
	defer n.mu.Unlock()

	reply := AppendEntriesReply{Term: n.term, Success: false}

	if args.Term < n.term {
		return reply
	}

	if args.Term > n.term {
		n.becomeFollower(args.Term)
	}

	n.electionTimer.Reset(randomElectionTimeout())

	if args.PrevLogIndex >= len(n.logs) {
		return reply
	}

	if n.logs[args.PrevLogIndex].Term != args.PrevLogTerm {
		return reply
	}

	for i, entry := range args.Entries {
		index := args.PrevLogIndex + 1 + i
		if index >= len(n.logs) {
			n.logs = append(n.logs, entry)
		} else if n.logs[index].Term != entry.Term {
			n.logs = n.logs[:index]
			n.logs = append(n.logs, entry)
		}
	}

	if args.LeaderCommit > n.commitIndex {
		if args.LeaderCommit > len(n.logs)-1 {
			n.commitIndex = len(n.logs) - 1
		} else {
			n.commitIndex = args.LeaderCommit
		}
		n.applyCommitted()
	}

	reply.Success = true
	return reply
}

func (n *Node) sendRequestVote(addr string, args RequestVoteArgs) RequestVoteReply {
	var reply RequestVoteReply
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(args)

	resp, err := http.Post(fmt.Sprintf("http://%s/raft/vote", addr), "application/octet-stream", &buf)
	if err != nil {
		return reply
	}
	defer resp.Body.Close()

	dec := gob.NewDecoder(resp.Body)
	dec.Decode(&reply)
	return reply
}

func (n *Node) sendAppendEntries(addr string, args AppendEntriesArgs) AppendEntriesReply {
	var reply AppendEntriesReply
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(args)

	resp, err := http.Post(fmt.Sprintf("http://%s/raft/append", addr), "application/octet-stream", &buf)
	if err != nil {
		return reply
	}
	defer resp.Body.Close()

	dec := gob.NewDecoder(resp.Body)
	dec.Decode(&reply)
	return reply
}

func (n *Node) SubmitCommand(command interface{}) (int, int, bool) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.state != Leader {
		return 0, 0, false
	}

	entry := LogEntry{
		Term:    n.term,
		Index:   len(n.logs),
		Command: command,
	}
	n.logs = append(n.logs, entry)
	n.matchIndex[n.id] = entry.Index

	return entry.Index, entry.Term, true
}

func (n *Node) GetState() (int, bool) {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.term, n.state == Leader
}

func (n *Node) GetID() string {
	return n.id
}

func (n *Node) GetStateString() string {
	n.mu.Lock()
	defer n.mu.Unlock()
	return string(n.state)
}

func (n *Node) GetLogs() []LogEntry {
	n.mu.Lock()
	defer n.mu.Unlock()
	logs := make([]LogEntry, len(n.logs))
	copy(logs, n.logs)
	return logs
}

func ParseURL(addr string) string {
	u, err := url.Parse(addr)
	if err != nil {
		return addr
	}
	if u.Host != "" {
		return u.Host
	}
	return addr
}
