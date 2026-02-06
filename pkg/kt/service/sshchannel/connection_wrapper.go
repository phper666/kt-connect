package sshchannel

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// TimeoutConn wraps a net.Conn with idle timeout and proper close handling
type TimeoutConn struct {
	net.Conn
	idleTimeout time.Duration
	lastActive  time.Time
	mu          sync.RWMutex
	closed      bool
	remoteAddr  string
}

// NewTimeoutConn creates a new connection wrapper with idle timeout
func NewTimeoutConn(conn net.Conn, idleTimeout time.Duration) *TimeoutConn {
	tc := &TimeoutConn{
		Conn:        conn,
		idleTimeout: idleTimeout,
		lastActive:  time.Now(),
		remoteAddr:  conn.RemoteAddr().String(),
	}

	// Start idle timeout checker
	go tc.checkIdleTimeout()

	return tc
}

// Read implements net.Conn.Read with idle timeout
func (t *TimeoutConn) Read(b []byte) (n int, err error) {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return 0, net.ErrClosed
	}
	t.lastActive = time.Now()
	t.mu.Unlock()

	// Set read deadline
	if t.idleTimeout > 0 {
		t.Conn.SetReadDeadline(time.Now().Add(t.idleTimeout))
	}

	n, err = t.Conn.Read(b)
	if err != nil && !t.isTimeout(err) {
		log.Debug().Err(err).Msgf("Read error on connection to %s", t.remoteAddr)
	}
	return n, err
}

// Write implements net.Conn.Write with idle timeout
func (t *TimeoutConn) Write(b []byte) (n int, err error) {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return 0, net.ErrClosed
	}
	t.lastActive = time.Now()
	t.mu.Unlock()

	// Set write deadline
	if t.idleTimeout > 0 {
		t.Conn.SetWriteDeadline(time.Now().Add(t.idleTimeout))
	}

	n, err = t.Conn.Write(b)
	if err != nil && !t.isTimeout(err) {
		log.Debug().Err(err).Msgf("Write error on connection to %s", t.remoteAddr)
	}
	return n, err
}

// Close implements net.Conn.Close with proper cleanup
func (t *TimeoutConn) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil
	}

	t.closed = true
	log.Debug().Msgf("Closing connection to %s", t.remoteAddr)

	err := t.Conn.Close()
	if err != nil {
		log.Warn().Err(err).Msgf("Failed to close connection to %s", t.remoteAddr)
	}

	return err
}

// checkIdleTimeout monitors connection idle time and closes if exceeded
func (t *TimeoutConn) checkIdleTimeout() {
	if t.idleTimeout <= 0 {
		return
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		t.mu.RLock()
		if t.closed {
			t.mu.RUnlock()
			return
		}

		idleTime := time.Since(t.lastActive)
		t.mu.RUnlock()

		if idleTime > t.idleTimeout {
			log.Debug().Msgf("Connection to %s idle for %v, closing", t.remoteAddr, idleTime)
			t.Close()
			return
		}
	}
}

// isTimeout checks if error is a timeout error
func (t *TimeoutConn) isTimeout(err error) bool {
	if err == nil {
		return false
	}

	if netErr, ok := err.(net.Error); ok {
		return netErr.Timeout()
	}
	return false
}

// DialerWrapper wraps SSH dialer with connection management
type DialerWrapper struct {
	dialer      DialContext
	idleTimeout time.Duration
	mu          sync.RWMutex
	connCount   int
}

// DialContext is the interface for dialing connections
type DialContext interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

// NewDialerWrapper creates a new dialer wrapper
func NewDialerWrapper(dialer DialContext, idleTimeout time.Duration) *DialerWrapper {
	return &DialerWrapper{
		dialer:      dialer,
		idleTimeout: idleTimeout,
	}
}

// DialContext wraps the dial operation with timeout connection
func (w *DialerWrapper) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	// Add dial timeout
	dialCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	conn, err := w.dialer.DialContext(dialCtx, network, address)
	if err != nil {
		log.Debug().Err(err).Msgf("Failed to dial %s", address)
		return nil, err
	}

	w.mu.Lock()
	w.connCount++
	count := w.connCount
	w.mu.Unlock()

	log.Debug().Msgf("Established connection #%d to %s", count, address)

	// Wrap with timeout connection
	return NewTimeoutConn(conn, w.idleTimeout), nil
}

// GetConnectionCount returns current connection count
func (w *DialerWrapper) GetConnectionCount() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.connCount
}
