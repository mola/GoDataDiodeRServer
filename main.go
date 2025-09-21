package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Application struct {
	opcuaServer    *OPCUAServer
	udpReceiver    *UDPReceiver
	splunkLogger   *SplunkLogger
	running        bool
	configMutex    sync.Mutex
	shutdownSignal chan os.Signal
}

func NewApplication() *Application {
	return &Application{
		running:        false,
		shutdownSignal: make(chan os.Signal, 1),
	}
}

func (app *Application) Start() error {
	app.running = true

	// Start OPC UA server
	app.opcuaServer = NewOPCUAServer(10)
	if err := app.opcuaServer.Start(); err != nil {
		return fmt.Errorf("failed to start OPC UA server: %w", err)
	}

	// Start UDP receiver
	app.udpReceiver = NewUDPReceiver(app.handleUDPData)
	if err := app.udpReceiver.Start(); err != nil {
		return fmt.Errorf("failed to start UDP receiver: %w", err)
	}

	// Start Splunk logger
	app.splunkLogger = NewSplunkLogger("192.168.2.112", 1514)
	if err := app.splunkLogger.Start(); err != nil {
		return fmt.Errorf("failed to start Splunk logger: %w", err)
	}

	log.Println("Application started successfully")
	return nil
}

func (app *Application) Stop() error {
	app.running = false
	log.Println("Shutting down application...")

	if app.udpReceiver != nil {
		app.udpReceiver.Stop()
	}

	if app.splunkLogger != nil {
		app.splunkLogger.Stop()
	}

	if app.opcuaServer != nil {
		app.opcuaServer.Stop()
	}

	log.Println("Application stopped")
	return nil
}

func (app *Application) handleUDPData(data map[string]interface{}) {
	app.configMutex.Lock()
	defer app.configMutex.Unlock()

	dataType, ok := data["type"].(string)
	if !ok {
		log.Println("Missing or invalid data type")
		return
	}

	switch dataType {
		case "opcua":
			app.handleOPCUAData(data)
		case "opcua_auth_config":
			app.handleAuthConfig(data)
		case "opcua_users":
			app.handleUsers(data)
		case "log":
			app.handleLog(data)
		case "secondary_log_config":
			app.handleLogConfig(data)
		case "receive_interface_config":
			app.handleInterfaceConfig(data)
		default:
			log.Printf("Unknown data type: %s", dataType)
	}
}

func (app *Application) handleOPCUAData(data map[string]interface{}) {
	tag, _ := data["tag"].(string)
	path, _ := data["path"].(string)
	value := data["value"]
	statusCode, _ := data["status_code"].(float64)
	namespace, _ := data["namespace"].(string)

	app.opcuaServer.HandleVariableUpdate(tag, path, value, uint32(statusCode), namespace)
}

func (app *Application) handleAuthConfig(data map[string]interface{}) {
	// Implement auth config handling
	log.Println("Handling auth config:", data)
}

func (app *Application) handleUsers(data map[string]interface{}) {
	// Implement user management
	log.Println("Handling users:", data)
}

func (app *Application) handleLog(data map[string]interface{}) {
	app.splunkLogger.SendLog(data)
}

func (app *Application) handleLogConfig(data map[string]interface{}) {
	// Implement log config handling
	log.Println("Handling log config:", data)
}

func (app *Application) handleInterfaceConfig(data map[string]interface{}) {
	// Implement interface config handling
	log.Println("Handling interface config:", data)
}

func main() {
	app := NewApplication()

	// Setup signal handling
	signal.Notify(app.shutdownSignal, syscall.SIGINT, syscall.SIGTERM)

	// Start application
	if err := app.Start(); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}

	// Wait for shutdown signal
	<-app.shutdownSignal

	// Stop application
	if err := app.Stop(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
}
