// Package server is responsible for the network transport layer of IgnisKV.
//
// # Purpose
//
// The server package owns TCP networking and client connection management. It
// acts as the outer shell of the database, bridging external clients communicating
// over the network with the internal database engine.
//
// # Responsibilities
//
// This package manages the entire connection lifecycle. Specifically, it must:
//
//   - Start the TCP server and bind to a port.
//   - Accept incoming client connections.
//   - Read raw RESP bytes from the network stream.
//   - Pass complete RESP messages to the protocol layer for parsing.
//   - Pass parsed types.Command objects to the command dispatcher.
//   - Encode the resulting types.Response objects back into RESP using the protocol layer.
//   - Write the formatted response bytes back to the client socket.
//   - Manage the lifecycle (connect/disconnect) of client connections.
//
// # What Does NOT Belong Here
//
// To preserve the strict separation of concerns established across IgnisKV,
// this package explicitly does not:
//
//   - Parse RESP syntax (owned by internal/protocol).
//   - Execute commands (owned by internal/commands).
//   - Access the Store directly (owned by internal/store).
//   - Contain any database business logic.
//   - Implement persistence or AOF logging.
//   - Handle concurrency beyond the scope of Sprint 6.
//
// # Architecture Position
//
// The server package sits at the very edge of the application boundary. The
// complete request lifecycle flows through it as follows:
//
//	Client
//	  ↓
//	TCP Connection
//	  ↓
//	Accept Connection
//	  ↓
//	Read RESP Message
//	  ↓
//	Protocol Parser
//	  ↓
//	types.Command
//	  ↓
//	Dispatcher
//	  ↓
//	Command Handler
//	  ↓
//	types.Response
//	  ↓
//	Protocol Encoder
//	  ↓
//	Write Response
//	  ↓
//	Client
//
// # Current Scope (Sprint 6)
//
// The current implementation is intentionally minimal to establish the baseline
// architecture. The Sprint 6 scope is strictly limited to:
//
//   - Single client support (one connection at a time).
//   - Sequential request processing.
//   - No goroutines (concurrency will be introduced in a future milestone).
//   - No graceful server shutdown.
//   - No TLS or transport security.
//   - No persistence.
//   - No authentication.
//
// These limitations are deliberate architectural decisions designed to keep
// the initial networking integration simple and verifiable. Future milestones
// will introduce concurrent client handling, graceful shutdown, connection
// timeouts, and transport security without changing the core request lifecycle.
//
// # Design Philosophy
//
// The server package behaves strictly as an orchestration layer. It coordinates
// components but owns no business logic itself. By connecting the networking I/O
// with the existing IgnisKV architecture, it translates network traffic into
// database action without blurring the boundaries between domains.
//
// # Dependencies
//
// The server package depends directly on:
//   - internal/protocol (for parsing and encoding)
//   - internal/commands (for the Dispatcher interface)
//   - internal/types (for the Command and Response models)
//
// It intentionally does NOT depend on the storage implementation (`internal/store`).
// It requires only an initialized Dispatcher, ensuring that the network layer
// remains completely oblivious to how data is stored or retrieved.
package server
