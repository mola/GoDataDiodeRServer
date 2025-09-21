package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
	"os"

	"github.com/gopcua/opcua/server"
	"github.com/gopcua/opcua/ua"
)

type OPCUAServer struct {
	port        int
	packetCount int
	running     bool
	numGroups   int
	server      *server.Server
	nodes       map[string]interface{}
	mu          sync.Mutex
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewOPCUAServer(numGroups int) *OPCUAServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &OPCUAServer{
		port:      49320,
		numGroups: numGroups,
		nodes:     make(map[string]interface{}),
		running:   false,
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (s *OPCUAServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("OPC UA server already running")
	}

	var opts []server.Option

	opts = append(opts,
		      server.EnableSecurity("None", ua.MessageSecurityModeNone),
		      server.EnableSecurity("Basic128Rsa15", ua.MessageSecurityModeSign),
		      server.EnableSecurity("Basic128Rsa15", ua.MessageSecurityModeSignAndEncrypt),
		      server.EnableSecurity("Basic256", ua.MessageSecurityModeSign),
		      server.EnableSecurity("Basic256", ua.MessageSecurityModeSignAndEncrypt),
		      server.EnableSecurity("Basic256Sha256", ua.MessageSecurityModeSignAndEncrypt),
		      server.EnableSecurity("Basic256Sha256", ua.MessageSecurityModeSign),
		      server.EnableSecurity("Aes128_Sha256_RsaOaep", ua.MessageSecurityModeSign),
		      server.EnableSecurity("Aes128_Sha256_RsaOaep", ua.MessageSecurityModeSignAndEncrypt),
		      server.EnableSecurity("Aes256_Sha256_RsaPss", ua.MessageSecurityModeSign),
		      server.EnableSecurity("Aes256_Sha256_RsaPss", ua.MessageSecurityModeSignAndEncrypt),
	)

	opts = append(opts,
		      server.EnableAuthMode(ua.UserTokenTypeAnonymous),
		      server.EnableAuthMode(ua.UserTokenTypeUserName),
		      server.EnableAuthMode(ua.UserTokenTypeCertificate),
		      //		server.EnableAuthWithoutEncryption(), // Dangerous and not recommended, shown for illustration only
	)

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("Error getting host name %v", err)
	}

	opts = append(opts,
		      server.EndPoint("0.0.0.0", s.port),
		      server.EndPoint("localhost", s.port),
		      server.EndPoint(hostname, s.port),
	)


	// Create server with endpoint
	srv := server.New(opts...)

	s.server = srv

	if err := s.server.Start(context.Background()); err != nil {
		log.Fatalf("Error starting server, exiting: %s", err)
	}


	s.running = true
	s.wg.Add(1)
	go s.monitorRates()


	log.Printf("OPC UA Server started on port %d", s.port)
	return nil
}

func (s *OPCUAServer) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.running = false
	s.cancel()

	if s.server != nil {
		s.server.Close()
	}
	s.wg.Wait()

	log.Println("OPC UA Server stopped")
}

func (s *OPCUAServer) monitorRates() {
	defer s.wg.Done()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
			case <-ticker.C:
				s.mu.Lock()
				currentCount := s.packetCount
				s.packetCount = 0
				s.mu.Unlock()

				if currentCount > 0 {
					log.Printf("Packet rate: %d packets/second", currentCount)
				}
			case <-s.ctx.Done():
				return
		}
	}
}

func (s *OPCUAServer) HandleVariableUpdate(tag, path string, value interface{}, statusCode uint32, namespace string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.packetCount++

	// Log the variable update
	log.Printf("Variable update - Tag: %s, Path: %s, Value: %v, Status: %d, Namespace: %s",
		   tag, path, value, statusCode, namespace)
}

func (s *OPCUAServer) ManageUsers(data map[string]interface{}) {
	log.Println("Managing OPC UA users:", data)
}
