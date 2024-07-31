package quorum

import (
	"fmt"
	"sync"
)

// Quorums type builds and maintains the cluster quorums using the bintree package to maintain the Node.
type Quorums struct {
	MyQuorums   map[int][]string
	Peers       map[string]bool
	NumPeers    int
	nodes       []*Node
	loaclAddre  string
	readQuorum  int
	writeQuorum int
	mu          sync.Mutex
}

type Node struct {
	ID       string
	isAlive  bool
	mu       sync.Mutex
	AllPaths string
}

// NewQuorums initializes the quorums from the local node's
func NewQuorum(readQuorum, writeQuorum int, nodeCount int) *Quorums {
	q := &Quorums{
		readQuorum:  readQuorum,
		writeQuorum: writeQuorum,
		nodes:       make([]*Node, nodeCount),
		Peers:       map[string]bool{},
		NumPeers:    nodeCount,
	}

	for i := 0; i < nodeCount; i++ {
		q.nodes[i] = &Node{
			ID:      fmt.Sprintf("node%d", i+1),
			isAlive: true,
		}
	}

	return q
}

func (q *Quorums) read() bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	count := 0
	for _, node := range q.nodes {
		node.mu.Lock()
		if node.isAlive {
			count++
		}
		node.mu.Unlock()
		if count >= q.readQuorum {
			return true
		}
	}
	return false
}

func (q *Quorums) write() bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	count := 0
	for _, node := range q.nodes {
		node.mu.Lock()
		if node.isAlive {
			count++
		}
		node.mu.Unlock()
		if count >= q.writeQuorum {
			return true
		}
	}
	return false
}

func (q *Quorums) setNodeStatus(id string, status bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, node := range q.nodes {
		if node.ID == id {
			node.mu.Lock()
			node.isAlive = status
			node.mu.Unlock()
			return
		}
	}
}

// // SubstitutePaths returns possible paths starting from the site’s two children and ending in leaf node.
// // Used in the distributed mutex algorithm when a node in the quorum has failed.
// func (q *Quorums) SubstitutePaths(node string) [][]string {
// 	var paths [][]string
// 	for _, treePath := range q.fullTree.NodePaths(node) {
// 		var path []string
// 		for _, node := range treePath {
// 			path = append(path, node)
// 		}
// 		paths = append(paths, path)
// 	}
// 	return paths
// }

// SubstitutePaths returns possible paths starting from the site’s two children and ending in leaf node.
func (q *Quorums) SubstitutePaths(node string) [][]string {
	// Mock implementation for NodePaths function.
	// This should return all paths starting from a given node in your tree structure.
	NodePaths := func(node string) [][]string {
		// This is a mock implementation, replace with your actual tree traversal logic.
		return [][]string{
			{node, "node2", "node3"},
			{node, "node4", "node5"},
		}
	}

	var paths [][]string
	for _, treePath := range NodePaths(node) {
		var path []string
		for _, n := range treePath {
			path = append(path, n)
		}
		paths = append(paths, path)
	}
	return paths
}
