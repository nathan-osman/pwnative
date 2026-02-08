package socket

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

var (
	defaultTimeout = 30 * time.Second
	dialer         = net.Dialer{
		Timeout: defaultTimeout,
	}
)

// Socket maintains a connection with a Unix socket. This type is designed
// specifically to be resilient in the face of errors, reconnecting on
// disconnect and retrying with backoff as needed.
type Socket struct {
	cfg        Config
	cancel     context.CancelFunc
	chanClosed chan any
}

func (s *Socket) connect(ctx context.Context) (net.Conn, error) {
	for _, f := range s.cfg.Filenames {
		c, err := dialer.DialContext(ctx, "unix", f)
		if err != nil {
			s.cfg.Logger.Debug(
				err.Error(),
				"filename",
				f,
			)
			continue
		}
		return c, nil
	}
	return nil, errors.New("unable to connect to a socket")
}

func (s *Socket) loop(ctx context.Context) (err error) {

	// Invoke the error handler when the function returns
	defer func() {
		if s.cfg.HandleError != nil && !errors.Is(err, context.Canceled) {
			s.cfg.HandleError(err)
		}
	}()

	// Connect to the first socket that succeeds
	c, err := s.connect(ctx)
	if err != nil {
		return err
	}

	// Socket is connected, invoke the connected handler
	if s.cfg.HandleConnected != nil {
		s.cfg.HandleConnected()
	}

	// Doing this in a proper "unix" way is tricky; we need to read() from the
	// socket but we also need to interrupt the read when the context is
	// canceled; therefore we create two file descriptors — one from the
	// socket and the other from a "wakeup" pipe we create; both of these are
	// passed to a poll() call, allowing either incoming data or the context
	// to interrupt

	// Create the pipe
	r, w, err := os.Pipe()
	if err != nil {
		return err
	}

	// Create a goroutine that monitors the context; if it is canceled, send
	// on the pipe to terminate the poll call
	chanDone := make(chan any)
	defer close(chanDone)
	go func() {
		select {
		case <-chanDone:
		case <-ctx.Done():
			w.Write([]byte{0})
		}
	}()

	// Create file descriptors for polling
	fds := []unix.PollFd{
		{
			Fd:     int32(0),
			Events: unix.POLLIN,
		},
		{
			Fd:     int32(r.Fd()),
			Events: unix.POLLIN,
		},
	}

	// Poll in a loop
	b := make([]byte, 1024)
	for {
		_, err := unix.Poll(fds, -1)
		if err != nil {
			return err
		}
		if fds[1].Revents&unix.POLLIN != 0 {
			return context.Canceled
		}
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

// New initializes a socket and starts the connection process.
func New(cfg *Config) *Socket {
	var (
		ctx, cancel = context.WithCancel(context.Background())
		s           = &Socket{
			cfg:        *cfg,
			cancel:     cancel,
			chanClosed: make(chan any),
		}
	)
	if s.cfg.Logger == nil {
		s.cfg.Logger = slog.Default()
	}
	go s.run(ctx)
	return s
}

// Close shuts down the connection.
func (s *Socket) Close() {
	s.cancel()
	<-s.chanClosed
}
