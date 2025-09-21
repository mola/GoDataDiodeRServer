
# GoDataDiodeRServer 
A high-performance, secure data gateway application written in Go (Golang). This server acts as a unidirectional data diode, receiving JSON-formatted telemetry data via UDP and seamlessly exposing it as a live OPC UA server for industrial automation and monitoring systems.

## Features

*   **Unidirectional Data Flow (Data Diode Pattern):** Emulates a hardware data diode by only accepting incoming UDP data, enforcing a strict one-way data transfer for enhanced security. The server does not send data back to the UDP source.
*   **UDP JSON Listener:** Efficiently listens for incoming datagrams on a configurable UDP port. The payload is expected to be a JSON object containing key-value pairs of telemetry data (e.g., `{"temperature": 42.7, "pressure": 1013.25, "status": "RUNNING"}`).
*   **Integrated OPC UA Server:** Hosts a full OPC UA server that makes the received data available to OPC UA clients (e.g., SCADA systems, HMIs, historians, or custom clients).
*   **Real-Time Data Bridging:** Parses incoming JSON messages and immediately updates the corresponding OPC UA node values, providing near real-time access to the data.
*   **Written in Go:** Benefits from the advantages of the Go language:
    *   **High Performance:** Excellent concurrency handling for multiple OPC UA connections and high-frequency UDP packets.
    *   **Cross-Platform:** Single binary that runs on Windows, Linux, and other platforms without external dependencies.
    *   **Robustness:** Strong typing and built-in error handling for a stable and reliable service.

## How It Works

1.  **Ingestion:** The server initializes and starts listening for incoming UDP packets on a predefined port.
2.  **Parsing:** Upon receiving a packet, it parses the JSON payload into a structured data map.
3.  **OPC UA Mapping:** The keys from the JSON object are mapped to corresponding nodes in the OPC UA server's address space. New nodes are created dynamically upon first sight of a new key.
4.  **Serving:** OPC UA clients can connect to the server, browse the address space, and subscribe to these nodes to receive real-time value updates whenever new UDP data arrives.

## Use Case

This server is ideal for securely transporting data from a source in a less trusted network zone (e.g., a field network with IoT devices or PLCs sending JSON via UDP) to a more secure control room network. Clients in the secure network can read the data via the standard OPC UA protocol without the ability to write back commands to the source through this channel.

## Build & Run

go mod tidy
go run .
