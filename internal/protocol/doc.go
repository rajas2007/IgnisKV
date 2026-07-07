// Package protocol is responsible for translating external wire-protocol messages
// into IgnisKV's internal domain models, and vice versa.
//
// Its primary purpose is to decouple the network communication format from the
// core database logic. By acting as a translation layer, it ensures that the
// command execution engine and storage layer never need to know whether a
// request arrived via a CLI REPL, an HTTP endpoint, or a RESP TCP stream.
//
// # Responsibilities
//
// This package owns the serialisation and deserialisation of the Redis
// Serialization Protocol (RESP). It parses complete RESP messages, validates
// their syntax, and produces internal types.Command values. In the future, it
// will also provide an encoder to translate types.Response values back into RESP
// byte streams.
//
// # What does not belong here
//
// The protocol layer contains no business logic. It does not execute commands,
// access the storage engine, or validate whether a command has the correct
// number of arguments for its specific operation. Furthermore, it does not
// manage TCP connections or read directly from network sockets. Connection
// lifecycle and stream buffering belong to the network layer; this package
// simply processes the bytes it is given.
//
// # Architecture Position
//
// The protocol layer sits squarely between the network transport and the command
// dispatcher. The complete request lifecycle flows through it as follows:
//
//	TCP Connection
//	      ↓
//	Protocol Layer (Parser)
//	      ↓
//	types.Command
//	      ↓
//	Dispatcher
//	      ↓
//	Command Handlers
//	      ↓
//	types.Response
//	      ↓
//	Protocol Layer (Encoder)
//	      ↓
//	Client
//
// # Current Scope and Design Philosophy
//
// Sprint 5 focuses solely on implementing the RESP parser. Initially, the parser
// supports only Arrays and Bulk Strings, as every supported Redis command can be
// represented entirely using those two types. Additional RESP types will be
// implemented incrementally as needed.
//
// A core design principle of this package is that the parser translates exactly
// one complete RESP message into exactly one types.Command. It relies on the
// caller to provide a complete message. Decoupling parsing from I/O reading
// ensures the parser remains independently testable without requiring mock
// network connections, strictly adhering to the Single Responsibility Principle.
//
// # Interactions
//
// This package imports internal/types to access the Command and Response domain
// models. It is designed to be consumed by the future networking layer, which will
// provide complete RESP messages to the parser and route the resulting types.Command
// values to the command dispatcher.
package protocol
