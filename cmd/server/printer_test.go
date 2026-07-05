package main

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/rajas2007/IgnisKV/internal/types"
)

func TestPrintResponseStatusOKMessage(t *testing.T) {
	// Arrange
	resp := types.Response{
		Status:  types.StatusOK,
		Message: "PONG",
	}

	// Redirect stdout.
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error: %v", err)
	}
	os.Stdout = w

	// Act
	PrintResponse(resp)

	// Close write end and restore stdout before reading.
	w.Close()
	os.Stdout = orig

	// Assert
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("io.Copy error: %v", err)
	}
	got := buf.String()
	want := "PONG\n"
	if got != want {
		t.Fatalf("PrintResponse output = %q; want %q", got, want)
	}
}

func TestPrintResponseStatusOKData(t *testing.T) {
	// Arrange
	resp := types.Response{
		Status: types.StatusOK,
		Data:   "Rajas",
	}

	// Redirect stdout.
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error: %v", err)
	}
	os.Stdout = w

	// Act
	PrintResponse(resp)

	// Close write end and restore stdout before reading.
	w.Close()
	os.Stdout = orig

	// Assert
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("io.Copy error: %v", err)
	}
	got := buf.String()
	want := "Rajas\n"
	if got != want {
		t.Fatalf("PrintResponse output = %q; want %q", got, want)
	}
}

func TestPrintResponseStatusNil(t *testing.T) {
	// Arrange
	resp := types.Response{
		Status: types.StatusNil,
	}

	// Redirect stdout.
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error: %v", err)
	}
	os.Stdout = w

	// Act
	PrintResponse(resp)

	// Close write end and restore stdout before reading.
	w.Close()
	os.Stdout = orig

	// Assert
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("io.Copy error: %v", err)
	}
	got := buf.String()
	want := "(nil)\n"
	if got != want {
		t.Fatalf("PrintResponse output = %q; want %q", got, want)
	}
}

func TestPrintResponseStatusError(t *testing.T) {
	// Arrange
	resp := types.Response{
		Status:  types.StatusError,
		Message: "wrong number of arguments",
	}

	// Redirect stdout.
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error: %v", err)
	}
	os.Stdout = w

	// Act
	PrintResponse(resp)

	// Close write end and restore stdout before reading.
	w.Close()
	os.Stdout = orig

	// Assert
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("io.Copy error: %v", err)
	}
	got := buf.String()
	want := "wrong number of arguments\n"
	if got != want {
		t.Fatalf("PrintResponse output = %q; want %q", got, want)
	}
}

func TestPrintResponseStatusExit(t *testing.T) {
	// Arrange — StatusExit must be treated identically to StatusOK for presentation.
	resp := types.Response{
		Status:  types.StatusExit,
		Message: "BYE",
	}

	// Redirect stdout.
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error: %v", err)
	}
	os.Stdout = w

	// Act
	PrintResponse(resp)

	// Close write end and restore stdout before reading.
	w.Close()
	os.Stdout = orig

	// Assert
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("io.Copy error: %v", err)
	}
	got := buf.String()
	want := "BYE\n"
	if got != want {
		t.Fatalf("PrintResponse output = %q; want %q", got, want)
	}
}
