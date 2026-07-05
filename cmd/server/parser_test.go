package main

import (
	"testing"
)

func TestParseCommandEmptyString(t *testing.T) {
	// Arrange
	input := ""

	// Act
	cmd, ok := ParseCommand(input)

	// Assert
	if ok {
		t.Fatalf("ParseCommand(%q) returned ok=true; want false", input)
	}
	if cmd.Name != "" {
		t.Fatalf("ParseCommand(%q) returned Name=%q; want empty string", input, cmd.Name)
	}
	if len(cmd.Args) != 0 {
		t.Fatalf("ParseCommand(%q) returned Args=%v; want nil or empty", input, cmd.Args)
	}
}

func TestParseCommandWhitespaceOnly(t *testing.T) {
	// Arrange
	input := "        "

	// Act
	cmd, ok := ParseCommand(input)

	// Assert
	if ok {
		t.Fatalf("ParseCommand(%q) returned ok=true; want false", input)
	}
	if cmd.Name != "" {
		t.Fatalf("ParseCommand(%q) returned Name=%q; want empty string", input, cmd.Name)
	}
	if len(cmd.Args) != 0 {
		t.Fatalf("ParseCommand(%q) returned Args=%v; want nil or empty", input, cmd.Args)
	}
}

func TestParseCommandNoArguments(t *testing.T) {
	// Arrange
	input := "PING"

	// Act
	cmd, ok := ParseCommand(input)

	// Assert
	if !ok {
		t.Fatalf("ParseCommand(%q) returned ok=false; want true", input)
	}
	if cmd.Name != "PING" {
		t.Fatalf("ParseCommand(%q) returned Name=%q; want %q", input, cmd.Name, "PING")
	}
	if len(cmd.Args) != 0 {
		t.Fatalf("ParseCommand(%q) returned Args=%v; want no arguments", input, cmd.Args)
	}
}

func TestParseCommandLowercaseCommand(t *testing.T) {
	// Arrange
	input := "set"

	// Act
	cmd, ok := ParseCommand(input)

	// Assert
	if !ok {
		t.Fatalf("ParseCommand(%q) returned ok=false; want true", input)
	}
	if cmd.Name != "SET" {
		t.Fatalf("ParseCommand(%q) returned Name=%q; want %q", input, cmd.Name, "SET")
	}
}

func TestParseCommandArguments(t *testing.T) {
	// Arrange
	input := "SET name Rajas"

	// Act
	cmd, ok := ParseCommand(input)

	// Assert
	if !ok {
		t.Fatalf("ParseCommand(%q) returned ok=false; want true", input)
	}
	if cmd.Name != "SET" {
		t.Fatalf("ParseCommand(%q) returned Name=%q; want %q", input, cmd.Name, "SET")
	}
	if len(cmd.Args) != 2 {
		t.Fatalf("ParseCommand(%q) returned %d args; want 2", input, len(cmd.Args))
	}
	if cmd.Args[0] != "name" {
		t.Fatalf("ParseCommand(%q) Args[0]=%q; want %q", input, cmd.Args[0], "name")
	}
	if cmd.Args[1] != "Rajas" {
		t.Fatalf("ParseCommand(%q) Args[1]=%q; want %q", input, cmd.Args[1], "Rajas")
	}
}

func TestParseCommandWhitespaceNormalization(t *testing.T) {
	// Arrange
	input := "    GET      myKey      "

	// Act
	cmd, ok := ParseCommand(input)

	// Assert
	if !ok {
		t.Fatalf("ParseCommand(%q) returned ok=false; want true", input)
	}
	if cmd.Name != "GET" {
		t.Fatalf("ParseCommand(%q) returned Name=%q; want %q", input, cmd.Name, "GET")
	}
	if len(cmd.Args) != 1 {
		t.Fatalf("ParseCommand(%q) returned %d args; want 1", input, len(cmd.Args))
	}
	if cmd.Args[0] != "myKey" {
		t.Fatalf("ParseCommand(%q) Args[0]=%q; want %q", input, cmd.Args[0], "myKey")
	}
}

func TestParseCommandPreservesArgumentCase(t *testing.T) {
	// Arrange
	input := "SET Name Rajas"

	// Act
	cmd, ok := ParseCommand(input)

	// Assert
	if !ok {
		t.Fatalf("ParseCommand(%q) returned ok=false; want true", input)
	}
	if cmd.Name != "SET" {
		t.Fatalf("ParseCommand(%q) returned Name=%q; want %q", input, cmd.Name, "SET")
	}
	if len(cmd.Args) != 2 {
		t.Fatalf("ParseCommand(%q) returned %d args; want 2", input, len(cmd.Args))
	}
	if cmd.Args[0] != "Name" {
		t.Fatalf("ParseCommand(%q) Args[0]=%q; want %q (argument case must be preserved)", input, cmd.Args[0], "Name")
	}
	if cmd.Args[1] != "Rajas" {
		t.Fatalf("ParseCommand(%q) Args[1]=%q; want %q (argument case must be preserved)", input, cmd.Args[1], "Rajas")
	}
}
