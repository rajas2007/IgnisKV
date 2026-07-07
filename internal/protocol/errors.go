package protocol

import "errors"

var (
	// ErrMalformedRESP indicates that the RESP payload violates protocol syntax,
	// such as missing CRLF, invalid lengths, or truncated data.
	ErrMalformedRESP = errors.New("malformed RESP message")

	// ErrUnsupportedType indicates that the RESP message uses a type that the
	// parser does not currently support (e.g., Simple Strings, Integers).
	ErrUnsupportedType = errors.New("unsupported RESP type")
)
