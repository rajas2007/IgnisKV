package protocol

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/rajas2007/IgnisKV/internal/types"
)

var crlf = []byte("\r\n")

// ParseRESP converts a complete RESP byte payload into an internal Command.
// It fails if the payload uses unsupported types, violates RESP syntax, or
// contains trailing unread bytes after a complete array.
func ParseRESP(data []byte) (types.Command, error) {
	if len(data) == 0 {
		return types.Command{}, fmt.Errorf("%w: empty input", ErrMalformedRESP)
	}

	p := &respParser{
		data:   data,
		offset: 0,
		length: len(data),
	}

	count, err := p.parseArrayHeader()
	if err != nil {
		return types.Command{}, err
	}

	if count < 1 {
		return types.Command{}, fmt.Errorf("%w: array length must be >= 1 for commands", ErrMalformedRESP)
	}

	args := make([]string, count)
	for i := 0; i < count; i++ {
		str, err := p.parseBulkString()
		if err != nil {
			return types.Command{}, err
		}
		args[i] = str
	}

	// Reject trailing unread bytes after successfully parsing the declared array.
	if p.offset < p.length {
		return types.Command{}, fmt.Errorf("%w: trailing unread bytes at offset %d", ErrMalformedRESP, p.offset)
	}

	return types.Command{
		Name: strings.ToUpper(args[0]),
		Args: args[1:],
	}, nil
}

type respParser struct {
	data   []byte
	offset int
	length int
}

func (p *respParser) readLine() ([]byte, error) {
	if p.offset >= p.length {
		return nil, fmt.Errorf("%w: unexpected end of message", ErrMalformedRESP)
	}

	// Find the next CRLF.
	idx := bytes.Index(p.data[p.offset:], crlf)
	if idx == -1 {
		return nil, fmt.Errorf("%w: missing CRLF at offset %d", ErrMalformedRESP, p.offset)
	}

	line := p.data[p.offset : p.offset+idx]
	p.offset += idx + 2 // skip over \r\n
	return line, nil
}

func (p *respParser) parseArrayHeader() (int, error) {
	line, err := p.readLine()
	if err != nil {
		return 0, err
	}

	if len(line) == 0 || line[0] != '*' {
		return 0, fmt.Errorf("%w: expected array prefix '*' at start", ErrUnsupportedType)
	}

	count, err := strconv.Atoi(string(line[1:]))
	if err != nil {
		return 0, fmt.Errorf("%w: invalid array length %q", ErrMalformedRESP, string(line[1:]))
	}

	return count, nil
}

func (p *respParser) parseBulkString() (string, error) {
	line, err := p.readLine()
	if err != nil {
		return "", err
	}

	if len(line) == 0 || line[0] != '$' {
		return "", fmt.Errorf("%w: expected bulk string prefix '$'", ErrUnsupportedType)
	}

	strLen, err := strconv.Atoi(string(line[1:]))
	if err != nil {
		return "", fmt.Errorf("%w: invalid bulk string length %q", ErrMalformedRESP, string(line[1:]))
	}

	if strLen < 0 {
		return "", fmt.Errorf("%w: Null Bulk String unsupported", ErrUnsupportedType)
	}

	// Check bounds to ensure we have enough bytes for the string payload AND the trailing CRLF.
	if p.offset+strLen+2 > p.length {
		return "", fmt.Errorf("%w: truncated bulk string payload", ErrMalformedRESP)
	}

	// Verify the payload is terminated with CRLF.
	if p.data[p.offset+strLen] != '\r' || p.data[p.offset+strLen+1] != '\n' {
		return "", fmt.Errorf("%w: bulk string payload not terminated with CRLF", ErrMalformedRESP)
	}

	str := string(p.data[p.offset : p.offset+strLen])
	p.offset += strLen + 2

	return str, nil
}
