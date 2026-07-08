package protocol_test

import (
	"bytes"
	"testing"

	"github.com/rajas2007/IgnisKV/internal/protocol"
	"github.com/rajas2007/IgnisKV/internal/types"
)

func TestEncodeRESP_StatusOK_WithMessage(t *testing.T) {
	resp := types.Response{
		Status:  types.StatusOK,
		Message: "PONG",
	}

	expected := []byte("+PONG\r\n")
	actual := protocol.EncodeRESP(resp)

	if !bytes.Equal(actual, expected) {
		t.Fatalf("Expected %q, got %q", expected, actual)
	}
}

func TestEncodeRESP_StatusOK_WithData(t *testing.T) {
	resp := types.Response{
		Status: types.StatusOK,
		Data:   "Rajas",
	}

	expected := []byte("$5\r\nRajas\r\n")
	actual := protocol.EncodeRESP(resp)

	if !bytes.Equal(actual, expected) {
		t.Fatalf("Expected %q, got %q", expected, actual)
	}
}

func TestEncodeRESP_StatusOK_Empty(t *testing.T) {
	resp := types.Response{
		Status: types.StatusOK,
	}

	expected := []byte("+OK\r\n")
	actual := protocol.EncodeRESP(resp)

	if !bytes.Equal(actual, expected) {
		t.Fatalf("Expected %q, got %q", expected, actual)
	}
}

func TestEncodeRESP_StatusNil(t *testing.T) {
	resp := types.Response{
		Status: types.StatusNil,
	}

	expected := []byte("$-1\r\n")
	actual := protocol.EncodeRESP(resp)

	if !bytes.Equal(actual, expected) {
		t.Fatalf("Expected %q, got %q", expected, actual)
	}
}

func TestEncodeRESP_StatusError(t *testing.T) {
	resp := types.Response{
		Status:  types.StatusError,
		Message: "malformed RESP message",
	}

	expected := []byte("-ERR malformed RESP message\r\n")
	actual := protocol.EncodeRESP(resp)

	if !bytes.Equal(actual, expected) {
		t.Fatalf("Expected %q, got %q", expected, actual)
	}
}

func TestEncodeRESP_StatusExit(t *testing.T) {
	resp := types.Response{
		Status: types.StatusExit,
	}

	expected := []byte("+BYE\r\n")
	actual := protocol.EncodeRESP(resp)

	if !bytes.Equal(actual, expected) {
		t.Fatalf("Expected %q, got %q", expected, actual)
	}
}

func TestEncodeRESP_UnknownStatus(t *testing.T) {
	resp := types.Response{
		Status: types.ResponseStatus(99),
	}

	expected := []byte("-ERR unknown status code 99\r\n")
	actual := protocol.EncodeRESP(resp)

	if !bytes.Equal(actual, expected) {
		t.Fatalf("Expected %q, got %q", expected, actual)
	}
}
