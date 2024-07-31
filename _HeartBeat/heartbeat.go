package heartbeat

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type Client struct {
	Id        string
	address   string
	LastBeat  time.Time
	IsAlive   bool
	mu        sync.Mutex
	NewClient error
}

type HeartbeatClient struct {
	nodes       map[string]*Client
	timeout     time.Duration
	checkPeriod time.Duration
	mu          sync.Mutex
}

// New creates new Heartbeat with specified duration. timeoutFunc will be called
// if timeout for heartbeat is expired. Note that in case of timeout you need to
// call Beat() to reactivate Heartbeat.
func New(timeout time.Duration, timeoutFunc func()) *HeartbeatClient {
	hb := &HeartbeatClient{
		nodes:       map[string]*Client{},
		timeout:     timeout,
		checkPeriod: time.Duration(1),
		mu:          sync.Mutex{},
	}
	return hb
}

func NewHeartbeatClient(timeout, checkPeriod time.Duration) *HeartbeatClient {
	return &HeartbeatClient{
		nodes:       make(map[string]*Client),
		timeout:     timeout,
		checkPeriod: checkPeriod,
	}
}

// func (c *HeartbeatClient) Get(context context.Context, node []string) (*HeartbeatClient, error){
// 	get := &Client{}
// 	err := c.Get(context, node)
// 	if err != nil{
// 		return nil, err
// 	}
// 	return get, nil
// }

func (hb *HeartbeatClient) Beat() {
	hb.checkPeriod.Hours()
}

func (hm *HeartbeatClient) AddNode(id, address string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.nodes[id] = &Client{
		Id:      id,
		address: address,
		IsAlive: true,
	}
}

func (hm *HeartbeatClient) StartHeartbeatListener(port string) {
	ln, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("Error starting heartbeat listener:", err)
		return
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go hm.handleHeartbeat(conn)
	}
}

func (hm *HeartbeatClient) handleHeartbeat(conn net.Conn) {
	defer conn.Close()
	var id string
	fmt.Fscanf(conn, "%s\n", &id)

	hm.mu.Lock()
	if node, exists := hm.nodes[id]; exists {
		node.mu.Lock()
		node.LastBeat = time.Now()
		node.IsAlive = true
		node.mu.Unlock()
	}
	hm.mu.Unlock()
}

func (hm *HeartbeatClient) StartMonitoring() {
	ticker := time.NewTicker(hm.checkPeriod)
	for range ticker.C {
		hm.checkNodes()
	}
}

func (hm *HeartbeatClient) checkNodes() {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	for _, node := range hm.nodes {
		node.mu.Lock()
		if time.Since(node.LastBeat) > hm.timeout {
			node.IsAlive = false
		}
		node.mu.Unlock()
	}
}
