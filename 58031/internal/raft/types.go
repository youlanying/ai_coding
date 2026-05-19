package raft

import (
	"sync"
	"time"
)

type NodeState string

const (
	Follower  NodeState = "follower"
	Candidate NodeState = "candidate"
	Leader    NodeState = "leader"
)

type LogEntry struct {
	Term    int
	Index   int
	Command interface{}
}

type CommandType string

const (
	CommandSet    CommandType = "set"
	CommandDelete CommandType = "delete"
)

type KVCommand struct {
	Type  CommandType
	Key   string
	Value string
}

type RequestVoteArgs struct {
	Term         int
	CandidateID  string
	LastLogIndex int
	LastLogTerm  int
}

type RequestVoteReply struct {
	Term        int
	VoteGranted bool
}

type AppendEntriesArgs struct {
	Term         int
	LeaderID     string
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	LeaderCommit int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
}

type Node struct {
	mu sync.Mutex

	id      string
	state   NodeState
	term    int
	voteFor string

	logs         []LogEntry
	commitIndex  int
	lastApplied  int
	nextIndex    map[string]int
	matchIndex   map[string]int

	peers    map[string]string
	httpAddr string

	electionTimer  *time.Timer
	heartbeatTimer *time.Timer

	applyCh chan ApplyMsg

	stopCh chan struct{}
}

type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int
}

const (
	electionTimeoutMin = 150 * time.Millisecond
	electionTimeoutMax = 300 * time.Millisecond
	heartbeatInterval  = 100 * time.Millisecond
)
