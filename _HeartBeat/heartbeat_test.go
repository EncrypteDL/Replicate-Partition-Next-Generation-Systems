package heartbeat

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestAddNode(t *testing.T) {
	hb := New(5*time.Second, nil)
	hb.AddNode("node1", "localhost:9001")
	hb.AddNode("node2", "localhost:9002")

	if len(hb.nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(hb.nodes))
	}

	if _, exists := hb.nodes["node1"]; !exists {
		t.Errorf("expected node1 to be present")
	}

	if _, exists := hb.nodes["node2"]; !exists {
		t.Errorf("expected node2 to be present")
	}
}

func TestCheckNodes(t *testing.T) {
	hb := New(1*time.Second, nil)
	hb.AddNode("node1", "localhost:9001")

	hb.nodes["node1"].LastBeat = time.Now().Add(-2 * time.Second)

	hb.checkNodes()

	if hb.nodes["node1"].IsAlive {
		t.Errorf("expected node1 to be dead")
	}
}

func TestHandleHeartbeat(t *testing.T) {
	hb := New(5*time.Second, nil)
	hb.AddNode("node1", "localhost:9001")

	// Mock heartbeat
	go func() {
		ln, _ := net.Listen("tcp", ":9000")
		conn, _ := ln.Accept()
		fmt.Fprintf(conn, "node1\n")
		conn.Close()
	}()

	hb.StartHeartbeatListener(":9000")

	time.Sleep(1 * time.Second)

	if !hb.nodes["node1"].IsAlive {
		t.Errorf("expected node1 to be alive")
	}
}
