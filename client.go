package pwnative

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nathan-osman/pwnative/socket"
)

var (
	socketDirEnvVars = []string{
		"PIPEWIRE_RUNTIME_DIR",
		"XDG_RUNTIME_DIR",
		"USERPROFILE",
	}
	socketName = "pipewire-0"
)

// Client maintains a connection with a PipeWire server.
type Client struct {
	s *socket.Socket
}

func (c *Client) handleConnected() {
	fmt.Println("Connected.")
}

func (c *Client) handleError(err error) {
	fmt.Printf("Err: %s\n", err.Error())
}

func (c *Client) handleData(b []byte) {
	fmt.Printf("  Data: (%d) %v\n", len(b), b)
}

// New initializes the client and connects to the provided server.
func New(cfg *Config) *Client {

	// Populate filenames based on what was provided
	var filenames []string
	if cfg.Filename != "" {
		filenames = append(filenames, cfg.Filename)
	} else {
		for _, p := range socketDirEnvVars {
			path := os.Getenv(p)
			if path != "" {
				filenames = append(
					filenames,
					filepath.Join(path, socketName),
				)
			}
		}
	}

	// Create the socket
	c := &Client{}
	c.s = socket.New(&socket.Config{
		Filenames:       filenames,
		HandleConnected: c.handleConnected,
		HandleError:     c.handleError,
		HandleData:      c.handleData,
	})

	// Return the client
	return c
}

// Close shuts down the connection to the PipeWire server.
func (c *Client) Close() {
	c.s.Close()
}
