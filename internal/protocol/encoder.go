package protocol

import (
	"fmt"

	"github.com/rajas2007/IgnisKV/internal/types"
)

// EncodeRESP converts an internal types.Response into a valid RESP byte array.
// It guarantees that every valid Response produces exactly one valid RESP message.
func EncodeRESP(resp types.Response) []byte {
	switch resp.Status {
	case types.StatusOK:
		if resp.Message != "" {
			return []byte(fmt.Sprintf("+%s\r\n", resp.Message))
		}
		if resp.Data != nil {
			str := fmt.Sprintf("%v", resp.Data)
			return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(str), str))
		}
		return []byte("+OK\r\n")

	case types.StatusNil:
		return []byte("$-1\r\n")

	case types.StatusError:
		return []byte(fmt.Sprintf("-ERR %s\r\n", resp.Message))

	case types.StatusInteger:
		return []byte(fmt.Sprintf(":%v\r\n", resp.Data))

	case types.StatusExit:
		return []byte("+BYE\r\n")

	default:
		// Defensive fallback.
		// Should never happen for valid Response values.
		return []byte(fmt.Sprintf("-ERR unknown status code %d\r\n", resp.Status))
	}
}
