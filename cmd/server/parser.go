package main

import (
	"strings"

	"github.com/rajas2007/IgnisKV/internal/types"
)

// ParseCommand converts a single line of CLI input into a types.Command.
//
// It tokenises the line using whitespace as the delimiter. The first token
// becomes the command name, uppercased so that command lookup is
// case-insensitive. All remaining tokens are collected as arguments in their
// original case, since argument values are user data and must not be modified.
//
// ParseCommand returns false when the input is empty or contains only
// whitespace, indicating that no command was present and the REPL should
// prompt again without dispatching. It returns true for any non-empty input,
// including single-token lines with no arguments.
//
// Validation of argument count and content belongs in the command handlers,
// not here. The parser's only responsibility is lexical: turn a string into a
// structured Command. Separating these concerns means the parser never needs
// to know which commands exist or what arguments they require.
func ParseCommand(line string) (types.Command, bool) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return types.Command{}, false
	}

	return types.Command{
		Name: strings.ToUpper(fields[0]),
		Args: fields[1:],
	}, true
}
