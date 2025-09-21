package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type UDPReceiver struct {
	dataCallback func(map[string]interface{})
	host         string
	port         int
	conn         *net.UDPConn
	running      bool
	wg           sync.WaitGroup
	mu           sync.Mutex
}

func NewUDPReceiver(callback func(map[string]interface{})) *UDPReceiver {
	return &UDPReceiver{
		dataCallback: callback,
		host:         "0.0.0.0",
		port:         8000,
		running:      false,
	}
}

func (r *UDPReceiver) Start() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.running {
		return fmt.Errorf("UDP receiver already running")
	}

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", r.host, r.port))
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address: %w", err)
	}

	r.conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on UDP: %w", err)
	}

	r.running = true
	r.wg.Add(1)

	go r.receiveLoop()

	log.Printf("UDP receiver started on %s:%d", r.host, r.port)
	return nil
}

func (r *UDPReceiver) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.running {
		return
	}

	r.running = false
	if r.conn != nil {
		r.conn.Close()
	}
	r.wg.Wait()

	log.Println("UDP receiver stopped")
}

func (r *UDPReceiver) receiveLoop() {
	defer r.wg.Done()

	buffer := make([]byte, 4096)

	for {
		r.mu.Lock()
		if !r.running {
			r.mu.Unlock()
			break
		}
		r.mu.Unlock()

		// Set read timeout to allow checking running status
		r.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

		n, addr, err := r.conn.ReadFromUDP(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			log.Printf("Error reading from UDP: %v", err)
			continue
		}

		go r.processData(buffer[:n], addr)
	}
}

func (r *UDPReceiver) processData(data []byte, addr *net.UDPAddr) {
	var jsonData map[string]interface{}

	if err := json.Unmarshal(data, &jsonData); err != nil {
		log.Printf("Failed to decode JSON from %s: %v", addr.String(), err)
		return
	}

	r.dataCallback(jsonData)
}
