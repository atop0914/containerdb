// Package container provides base container management utilities.
package container

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"
)

// AvailablePort finds an available port on localhost.
func AvailablePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// WaitForPort waits for a port to become available with timeout.
func WaitForPort(host string, port int, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	endpoint := net.JoinHostPort(host, strconv.Itoa(port))
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for %s", endpoint)
		default:
			conn, err := net.DialTimeout("tcp", endpoint, 500*time.Millisecond)
			if err == nil {
				conn.Close()
				return nil
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}
