package pwnative

import (
	"log/slog"
)

// Config provides configuration options to the New() function.
type Config struct {

	// Filename is the path to PipeWire's socket. If this is left empty, an
	// attempt is made to read this from environment variables.
	Filename string

	// Logger is used for handling log messages.
	Logger *slog.Logger
}
