package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type SplunkLogger struct {
	splunkHost string
	splunkPort int
	conn       *net.UDPConn
	running    bool
	mu         sync.Mutex
	wg         sync.WaitGroup
}

func NewSplunkLogger(host string, port int) *SplunkLogger {
	return &SplunkLogger{
		splunkHost: host,
		splunkPort: port,
		running:    false,
	}
}

func (l *SplunkLogger) Start() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.running {
		return fmt.Errorf("Splunk logger already running")
	}

	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.ParseIP(l.splunkHost),
				 Port: l.splunkPort,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to Splunk: %w", err)
	}

	l.conn = conn
	l.running = true
	l.wg.Add(1)

	// Start heartbeat goroutine
	go l.heartbeatLoop()

	log.Printf("Splunk logger started for %s:%d", l.splunkHost, l.splunkPort)
	return nil
}

func (l *SplunkLogger) Stop() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.running {
		return
	}

	l.running = false
	if l.conn != nil {
		l.conn.Close()
	}
	l.wg.Wait()

	log.Println("Splunk logger stopped")
}

func (l *SplunkLogger) heartbeatLoop() {
	defer l.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
			case <-ticker.C:
				heartbeat := map[string]interface{}{
					"timestamp": time.Now().Format(time.RFC3339),
					"level":     "INFO",
					"message":   "Heartbeat from Splunk logger",
				}
				l.SendLog(heartbeat)
		}

		l.mu.Lock()
		if !l.running {
			l.mu.Unlock()
			break
		}
		l.mu.Unlock()
	}
}

func (l *SplunkLogger) SendLog(logData map[string]interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.running || l.conn == nil {
		return
	}

	// Ensure required fields
	if _, ok := logData["timestamp"]; !ok {
		logData["timestamp"] = time.Now().Format(time.RFC3339)
	}
	if _, ok := logData["level"]; !ok {
		logData["level"] = "INFO"
	}
	if _, ok := logData["message"]; !ok {
		logData["message"] = "No message provided"
	}

	jsonData, err := json.Marshal(logData)
	if err != nil {
		log.Printf("Failed to marshal log data: %v", err)
		return
	}

	if _, err := l.conn.Write(jsonData); err != nil {
		log.Printf("Failed to send log to Splunk: %v", err)
	}
}

func (l *SplunkLogger) ConfigLog(configData map[string]interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if newHost, ok := configData["ip"].(string); ok && newHost != l.splunkHost {
		l.splunkHost = newHost
	}
	if newPort, ok := configData["port"].(float64); ok {
		l.splunkPort = int(newPort)
	}

	log.Printf("Updated logger config: %s:%d", l.splunkHost, l.splunkPort)
}
