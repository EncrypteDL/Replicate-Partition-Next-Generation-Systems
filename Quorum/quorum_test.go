package quorum

import (
	"testing"
)

func TestNewQuorums(t *testing.T) {
	readQuorum := 3
	writeQuorum := 4
	nodeCount := 5

	quorum := NewQuorum(readQuorum, writeQuorum, nodeCount)

	if len(quorum.nodes) != nodeCount {
		t.Errorf("Expected %d nodes, got %d", nodeCount, len(quorum.nodes))
	}

	if quorum.readQuorum != readQuorum {
		t.Errorf("Expected readQuorum %d, got %d", readQuorum, quorum.readQuorum)
	}

	if quorum.writeQuorum != writeQuorum {
		t.Errorf("Expected writeQuorum %d, got %d", writeQuorum, quorum.writeQuorum)
	}
}

func TestRead(t *testing.T) {
	readQuorum := 3
	writeQuorum := 4
	nodeCount := 5

	quorum := NewQuorum(readQuorum, writeQuorum, nodeCount)

	if !quorum.read() {
		t.Errorf("Expected read quorum to be successful")
	}

	quorum.setNodeStatus("node1", false)
	quorum.setNodeStatus("node2", false)
	quorum.setNodeStatus("node3", false)

	if quorum.read() {
		t.Errorf("Expected read quorum to fail")
	}
}

func TestWrite(t *testing.T) {
	readQuorum := 3
	writeQuorum := 4
	nodeCount := 5

	quorum := NewQuorum(readQuorum, writeQuorum, nodeCount)

	if !quorum.write() {
		t.Errorf("Expected write quorum to be successful")
	}

	quorum.setNodeStatus("node1", false)
	quorum.setNodeStatus("node2", false)
	quorum.setNodeStatus("node3", false)

	if quorum.write() {
		t.Errorf("Expected write quorum to fail")
	}
}

func TestSetNodeStatus(t *testing.T) {
	readQuorum := 3
	writeQuorum := 4
	nodeCount := 5

	quorum := NewQuorum(readQuorum, writeQuorum, nodeCount)

	quorum.setNodeStatus("node1", false)
	if quorum.nodes[0].isAlive {
		t.Errorf("Expected node1 to be marked as not alive")
	}

	quorum.setNodeStatus("node1", true)
	if !quorum.nodes[0].isAlive {
		t.Errorf("Expected node1 to be marked as alive")
	}
}

func TestSubstitutePaths(t *testing.T) {
	quorum := NewQuorum(2, 3, 5)

	paths := quorum.SubstitutePaths("node1")

	expectedPaths := [][]string{
		{"node1", "node2", "node3"},
		{"node1", "node4", "node5"},
	}

	if len(paths) != len(expectedPaths) {
		t.Errorf("Expected %d paths, got %d", len(expectedPaths), len(paths))
	}

	for i, path := range paths {
		for j, node := range path {
			if node != expectedPaths[i][j] {
				t.Errorf("Expected path %v, got %v", expectedPaths[i], path)
			}
		}
	}
}
