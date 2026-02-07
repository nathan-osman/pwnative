package pwnative

import (
	"context"
	"errors"
	"net"
	"time"
)

var (
	defaultTimeout = 30 * time.Second
	dialer         = net.Dialer{
		Timeout: defaultTimeout,
	}
)

// SocketConfig provides configuration options to the NewSocket() function.
type SocketConfig struct {

	// Filename is the path to the Unix socket to connect to.
	Filename string

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

// Socket maintains a connection with a Unix socket. This type is designed
// specifically to be resilient in the face of errors, reconnecting on
// disconnect and retrying with backoff as needed.
type Socket struct {
	cfg        *SocketConfig
	cancel     context.CancelFunc
	chanClosed chan any
}

func (s *Socket) loop(ctx context.Context) (err error) {

	// Invoke the error handler when the function returns
	defer func() {
		if s.cfg.HandleError != nil && !errors.Is(err, context.Canceled) {
			s.cfg.HandleError(err)
		}
	}()

	// Connect to the socket
	c, err := dialer.DialContext(ctx, "unix", s.cfg.Filename)
	if err != nil {
		return err
	}

	// The socket remains connected at this point until either the context is
	// cancelled or a read error is encountered; invoke the connected handler
	if s.cfg.HandleConnected != nil {
		s.cfg.HandleConnected()
	}

	// In order to shut things down when the context is cancelled, create a
	// separate goroutine that will close the socket, terminating the Read()
	// call
	chanError := make(chan any)
	defer close(chanError)
	go func() {
		select {
		case <-chanError:
		case <-ctx.Done():
			err = context.Canceled
			c.Close()
		}
	}()

	// Read continuously from the socket in 1024-byte chunks
	b := make([]byte, 1024)
	for {
		n, err := c.Read(b)
		if err != nil {
			return err
		}
		if s.cfg.HandleData != nil {
			s.cfg.HandleData(b[:n])
		}
	}
}

func (s *Socket) run(ctx context.Context) {
	defer close(s.chanClosed)
	for {
		if err := s.loop(ctx); errors.Is(err, context.Canceled) {
			return
		}
		select {
		case <-time.After(defaultTimeout):
		case <-ctx.Done():
			return
		}
	}
}

// NewSocket initializes the socket and starts the connection process.
func NewSocket(cfg *SocketConfig) *Socket {
	var (
		ctx, cancel = context.WithCancel(context.Background())
		s           = &Socket{
			cfg:        cfg,
			cancel:     cancel,
			chanClosed: make(chan any),
		}
	)
	go s.run(ctx)
	return s
}

// Close shuts down the connection.
func (s *Socket) Close() {
	s.cancel()
	<-s.chanClosed
}
