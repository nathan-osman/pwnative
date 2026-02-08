package socket

import (
	"log/slog"
)

// Config provides configuration options to the New() function.
type Config struct {

	// Filenames is a list of paths to Unix sockets to connect to.
	Filenames []string

	// Logger is used for handling log messages.
	Logger *slog.Logger

	// HandleConnected is invoked when the socket is connected. This field can
	// be left at its zero value.
	HandleConnected func()

	// HandleError is invoked when the socket encounters an error. This field
	// can be left at its zero value.
	HandleError func(error)

	// HandleData is invoked when data is read from the socket. Note that this
	// method is invoked from a different goroutine than the one that
	// originally created the Socket.
	HandleData func([]byte)
}
