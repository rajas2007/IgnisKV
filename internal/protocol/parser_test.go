package protocol

import (
	"errors"
	"testing"
)

// Successful Parsing Tests

func TestParseSimplePING(t *testing.T) {
	// Arrange
	input := []byte("*1\r\n$4\r\nPING\r\n")

	// Act
	cmd, err := ParseRESP(input)

	// Assert
	if err != nil {
		t.Fatalf("ParseRESP returned unexpected error: %v", err)
	}
	if cmd.Name != "PING" {
		t.Fatalf("Command.Name = %q; want %q", cmd.Name, "PING")
	}
	if len(cmd.Args) != 0 {
		t.Fatalf("Command.Args length = %d; want 0", len(cmd.Args))
	}
}

func TestParseLowercaseCommand(t *testing.T) {
	// Arrange
	input := []byte("*1\r\n$4\r\nping\r\n")

	// Act
	cmd, err := ParseRESP(input)

	// Assert
	if err != nil {
		t.Fatalf("ParseRESP returned unexpected error: %v", err)
	}
	if cmd.Name != "PING" {
		t.Fatalf("Command.Name = %q; want %q (should be uppercase)", cmd.Name, "PING")
	}
	if len(cmd.Args) != 0 {
		t.Fatalf("Command.Args length = %d; want 0", len(cmd.Args))
	}
}

func TestParseGET(t *testing.T) {
	// Arrange
	input := []byte("*2\r\n$3\r\nGET\r\n$4\r\nname\r\n")

	// Act
	cmd, err := ParseRESP(input)

	// Assert
	if err != nil {
		t.Fatalf("ParseRESP returned unexpected error: %v", err)
	}
	if cmd.Name != "GET" {
		t.Fatalf("Command.Name = %q; want %q", cmd.Name, "GET")
	}
	if len(cmd.Args) != 1 {
		t.Fatalf("Command.Args length = %d; want 1", len(cmd.Args))
	}
	if cmd.Args[0] != "name" {
		t.Fatalf("Command.Args[0] = %q; want %q", cmd.Args[0], "name")
	}
}

func TestParseSET(t *testing.T) {
	// Arrange
	input := []byte("*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$5\r\nRajas\r\n")

	// Act
	cmd, err := ParseRESP(input)

	// Assert
	if err != nil {
		t.Fatalf("ParseRESP returned unexpected error: %v", err)
	}
	if cmd.Name != "SET" {
		t.Fatalf("Command.Name = %q; want %q", cmd.Name, "SET")
	}
	if len(cmd.Args) != 2 {
		t.Fatalf("Command.Args length = %d; want 2", len(cmd.Args))
	}
	if cmd.Args[0] != "name" {
		t.Fatalf("Command.Args[0] = %q; want %q", cmd.Args[0], "name")
	}
	if cmd.Args[1] != "Rajas" {
		t.Fatalf("Command.Args[1] = %q; want %q", cmd.Args[1], "Rajas")
	}
}

func TestParseDEL(t *testing.T) {
	// Arrange
	input := []byte("*2\r\n$3\r\nDEL\r\n$4\r\nname\r\n")

	// Act
	cmd, err := ParseRESP(input)

	// Assert
	if err != nil {
		t.Fatalf("ParseRESP returned unexpected error: %v", err)
	}
	if cmd.Name != "DEL" {
		t.Fatalf("Command.Name = %q; want %q", cmd.Name, "DEL")
	}
	if len(cmd.Args) != 1 {
		t.Fatalf("Command.Args length = %d; want 1", len(cmd.Args))
	}
	if cmd.Args[0] != "name" {
		t.Fatalf("Command.Args[0] = %q; want %q", cmd.Args[0], "name")
	}
}

func TestParseHELP(t *testing.T) {
	// Arrange
	input := []byte("*1\r\n$4\r\nHELP\r\n")

	// Act
	cmd, err := ParseRESP(input)

	// Assert
	if err != nil {
		t.Fatalf("ParseRESP returned unexpected error: %v", err)
	}
	if cmd.Name != "HELP" {
		t.Fatalf("Command.Name = %q; want %q", cmd.Name, "HELP")
	}
	if len(cmd.Args) != 0 {
		t.Fatalf("Command.Args length = %d; want 0", len(cmd.Args))
	}
}

func TestParseQUIT(t *testing.T) {
	// Arrange
	input := []byte("*1\r\n$4\r\nQUIT\r\n")

	// Act
	cmd, err := ParseRESP(input)

	// Assert
	if err != nil {
		t.Fatalf("ParseRESP returned unexpected error: %v", err)
	}
	if cmd.Name != "QUIT" {
		t.Fatalf("Command.Name = %q; want %q", cmd.Name, "QUIT")
	}
	if len(cmd.Args) != 0 {
		t.Fatalf("Command.Args length = %d; want 0", len(cmd.Args))
	}
}

func TestParseEmptyBulkString(t *testing.T) {
	// Arrange
	input := []byte("*2\r\n$3\r\nGET\r\n$0\r\n\r\n")

	// Act
	cmd, err := ParseRESP(input)

	// Assert
	if err != nil {
		t.Fatalf("ParseRESP returned unexpected error: %v", err)
	}
	if cmd.Name != "GET" {
		t.Fatalf("Command.Name = %q; want %q", cmd.Name, "GET")
	}
	if len(cmd.Args) != 1 {
		t.Fatalf("Command.Args length = %d; want 1", len(cmd.Args))
	}
	if cmd.Args[0] != "" {
		t.Fatalf("Command.Args[0] = %q; want empty string", cmd.Args[0])
	}
}

// Error Tests

func TestParseEmptyInput(t *testing.T) {
	// Arrange
	input := []byte("")

	// Act
	_, err := ParseRESP(input)

	// Assert
	if err == nil {
		t.Fatalf("ParseRESP expected error but got nil")
	}
	if !errors.Is(err, ErrMalformedRESP) {
		t.Fatalf("ParseRESP returned error %v; want %v", err, ErrMalformedRESP)
	}
}

func TestParseRootNotArray(t *testing.T) {
	// Arrange
	input := []byte("+OK\r\n")

	// Act
	_, err := ParseRESP(input)

	// Assert
	if err == nil {
		t.Fatalf("ParseRESP expected error but got nil")
	}
	if !errors.Is(err, ErrUnsupportedType) {
		t.Fatalf("ParseRESP returned error %v; want %v", err, ErrUnsupportedType)
	}
}

func TestParseArrayLengthZero(t *testing.T) {
	// Arrange
	input := []byte("*0\r\n")

	// Act
	_, err := ParseRESP(input)

	// Assert
	if err == nil {
		t.Fatalf("ParseRESP expected error but got nil")
	}
	if !errors.Is(err, ErrMalformedRESP) {
		t.Fatalf("ParseRESP returned error %v; want %v", err, ErrMalformedRESP)
	}
}

func TestParseMissingCRLF(t *testing.T) {
	// Arrange
	input := []byte("*1\n$4\nPING\n") // Missing \r

	// Act
	_, err := ParseRESP(input)

	// Assert
	if err == nil {
		t.Fatalf("ParseRESP expected error but got nil")
	}
	if !errors.Is(err, ErrMalformedRESP) {
		t.Fatalf("ParseRESP returned error %v; want %v", err, ErrMalformedRESP)
	}
}

func TestParseInvalidArrayLength(t *testing.T) {
	// Arrange
	input := []byte("*abc\r\n")

	// Act
	_, err := ParseRESP(input)

	// Assert
	if err == nil {
		t.Fatalf("ParseRESP expected error but got nil")
	}
	if !errors.Is(err, ErrMalformedRESP) {
		t.Fatalf("ParseRESP returned error %v; want %v", err, ErrMalformedRESP)
	}
}

func TestParseInvalidBulkStringLength(t *testing.T) {
	// Arrange
	input := []byte("*1\r\n$xyz\r\n")

	// Act
	_, err := ParseRESP(input)

	// Assert
	if err == nil {
		t.Fatalf("ParseRESP expected error but got nil")
	}
	if !errors.Is(err, ErrMalformedRESP) {
		t.Fatalf("ParseRESP returned error %v; want %v", err, ErrMalformedRESP)
	}
}

func TestParseDeclaredBulkLengthLargerThanPayload(t *testing.T) {
	// Arrange
	// Declares 4 bytes but only 3 are provided ("PIN")
	input := []byte("*1\r\n$4\r\nPIN\r\n")

	// Act
	_, err := ParseRESP(input)

	// Assert
	if err == nil {
		t.Fatalf("ParseRESP expected error but got nil")
	}
	if !errors.Is(err, ErrMalformedRESP) {
		t.Fatalf("ParseRESP returned error %v; want %v", err, ErrMalformedRESP)
	}
}

func TestParseDeclaredBulkLengthSmallerThanPayload(t *testing.T) {
	// Arrange
	// Declared payload length does not match actual payload.
	input := []byte("*1\r\n$3\r\nPING\r\n")

	// Act
	_, err := ParseRESP(input)

	// Assert
	if err == nil {
		t.Fatalf("ParseRESP expected error but got nil")
	}
	if !errors.Is(err, ErrMalformedRESP) {
		t.Fatalf("ParseRESP returned error %v; want %v", err, ErrMalformedRESP)
	}
}

func TestParseBulkStringMissingDollar(t *testing.T) {
	// Arrange
	// Pass an array instead of a bulk string for the command name.
	input := []byte("*1\r\n*1\r\n$4\r\nPING\r\n")

	// Act
	_, err := ParseRESP(input)

	// Assert
	if err == nil {
		t.Fatalf("ParseRESP expected error but got nil")
	}
	if !errors.Is(err, ErrUnsupportedType) {
		t.Fatalf("ParseRESP returned error %v; want %v", err, ErrUnsupportedType)
	}
}

func TestParseUnsupportedRESPTypeInsideArray(t *testing.T) {
	// Arrange
	// Simple string (+PING) passed as an element instead of a bulk string.
	input := []byte("*1\r\n+PING\r\n")

	// Act
	_, err := ParseRESP(input)

	// Assert
	if err == nil {
		t.Fatalf("ParseRESP expected error but got nil")
	}
	if !errors.Is(err, ErrUnsupportedType) {
		t.Fatalf("ParseRESP returned error %v; want %v", err, ErrUnsupportedType)
	}
}

func TestParseNegativeBulkStringLength(t *testing.T) {
	// Arrange
	// Null bulk string is technically valid RESP, but unsupported in this scope.
	input := []byte("*1\r\n$-1\r\n")

	// Act
	_, err := ParseRESP(input)

	// Assert
	if err == nil {
		t.Fatalf("ParseRESP expected error but got nil")
	}
	if !errors.Is(err, ErrUnsupportedType) {
		t.Fatalf("ParseRESP returned error %v; want %v", err, ErrUnsupportedType)
	}
}

func TestParseUnexpectedEndOfMessage(t *testing.T) {
	// Arrange
	// Ends abruptly before providing the bulk string element.
	input := []byte("*1\r\n")

	// Act
	_, err := ParseRESP(input)

	// Assert
	if err == nil {
		t.Fatalf("ParseRESP expected error but got nil")
	}
	if !errors.Is(err, ErrMalformedRESP) {
		t.Fatalf("ParseRESP returned error %v; want %v", err, ErrMalformedRESP)
	}
}

func TestParseTrailingUnreadBytes(t *testing.T) {
	// Arrange
	// Array of length 1, but payload provides 2 bulk strings.
	input := []byte("*1\r\n$4\r\nPING\r\n$4\r\ntest\r\n")

	// Act
	_, err := ParseRESP(input)

	// Assert
	if err == nil {
		t.Fatalf("ParseRESP expected error but got nil")
	}
	if !errors.Is(err, ErrMalformedRESP) {
		t.Fatalf("ParseRESP returned error %v; want %v", err, ErrMalformedRESP)
	}
}
