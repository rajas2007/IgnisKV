package server

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net"
	"strconv"

	"github.com/rajas2007/IgnisKV/internal/protocol"
	"github.com/rajas2007/IgnisKV/internal/types"
)

const defaultBufferSize = 4096

// handleConnection manages the request lifecycle for a single client connection.
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReaderSize(conn, defaultBufferSize)

	for {
		// Read one complete RESP message
		reqData, err := s.readRequest(reader)
		if err != nil {
			// Network errors terminate the client connection immediately
			if errors.Is(err, io.EOF) {
				return
			}
			return
		}

		// ParseRESP()
		cmd, err := protocol.ParseRESP(reqData)
		if err != nil {
			// Protocol parsing errors are encoded as RESP error responses
			resp := types.Response{
				Status:  types.StatusError,
				Message: err.Error(),
			}
			_ = s.writeResponse(conn, resp)
			continue
		}

		// Dispatcher.Execute()
		resp := s.dispatcher.Dispatch(cmd)

		// Write response
		err = s.writeResponse(conn, resp)
		if err != nil {
			// Network error writing response, terminate
			return
		}

		// Session termination check
		if resp.Status == types.StatusExit {
			return
		}
	}
}

// readRequest reads bytes from the connection using a bufio.Reader until one
// complete RESP message has been assembled. This keeps the protocol parser
// completely stateless and separate from the network stream handling.
func (s *Server) readRequest(reader *bufio.Reader) ([]byte, error) {
	// Read the array header
	header, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	buf.Write(header)

	// If it doesn't look like an array, return what we have and let the parser reject it.
	if len(header) < 4 || header[0] != '*' {
		return buf.Bytes(), nil
	}

	// Extract array length (ignoring the \r\n)
	countStr := string(header[1 : len(header)-2])
	count, err := strconv.Atoi(countStr)
	if err != nil || count < 0 {
		return buf.Bytes(), nil
	}

	// Read each bulk string element
	for i := 0; i < count; i++ {
		lenLine, err := reader.ReadBytes('\n')
		if err != nil {
			return buf.Bytes(), err
		}
		buf.Write(lenLine)

		if len(lenLine) < 4 || lenLine[0] != '$' {
			break // Malformed; let the parser handle the error
		}

		strLenStr := string(lenLine[1 : len(lenLine)-2])
		strLen, err := strconv.Atoi(strLenStr)
		if err != nil || strLen < 0 {
			break // Malformed; let the parser handle the error
		}

		// Read the exact payload length plus the trailing \r\n
		payload := make([]byte, strLen+2)
		_, err = io.ReadFull(reader, payload)
		if err != nil {
			return buf.Bytes(), err
		}
		buf.Write(payload)
	}

	return buf.Bytes(), nil
}

// writeResponse encodes the internal types.Response into RESP and writes it to the connection.
func (s *Server) writeResponse(conn net.Conn, resp types.Response) error {
	encoded := protocol.EncodeRESP(resp)
	_, err := conn.Write(encoded)
	return err
}
